package exec

import (
	"fmt"
	glog "github.com/snail007/gmc/module/log"
	gcond "github.com/snail007/gmc/util/cond"
	gexec "github.com/snail007/gmc/util/exec"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		execCMD := &cobra.Command{
			Use: "exec",
		}
		cmdRetry := &cobra.Command{
			Use:     "retry",
			Long:    "loop exec the command unless it exited with code 0",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				if len(a) == 0 {
					return fmt.Errorf("command string is required")
				}
				daemon, _ := c.Flags().GetBool("daemon")
				output, _ := c.Flags().GetString("output")
				maxCount, _ := c.Flags().GetInt("count")
				seconds, _ := c.Flags().GetInt("sleep")
				timeout, _ := c.Flags().GetInt("timeout")
				seconds = gcond.Cond(seconds <= 0, 5, seconds).(int)
				timeout = gcond.Cond(timeout <= 0, 0, timeout).(int)
				w, wr := os.Stdout, os.Stderr
				if output != "" {
					f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
					if err != nil {
						return err
					}
					w, wr = f, f
					glog.SetOutput(glog.NewLoggerWriter(f))
				}
				if daemon {
					var args []string
					for _, v := range os.Args[1:] {
						if v != "-d" && v != "--daemon" {
							args = append(args, v)
						}
					}
					dCmd := exec.Command(os.Args[0], args...)
					err := dCmd.Start()
					if err != nil {
						glog.Errorf("fail to running in background, error: %v", err.Error())
					} else {
						glog.Infof("running in background, pid: %v", dCmd.Process.Pid)
					}
					os.Exit(0)
					return nil
				}
				tryCount := 0
				cmdStr := a[0]
				signalChan := make(chan os.Signal, 1)
				signal.Notify(signalChan,
					os.Interrupt,
					syscall.SIGHUP,
					syscall.SIGINT,
					syscall.SIGTERM,
					syscall.SIGQUIT)
				var sig os.Signal
				var kill bool
				defer func() {
					if kill {
						glog.Infof("exited with signal: %v", sig.String())
					}
				}()
				var cmd *gexec.Command
				go func() {
					defer func() {
						select {
						case signalChan <- syscall.SIGQUIT:
							break
						default:
						}
					}()
					for {
						if maxCount > 0 && tryCount >= maxCount {
							glog.Infof("max try count %v reached", maxCount)
							os.Exit(0)
						}
						if kill {
							return
						}
						cmd = gexec.NewCommand(cmdStr).
							Timeout(time.Second * time.Duration(timeout)).
							BeforeExec(func(command *gexec.Command, c *exec.Cmd) {
								c.Stdout = w
								c.Stderr = wr
								c.Stdin = os.Stdin
							}).
							AfterExec(func(command *gexec.Command, cmd *exec.Cmd, err error) {
								glog.Infof("running pid: %d", cmd.Process.Pid)
							})
						_, err := cmd.Exec()
						if kill {
							return
						}
						if err != nil {
							tryCount++
							glog.Infof("process exited with %d, restarting...", cmd.Cmd().ProcessState.ExitCode())
							time.Sleep(time.Second * time.Duration(seconds))
						} else {
							glog.Infof("process exited with %d", cmd.Cmd().ProcessState.ExitCode())
							return
						}
					}
				}()
				sig = <-signalChan
				kill = true
				if cmd != nil && cmd.Cmd() != nil && cmd.Cmd().Process != nil {
					killCmd(cmd.Cmd())
				}
				return nil
			},
		}
		cmdRetry.Flags().IntP("timeout", "t", 0, "command timeout seconds, exceeded will be kill. 0: unlimited")
		cmdRetry.Flags().IntP("sleep", "s", 5, "sleep seconds before restarting")
		cmdRetry.Flags().StringP("output", "o", "", "the file logging output to")
		cmdRetry.Flags().BoolP("daemon", "d", false, "running in background")
		cmdRetry.Flags().IntP("count", "c", 0, "maximum try count, 0 means no limit")
		execCMD.AddCommand(cmdRetry)
		root.AddCommand(execCMD)
	})
}
