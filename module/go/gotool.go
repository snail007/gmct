package gotool

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	glog "github.com/snail007/gmc/module/log"
	ghttp "github.com/snail007/gmc/util/http"
	gmap "github.com/snail007/gmc/util/map"
	goinstall "github.com/snail007/gmct/scripts/go/install"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util/config"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	installPkg string
	apiName    string
	aliasData  = map[string]string{
		"golint":   "golang.org/x/lint/golint",
		"dlv":      "github.com/go-delve/delve/cmd/dlv",
		"gomobile": "golang.org/x/mobile/cmd/gomobile;gomobile init",
	}
)
var checkArgs = func(subName string, callback func(arg string)) {
	if len(os.Args) >= 3 && os.Args[1] == "go" && os.Args[2] == subName {
		if len(os.Args) == 4 {
			callback(os.Args[3])
			os.Args = os.Args[:3]
		}
	}
}

func init() {
	checkArgs("install", func(arg string) {
		installPkg = arg
	})
	checkArgs("api", func(arg string) {
		apiName = arg
	})
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
	fmtCmd := `gofmt -s -w ./`
	switch *s.args.SubName {
	case "lint", "check":
		if !s.commandIsExists("golint") {
			s.install("golang.org/x/lint/golint")
		}
	}
	goodCode := "^_^ The project has a good coding. ^_^"
	switch *s.args.SubName {
	case "lint":
		if !s.do(lintCmd) {
			fmt.Println(goodCode)
		}
	case "vet":
		if !s.do(vetCmd) {
			fmt.Println(goodCode)
		}
	case "fmt":
		s.do(fmtCmd)
	case "check":
		a := s.do(lintCmd)
		b := s.do(vetCmd)
		s.do(fmtCmd)
		if !a && !b {
			fmt.Println(goodCode)
		}
	case "install":
		if installPkg == "" {
			return fmt.Errorf("package required")
		}
		return s.install("")
	case "api":
		if apiName == "" || !strings.Contains(apiName, ".") {
			return fmt.Errorf("api path required")
		}
		s.api()
	}
	return
}

func (s *GoTool) Stop() {
	return
}

func (s *GoTool) api() {
	info := strings.Split(apiName, ".")
	path := info[0]
	method := info[1]
	queryURL := fmt.Sprintf("https://pkg.go.dev/%s", path)
	client := ghttp.NewHTTPClient()
	client.SetDNS("8.8.8.8:53")
	client.SetProxyFromEnv(true)
	d, err := client.Download(queryURL, time.Second*30, nil)
	if err != nil {
		glog.Errorf("fetch api info fail, maybe you need to set HTTP_PROXY=<PROXY_SERVER_URL> environment, access url error: %s", queryURL)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(d))
	if err != nil {
		glog.Errorf("parse api info fail, error: %s,\ncontent: %s", err, string(d))
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
	if pkg != "" {
		installPkg = pkg
	}

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

func CmdList() []string {
	return gmap.New().MergeStrStrMap(aliasData).StringKeys()
}
