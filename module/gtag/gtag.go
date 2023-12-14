package gtag

import (
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:  "gtag",
			Long: "print go mod require tag of git repository in current directory",
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewGTag()
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		root.AddCommand(cmd)
	})
}

type GTag struct {
}

func NewGTag() *GTag {
	return &GTag{}
}

func (s *GTag) init() (err error) {
	return
}

func (s *GTag) Start() (err error) {
	//git -c log.showsignature=false log --no-decorate -n1 "--format=format:%H %ct %D"
	cmd := exec.Command("git", "-c", "log.showsignature=false", "log", "--no-decorate", "-n1", "--format=format:%H %ct %D")
	b, e := cmd.CombinedOutput()
	if e != nil {
		fmt.Println(e, "\n", string(b))
		return
	}
	str := string(b)
	hash := ""

	f := strings.Fields(string(str))
	if len(f) < 2 {
		return fmt.Errorf("unexpected response from git log: %q", str)
	}
	hash = f[0]
	t, err := strconv.ParseInt(f[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid time from git log: %q", str)
	}
	date := time.Unix(t, 0).UTC()
	fmt.Printf("v0.0.0-%s-%s\n", date.In(time.UTC).Format("20060102150405"), hash[:12])
	return
}

func (s *GTag) Stop() {
	return
}
