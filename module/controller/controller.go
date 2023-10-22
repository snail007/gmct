package controller

import (
	"fmt"
	"github.com/snail007/gmc"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os/exec"
	"strings"
)

var (
	defaultFile = "controllers.toml"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "controller",
			Long:    "create a controller in current directory",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewController(Args{
					ForceCreate:    util.Must(c.Flags().GetBool("force")).Bool(),
					ControllerName: util.Must(c.Flags().GetString("name")).String(),
					TableName:      util.Must(c.Flags().GetString("table")).String(),
					fromFile:       false,
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().StringP("name", "n", "", "controller struct name")
		cmd.Flags().StringP("table", "t", "", "table name without prefix")
		cmd.Flags().BoolP("force", "f", false, "overwrite controller file, if it exists.")
		root.AddCommand(cmd)

	})
}

type Args struct {
	ForceCreate    bool
	ControllerName string
	TableName      string
	SubName        string
	fromFile       bool
}

type Controller struct {
	args Args
}

func NewController(args Args) *Controller {
	return &Controller{args: args}
}

func (s *Controller) init() (err error) {
	if s.args.ControllerName == "" {
		s.args.fromFile = util.Exists(defaultFile)
		if s.args.fromFile {
			return
		}
		return fmt.Errorf("controller name required")
	} else if s.args.TableName == "" {
		s.args.TableName = strings.ToLower(s.args.ControllerName)
	}
	return
}

func (s *Controller) Start() (err error) {
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
	table := s.args.TableName
	tpl := fmt.Sprintf(tpl, packageName)
	tpl = strings.Replace(tpl, "{{HOLDER}}", s.args.ControllerName, -1)
	tpl = strings.Replace(tpl, "{{TABLE}}", table, -1)
	filename := strings.ToLower(s.args.ControllerName) + ".go"
	if util.Exists(filename) && !s.args.ForceCreate {
		return fmt.Errorf("file %s aleadly exists, please using option `-f` to overwrite it", filename)
	}
	ioutil.WriteFile(filename, []byte(tpl), 0755)
	return
}
