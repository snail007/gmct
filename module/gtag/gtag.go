package gtag

import (
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/spf13/cobra"
	"os/exec"
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
	cmd := exec.Command("git", "log", "-n", "1", "--date", "format:%Y-%m-%d %H:%M:%S")
	b, e := cmd.CombinedOutput()
	if e != nil {
		fmt.Println(e, "\n", string(b))
		return
	}
	str := string(b)
	hash := ""
	var date time.Time
	for _, v := range strings.Split(str, "\n") {
		line := strings.Fields(v)
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case "commit":
			hash = line[1]
		case "Date:":
			date, err = time.ParseInLocation(time.DateTime, strings.Join(line[1:], " "), time.UTC)
			if err != nil {
				return err
			}
		}
	}
	if hash == "" || date.IsZero() {
		fmt.Printf("can not find git log in: \n%s", str)
		return
	}
	fmt.Printf("v0.0.0-%s-%s\n", date.Format("20060102150405"), hash[:12])
	return
}

func (s *GTag) Stop() {
	return
}
