package gotool

import (
	"fmt"
	glog "github.com/snail007/gmc/module/log"
	ghttp "github.com/snail007/gmc/util/http"
	gmap "github.com/snail007/gmc/util/map"
	"github.com/snail007/gmct/tool"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	installPkg string
	aliasData  = map[string]string{
		"golint":   "golang.org/x/lint/golint",
		"dlv":      "github.com/go-delve/delve/cmd/dlv",
		"gomobile": "golang.org/x/mobile/cmd/gomobile;gomobile init",
	}
)

func init() {
	if len(os.Args) >= 3 && os.Args[1] == "go" && os.Args[2] == "install" {
		if len(os.Args) == 4 {
			installPkg = os.Args[3]
			os.Args = os.Args[:3]
		}
	}
}

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
	case "lint", "check":
		if !s.commandIsExists("golint") {
			s.install("golang.org/x/lint/golint")
		}
	}
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
	case "install":
		if installPkg == "" {
			return fmt.Errorf("package required")
		}
		return s.install("")
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

func (s *GoTool) commandIsExists(bin string) bool {
	b, _ := exec.Command("which", bin).CombinedOutput()
	return len(b) > 0
}

func (s *GoTool) install(pkg string) (err error) {
	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir(os.TempDir())
	if pkg != "" {
		installPkg = pkg
	}

	if v := aliasData[installPkg]; v != "" {
		installPkg = v
	}
	cmd := ""
	if !strings.Contains(installPkg, "/") {
		// the short name not found locally, try fetch from https://github.com/snail007/gmct/
		glog.Infof("[ %s ] not found locally, fetch from snail007/gmct ...", installPkg)
		u := "https://github.host900.com/snail007/gmct/raw/master/scripts/go/install/" + installPkg + ".sh"
		c := ghttp.NewHTTPClient()
		c.SetDNS("8.8.8.8:53")
		b, code, _, e := c.Get(u, time.Second*30, nil)
		if code != 200 {
			m := ""
			if e != nil {
				m = ", error: " + e.Error()
			}
			return fmt.Errorf("request fail, code: %d%s", code, m)
		}
		cmd = string(b)
	}
	if cmd == "" {
		cmd = `go get -u ` + installPkg
	}
	cmd = s.exportString() + cmd
	glog.Info("Install >>> " + strings.SplitN(installPkg, ";", 2)[0])
	b, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("install fail, OUTPUT: %s, ERROR: %s", string(b), err)
	}
	if len(b) > 0 {
		fmt.Println(string(b))
	}
	return
}

func (s *GoTool) exportString() string {
	env := []string{
		"GOPROXY", "https://goproxy.io,direct",
		"GONOSUMDB", "github.com,golang.org,gopkg.in"}
	for i := 0; i < len(env)-1; i += 2 {
		if os.Getenv(env[i]) == "" {
			os.Setenv(env[i], env[i+1])
		}
	}
	var export []string
	for _, v := range os.Environ() {
		if strings.ContainsAny(v, `./\ &^*()`) {
			continue
		}
		ev := strings.SplitN(v, "=", 2)
		if len(ev) != 2 {
			continue
		}
		if ev[0] == "" || ev[1] == "" {
			continue
		}
		export = append(export, "export "+v)
	}
	return strings.Join(export, ";") + ";"
}

func CmdList() []string {
	return gmap.New().MergeStrStrMap(aliasData).StringKeys()
}
