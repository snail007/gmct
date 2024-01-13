package misc

import (
	"errors"
	"fmt"
	gfile "github.com/snail007/gmc/util/file"
	ghash "github.com/snail007/gmc/util/hash"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		ipCMD := &cobra.Command{
			Use:     "md5",
			Long:    "show md5 hash of input string or file",
			Example: "gmct md5 a.txt （a.txt is file）\ngmct md5 abcd",
			RunE: func(c *cobra.Command, a []string) error {
				if len(a) != 1 {
					return errors.New("arg required")
				}
				str := a[0]
				var h string
				var err error
				if gfile.IsFile(str) {
					h, err = ghash.MD5File(str)
				} else {
					h = ghash.MD5(str)
				}
				if err != nil {
					return err
				}
				fmt.Println(h)
				return nil
			},
		}
		root.AddCommand(ipCMD)
	})
}
