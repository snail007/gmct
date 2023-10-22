package installtool

import (
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	URL "net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	ghttp "github.com/snail007/gmc/util/http"
	gmctinstall "github.com/snail007/gmct/scripts/install"
)

var (
	installBaseURLEnvKey  = "GMCT_INSTALL_BASE_URL"
	installPkg            string
	defaultInstallBaseURL = "https://mirrors.host900.com/https://raw.githubusercontent.com/snail007/gmct/master/scripts/install/"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		s := NewInstallTool()
		cmd := &cobra.Command{
			Use:     "install",
			Long:    "install toolkit",
			Aliases: []string{"install-force"},
			RunE: func(c *cobra.Command, a []string) error {
				if len(a) == 0 {
					return fmt.Errorf("args required")
				}
				force := c.Name() == "install-force"
				appVersion, installer := s.getInstaller(a[0])
				if force {
					if installer != nil {
						return installer.InstallForce(appVersion)
					}
					return s.do("install", "", true)
				} else {
					if installer != nil {
						return installer.Install(appVersion)
					}
					return s.do("install", "", false)
				}
			},
		}
		root.AddCommand(&cobra.Command{
			Use:  "uninstall",
			Long: "uninstall staff installed by install toolkit",
			RunE: func(c *cobra.Command, a []string) error {
				if len(a) == 0 {
					return fmt.Errorf("args required")
				}
				appVersion, installer := s.getInstaller(a[0])
				if installer != nil {
					return installer.Uninstall(appVersion)
				}
				return s.do("uninstall", "", false)
			},
		})
		root.AddCommand(cmd)
	})
}

type InstallTool struct {
}

func NewInstallTool() *InstallTool {
	return &InstallTool{}
}

func (s *InstallTool) getInstaller(installPkg string) (appVersion string, installer APPInstaller) {
	for k, v := range appInstallers {
		if !strings.HasPrefix(installPkg, k) {
			continue
		}
		if ok, _ := regexp.MatchString(`^\w+\d+\.\d+`, installPkg); installPkg != k && !ok {
			continue
		}
		appVersion = strings.TrimPrefix(installPkg, k)
		installer = v
		s.checkRoot(v)
		break
	}
	return
}

func (s *InstallTool) checkRoot(installer APPInstaller) {
	if installer.NeedRoot() && runtime.GOOS != "windows" && os.Getuid() != 0 {
		glog.Fatal("install need root permission")
	}
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
		installBaseURL := os.Getenv(installBaseURLEnvKey)
		if installBaseURL == "" {
			installBaseURL = defaultInstallBaseURL
		}
		glog.Infof("fetch [ %s ] from %s", installPkg, installBaseURL)
		u := strings.TrimSuffix(installBaseURL, "/") + "/" + installPkg + ".sh"
		url, e := URL.Parse(u)
		if e != nil {
			return fmt.Errorf("parse url fail, error: %s", e)
		}
		switch url.Scheme {
		case "file":
			content := gfile.Bytes(url.Path)
			if len(content) == 0 {
				return fmt.Errorf("get content fail, file: %s", url.Path)
			}
			cmd = string(content)
		case "http", "https":
			c := ghttp.NewHTTPClient()
			c.SetDNS("8.8.8.8:53")
			c.SetProxyFromEnv(true)
			b, code, _, e := c.Get(u, time.Second*30, nil, nil)
			if code != 200 {
				m := ""
				if e != nil {
					m = ", error: " + e.Error()
				}
				return fmt.Errorf("request fail, code: %d%s", code, m)
			}
			cmd = string(b)
		default:
			return fmt.Errorf("unknown url scheme [%s]", u)
		}
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
	glog.Infof("installing [ %s ] ...", installPkg)
	command := exec.Command("bash", "-c", cmd)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		return fmt.Errorf("install FAIL, ERROR: %s", err)
	}
	glog.Infof("[ %s ] install DONE.", installPkg)
	return
}

func (s *InstallTool) exportString() string {
	env := []string{
		"GOPROXY", "https://goproxy.io,direct",
		"GONOSUMDB", "github.com,golang.org,gopkg.in",
		"GOOS", runtime.GOOS,
		"GOARCH", runtime.GOARCH,
	}
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
