package controller

import (
	"fmt"
	"github.com/snail007/gmc"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util"
	"io/ioutil"
	"os/exec"
	"strings"
)

var (
	defaultFile = "controllers.toml"
)

type ControllerArgs struct {
	ForceCreate    *bool
	ControllerName *string
	TableName      *string
	SubName        *string
	fromFile       bool
}

func NewControllerArgs() ControllerArgs {
	return ControllerArgs{
		ControllerName: new(string),
		TableName:      new(string),
		SubName:        new(string),
		ForceCreate:    new(bool),
	}
}

type Controller struct {
	tool.GMCTool
	args ControllerArgs
}

func NewController() *Controller {
	return &Controller{
	}
}

func (s *Controller) init(args0 interface{}) (err error) {
	s.args = args0.(ControllerArgs)
	if *s.args.ControllerName == "" {
		s.args.fromFile = util.Exists(defaultFile)
		if s.args.fromFile {
			return
		}
		return fmt.Errorf("controller name required")
	} else if *s.args.TableName == "" {
		*s.args.TableName = strings.ToLower(*s.args.ControllerName)
	}
	return
}

func (s *Controller) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	if s.args.fromFile {
		err = s.fromFile()
	} else {
		err = s.create()
	}
	return
}

func (s *Controller) Stop() {
	return
}

func (s *Controller) exec(name string) string {
	nameArr := strings.Split(name, ":")
	c := ""
	t := ""
	if len(nameArr) >= 1 {
		c = nameArr[0]
	}
	if len(nameArr) >= 2 {
		t = nameArr[1]
	}
	if t == "" {
		t = c
	}
	cmd := exec.Command("gmct", "controller", "-n", name, "-t", t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}
	txt := string(out)
	if txt == "" {
		txt = "success"
	}
	return strings.TrimSuffix(txt, "\n")
}

func (s *Controller) fromFile() (err error) {
	cfg, err := gmc.New.ConfigFile(defaultFile)
	if err != nil {
		return
	}
	c := cfg.GetStringSlice("controllers")

	if len(c) == 0 {
		err = fmt.Errorf("done, create controller files fail, controller names is empty")
		return
	}

	out := []string{}
	out = append(out, "start create controller files: "+strings.Join(c, ", "))
	for _, v := range c {
		out = append(out, fmt.Sprintf("    %s >>> %s", v, s.exec(v)))
	}

	fmt.Println(strings.Join(out, "\n"))
	return nil
}

func (s *Controller) create() (err error) {
	packageName := util.GetPackageName("./")
	table := *s.args.TableName
	tpl := fmt.Sprintf(tpl, packageName)
	tpl = strings.Replace(tpl, "{{HOLDER}}", *s.args.ControllerName, -1)
	tpl = strings.Replace(tpl, "{{TABLE}}", table, -1)
	filename := strings.ToLower(*s.args.ControllerName) + ".go"
	if util.Exists(filename) && !*s.args.ForceCreate {
		return fmt.Errorf("file %s aleadly exists, please using option `-f` to overwrite it", filename)
	}
	ioutil.WriteFile(filename, []byte(tpl), 0755)
	return
}
