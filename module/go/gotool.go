package gotool

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	glog "github.com/snail007/gmc/module/log"
	gexec "github.com/snail007/gmc/util/exec"
	gfile "github.com/snail007/gmc/util/file"
	ghttp "github.com/snail007/gmc/util/http"
	gmap "github.com/snail007/gmc/util/map"
	ghook "github.com/snail007/gmc/util/process/hook"
	grand "github.com/snail007/gmc/util/rand"
	gvalue "github.com/snail007/gmc/util/value"
	"github.com/snail007/gmct/module/module"
	goinstall "github.com/snail007/gmct/scripts/go/install"
	"github.com/snail007/gmct/util/config"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	aliasData = map[string]string{
		"golint":   "golang.org/x/lint/golint",
		"dlv":      "github.com/go-delve/delve/cmd/dlv",
		"gomobile": "golang.org/x/mobile/cmd/gomobile;gomobile init",
	}
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		goodCode := "^_^ The project has a good coding. ^_^"
		s := NewGoTool()
		goCMD := &cobra.Command{
			Use: "go",
			PersistentPreRunE: func(c *cobra.Command, a []string) error {
				if (c.Name() == "install" || c.Name() == "api") && len(a) != 1 {
					return fmt.Errorf("arg required")
				}
				if c.Name() == "link" || c.Name() == "check" {
					if !s.commandIsExists("golint") {
						s.install("golang.org/x/lint/golint")
					}
				}
				return nil
			},
		}
		goCMD.AddCommand(&cobra.Command{
			Use: "install",
			RunE: func(c *cobra.Command, a []string) error {
				return s.install(a[0])
			},
		})
		goCMD.AddCommand(&cobra.Command{
			Use: "api",
			RunE: func(c *cobra.Command, a []string) error {
				if !strings.Contains(a[0], ".") {
					return fmt.Errorf("arg required")
				}
				s.api(a[0])
				return nil
			},
		})

		goCMD.AddCommand(&cobra.Command{
			Use: "lint",
			RunE: func(c *cobra.Command, a []string) error {
				if !s.do(s.getLintCmd()) {
					fmt.Println(goodCode)
				}
				return nil
			},
		})

		goCMD.AddCommand(&cobra.Command{
			Use: "check",
			RunE: func(c *cobra.Command, args []string) error {
				a := s.do(s.getLintCmd())
				b := s.do(s.getVetCmd())
				s.do(s.getFmtCmd())
				if !a && !b {
					fmt.Println(goodCode)
				}
				return nil
			},
		})

		goCMD.AddCommand(&cobra.Command{
			Use: "vet",
			RunE: func(c *cobra.Command, a []string) error {
				if !s.do(s.getVetCmd()) {
					fmt.Println(goodCode)
				}
				return nil
			},
		})

		goCMD.AddCommand(&cobra.Command{
			Use: "fmt",
			RunE: func(c *cobra.Command, a []string) error {
				s.do(s.getFmtCmd())
				return nil
			},
		})
		pprofCMD := &cobra.Command{
			Use: "pprof",
			RunE: func(c *cobra.Command, a []string) error {
				if !strings.Contains(a[0], ".") {
					return fmt.Errorf("arg required")
				}
				for idx, v := range a {
					a[idx] = gfile.Abs(v)
				}
				tmpPath := "/tmp/pprof" + grand.String(32)
				err := os.Mkdir(tmpPath, 0755)
				if err != nil {
					return err
				}
				defer gexec.NewCommand(fmt.Sprintf("rm -rf %s", tmpPath)).Exec()
				for _, v := range a {
					_, err = gexec.NewCommand(fmt.Sprintf("cp %s %s", v, tmpPath)).Exec()
					if err != nil {
						return err
					}
				}
				os.Chdir(tmpPath)
				goVersion := gvalue.Must(c.Flags().GetString("go")).String()
				port := gvalue.Must(c.Flags().GetString("port")).String()
				volume, _ := c.Flags().GetStringSlice("volume")
				volumeStr := ""
				for _, v := range volume {
					volumeStr += " -v " + v
				}
				info, err := getProfileInfo(a)
				if err != nil {
					return err
				}
				env := gmap.Mss{
					"GOSUMDB": "off",
				}
				if p := os.Getenv("GOPROXY"); p == "" {
					env["GOPROXY"] = "https://goproxy.io,direct"
				}
				for _, pkg := range info.ImportLibraryList {
					glog.Infof("download dependency %s", pkg.ModPath())
					err = goGet(pkg.ModPath(), env, 2)
					if err != nil {
						glog.Fatalf("download dependency %s FAIL, error: %s", pkg.ModPath(), err)
					}
				}
				img := "snail007/golang:" + goVersion
				chGoRoot := ""
				if info.GoRoot != "/usr/local/go" {
					chGoRoot = "ln -s /usr/local/go " + info.GoRoot + "; "
				}
				pwd := gvalue.Must(os.Getwd()).String()
				cmd := ` bash -c "` + chGoRoot + ` go tool pprof -no_browser -http 0.0.0.0:8080 ` + gfile.BaseName(a[0]) + `"`
				dockerName := "pprof_" + grand.String(12)
				dockerCMD := fmt.Sprintf(
					`docker run --name %s --rm -i %s -v %s:/mnt -w /mnt -v %s:%s -e GOPATH=%s -p %s:8080 %s %s`,
					dockerName, volumeStr, pwd, os.Getenv("GOPATH"), info.GoPath, info.GoPath, port, img, cmd)
				command := gexec.NewCommand(dockerCMD)
				ghook.RegistShutdown(func() {
					gexec.NewCommand("docker stop " + dockerName).Detach(true).ExecAsync()
					command.Kill()
				})
				go ghook.WaitShutdown()
				glog.Infof("server on http://127.0.0.1:%s/", port)
				_, err = command.Exec()
				return err
			},
		}

		pprofCMD.Flags().String("go", "latest", "go version, example: 1.19, 1.20")
		pprofCMD.Flags().String("port", "8020", "profile web port")
		pprofCMD.Flags().StringSliceP("volume", "v", []string{}, "equal to docker -v")
		goCMD.AddCommand(pprofCMD)
		root.AddCommand(goCMD)
	})
}

type GoTool struct {
}

func NewGoTool() *GoTool {
	return &GoTool{}
}
func (s *GoTool) getVetCmd() string {
	vetSkipWords := []string{
		"vendor/",
		"should have signature",
		"_test.go:",
		"^# ",
		"possible misuse of",
	}
	vetSkipWords = append(vetSkipWords, config.Options.GetStringSlice("go.vet.skip")...)
	vetCmd := "go vet ./... 2>&1 "
	for _, v := range vetSkipWords {
		vetCmd += "| grep -v \"" + v + "\" "
	}
	return vetCmd
}

func (s *GoTool) getFmtCmd() string {
	fmtCmd := `gofmt -s -w ./`
	return fmtCmd
}

func (s *GoTool) getLintCmd() string {
	lintSkipWords := []string{
		"vendor/",
		"receiver name should be a reflection of its identity",
		"should have comment",
		"string as key in context.WithValue",
		"comment on exported function",
		"don't use underscores in Go names",
		"a blank import should be only in a main or test package",
		"_test.go:",
	}
	lintSkipWords = append(lintSkipWords, config.Options.GetStringSlice("go.lint.skip")...)
	lintCmd := "golint -min_confidence 0.9 ./... "
	for _, v := range lintSkipWords {
		lintCmd += "| grep -v \"" + v + "\" "
	}
	return lintCmd
}

func (s *GoTool) api(apiName string) {
	info := strings.Split(apiName, ".")
	path := info[0]
	method := info[1]
	queryURL := fmt.Sprintf("https://pkg.go.dev/%s", path)
	client := ghttp.NewHTTPClient()
	client.SetDNS("8.8.8.8:53")
	client.SetProxyFromEnv(true)
	d, _, err := client.Download(queryURL, time.Second*30, nil, nil)
	if err != nil {
		glog.Fatalf("fetch api info fail, maybe you need to set HTTP_PROXY=<PROXY_SERVER_URL> environment, access url error: %s", queryURL)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(d))
	if err != nil {
		glog.Fatalf("parse api info fail, error: %s,\ncontent: %s", err, string(d))
	}
	sel := doc.Find("#" + method)
	if len(sel.Nodes) == 0 {
		glog.Warnf("not found %s info in url: %s", apiName, queryURL)
		return
	}
	exists := sel.Find(".Documentation-sinceVersionVersion")
	if len(exists.Nodes) == 0 {
		fmt.Println(apiName + " added in all go version")
		return
	}
	label := sel.Find(".Documentation-sinceVersionLabel")
	version := sel.Find(".Documentation-sinceVersionVersion")
	fmt.Printf("%s %s %s\n", apiName, label.Text(), version.Text())
}

func (s *GoTool) do(c string) (hasOutput bool) {
	cmd := exec.Command("bash", "-c", c)
	b, _ := cmd.CombinedOutput()
	if len(b) > 0 {
		fmt.Print(strings.Trim(string(b), "\n \t"), "\n")
	}
	hasOutput = len(bytes.Trim(b, "\r\n")) > 0
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
	installPkg := pkg
	if v := aliasData[installPkg]; v != "" {
		installPkg = v
	}
	cmd := ""
	if !strings.Contains(installPkg, "/") {
		// installPkg is short name.
		// find short name in goscripts.Scripts
		if v := goinstall.Scripts[installPkg]; v != "" {
			cmd = v
		} else {
			// the short name not found locally, try fetch from https://github.com/snail007/gmct/
			glog.Infof("[ %s ] not found locally, fetch from snail007/gmct ...", installPkg)
			u := "https://mirrors.host900.com/https://raw.githubusercontent.com/snail007/gmct/master/scripts/go/install/" + installPkg + ".sh"
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
		}
	}
	if cmd == "" {
		cmd = `go get -u ` + installPkg
	}
	cmd = "export ACTION=install;" + s.exportString() + cmd
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
	return strings.Join(export, ";") + ";\n"
}
