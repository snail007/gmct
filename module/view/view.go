package view

import (
	"fmt"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util"
	"io/ioutil"
	"strings"
)

type ViewArgs struct {
	ForceCreate    *bool
	SubName        *string
	Table          *string
	ControllerPath *string
}

func NewViewArgs() ViewArgs {
	return ViewArgs{
		SubName:        new(string),
		ControllerPath: new(string),
		Table:          new(string),
		ForceCreate:    new(bool),
	}
}

type View struct {
	tool.GMCTool
	args ViewArgs
}

func NewView() *View {
	return &View{}
}

func (s *View) init(args0 interface{}) (err error) {
	s.args = args0.(ViewArgs)
	if *s.args.ControllerPath == "" {
		return fmt.Errorf("option '-n' required")
	}
	if *s.args.Table == "" {
		return fmt.Errorf("option '-t' required")
	}
	return
}

func (s *View) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	err = s.create()
	return
}

func (s *View) Stop() {
	return
}

func (s *View) create() (err error) {
	controllerPath := *s.args.ControllerPath
	table := *s.args.Table
	for _, v := range []struct {
		tpl  string
		name string
	}{
		{listTpl, "list"},
		{formTpl, "form"},
		{detailTpl, "detail"},
	} {
		tpl := strings.Replace(v.tpl, "{{CONTROLLER}}", controllerPath, -1)
		tpl = strings.Replace(v.tpl, "{{TABLE}}", table, -1)
		filename := v.name + ".html"
		if util.Exists(filename) && !*s.args.ForceCreate {
			return fmt.Errorf("file %s aleadly exists, please using option `-f` to overwrite it", filename)
		}
		ioutil.WriteFile(filename, []byte(tpl), 0755)
	}
	return
}
