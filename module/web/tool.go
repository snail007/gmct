package web

import (
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"net"
	"strings"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		s := NewTool()
		httpCMD := &cobra.Command{
			Use:     "web",
			Long:    "simple http server",
			Aliases: []string{"http", "www"},
			Run: func(c *cobra.Command, a []string) {
				httpServer(HTTPArgs{
					Addr:     util.Must(c.Flags().GetString("addr")).String(),
					RootDir:  util.Must(c.Flags().GetString("root")).String(),
					Auth:     util.Must(c.Flags().GetStringSlice("auth")).StringSlice(),
					Upload:   util.Must(c.Flags().GetString("upload")).String(),
					ServerID: util.Must(c.Flags().GetString("id")).String(),
				})
			},
		}
		httpCMD.Flags().StringP("addr", "l", ":9669", "simple http server listen on")
		httpCMD.Flags().StringP("root", "d", ".", "simple http server root directory")
		httpCMD.Flags().StringP("auth", "a", "", "simple http server basic auth username:password, such as : foouser:foopassowrd")
		httpCMD.Flags().StringP("upload", "u", "", "simple http server upload url path, default `random`")
		httpCMD.Flags().StringP("id", "i", "", "set the server id name, example: server01")

		downloadCMD := &cobra.Command{
			Use:     "download",
			Long:    "download file from gmct simple http server",
			Aliases: []string{"dl"},
			Run: func(c *cobra.Command, a []string) {
				host := util.Must(c.Flags().GetStringSlice("host")).StringSlice()
				file := util.Must(c.Flags().GetString("file")).String()
				id := util.Must(c.Flags().GetString("id")).String()
				if len(a) == 1 {
					ip := net.ParseIP(a[0])
					if ip != nil && ip.To4() != nil {
						// a[0] is ip
						host = []string{a[0]}
					} else if strings.Contains(a[0], ":") {
						// a[0] is serverID:file
						b := strings.SplitN(a[0], ":", 2)
						id = b[0]
						file = b[1]
					} else {
						// a[0] is a file
						file = a[0]
					}
				} else if len(a) == 2 {
					host = []string{a[0]}
					file = a[1]
				}
				s.download(s.initDownload(&DownloadArgs{
					Net:          util.Must(c.Flags().GetStringSlice("net")).StringSlice(),
					Port:         util.Must(c.Flags().GetStringSlice("port")).StringSlice(),
					File:         file,
					Name:         util.Must(c.Flags().GetString("name")).String(),
					MaxDeepLevel: util.Must(c.Flags().GetInt("deep")).Int(),
					Host:         host,
					Auth:         util.Must(c.Flags().GetString("auth")).String(),
					ServerID:     id,
					DownloadAll:  util.Must(c.Flags().GetBool("all")).Bool(),
					Timeout:      util.Must(c.Flags().GetInt("timeout")).Int(),
					DownloadDir:  util.Must(c.Flags().GetString("dir")).String(),
				}))
			},
		}
		downloadCMD.Flags().StringP("net", "n", "", "network to scan, format: 192.168.1.0")
		downloadCMD.Flags().StringP("port", "p", "9669", "gmct tool http port")
		downloadCMD.Flags().StringP("file", "f", "*", "filename to download")
		downloadCMD.Flags().StringP("name", "m", "", "rename download file to")
		downloadCMD.Flags().IntP("deep", "d", 1, "max directory deep level to list server files, value 0: no limit")
		downloadCMD.Flags().StringSliceP("host", "x", []string{}, "specify a domain or ip to download, example: 192.168.1.1 or 192.168.1.1:9090. \nyou can specify auth info, example: foo_user:foo_pass@192.168.1.2")
		downloadCMD.Flags().StringP("auth", "a", "", "basic auth info, example: username:password")
		downloadCMD.Flags().StringP("id", "i", "", "server id name to download files")
		downloadCMD.Flags().Bool("all", false, "download all files matched")
		downloadCMD.Flags().IntP("timeout", "t", 3, "timeout seconds to connect to server")
		downloadCMD.Flags().StringP("dir", "c", "download_files", "path to download all files")

		root.AddCommand(httpCMD)
		root.AddCommand(downloadCMD)
	})
}

type Tool struct {
}

func NewTool() *Tool {
	return &Tool{}
}
