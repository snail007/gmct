package gotool

import (
	"fmt"
	"github.com/snail007/gmct/tool"
	"os/exec"
	"strings"
)

type GoToolArgs struct {
	GoToolName *string
	SubName    *string
}

func NewGoToolArgs() GoToolArgs {
	return GoToolArgs{
		GoToolName: new(string),
		SubName:    new(string),
	}
}

type GoTool struct {
	tool.GMCTool
	args GoToolArgs
}

func NewGoTool() *GoTool {
	return &GoTool{}
}

func (s *GoTool) init(args0 interface{}) (err error) {
	s.args = args0.(GoToolArgs)
	return
}

func (s *GoTool) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	lintCmd := `golint ./... | grep -v "receiver name should be a reflection of its identity" | grep -v "should have comment"| grep -v "don't use underscores in Go names"`
	vetCmd := `go vet ./... | grep -v "uses unkeyed fields"`
	fmtCmd := `gofmt -s -w ./`
	switch *s.args.SubName {
	case "lint":
		s.do(lintCmd)
	case "vet":
		s.do(vetCmd)
	case "fmt":
		s.do(fmtCmd)
	case "check":
		s.do(lintCmd)
		s.do(vetCmd)
		s.do(fmtCmd)
	}
	return
}

func (s *GoTool) Stop() {
	return
}

func (s *GoTool) do(c string) {
	cmd := exec.Command("bash", "-c", c)
	b, _ := cmd.CombinedOutput()
	if len(b) > 0 {
		fmt.Print(strings.Trim(string(b), "\n \t"), "\n")
	}
	return
}
