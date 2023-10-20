package main

import (
	"fmt"
	gcore "github.com/snail007/gmc/core"
	glog "github.com/snail007/gmc/module/log"
	"github.com/snail007/gmct/module/controller"
	"github.com/snail007/gmct/module/cover"
	"github.com/snail007/gmct/module/docker"
	"github.com/snail007/gmct/module/file_repeat"
	gotool "github.com/snail007/gmct/module/go"
	"github.com/snail007/gmct/module/gtag"
	"github.com/snail007/gmct/module/i18n"
	installtool "github.com/snail007/gmct/module/install"
	"github.com/snail007/gmct/module/model"
	newx "github.com/snail007/gmct/module/new"
	"github.com/snail007/gmct/module/run"
	ssht "github.com/snail007/gmct/module/ssh"
	tlstool "github.com/snail007/gmct/module/tls"
	toolx "github.com/snail007/gmct/module/tool"
	"github.com/snail007/gmct/module/update"
	gurl "github.com/snail007/gmct/module/url"
	"github.com/snail007/gmct/module/view"
	"os"
	"strings"

	"github.com/snail007/gmct/module/static"

	"github.com/snail007/gmct/module/template"
	"github.com/snail007/gmct/tool"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	tool.Version = version
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println(version)
		os.Exit(0)
	}
	glog.SetFlag(gcore.LogFlagNormal)
}

func main() {
	runCmdArgs := []string{}
	if len(os.Args) >= 2 && os.Args[1] == "run" {
		if len(os.Args) > 2 {
			runCmdArgs = os.Args[2:]
		}
		os.Args = os.Args[:2]
	}
	gmctApp := kingpin.New("gmct", "toolchain for go web framework gmc, https://github.com/snail007/gmc")
	gmctApp.Version(version)

	// #1

	// all subtool args defined here
	templateArgs := template.NewTemplateArgs()
	staticArgs := static.NewStaticArgs()
	runArgs := run.NewRunArgs()
	newArgs := newx.NewArgs()
	i18nArgs := i18n.NewI18nArgs()
	controllerArgs := controller.NewControllerArgs()
	modelArgs := model.NewModelArgs()
	gtagArgs := gtag.NewGTagArgs()
	coverArgs := cover.NewCoverArgs()
	viewArgs := view.NewViewArgs()
	dockerArgs := docker.NewDockerArgs()
	toolArgs := toolx.NewToolArgs()
	sshArgs := ssht.NewSshArgs()
	updateArgs := update.NewUpdateArgs()
	goToolArgs := gotool.NewGoToolArgs()
	installToolArgs := installtool.NewInstallToolArgs()
	tlsToolArgs := tlstool.NewTLSArgs()
	urlArgs := gurl.NewArgs()
	fileRepeatArgs := file_repeat.NewFileRepeatArgs()
	//all command defined here

	// #2
	// command template
	templateCmd := gmctApp.Command("tpl", "pack or clean templates go file")
	templateArgs.Dir = templateCmd.Flag("dir", "template's template directory path, gmct will convert all template files in the folder to one go file").Default(".").String()
	templateArgs.Extension = templateCmd.Flag("ext", "extension of template file").Default(".html").String()
	templateArgs.Clean = templateCmd.Flag("clean", "clean packed file, if exists").Default("false").Bool()

	// command static
	staticCmd := gmctApp.Command("static", "pack or clean static go file")
	staticArgs.Dir = staticCmd.Flag("dir", "template's static directory path, gmct will convert all static files in the folder to one go file").Default(".").String()
	staticArgs.NotExtension = staticCmd.Flag("ext", "extension of exclude static files ").Default("").String()
	staticArgs.Clean = staticCmd.Flag("clean", "clean packed file, if exists").Default("false").Bool()

	// command run
	gmctApp.Command("run", "run gmc project with auto build when project's file changed")

	// command new
	newCMD := gmctApp.Command("new", "new a gmc web/api project")
	newArgsWebCMD := newCMD.Command("web", "new a gmc web project")
	newArgsAPICMD := newCMD.Command("api", "new a gmc api project")
	newArgsSimpleAPICMD := newCMD.Command("api-simple", "new a simple gmc api project")
	newArgsAdminCMD := newCMD.Command("admin", "new a gmc admin project")
	// new web args
	newArgs.Web.Package = newArgsWebCMD.Flag("pkg", "package path of project in GOPATH").Default("").String()
	// new api args
	newArgs.API.Package = newArgsAPICMD.Flag("pkg", "package path of project in GOPATH").Default("").String()
	// new simple api args
	newArgs.SimpleAPI.Package = newArgsSimpleAPICMD.Flag("pkg", "package path of project in GOPATH").Default("").String()
	// new admin args
	newArgs.Admin.Package = newArgsAdminCMD.Flag("pkg", "package path of project in GOPATH").Default("").String()

	// command i18n
	i18nCmd := gmctApp.Command("i18n", "pack or clean i18n go file")
	i18nArgs.Dir = i18nCmd.Flag("dir", "i18n's template directory path, gmct will convert all i18n files in the folder to one go file").Default(".").String()
	i18nArgs.Clean = i18nCmd.Flag("clean", "clean packed file, if exists").Default("false").Bool()

	// command controller
	controllerCmd := gmctApp.Command("controller", "create a controller in current directory")
	controllerArgs.ControllerName = controllerCmd.Flag("name", "controller struct name").Short('n').Default("").String()
	controllerArgs.TableName = controllerCmd.Flag("table", "table name without prefix").Short('t').Default("").String()
	controllerArgs.ForceCreate = controllerCmd.Flag("force", "overwrite controller file, if it exists.").Short('f').Default("false").Bool()

	// command model
	modelCmd := gmctApp.Command("model", "create a model in current directory")
	modelArgs.Table = modelCmd.Flag("table", "table name without suffix").Short('n').Default("").String()
	modelArgs.ForceCreate = modelCmd.Flag("force", "overwrite model file, if it exists.").Short('f').Default("false").Bool()

	// command gtag
	gmctApp.Command("gtag", "print go mod require tag of git repository in current directory")

	// command cover
	coverCmd := gmctApp.Command("cover", "print go mod require tag of git repository in current directory")
	coverArgs.Race = coverCmd.Flag("race", "enable race checking").Short('r').Default("false").Bool()
	coverArgs.Verbose = coverCmd.Flag("verbose", "verbose testing logging output").Short('v').Default("false").Bool()
	coverArgs.KeepResult = coverCmd.Flag("keep", "kept the coverage result: coverage.txt ").Short('k').Default("false").Bool()
	coverArgs.Silent = coverCmd.Flag("silent", "silent mode, not to open a browser").Short('s').Default("false").Bool()
	coverArgs.ForceCheck = coverCmd.Flag("force", "force check the package even it not contains any *_test.go file").Short('f').Default("false").Bool()
	coverArgs.Ordered = coverCmd.Flag("order", "disable parallel run").Short('o').Default("false").Bool()
	coverArgs.Only = coverCmd.Flag("only", "only testing current directory without sub directory").Default("false").Bool()
	coverArgs.CoverPkg = coverCmd.Flag("coverpkg", "additional cover packages split by comma").Default("").String()
	coverArgs.Timeout = coverCmd.Flag("timeout", `timeout flag accept any input valid for time.ParseDuration.A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`).Default("15m").String()
	coverArgs.Debug = coverCmd.Flag("debug", "in debug mode will logging steps of testing").Default("false").Bool()

	// command view
	viewCmd := gmctApp.Command("view", "create a controller in current directory")
	viewArgs.ControllerPath = viewCmd.Flag("controller", "controller name in url path").Short('n').Default("").String()
	viewArgs.Table = viewCmd.Flag("table", "table name without prefix").Short('t').Default("").String()
	viewArgs.ForceCreate = viewCmd.Flag("force", "overwrite model file, if it exists.").Short('f').Default("false").Bool()

	// command docker
	dockerCmd := gmctApp.Command("docker", "create a model in current directory, all run arguments after -- \n "+
		"Example:  \n "+
		"gmct docker -- ./foo -u xxx \n "+
		"gmct docker -g -- go build \n "+
		"gmct docker -g -e GO111MODULE=off -- go build \n "+
		"gmct docker -g -e GO111MODULE=off -- go build -buildmode=c-archive *.go \n "+
		"gmct docker -g -- go build -buildmode=c-archive *.go \n",
	)
	dockerArgs.Image = dockerCmd.Flag("img", "image used to run program").Default("snail007/golang:1.16").String()
	dockerArgs.DArg_v = dockerCmd.Flag("volume", "volume").Short('v').Strings()
	dockerArgs.DArg_p = dockerCmd.Flag("port", "port").Short('p').Strings()
	dockerArgs.DArg_e = dockerCmd.Flag("env", "environment variable").Short('e').Strings()
	dockerArgs.DArg_name = dockerCmd.Flag("name", "set container name").Default("").Short('n').String()
	dockerArgs.IsDebug = dockerCmd.Flag("debug", "debug output").Bool()
	dockerArgs.Golang = dockerCmd.Flag("golang", "sets some golang environment variables").Short('g').Bool()
	dockerArgs.WorkDir = dockerCmd.Flag("work", "set work dir").Default("/mnt").Short('w').String()

	// command tool
	toolCMD := gmctApp.Command("tool", "gmct tools collection")

	toolIPCMD := toolCMD.Command("ip", "ip toolkit")
	_ = toolIPCMD
	//command tool http
	toolHTTPCMD := toolCMD.Command("http", "simple http server")
	toolHTTPCMD.Alias("web").Alias("www")
	toolArgs.HTTP.Addr = toolHTTPCMD.Flag("addr", "simple http server listen on").Short('l').Default(":" + toolx.DefaultPort).String()
	toolArgs.HTTP.RootDir = toolHTTPCMD.Flag("root", "simple http server root directory").Short('d').Default("./").String()
	toolArgs.HTTP.Auth = toolHTTPCMD.Flag("auth", "simple http server basic auth username:password, such as : foouser:foopassowrd ").Short('a').Strings()
	toolArgs.HTTP.Upload = toolHTTPCMD.Flag("upload", "simple http server upload url path, default `random`").Short('u').String()
	toolArgs.HTTP.ServerID = toolHTTPCMD.Flag("id", "set the server id name, example: server01").Short('i').String()

	//command tool download
	toolDownloadCMD := toolCMD.Command("download", "download file from gmct simple http server")
	toolDownloadCMD.Alias("dl")
	toolArgs.Download.Net = toolDownloadCMD.Flag("net", "network to scan, format: 192.168.1.0").Short('n').Strings()
	toolArgs.Download.Port = toolDownloadCMD.Flag("port", "gmct tool http port").Short('p').Strings()
	toolArgs.Download.File = toolDownloadCMD.Flag("file", "filename to download").Short('f').Default("*").String()
	toolArgs.Download.Name = toolDownloadCMD.Flag("name", "rename download file to").Short('m').String()
	toolArgs.Download.MaxDeepLevel = toolDownloadCMD.Flag("deep", "max directory deep level to list server files, value 0: no limit").Default("1").Short('d').Int()
	toolArgs.Download.Host = toolDownloadCMD.Flag("host", "specify a domain or ip to download, example: 192.168.1.1 or 192.168.1.1:9090. \nyou can specify auth info, example: foo_user:foo_pass@192.168.1.2").Short('h').Strings()
	toolArgs.Download.Auth = toolDownloadCMD.Flag("auth", "basic auth info, example: username:password").Short('a').String()
	toolArgs.Download.ServerID = toolDownloadCMD.Flag("id", "server id name to download files").Short('i').String()
	toolArgs.Download.DownloadAll = toolDownloadCMD.Flag("all", "download all files matched").Default("false").Bool()
	toolArgs.Download.Timeout = toolDownloadCMD.Flag("timeout", "timeout seconds to connect to server").Default("3").Short('t').Int()
	toolArgs.Download.DownloadDir = toolDownloadCMD.Flag("dir", "path to download all files").Default("download_files").Short('c').String()

	// command ssh
	toolSSH := gmctApp.Command("ssh", "ssh tool, copy  file to or execute command on remote host")
	sshArgs.File = toolSSH.Flag("copy", "<local_file>:<remote_file>, local file to copy").Short('c').String()
	sshArgs.Command = toolSSH.Flag("cmd", "command to execute, or '@file' exec script file").Short('e').String()
	sshArgs.SSHURL = toolSSH.Flag("url", "ssh info url").Short('u').String()

	// command update
	updateCMD := gmctApp.Command("update", "update gmct to the latest version")
	updateArgs.Force = updateCMD.Flag("force", "force update").Default("false").Short('f').Bool()

	// command go
	goToolCMD := gmctApp.Command("go", "go development toolkit")
	gotoolLintCMD := goToolCMD.Command("lint", "print go code issues are found. Install: go get -u golang.org/x/lint/golint")
	_ = gotoolLintCMD
	gotoolVetCMD := goToolCMD.Command("vet", "print go code issues are found")
	_ = gotoolVetCMD
	goToolFmtCMD := goToolCMD.Command("fmt", "format code in go files")
	_ = goToolFmtCMD
	goToolCheckCMD := goToolCMD.Command("check", "combine of vet, lint and fmt")
	_ = goToolCheckCMD
	goToolInstallCMD := goToolCMD.Command("install", "go package install toolkit, and short names are supported: "+strings.Join(gotool.CmdList(), ", "))
	_ = goToolInstallCMD
	goToolAPICMD := goToolCMD.Command("api", "check the standard library api added in which go version, format is full package path of a method, for example: 'gmct go api net/http.Serve', 'gmct go api io.ReadAll'")
	_ = goToolAPICMD

	// command install
	gmctApp.Command("install", "install toolkit")
	gmctApp.Command("install-force", "install toolkit")
	gmctApp.Command("uninstall", "uninstall staff installed by install toolkit")

	// command tls
	tlsToolCMD := gmctApp.Command("tls", "tls certificate toolkit")
	//tls info
	tlsInfoCMD := tlsToolCMD.Command("info", "print cert file or tls target host:port certificate info")
	tlsToolArgs.Info.Proxy = tlsInfoCMD.Flag("proxy", "proxy URL connect to address of tls target, example: http://127.0.0.1:8080").Short('p').Default("").String()
	tlsToolArgs.Info.Addr = tlsInfoCMD.Flag("addr", "address of tls target, ip:port").Short('a').Default("").String()
	tlsToolArgs.Info.File = tlsInfoCMD.Flag("file", "path of tls certificate file").Short('f').Default("").String()
	tlsToolArgs.Info.ServerName = tlsInfoCMD.Flag("servername", "the server name sent to tls server").Short('s').Default("").String()
	//tls save
	tlsSaveCMD := tlsToolCMD.Command("save", "save tls target host:port certificate to file")
	tlsToolArgs.Save.Addr = tlsSaveCMD.Flag("addr", "address of tls target, ip:port").Short('a').Default("").String()
	tlsToolArgs.Save.Proxy = tlsSaveCMD.Flag("proxy", "proxy URL connect to address of tls target, example: http://127.0.0.1:8080").Short('p').Default("").String()
	tlsToolArgs.Save.ServerName = tlsSaveCMD.Flag("servername", "the server name sent to tls server").Short('s').Default("").String()
	tlsToolArgs.Save.FolderName = tlsSaveCMD.Flag("name", "save certificate folder name").Short('n').Default("").String()

	// command url
	urlCmd := gmctApp.Command("url", "url toolkit")
	urlArgs.EncodeStr = urlCmd.Flag("encode", "escape the string").Short('e').Default("").String()
	urlArgs.DecodeStr = urlCmd.Flag("decode", "unescape the string").Short('d').Default("").String()

	// command file_repeat
	fileRepeatCmd := gmctApp.Command("file_repeat", "scan repeated files")
	fileRepeatArgs.Delete = fileRepeatCmd.Flag("delete", "delete the repeat file").Bool()
	fileRepeatArgs.Debug = fileRepeatCmd.Flag("debug", "output debug logging").Bool()
	fileRepeatArgs.Dir = fileRepeatCmd.Flag("dir", "path to scan").Short('d').String()
	fileRepeatArgs.Logfile = fileRepeatCmd.Flag("log", "log file to store scan result, if empty path, the default will be used").String()
	fileRepeatArgs.Workers = fileRepeatCmd.Flag("workers", "delete the repeat file").Int()

	//check command line args
	if len(os.Args) == 0 {
		os.Args = []string{""}
		gmctApp.Usage(os.Args)
		return
	}

	subToolName, err := gmctApp.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	subToolSubName := ""
	a := strings.Split(subToolName, " ")
	if len(a) > 1 {
		subToolName = a[0]
		subToolSubName = a[1]
		subToolSubName = a[1]
	}

	// #3
	var gmcToolObj tool.GMCTool
	var args interface{}
	switch subToolName {
	case "tpl":
		templateArgs.SubName = &subToolSubName
		args = templateArgs
		gmcToolObj = template.NewTemplate()
	case "static":
		staticArgs.SubName = &subToolSubName
		args = staticArgs
		gmcToolObj = static.NewStatic()
	case "run":
		runArgs.SubName = &subToolSubName
		runArgs.Args = runCmdArgs
		args = runArgs
		gmcToolObj = run.NewRun()
	case "new":
		newArgs.SubName = &subToolSubName
		args = newArgs
		gmcToolObj = newx.NewX()
	case "i18n":
		i18nArgs.SubName = &subToolSubName
		args = i18nArgs
		gmcToolObj = i18n.NewI18n()
	case "controller":
		controllerArgs.SubName = &subToolSubName
		args = controllerArgs
		gmcToolObj = controller.NewController()
	case "model":
		modelArgs.SubName = &subToolSubName
		args = modelArgs
		gmcToolObj = model.NewModel()
	case "gtag":
		gtagArgs.SubName = &subToolSubName
		args = gtagArgs
		gmcToolObj = gtag.NewGTag()
	case "cover":
		coverArgs.SubName = &subToolSubName
		args = coverArgs
		gmcToolObj = cover.NewCover()
	case "view":
		viewArgs.SubName = &subToolSubName
		args = viewArgs
		gmcToolObj = view.NewView()
	case "docker":
		dockerArgs.SubName = &subToolSubName
		args = dockerArgs
		gmcToolObj = docker.NewDocker()
	case "tool":
		toolArgs.SubName = &subToolSubName
		args = toolArgs
		gmcToolObj = toolx.NewTool()
	case "ssh":
		args = sshArgs
		gmcToolObj = ssht.NewSsh()
	case "update":
		args = updateArgs
		gmcToolObj = update.NewUpdate()
	case "go":
		goToolArgs.SubName = &subToolSubName
		args = goToolArgs
		gmcToolObj = gotool.NewGoTool()
	case "install", "install-force", "uninstall":
		installToolArgs.Action = subToolName
		args = installToolArgs
		gmcToolObj = installtool.NewInstallTool()
	case "tls":
		tlsToolArgs.SubName = &subToolSubName
		args = tlsToolArgs
		gmcToolObj = tlstool.NewTLS()
	case "url":
		urlArgs.SubName = &subToolSubName
		args = urlArgs
		gmcToolObj = gurl.New()
	case "file_repeat":
		fileRepeatArgs.SubName = &subToolSubName
		args = fileRepeatArgs
		gmcToolObj = file_repeat.NewFileRepeat()
	default:
		fmt.Printf("sub command '%s' not found\n", subToolName)
		return
	}
	err = gmcToolObj.Start(args)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}
}
