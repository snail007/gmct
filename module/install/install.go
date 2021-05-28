package installtool

import (
	"fmt"
	glog "github.com/snail007/gmc/module/log"
	ghttp "github.com/snail007/gmc/util/http"
	gmctinstall "github.com/snail007/gmct/scripts/install"
	"github.com/snail007/gmct/tool"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	installPkg string
)

func init() {
	if len(os.Args) >= 2 && (os.Args[1] == "install" || os.Args[1] == "install-force" || os.Args[1] == "uninstall") {
		if len(os.Args) == 3 {
			installPkg = os.Args[2]
			os.Args = os.Args[:2]
		}
	}
}

type Args struct {
	InstallToolName *string
	// Action install, install-force, uninstall
	Action string
}

func NewInstallToolArgs() Args {
	return Args{
		InstallToolName: new(string),
	}
}

type InstallTool struct {
	tool.GMCTool
	args Args
}

func NewInstallTool() *InstallTool {
	return &InstallTool{}
}

func (s *InstallTool) init(args0 interface{}) (err error) {
	s.args = args0.(Args)
	return
}

func (s *InstallTool) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	if installPkg == "" {
		return fmt.Errorf("install name required")
	}
	switch s.args.Action {
	case "install":
		return s.do(s.args.Action, "", false)
	case "install-force":
		return s.do("install", "", true)
	case "uninstall":
		return s.do(s.args.Action, "", false)
	}
	return
}

func (s *InstallTool) Stop() {
	return
}

func (s *InstallTool) do(action, pkg string, force bool) (err error) {
	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir(os.TempDir())
	if pkg != "" {
		installPkg = pkg
	}
	var cmd string
	if v := gmctinstall.Scripts[installPkg]; v != "" {
		// installPkg found in locally goinstall.Scripts
		cmd = v
	} else {
		// fetch install script from https://github.com/snail007/gmct/
		glog.Infof("fetch [ %s ] from snail007/gmct ...", installPkg)
		u := "https://github.host900.com/snail007/gmct/raw/master/scripts/install/" + installPkg + ".sh"
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
	if action == "install" && force == false {
		//check if installPkg already installed
		c := "export ACTION=installed;" + s.exportString() + cmd
		_, e := exec.Command("bash", "-c", c).CombinedOutput()
		if e != nil {
			return fmt.Errorf("install fail, already installed, try run: install-force %s", installPkg)
		}
	}
	cmd = "export ACTION=" + action + ";" + s.exportString() + cmd
	glog.Infof("running [ %s ] install script ...", installPkg)
	b, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("install fail, OUTPUT: %s, ERROR: %s", string(b), err)
	}
	if len(b) > 0 {
		glog.Infof("[ %s ] install script done.", installPkg)
		fmt.Println(string(b))
	}
	return
}

func (s *InstallTool) exportString() string {
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
	return strings.Join(export, ";") + ";\n"
}
