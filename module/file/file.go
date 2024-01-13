package file

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
		fileCMD := &cobra.Command{
			Use: "file",
		}
		cmdRepeat := &cobra.Command{
			Use:     "repeat",
			Long:    "scan repeated files",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				workers := util.Must(c.Flags().GetInt("workers")).Int()
				if workers == 0 {
					workers = runtime.NumCPU()
				}
				dir := util.Must(c.Flags().GetString("dir")).String()
				dir = gcond.Cond(dir != "", dir, "./").String()
				file_repeat_scanner.NewRepeatFileScanner(dir).
					Delete(util.Must(c.Flags().GetBool("delete")).Bool()).
					DeleteOld(util.Must(c.Flags().GetBool("old")).Bool()).
					WorkersCount(workers).
					Logfile(util.Must(c.Flags().GetString("log")).String()).
					Debug(util.Must(c.Flags().GetBool("debug")).Bool()).
					OutputJSON(util.Must(c.Flags().GetBool("json")).Bool()).
					Scan().
					DeleteRepeat()
				return nil
			},
		}
		cmdRepeat.Flags().StringP("dir", "d", "./", "path to scan")
		cmdRepeat.Flags().String("log", "", "log file to store scan result, if empty path, the default will be used")
		cmdRepeat.Flags().Bool("delete", false, "delete the repeat file")
		cmdRepeat.Flags().Bool("debug", false, "output debug logging")
		cmdRepeat.Flags().Bool("json", false, "output json result")
		cmdRepeat.Flags().Bool("old", false, "delete the old files, default delete the newer files")
		cmdRepeat.Flags().Int("workers", runtime.NumCPU(), "scan workers count")
		fileCMD.AddCommand(cmdRepeat)
		root.AddCommand(fileCMD)

	})
}
