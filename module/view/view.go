package view

import (
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "view",
			Long:    "create views in current directory",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewView(Args{
					Table:          util.Must(c.Flags().GetString("dir")).String(),
					ControllerPath: util.Must(c.Flags().GetString("ext")).String(),
					ForceCreate:    util.Must(c.Flags().GetBool("clean")).Bool(),
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().StringP("controller", "n", "", "controller name in url path")
		cmd.Flags().StringP("table", "t", "", "table name without prefix")
		cmd.Flags().BoolP("force", "f", false, "overwrite model file, if it exists")
		root.AddCommand(cmd)
	})
}

type Args struct {
	ForceCreate    bool
	Table          string
	ControllerPath string
}

type View struct {
	args Args
}

func NewView(args Args) *View {
	return &View{args: args}
}

func (s *View) init() (err error) {
	if s.args.ControllerPath == "" {
		return fmt.Errorf("option '-n' required")
	}
	if s.args.Table == "" {
		return fmt.Errorf("option '-t' required")
	}
	return
}

func (s *View) Start() (err error) {
	err = s.create()
	return
}

func (s *View) Stop() {
	return
}

func (s *View) create() (err error) {
	controllerPath := s.args.ControllerPath
	table := s.args.Table
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
		if util.Exists(filename) && !s.args.ForceCreate {
			return fmt.Errorf("file %s aleadly exists, please using option `-f` to overwrite it", filename)
		}
		ioutil.WriteFile(filename, []byte(tpl), 0755)
	}
	return
}
