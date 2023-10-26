package exec

import (
	"fmt"
	glog "github.com/snail007/gmc/module/log"
	gexec "github.com/snail007/gmc/util/exec"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/signal"
	"sync"
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
				noDaemon, _ := c.Flags().GetBool("no-daemon")
				if !noDaemon && daemon {
					//daemon
					fmt.Println(os.Args[1:3], os.Args[3:], len(os.Args[3:]))
					args := []string{}
					args = append(append(append(args, os.Args[1:3]...), "--disable-daemon"), os.Args[3:]...)
					dCmd := exec.Command(os.Args[0], args...)
					fmt.Println(args)
					//dCmd.Stdout = io.Discard
					//dCmd.Stderr = io.Discard
					err := dCmd.Start()
					if err != nil {
						glog.Errorf("fail to running in background, error: %v", err.Error())
					} else {
						glog.Infof("running in background, pid: %v", dCmd.Process.Pid)
					}
					os.Exit(0)
					return nil
				}
				maxCount, _ := c.Flags().GetInt("count")
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
				g := sync.WaitGroup{}
				g.Add(1)
				go func() {
					defer g.Done()
					for {
						if maxCount > 0 && tryCount >= maxCount {
							glog.Infof("max try count %v reached", maxCount)
							os.Exit(0)
						}
						if kill {
							return
						}
						cmd = gexec.NewCommand(cmdStr).BeforeExec(func(command *gexec.Command, c *exec.Cmd) {
							c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
							c.Stdout = os.Stdout
							c.Stderr = os.Stderr
							c.Stdin = os.Stdin
						}).AfterExec(func(command *gexec.Command, cmd *exec.Cmd, err error) {
							glog.Infof("running pid: %d", cmd.Process.Pid)
						})
						_, err := cmd.Exec()
						if kill {
							return
						}
						if err != nil {
							tryCount++
							glog.Infof("process exited with %d, restarting...", cmd.Cmd().ProcessState.ExitCode())
							time.Sleep(time.Second * 5)
						} else {
							glog.Infof("process exited with %d", cmd.Cmd().ProcessState.ExitCode())
							return
						}
					}
				}()
				sig = <-signalChan
				kill = true
				if cmd != nil && cmd.Cmd() != nil && cmd.Cmd().Process != nil {
					syscall.Kill(-cmd.Cmd().Process.Pid, syscall.SIGKILL)
				}
				g.Wait()
				return nil
			},
		}
		cmdRetry.Flags().Bool("disable-daemon", false, "disable --daemon")
		cmdRetry.Flags().MarkHidden("disable-daemon")
		cmdRetry.Flags().BoolP("daemon", "d", false, "running in background")
		cmdRetry.Flags().IntP("count", "c", 0, "maximum try count, 0 means no limit")
		execCMD.AddCommand(cmdRetry)
		root.AddCommand(execCMD)
	})
}
