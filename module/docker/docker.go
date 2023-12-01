package docker

import (
	"bufio"
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	gfile "github.com/snail007/gmc/util/file"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use: "docker",
			Long: "create a model in current directory, all run arguments after -- \n " +
				"Example:  \n " +
				"gmct docker -- ./foo -u xxx \n " +
				"gmct docker -g -- go build \n " +
				"gmct docker -g -e GO111MODULE=off -- go build \n " +
				"gmct docker -g -e GO111MODULE=off -- go build -buildmode=c-archive *.go \n " +
				"gmct docker -g -- go build -buildmode=c-archive *.go \n",
			RunE: func(c *cobra.Command, a []string) error {
				if len(a) == 0 {
					return fmt.Errorf("execute arguments required")
				}
				srv := NewDocker(Args{
					DArg_v:    util.Must(c.Flags().GetStringSlice("volume")).StringSlice(),
					DArg_e:    util.Must(c.Flags().GetStringSlice("env")).StringSlice(),
					DArg_p:    util.Must(c.Flags().GetStringSlice("port")).StringSlice(),
					DArg_name: util.Must(c.Flags().GetString("name")).String(),
					Image:     util.Must(c.Flags().GetString("img")).String(),
					IsDebug:   util.Must(c.Flags().GetBool("debug")).Bool(),
					Golang:    util.Must(c.Flags().GetBool("golang")).Bool(),
					WorkDir:   util.Must(c.Flags().GetString("work")).String(),
					CMD:       strings.Join(a, " "),
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}

		cmd.Flags().StringSliceP("volume", "v", []string{}, "volume")
		cmd.Flags().StringSliceP("env", "e", []string{}, "environment variable")
		cmd.Flags().StringSliceP("port", "p", []string{}, "port")
		cmd.Flags().StringP("name", "n", "", "set container name")
		cmd.Flags().String("img", "snail007/golang:latest", "image used to run program")
		cmd.Flags().Bool("debug", false, "debug output")
		cmd.Flags().BoolP("golang", "g", false, "sets some golang environment variables")
		cmd.Flags().StringP("work", "w", "/mnt", "set work dir")
		root.AddCommand(cmd)
	})
}

type Args struct {
	DArg_v    []string
	DArg_e    []string
	DArg_p    []string
	DArg_name string
	CMD       string
	Image     string
	IsDebug   bool
	Golang    bool
	WorkDir   string
}

type Docker struct {
	args Args
}

func NewDocker(args Args) *Docker {
	return &Docker{args: args}
}

func (s *Docker) init() (err error) {
	if s.args.Golang {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath, _ = os.UserHomeDir()
			gopath = filepath.Join(gopath, "go")
		}
		//d, _ := os.Getwd()
		//if !strings.HasPrefix(d, gopath) {
		//	return fmt.Errorf("you must run command in GOPATH")
		//}
		//pkg := strings.Trim(strings.Replace(d, filepath.Join(gopath, "src"), "", 1), "/")
		s.args.DArg_v = append(s.args.DArg_v, gopath+":/go")
		//s.args.DArg_e = append(s.args.DArg_e, "BUILDDIR="+pkg, "GOSUMDB=off", "CGO_ENABLED=1")
		s.args.DArg_e = append(s.args.DArg_e, "GOSUMDB=off", "CGO_ENABLED=1")

	}

	for k, v := range s.args.DArg_v {
		(s.args.DArg_v)[k] = " -v " + v + " "
	}

	for k, v := range s.args.DArg_e {
		(s.args.DArg_e)[k] = " -e " + v + " "
	}

	for k, v := range s.args.DArg_p {
		(s.args.DArg_p)[k] = " -p " + v + " "
	}

	pwd, _ := os.Getwd()
	pwd = gfile.Abs(pwd)
	s.args.DArg_v = append(s.args.DArg_v, " -v "+pwd+":/mnt ")
	return
}

func (s *Docker) Start() (err error) {
	net := ""
	if runtime.GOOS == "linux" {
		net = "--network=host"
	}
	name := ""
	if len(s.args.DArg_name) > 0 {
		name = fmt.Sprintf("--name %s", s.args.DArg_name)
	}
	cmdStr := fmt.Sprintf("docker run -t --rm -w %s %s %s %s %s %s %s %s",
		s.args.WorkDir,
		name,
		net,
		strings.Join(s.args.DArg_p, ""),
		strings.Join(s.args.DArg_e, ""),
		strings.Join(s.args.DArg_v, ""),
		s.args.Image,
		s.args.CMD,
	)
	reg := regexp.MustCompile(`\s+`)
	cmdStr = reg.ReplaceAllString(cmdStr, " ")
	if s.args.IsDebug {
		fmt.Println("exec:", cmdStr)
	}
	err = s.exec("bash", "-c", cmdStr)
	return
}

func (s *Docker) exec(command string, args ...string) (err error) {
	cmd := exec.Command(command, args...)
	cmdReaderStderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(cmdReader)
	scannerStdErr := bufio.NewScanner(cmdReaderStderr)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	go func() {
		for scannerStdErr.Scan() {
			fmt.Println(scannerStdErr.Text())
		}
	}()
	err = cmd.Start()
	if err != nil {
		return
	}
	err = cmd.Wait()
	if err != nil {
		return
	}
	return
}

func (s *Docker) Stop() {
	return
}
