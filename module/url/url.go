package gurl

import (
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	"net/url"
	"strings"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use: "url",
			PersistentPreRunE: func(c *cobra.Command, a []string) error {
				if len(a) == 0 {
					return fmt.Errorf("args required")
				}
				return nil
			},
		}
		cmd.AddCommand(&cobra.Command{
			Use:  "encode",
			Long: "escape the string",
			Run: func(c *cobra.Command, a []string) {
				fmt.Println(url.QueryEscape(strings.Join(a, " ")))
			},
		})
		cmd.AddCommand(&cobra.Command{
			Use:  "decode",
			Long: "unescape the string",
			RunE: func(c *cobra.Command, a []string) error {
				result, e := url.QueryUnescape(strings.Join(a, " "))
				if e != nil {
					return e
				}
				fmt.Println(result)
				return nil
			},
		})
		root.AddCommand(cmd)
	})
}
