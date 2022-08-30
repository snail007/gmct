package tool

import (
	"net"
	"os"
	"strings"

	"github.com/snail007/gmct/tool"
)

func init() {
	// gmct xxx => gmct tool xxx
	flags := map[string]bool{"web": true, "http": true, "www": true, "download": true, "dl": true, "ip": true}
	if len(os.Args) >= 2 && flags[os.Args[1]] {
		newArgs := []string{}
		flagFound := false
		// insert tool
		for i, v := range os.Args {
			if strings.HasPrefix(v, "-") {
				flagFound = true
			}
			if i == 1 {
				newArgs = append(newArgs, "tool")
			}
			newArgs = append(newArgs, v)
		}
		// download compact
		if newArgs[2] == "download" || newArgs[2] == "dl" && !flagFound {
			switch len(newArgs) {
			case 4:
				//gmct tool download <host|file>
				//gmct tool download <serverID:file>
				ip := net.ParseIP(newArgs[3])
				if ip != nil && ip.To4() != nil {
					// newArgs[3] is ip
					newArgs = []string{newArgs[0], "tool", "dl", "-h", newArgs[3]}
				} else if strings.Contains(newArgs[3], ":") {
					// newArgs[3] is serverID:file
					a := strings.SplitN(newArgs[3], ":", 2)
					newArgs = []string{newArgs[0], "tool", "dl", "-i", a[0], "-f", a[1]}
				} else {
					// newArgs[3] is a file
					newArgs = []string{newArgs[0], "tool", "dl", "-f", newArgs[3]}
				}
			case 5:
				//gmct tool download <host> <file>
				newArgs = []string{newArgs[0], "tool", "dl", "-h", newArgs[3], "-f", newArgs[4]}
			}
		}
		os.Args = newArgs
	}
}

type ToolArgs struct {
	ToolName *string
	SubName  *string
	HTTP     *HTTPArgs
	Download *DownloadArgs
}

func NewToolArgs() ToolArgs {
	return ToolArgs{
		ToolName: new(string),
		SubName:  new(string),
		HTTP:     new(HTTPArgs),
		Download: new(DownloadArgs),
	}
}

type Tool struct {
	tool.GMCTool
	args ToolArgs
}

func NewTool() *Tool {
	return &Tool{}
}

func (s *Tool) init(args0 interface{}) (err error) {
	s.args = args0.(ToolArgs)
	return
}

func (s *Tool) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	switch *s.args.SubName {
	case "ip":
		s.ip()
	case "http":
		s.httpServer()
	case "download":
		s.initDownload()
		s.download()
	}
	return
}

func (s *Tool) Stop() {
	return
}


