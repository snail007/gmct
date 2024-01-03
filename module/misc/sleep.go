package misc

import (
	"fmt"
	"github.com/pkg/errors"
	gcast "github.com/snail007/gmc/util/cast"
	grand "github.com/snail007/gmc/util/rand"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		ipCMD := &cobra.Command{
			Use:     "sleep",
			Long:    "sleep seconds or sleep in range seconds",
			Example: "gmct sleep 10s\ngmct sleep 10s 100s",
			RunE: func(c *cobra.Command, a []string) error {
				var start time.Duration
				var end time.Duration
				switch len(a) {
				case 1:
					start = gcast.ToDuration(a[0]).Round(time.Second)
					fmt.Printf("sleep in %s\n", start)
					time.Sleep(start)
				case 2:
					start = gcast.ToDuration(a[0])
					end = gcast.ToDuration(a[1])
					if end <= start {
						return errors.New("range error")
					}
					dur := time.Duration(int64(start) + grand.New().Int63n(int64(end-start))).Round(time.Second)
					fmt.Printf("sleep in %s\n", dur)
					time.Sleep(dur)
				default:
					return errors.New("seconds required")
				}
				return nil
			},
		}
		root.AddCommand(ipCMD)
	})
}
