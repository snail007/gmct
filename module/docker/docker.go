package docker

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmct/tool"
)

type DockerArgs struct {
	SubName   *string
	DArg_v    *[]string
	DArg_e    *[]string
	DArg_p    *[]string
	DArg_name *string
	CMD       *string
	Image     *string
	IsDebug   *bool
	Golang    *bool
	WorkDir   *string
}

func NewDockerArgs() DockerArgs {
	return DockerArgs{
		SubName:   new(string),
		DArg_v:    new([]string),
		DArg_e:    new([]string),
		DArg_p:    new([]string),
		CMD:       new(string),
		Image:     new(string),
		IsDebug:   new(bool),
		Golang:    new(bool),
		WorkDir:   new(string),
		DArg_name: new(string),
	}
}

type Docker struct {
	tool.GMCTool
	args DockerArgs
}

func NewDocker() *Docker {
	return &Docker{}
}

func (s *Docker) init(args0 interface{}) (err error) {
	s.args = args0.(DockerArgs)
	if len(tool.Args) == 0 {
		return fmt.Errorf("execute arguments required")
	}
	*s.args.CMD = strings.Join(tool.Args, " ")

	if *s.args.Golang {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath, _ = os.UserHomeDir()
			gopath = filepath.Join(gopath, "go")
		}
		d, _ := os.Getwd()
		if !strings.HasPrefix(d, gopath) {
			return fmt.Errorf("you must run command in GOPATH")
		}
		pkg := strings.Trim(strings.Replace(d, filepath.Join(gopath, "src"), "", 1), "/")
		*s.args.DArg_v = append(*s.args.DArg_v, gopath+":/go")
		*s.args.DArg_e = append(*s.args.DArg_e, "BUILDDIR="+pkg, "GOSUMDB=off", "CGO_ENABLED=1")
	}

	for k, v := range *s.args.DArg_v {
		(*s.args.DArg_v)[k] = " -v " + v + " "
	}

	for k, v := range *s.args.DArg_e {
		(*s.args.DArg_e)[k] = " -e " + v + " "
	}

	for k, v := range *s.args.DArg_p {
		(*s.args.DArg_p)[k] = " -p " + v + " "
	}

	pwd, _ := os.Getwd()
	pwd = gfile.Abs(pwd)
	*s.args.DArg_v = append(*s.args.DArg_v, " -v "+pwd+":/mnt ")
	return
}

func (s *Docker) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	net := ""
	if runtime.GOOS == "linux" {
		net = "--network=host"
	}
	name := ""
	if len(*s.args.DArg_name) > 0 {
		name = fmt.Sprintf("--name %s", *s.args.DArg_name)
	}
	cmdStr := fmt.Sprintf("docker run -t --rm -w %s %s %s %s %s %s %s %s",
		*s.args.WorkDir,
		name,
		net,
		strings.Join(*s.args.DArg_p, ""),
		strings.Join(*s.args.DArg_e, ""),
		strings.Join(*s.args.DArg_v, ""),
		*s.args.Image,
		*s.args.CMD,
	)
	reg := regexp.MustCompile(`\s+`)
	cmdStr = reg.ReplaceAllString(cmdStr, " ")
	if *s.args.IsDebug {
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
