package tool

import (
	"crypto/rand"
	"fmt"
	gctx "github.com/snail007/gmc/module/ctx"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmct/tool"
	"io"
	"log"
	"net"
	"net/http"
	URL "net/url"
	"os"
	"path/filepath"
	"strings"
)

type ToolArgs struct {
	ToolName *string
	SubName  *string
	HTTP     *HTTPArgs
}
type HTTPArgs struct {
	Addr    *string
	RootDir *string
	Auth    *[]string
	Upload  *string
}

func NewToolArgs() ToolArgs {
	return ToolArgs{
		ToolName: new(string),
		SubName:  new(string),
		HTTP:     new(HTTPArgs),
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
	}
	return
}

func (s *Tool) httpServer() {
	fmt.Println(">>> Simple HTTP Server")
	_, port, _ := net.SplitHostPort(*s.args.HTTP.Addr)
	for _, v := range getLocalIP() {
		fmt.Printf("http://%s:%s/\n", v, port)
	}
	var randID = func(len int) string {
		b := make([]byte, len/2)
		rand.Read(b)
		return fmt.Sprintf("%x", b)
	}
	rid := randID(16)
	if *s.args.HTTP.Upload != "" {
		rid = *s.args.HTTP.Upload
	}
	fmt.Println(">>> Upload ")
	for _, v := range getLocalIP() {
		fmt.Printf("http://%s:%s/%s\n", v, port, rid)
	}
	var sendAuth = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm=""`)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised.\n"))
	}
	var sendError = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	http.Handle("/"+rid, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// upload
		if r.Method == http.MethodGet {
			w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html><html lang="zh-CN"><head>
<meta charset="UTF-8"><title>upload file</title></head><body><form action="%s" name="upload" method="post" enctype="multipart/form-data">
<input type="file" name="file" style="display: none" multiple/><button id="upload">Upload</button></form><script>
document.forms["upload"].file.onchange=function(){document.forms["upload"].submit();};
document.getElementById("upload").onclick=function () {document.forms["upload"].file.click();return false;}</script></body></html>`,
				rid)))
			return
		}
		if r.Method != http.MethodPost {
			return
		}
		ctx := gctx.NewCtx()
		ctx.SetRequest(r)
		ctx.SetResponse(w)
		fs, err := ctx.MultipartForm(8 << 20)
		if err != nil {
			sendError(w, r, err)
			return
		}
		for _, f := range fs.File["file"] {
			path := filepath.Join(*s.args.HTTP.RootDir, f.Filename)
			suffix := ""
			if gfile.Exists(path) {
				suffix = "." + randID(6)
				path += suffix
			}
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			log.Printf("%s UPLOAD %s", ip, f.Filename+suffix)
			srcFile, err := f.Open()
			if err != nil {
				sendError(w, r, err)
				return
			}
			defer srcFile.Close()
			dstFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				sendError(w, r, err)
				return
			}
			defer dstFile.Close()
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				sendError(w, r, err)
				return
			}
		}
		ctx.Write(`<html><head><meta http-equiv="refresh" content="2;url=/"></head><body>success</body></html>`)
	}))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		root := *s.args.HTTP.RootDir
		reqPath := filepath.Clean(r.URL.Path)
		rootAbs := gfile.Abs(root)
		reqPathAbs := gfile.Abs(filepath.Join(root, reqPath))
		if !strings.Contains(reqPathAbs, rootAbs) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(*s.args.HTTP.Auth) > 0 {
			authOkay := false
			u, p, ok := r.BasicAuth()
			if !ok {
				sendAuth(w, r)
				return
			}
			for _, v := range *s.args.HTTP.Auth {
				userInfo := strings.Split(v, ":")
				if len(userInfo) != 2 {
					continue
				}
				user, _ := URL.QueryUnescape(userInfo[0])
				pass, _ := URL.QueryUnescape(userInfo[1])
				if user == "" || pass == "" {
					sendAuth(w, r)
					return
				}
				if u == user && p == pass {
					authOkay = true
					break
				}
			}
			if !authOkay {
				sendAuth(w, r)
				return
			}
		}
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		log.Printf("%s %s %s", ip, r.Method, r.URL.Path)
		http.ServeFile(w, r, reqPathAbs)
	}))
	panic(http.ListenAndServe(*s.args.HTTP.Addr, nil))
}

func (s *Tool) ip() {
	for _, v := range getLocalIP() {
		fmt.Println(v)
	}
}

func (s *Tool) Stop() {
	return
}

func getLocalIP() (ips []string) {
	ifs, _ := net.Interfaces()
	for _, v := range ifs {
		addrs, err := v.Addrs()
		if err != nil {
			continue
		}
		for _, vv := range addrs {
			ip, _, err := net.ParseCIDR(vv.String())
			if err != nil {
				continue
			}
			if ip.To4() == nil || ip.IsLoopback() {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return
}
