package model

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
	defaultFile = "tables.toml"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "model",
			Long:    "create a model in current directory",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewModel(Args{
					ForceCreate: util.Must(c.Flags().GetBool("force")).Bool(),
					Table:       util.Must(c.Flags().GetString("table")).String(),
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().StringP("table", "n", "", "table name without suffix")
		cmd.Flags().BoolP("force", "f", false, "overwrite model file, if it exists.")
		root.AddCommand(cmd)
	})
}

type Args struct {
	ForceCreate     bool
	Table           string
	tablename       string
	tableStructName string
	fromFile        bool
}

type Model struct {
	args Args
}

func NewModel(args Args) *Model {
	return &Model{args: args}
}

func (s *Model) init() (err error) {
	if s.args.Table == "" {
		s.args.fromFile = util.Exists(defaultFile)
		if s.args.fromFile {
			return
		}
		return fmt.Errorf("table name required")
	}
	arr := strings.Split(s.args.Table, "_")
	for k, v := range arr {
		arr[k] = strings.Title(strings.ToLower(v))
	}
	s.args.tablename = s.args.Table
	s.args.tableStructName = strings.Join(arr, "")
	return
}

func (s *Model) Start() (err error) {
	if s.args.fromFile {
		err = s.fromFile()
	} else {
		err = s.create()
	}
	return
}

func (s *Model) Stop() {
	return
}
func (s *Model) exec(table string) string {
	cmd := exec.Command("gmct", "model", "-n", table)
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
func (s *Model) fromFile() (err error) {
	cfg, err := gmc.New.ConfigFile(defaultFile)
	if err != nil {
		return
	}
	tables := cfg.GetStringSlice("tables")
	if tables == nil || len(tables) == 0 {
		return fmt.Errorf("bad config file %s", defaultFile)
	}
	out := []string{}
	out = append(out, "start create model files: "+strings.Join(tables, ", "))
	for _, v := range tables {
		out = append(out, fmt.Sprintf("    %s >>> %s", v, s.exec(v)))
	}
	fmt.Println(strings.Join(out, "\n"))
	return nil
}
func (s *Model) create() (err error) {
	packageName := util.GetPackageName("./")
	for k, v := range map[string]string{
		"{{PKG}}":               packageName,
		"{{TABLE_STRUCT_NAME}}": s.args.tableStructName,
		"{{TABLE_NAME}}":        s.args.tablename,
		"{{TABLE_PKEY}}":        s.args.tablename + "_id",
	} {
		tpl = strings.Replace(tpl, k, v, -1)

	}
	filename := strings.ToLower(s.args.tablename) + ".go"
	if util.Exists(filename) && !s.args.ForceCreate {
		return fmt.Errorf("file %s aleadly exists, please using option `-f` to overwrite it", filename)
	}
	ioutil.WriteFile(filename, []byte(tpl), 0755)
	return
}
