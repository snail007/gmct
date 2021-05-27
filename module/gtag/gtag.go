package gtag

import (
	"fmt"
	"github.com/snail007/gmct/tool"
	"os/exec"
	"strings"
	"time"
)

type GTagArgs struct {
	SubName *string
}

func NewGTagArgs() GTagArgs {
	return GTagArgs{
		SubName: new(string),
	}
}

type GTag struct {
	tool.GMCTool
	args GTagArgs
}

func NewGTag() *GTag {
	return &GTag{}
}

func (s *GTag) init(args0 interface{}) (err error) {
	s.args = args0.(GTagArgs)
	return
}

func (s *GTag) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
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
			date, err = time.Parse("2006-01-02 15:04:05", strings.Join(line[1:], " "))
			if err != nil {
				return err
			}
		}
	}
	if hash == "" || date.IsZero() {
		fmt.Printf("can not find git log in: \n%s", str)
		return
	}
	date = date.In(time.FixedZone("GMT", -8*3600))
	fmt.Printf("v0.0.0-%s-%s\n", date.Format("20060102150405"), string(hash[:12]))
	return
}

func (s *GTag) Stop() {
	return
}
