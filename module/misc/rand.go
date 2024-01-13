package misc

import (
	"fmt"
	"github.com/pkg/errors"
	gcast "github.com/snail007/gmc/util/cast"
	grand "github.com/snail007/gmc/util/rand"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		randCMD := &cobra.Command{
			Use:  "rand",
			Long: "generate random string or integer with range",
		}
		randCMD.AddCommand(&cobra.Command{
			Use:     "string",
			Long:    "generate random string",
			Example: "gmct rand string 10",
			RunE: func(c *cobra.Command, a []string) error {
				if len(a) != 1 || gcast.ToInt(a[0]) == 0 {
					return errors.New("int argument is required")
				}
				fmt.Println(grand.String(gcast.ToInt(a[0])))
				return nil
			},
		})
		randCMD.AddCommand(&cobra.Command{
			Use:     "int",
			Long:    "generate integer string",
			Example: "gmct rand string 10",
			RunE: func(c *cobra.Command, a []string) error {
				start := int64(0)
				end := int64(0)
				switch len(a) {
				case 1:
					start = gcast.ToInt64(a[0])
				case 2:
					start = gcast.ToInt64(a[0])
					end = gcast.ToInt64(a[1])
					if start >= end {
						return errors.New("range error")
					}
				default:
					return errors.New("one or two arguments is required")
				}
				r := grand.New()
				if end > 0 {
					fmt.Println(start + r.Int63n(end-start+1))
				} else {
					fmt.Println(r.Int63n(start + 1))
				}
				return nil
			},
		})
		root.AddCommand(randCMD)
	})
}
