package gotool

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	glog "github.com/snail007/gmc/module/log"
	gbatch "github.com/snail007/gmc/util/batch"
	gcond "github.com/snail007/gmc/util/cond"
	gexec "github.com/snail007/gmc/util/exec"
	gfile "github.com/snail007/gmc/util/file"
	ghttp "github.com/snail007/gmc/util/http"
	gmap "github.com/snail007/gmc/util/map"
	gnet "github.com/snail007/gmc/util/net"
	ghook "github.com/snail007/gmc/util/process/hook"
	grand "github.com/snail007/gmc/util/rand"
	gvalue "github.com/snail007/gmc/util/value"
	"github.com/snail007/gmct/module/module"
	goinstall "github.com/snail007/gmct/scripts/go/install"
	"github.com/snail007/gmct/util/config"
	"github.com/snail007/gmct/util/profile"
	"github.com/spf13/cobra"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	aliasData = map[string]string{
		"golint":   "golang.org/x/lint/golint",
		"dlv":      "github.com/go-delve/delve/cmd/dlv",
		"gomobile": "golang.org/x/mobile/cmd/gomobile;gomobile init",
	}
	pprofTypes = [][]string{
		{"cpu", "profile"},
		{"mem", "heap"},
		{"goroutine", "goroutine"},
		{"allocs", "allocs"},
		{"block", "block"},
		{"mutex", "mutex"},
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
			RunE: func(c *cobra.Command, args []string) (err error) {
				files := args
				isURL := len(args) == 1 && strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://")
				if isURL {
					files, err = pprofFetch(c, args)
					if err != nil {
						return err
					}
					defer func() {
						for _, v := range files {
							os.Remove(v)
						}
					}()
					addr := gvalue.Must(c.Flags().GetString("analyze")).String()
					if addr != "" {
						fs := &PprofFiles{}
						var getFilename = func(n string) string {
							return n + ".pprof"
						}
						if f := getFilename("cpu"); gfile.Exists(f) {
							fs.CpuFile = f
						}
						if f := getFilename("mem"); gfile.Exists(f) {
							fs.MemFile = f
						}
						if f := getFilename("goroutine"); gfile.Exists(f) {
							fs.GoroutineFile = f
						}
						if f := getFilename("allocs"); gfile.Exists(f) {
							fs.AllocsFile = f
						}
						if f := getFilename("block"); gfile.Exists(f) {
							fs.BlockFile = f
						}
						if f := getFilename("mutex"); gfile.Exists(f) {
							fs.MutexFile = f
						}
						err = startAnalyzeServer(addr, fs)
						if err != nil {
							return err
						}
					}
				}
				cnt := 0
				for _, v := range pprofTypes {
					if gvalue.Must(c.Flags().GetBool(v[0])).Bool() {
						cnt++
					}
				}
				shouldEmpty := (isURL && cnt > 1) || (!isURL && len(args) > 1)
				if shouldEmpty && gvalue.Must(c.Flags().GetString("port")).String() != "" {
					return fmt.Errorf("port only for one profile")
				}
				return pprof(c, files)
			},
		}
		bindPprofCommonFlags := func(c *cobra.Command) {
			c.Flags().IntP("duration", "d", 30, "seconds of profiling")
			c.Flags().Bool("cpu", false, "fetch cpu profile in url mode")
			c.Flags().Bool("mem", false, "fetch heap profile in url mode")
			c.Flags().Bool("goroutine", false, "fetch goroutine profile in url mode")
			c.Flags().Bool("allocs", false, "fetch allocs profile in url mode")
			c.Flags().Bool("block", false, "fetch block profile in url mode")
			c.Flags().Bool("mutex", false, "fetch mutex profile in url mode")
			c.Flags().StringSliceP("ignore", "i", []string{}, "ignored library to download")
			c.Flags().String("pre", "", "access the url before fetch mutex and block profile")
			c.Flags().String("post", "", "access the url after fetch mutex and block profile")

		}

		pprofCMD.Flags().String("go", "latest", "go version, example: 1.19, 1.20")
		pprofCMD.Flags().String("port", "", "profile web port, default: random")
		pprofCMD.Flags().StringSliceP("volume", "v", []string{}, "equal to docker -v")
		pprofCMD.Flags().StringP("analyze", "a", "", "the ip:port analyze server listen on in url pprof mode")

		bindPprofCommonFlags(pprofCMD)

		fetchCMD := &cobra.Command{
			Use: "fetch",
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				_, err = pprofFetch(cmd, args)
				return
			},
		}
		bindPprofCommonFlags(fetchCMD)

		cleanPprofCMD := &cobra.Command{
			Use: "clean",
			Run: func(c *cobra.Command, a []string) {
				gexec.NewCommand(`
rm -rf /tmp/pprof_tmp_*
rm -rf /tmp/tmp_*.sh
rm -rf /tmp/gogetmod_*
`).Exec()
				out, _ := gexec.NewCommand("docker ps |grep gmct_pprof_").Exec()
				be := gbatch.NewBatchExecutor()
				for _, l := range strings.Split(out, "\n") {
					arr := strings.Fields(l)
					if len(arr) < 1 {
						continue
					}
					name := arr[len(arr)-1]
					be.AppendTask(func(ctx context.Context) (value interface{}, err error) {
						glog.Infof("killing %s", name)
						gexec.NewCommand("docker stop " + name).Exec()
						return nil, nil
					})
				}
				be.WaitAll()
				return
			},
		}
		analyzeCMD := &cobra.Command{
			Use: "analyze",
			RunE: func(c *cobra.Command, a []string) error {
				files := &PprofFiles{
					CpuFile:       gvalue.Must(c.Flags().GetString("cpu")).String(),
					MemFile:       gvalue.Must(c.Flags().GetString("mem")).String(),
					AllocsFile:    gvalue.Must(c.Flags().GetString("allocs")).String(),
					GoroutineFile: gvalue.Must(c.Flags().GetString("go")).String(),
					MutexFile:     gvalue.Must(c.Flags().GetString("mutex")).String(),
					BlockFile:     gvalue.Must(c.Flags().GetString("block")).String(),
				}
				addr := gvalue.Must(c.Flags().GetString("addr")).String()
				err := startAnalyzeServer(addr, files)
				if err != nil {
					return err
				}
				ghook.WaitShutdown()
				return nil
			},
		}
		analyzeCMD.Flags().String("cpu", "", "pprof cpu profile")
		analyzeCMD.Flags().String("mem", "", "pprof heap profile")
		analyzeCMD.Flags().String("allocs", "", "pprof allocs profile")
		analyzeCMD.Flags().String("go", "", "pprof goroutine profile")
		analyzeCMD.Flags().String("mutex", "", "pprof mutex profile")
		analyzeCMD.Flags().String("block", "", "pprof block profile")
		analyzeCMD.Flags().String("addr", "", "ip:port http server listen on")

		pprofCMD.AddCommand(cleanPprofCMD)
		pprofCMD.AddCommand(fetchCMD)
		pprofCMD.AddCommand(analyzeCMD)

		goCMD.AddCommand(pprofCMD)

		root.AddCommand(pprofCMD)
		root.AddCommand(goCMD)
	})
}

func startAnalyzeServer(addr string, files *PprofFiles) error {
	addr = gcond.Cond(addr == "", ":0", addr).String()
	addr = gcond.Cond(strings.Contains(addr, ":"), addr, ":"+addr).String()
	h, p, _ := net.SplitHostPort(addr)
	h = gcond.Cond(h == "", "0.0.0.0", h).String()
	p = gcond.Cond(p == "", "0", p).String()
	addr = net.JoinHostPort(h, p)
	if !files.IsExistsMemFile() || !files.IsExistsCpuFile() {
		return fmt.Errorf("cpu and mem profile are rqeuired")
	}
	server, err := NewPprofServer(addr, files)
	if err != nil {
		return err
	}
	server.server.Logger().SetOutput(glog.NewLoggerWriter(io.Discard))
	err = server.Start()
	if err != nil {
		return err
	}
	glog.Infof("analyze http server on http://127.0.0.1:%s", gnet.NewTCPAddr(server.server.Address()).Port())
	return nil
}

func pprofFetch(c *cobra.Command, a []string) (files []string, err error) {
	if len(a) == 0 {
		return nil, fmt.Errorf("arg required")
	}
	dur := gvalue.MustAny(c.Flags().GetInt("duration"))
	baseURL := strings.TrimSuffix(a[0], "/")
	preURL := gvalue.Must(c.Flags().GetString("pre")).String()
	postURL := gvalue.Must(c.Flags().GetString("post")).String()
	var isHTTPURL = func(str string) bool {
		return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
	}
	if preURL != "" && !isHTTPURL(preURL) {
		preURL = baseURL + "/" + strings.TrimPrefix(preURL, "/")
	}
	if postURL != "" && !isHTTPURL(postURL) {
		postURL = baseURL + "/" + strings.TrimPrefix(postURL, "/")
	}
	var accessURL = func(u, typ string) {
		if u == "" {
			return
		}
		_, code, _, e := ghttp.Get(u, time.Second*10, nil, nil)
		if e != nil {
			glog.Warnf("access "+typ+" url fail, error: %s", e)
		} else if code != 200 {
			glog.Warnf("access "+typ+" url fail, response code: %d", code)
		}
	}
	accessURL(preURL, "pre")
	be := gbatch.NewBatchExecutor()
	timeout := time.Second * time.Duration(dur.Int()*2)
	cnt := 0
	for _, v := range pprofTypes {
		typ := v[0]
		path := v[1]
		if !gvalue.Must(c.Flags().GetBool(typ)).Bool() {
			continue
		}
		filename := typ + ".pprof"
		files = append(files, filename)
		cnt++
		be.AppendTask(func(ctx context.Context) (value interface{}, err error) {
			link := baseURL + "/" + path + "?seconds=" + dur.String()
			resp, err := ghttp.DownloadToFile(link, timeout, nil, nil, filename)
			if resp != nil && resp.StatusCode != 200 {
				err = fmt.Errorf("reponse %d", resp.StatusCode)
				return
			}
			return
		})
	}
	if cnt == 0 {
		return nil, fmt.Errorf("at least one type to profiling, one of cpu,mem,goroutine,allocs,mutex,block")
	}
	rs := be.WaitAll()
	accessURL(postURL, "post")
	for _, r := range rs {
		if r.Err() != nil {
			return nil, r.Err()
		}
	}
	return
}

func isExistsDockerImage(name, version string) bool {
	out, _ := gexec.NewCommand(fmt.Sprintf("docker images %s:%s", name, version)).Exec()
	return strings.Contains(out, name) && strings.Contains(out, version)
}

func pprof(c *cobra.Command, a []string) error {
	if len(a) == 0 {
		return fmt.Errorf("arg required")
	}
	for idx, v := range a {
		a[idx] = gfile.Abs(v)
	}
	tmpPath := "/tmp/pprof_tmp_" + grand.String(32)
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
	info, err := gprofile.GetProfileInfo(a)
	if err != nil {
		return err
	}
	imgName := "snail007/golang"
	img := imgName + ":" + goVersion
	if !isExistsDockerImage(imgName, goVersion) {
		glog.Infof("docker pull %s", img)
		if _, e := gexec.NewCommand("docker pull " + img).Exec(); e != nil {
			return e
		}
	}
	env := gmap.Mss{
		"GOSUMDB": "off",
	}
	if p := os.Getenv("GOPROXY"); p == "" {
		env["GOPROXY"] = "https://goproxy.io,direct"
	}
	localGopath := os.Getenv("GOPATH")
	var importLibraryList []gprofile.LibraryInfoItem
	ignoredList := gvalue.Must(c.Flags().GetStringSlice("ignore")).StringSlice()
CON:
	for _, pkg := range info.ImportLibraryList {
		for _, v := range ignoredList {
			if v == pkg.ModPath() {
				continue CON
			}
		}
		if gfile.IsDir(filepath.Join(localGopath, "pkg/mod", pkg.ModFullPath())) {
			glog.Infof("found dependency %s exists", pkg.ModFullPath())
		} else {
			importLibraryList = append(importLibraryList, pkg)
			glog.Infof("found dependency %s downloading...", pkg.ModFullPath())
		}
	}
	err = gprofile.GoGet(importLibraryList, env, 2)
	if err != nil {
		glog.Fatalf("download dependency FAIL, error: %s", err)
	}
	chGoRoot := ""
	if info.GoRoot != "/usr/local/go" {
		chGoRoot = "ln -s /usr/local/go " + info.GoRoot + "; "
	}
	pwd := gvalue.Must(os.Getwd()).String()
	dockerNames := []string{}
	rid := grand.String(12)
	for _, v1 := range a {
		filename := filepath.Base(v1)
		dockerName := "gmct_pprof_" + gfile.FileName(filename) + "_" + rid
		dockerNames = append(dockerNames, dockerName)
		p := port
		if p == "" {
			p, _ = gnet.RandomPort()
		}

		cmd := ` bash -c "` + chGoRoot + ` go tool pprof -no_browser -http 0.0.0.0:8080 ` + filename + `"`
		gopathCmdStr := ""
		if info.GoPath != "" {
			gopathCmdStr = fmt.Sprintf(" -v %s:%s -e GOPATH=%s ", os.Getenv("GOPATH"), info.GoPath, info.GoPath)
		}
		dockerCMD := fmt.Sprintf(
			`docker run --name %s --rm -i %s -v %s:/mnt -w /mnt %s  -p %s:8080 %s %s`,
			dockerName, volumeStr, pwd, gopathCmdStr, p, img, cmd)
		buf := bytes.NewBuffer(nil)
		err = gexec.NewCommand(dockerCMD).Detach(true).Output(buf).ExecAsync()
		if err != nil {
			return err
		}
		time.Sleep(time.Second)
		if !strings.Contains(buf.String(), "0.0.0.0:8080") {
			glog.Warnf("[" + filename + "] failed")
		} else {
			glog.Infof("["+filename+"] on http://127.0.0.1:%s/", p)
		}
	}
	clean := func() {
		for _, n := range dockerNames {
			gexec.NewCommand("docker stop " + n).Detach(true).ExecAsync()
		}
	}
	defer clean()
	ghook.RegistShutdown(clean)
	ghook.WaitShutdown()
	return err
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
	pkg := info[0]
	method := strings.Join(info[1:], ".")
	queryURL := fmt.Sprintf("https://pkg.go.dev/%s", pkg)
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
	sel := doc.Find("#" + strings.Replace(method, ".", `\.`, -1))
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
