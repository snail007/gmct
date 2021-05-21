package docker

import (
	"bufio"
	"fmt"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmct/tool"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type DockerArgs struct {
	SubName *string
	DArg_v  *[]string
	DArg_e  *[]string
	DArg_p  *[]string
	CMD     *string
	Image   *string
	IsDebug *bool
}

func NewDockerArgs() DockerArgs {
	return DockerArgs{
		SubName: new(string),
		DArg_v:  new([]string),
		DArg_e:  new([]string),
		DArg_p:  new([]string),
		CMD:     new(string),
		Image:   new(string),
		IsDebug: new(bool),
	}
}

type Docker struct {
	tool.GMCTool
	args DockerArgs
}

func NewDocker() *Docker {
	return &Docker{
	}
}

func (s *Docker) init(args0 interface{}) (err error) {
	s.args = args0.(DockerArgs)
	if len(tool.Args) == 0 {
		return fmt.Errorf("execute arguments required")
	}
	*s.args.CMD = strings.Join(tool.Args, " ")

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
	cmdStr := fmt.Sprintf("docker run -t --rm -w /mnt %s %s %s %s %s %s",
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
