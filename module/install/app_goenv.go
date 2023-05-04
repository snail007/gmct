package installtool

import (
	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"path/filepath"
	"strings"
)

func init() {
	AddInstaller(GoEnvAPPName, NewGoEnvInstaller())
}

type GoEnvInstaller struct {
}

func NewGoEnvInstaller() *GoEnvInstaller {
	return &GoEnvInstaller{}
}

func (s *GoEnvInstaller) Install(version string) error {
	cmd := `
# added by gmct install
export GOPROXY=https://goproxy.cn,direct
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
# sets repository which bypass GOPROXY，split by comma
# export GOPRIVATE="*.example.com"
# sets repository which bypass checking certificate，split by comma
# export GOINSECURE="example.com"
# export GONOPROXY=""
# export GONOSUMDB=""
# export GONOSUMDB="off"
`
	bashProfileFile := filepath.Join(gfile.HomeDir(), ".bash_profile")
	d := gfile.Bytes(bashProfileFile)
	if len(d) == 0 || !strings.Contains(string(d), "GOROOT=/usr/local/go") {
		err := gfile.WriteString(bashProfileFile, cmd, true)
		if err != nil {
			glog.Fatalf("write environment info to %s, error: %s", bashProfileFile, err)
		}
	}

	glog.Infof("install %s SUCCESS, reopen bash to take changing be working, or exec: source /etc/profile. Change by edit it.", bashProfileFile)
	return nil
}

func (s *GoEnvInstaller) InstallForce(version string) error {
	return nil
}

func (s *GoEnvInstaller) Uninstall(version string) error {
	return nil
}

func (s *GoEnvInstaller) Exists(version string) bool {
	return false
}

func (s *GoEnvInstaller) NeedRoot() bool {
	return false
}
