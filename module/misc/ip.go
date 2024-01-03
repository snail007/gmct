package misc

import (
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		ipCMD := &cobra.Command{
			Use:  "ip",
			Long: "ip toolkit",
			Run: func(c *cobra.Command, a []string) {
				ip()
			},
		}
		root.AddCommand(ipCMD)
	})
}

func ip() {
	for _, v := range util.GetLocalIP() {
		fmt.Println(v)
	}
}
