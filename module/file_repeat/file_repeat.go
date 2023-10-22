package file_repeat

import (
	gcond "github.com/snail007/gmc/util/cond"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/snail007/gmct/util/file_repeat_scanner"
	"github.com/spf13/cobra"
	"runtime"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "file_repeat",
			Long:    "scan repeated files",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewFileRepeat(Args{
					Dir:       util.Must(c.Flags().GetString("dir")).String(),
					Logfile:   util.Must(c.Flags().GetString("log")).String(),
					Delete:    util.Must(c.Flags().GetBool("delete")).Bool(),
					Debug:     util.Must(c.Flags().GetBool("debug")).Bool(),
					DeleteNew: util.Must(c.Flags().GetBool("new")).Bool(),
					Workers:   util.Must(c.Flags().GetInt("workers")).Int(),
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().StringP("dir", "d", "./", "path to scan")
		cmd.Flags().String("log", "", "log file to store scan result, if empty path, the default will be used")
		cmd.Flags().Bool("delete", false, "delete the repeat file")
		cmd.Flags().Bool("debug", false, "output debug logging")
		cmd.Flags().Bool("new", false, "delete the newest file, default delete oldest file")
		cmd.Flags().Int("workers", 0, "scan workers count")
		root.AddCommand(cmd)

	})
}

type Args struct {
	Dir       string
	Logfile   string
	Delete    bool
	Debug     bool
	DeleteNew bool
	Workers   int
}

type FileRepeat struct {
	args Args
}

func NewFileRepeat(args Args) *FileRepeat {
	return &FileRepeat{args: args}
}

func (s *FileRepeat) init() (err error) {
	return
}

func (s *FileRepeat) Start() (err error) {
	workers := s.args.Workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}
	dir := gcond.Cond(s.args.Dir != "", s.args.Dir, "./").(string)
	file_repeat_scanner.NewRepeatFileScanner(dir).
		Delete(s.args.Delete).
		DeleteNew(s.args.DeleteNew).
		WorkersCount(workers).
		Logfile(s.args.Logfile).
		Debug(s.args.Debug).
		Scan().
		DeleteRepeat()
	return
}

func (s *FileRepeat) Stop() {
	return
}
