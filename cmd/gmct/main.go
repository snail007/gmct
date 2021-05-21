package main

import (
	"fmt"
	"github.com/snail007/gmct/module/controller"
	"github.com/snail007/gmct/module/cover"
	"github.com/snail007/gmct/module/docker"
	"github.com/snail007/gmct/module/gtag"
	"github.com/snail007/gmct/module/i18n"
	"github.com/snail007/gmct/module/model"
	newx "github.com/snail007/gmct/module/new"
	"github.com/snail007/gmct/module/run"
	ssht "github.com/snail007/gmct/module/ssh"
	toolx "github.com/snail007/gmct/module/tool"
	"github.com/snail007/gmct/module/update"
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
	//all subtool defined here

	// #2
	// subtool template
	templateCmd := gmctApp.Command("tpl", "pack or clean templates go file")
	templateArgs.Dir = templateCmd.Flag("dir", "template's template directory path, gmct will convert all template files in the folder to one go file").Default(".").String()
	templateArgs.Extension = templateCmd.Flag("ext", "extension of template file").Default(".html").String()
	templateArgs.Clean = templateCmd.Flag("clean", "clean packed file, if exists").Default("false").Bool()

	// subtool static
	staticCmd := gmctApp.Command("static", "pack or clean static go file")
	staticArgs.Dir = staticCmd.Flag("dir", "template's static directory path, gmct will convert all static files in the folder to one go file").Default(".").String()
	staticArgs.NotExtension = staticCmd.Flag("ext", "extension of exclude static files ").Default("").String()
	staticArgs.Clean = staticCmd.Flag("clean", "clean packed file, if exists").Default("false").Bool()

	// subtool run
	gmctApp.Command("run", "run gmc project with auto build when project's file changed")

	// subtool new
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

	// subtool i18n
	i18nCmd := gmctApp.Command("i18n", "pack or clean i18n go file")
	i18nArgs.Dir = i18nCmd.Flag("dir", "i18n's template directory path, gmct will convert all i18n files in the folder to one go file").Default(".").String()
	i18nArgs.Clean = i18nCmd.Flag("clean", "clean packed file, if exists").Default("false").Bool()

	// subtool controller
	controllerCmd := gmctApp.Command("controller", "create a controller in current directory")
	controllerArgs.ControllerName = controllerCmd.Flag("name", "controller struct name").Short('n').Default("").String()
	controllerArgs.TableName = controllerCmd.Flag("table", "table name without prefix").Short('t').Default("").String()
	controllerArgs.ForceCreate = controllerCmd.Flag("force", "overwrite controller file, if it exists.").Short('f').Default("false").Bool()

	// subtool model
	modelCmd := gmctApp.Command("model", "create a model in current directory")
	modelArgs.Table = modelCmd.Flag("table", "table name without suffix").Short('n').Default("").String()
	modelArgs.ForceCreate = modelCmd.Flag("force", "overwrite model file, if it exists.").Short('f').Default("false").Bool()

	// subtoolgtag
	gmctApp.Command("gtag", "print go mod require tag of git repository in current directory")

	// subtool test coverage
	coverCmd := gmctApp.Command("cover", "print go mod require tag of git repository in current directory")
	coverArgs.Race = coverCmd.Flag("race", "enable race checking").Short('r').Default("false").Bool()
	coverArgs.Verbose = coverCmd.Flag("verbose", "verbose testing logging output").Short('v').Default("false").Bool()
	coverArgs.KeepResult = coverCmd.Flag("keep", "kept the coverage result: coverage.txt ").Short('k').Default("false").Bool()
	coverArgs.Silent = coverCmd.Flag("silent", "silent mode, not to open a browser").Short('s').Default("false").Bool()
	coverArgs.ForceCheck = coverCmd.Flag("force", "force check the package even it not contains any *_test.go file").Short('f').Default("false").Bool()
	coverArgs.Ordered = coverCmd.Flag("order", "disable parallel run").Short('o').Default("false").Bool()

	// subtool controller
	viewCmd := gmctApp.Command("view", "create a controller in current directory")
	viewArgs.ControllerPath = viewCmd.Flag("controller", "controller name in url path").Short('n').Default("").String()
	viewArgs.Table = viewCmd.Flag("table", "table name without prefix").Short('t').Default("").String()
	viewArgs.ForceCreate = viewCmd.Flag("force", "overwrite model file, if it exists.").Short('f').Default("false").Bool()

	// subtool docker
	dockerCmd := gmctApp.Command("docker", "create a model in current directory, all run arguments after --, example: gmct docker -- ./foo -u xxx")
	dockerArgs.Image = dockerCmd.Flag("img", "image used to run program").Default("debian:8").String()
	dockerArgs.DArg_v = dockerCmd.Flag("volume", "volume").Short('v').Strings()
	dockerArgs.DArg_p = dockerCmd.Flag("port", "port").Short('p').Strings()
	dockerArgs.DArg_e = dockerCmd.Flag("env", "environment variable").Short('e').Strings()
	dockerArgs.IsDebug = dockerCmd.Flag("debug", "debug output").Bool()
	dockerArgs.Golang = dockerCmd.Flag("golang", "sets some golang environment variables").Bool()

	// subtool tool
	toolCMD := gmctApp.Command("tool", "gmct tools collection")

	toolIPCMD := toolCMD.Command("ip", "ip toolkit")
	_ = toolIPCMD

	toolHTTPCMD := toolCMD.Command("http", "simple http server")
	toolArgs.HTTP.Addr = toolHTTPCMD.Flag("addr", "simple http server listen on").Short('l').Default(":9669").String()
	toolArgs.HTTP.RootDir = toolHTTPCMD.Flag("root", "simple http server root directory").Short('d').Default("./").String()
	toolArgs.HTTP.Auth = toolHTTPCMD.Flag("auth", "simple http server basic auth username:password, such as : foouser:foopassowrd ").Short('a').Strings()
	toolArgs.HTTP.Upload = toolHTTPCMD.Flag("upload", "simple http server upload url path, default `random`").Short('u').String()

	// sub tool ssh
	toolSsh := gmctApp.Command("ssh", "ssh tool, copy  file to or execute command on remote host")
	sshArgs.File = toolSsh.Flag("copy", "<local_file>:<remote_file>, local file to copy").Short('c').String()
	sshArgs.Command = toolSsh.Flag("cmd", "command to execute, or '@file' exec script file").Short('e').String()
	sshArgs.SSHURL = toolSsh.Flag("url", "ssh info url").Short('u').String()

	// sub tool update
	updateCMD := gmctApp.Command("update", "update gmct to the latest version")
	updateArgs.Force = updateCMD.Flag("force", "force update").Default("false").Short('f').Bool()

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
