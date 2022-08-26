package tool

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	URL "net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
	gctx "github.com/snail007/gmc/module/ctx"
	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmc/util/gpool"
	ghttp "github.com/snail007/gmc/util/http"
	gset "github.com/snail007/gmc/util/set"
	"github.com/snail007/gmct/tool"
	"github.com/spf13/viper"
)

var (
	DefaultPort       = "9669"
	defaultConfigName = ".gmct_download"
)

type ToolArgs struct {
	ToolName *string
	SubName  *string
	HTTP     *HTTPArgs
	Download *DownloadArgs
}
type HTTPArgs struct {
	Addr    *string
	RootDir *string
	Auth    *[]string
	Upload  *string
}

type DownloadArgs struct {
	Port         *[]string
	Net          *[]string
	Name         *string
	File         *string
	MaxDeepLevel *int
	Host         *string
	Auth         *string
	client       *ghttp.HTTPClient
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
		s.download()
	}
	return
}
func (s *Tool) getBasicAuth() (user, pass string, ok bool) {
	if *s.args.Download.Auth != "" {
		a := strings.Split(*s.args.Download.Auth, ":")
		return a[0], a[1], true
	}
	return "", "", false
}
func (s *Tool) download() {
	if *s.args.Download.File == "" {
		glog.Error("download file name required, use option: -f xxx")
		return
	}
	s.args.Download.client = ghttp.NewHTTPClient()
	if u, p, ok := s.getBasicAuth(); ok {
		s.args.Download.client.SetBasicAuth(u, p)
	}
	basename := *s.args.Download.Name
	if basename != "" && !s.confirmOverwrite(basename) {
		return
	}

	foundFile := s.getFoundFile(s.getServerURL())
	if basename == "" {
		basename = filepath.Base(foundFile)
		if !s.confirmOverwrite(basename) {
			return
		}
	}

	s.downloadFile(basename, foundFile)
}

func (s *Tool) getSubnetArr() []string {
	var checkNet = func(ip string) string {
		n := net.ParseIP(ip)
		if n == nil || n.To4() == nil {
			glog.Errorf("parse network error, %s", ip)
		}
		return strings.Join(strings.Split(ip, ".")[:3], ".") + "."
	}
	subnetList := gset.New()
	for _, v := range *s.args.Download.Net {
		subnetList.Add(checkNet(v))
	}
	if subnetList.Len() == 0 {
		f := filepath.Join(gfile.HomeDir(), defaultConfigName)
		if gfile.Exists(f) {
			cfg := viper.New()
			cfg.SetConfigType("toml")
			cfg.SetConfigFile(f)
			e := cfg.ReadInConfig()
			if e != nil {
				glog.Warnf("read config error: %s, file:%s", e, f)
			} else {
				for _, v := range cfg.GetStringSlice("net") {
					subnetList.Add(checkNet(v))
				}
			}
		} else {
			gfile.WriteString(f, "# default networks to scan\n# example: net=[\"192.168.1.0\",\"192.168.2.0\"]\nnet=[]\n", false)
		}
	}
	if subnetList.Len() == 0 {
		for _, v := range getLocalIP() {
			subnetList.Add(checkNet(v))
		}
	}
	return subnetList.ToStringSlice()
}
func (s *Tool) getScanURLs() []string {
	scanURLArr := []string{}
	subnetArr := s.getSubnetArr()
	portList := gset.New()
	portList.MergeStringSlice(*s.args.Download.Port)
	portList.Add(DefaultPort)
	for _, subnet := range subnetArr {
		for i := 1; i <= 255; i++ {
			ip := subnet + fmt.Sprintf("%d", i)
			for _, port := range portList.ToStringSlice() {
				scanURLArr = append(scanURLArr, fmt.Sprintf("http://%s:%s/", ip, port))
			}
		}
	}
	return scanURLArr
}

func (s *Tool) getWebServerList() []string {
	scanURLArr := s.getScanURLs()
	length := len(scanURLArr)
	pool := gpool.New(length)
	g := sync.WaitGroup{}
	g.Add(length)
	gmctWebServerList := []string{}
	for _, u1 := range scanURLArr {
		u := u1
		pool.Submit(func() {
			defer g.Done()
			_, _, resp, e := s.args.Download.client.Get(u, time.Second*3, nil)
			if e != nil {
				return
			}
			if strings.EqualFold(resp.Header.Get("Powered-By"), "GMCT") {
				gmctWebServerList = append(gmctWebServerList, u)
			}
		})
	}
	g.Wait()
	return gmctWebServerList
}
func (s *Tool) getServerURL() string {
	host := *s.args.Download.Host
	if host != "" {
		if !strings.Contains(host, ":") {
			host = net.JoinHostPort(host, DefaultPort)
		}
		return fmt.Sprintf("http://%s/", host)
	}
	gmctWebServerList := s.getWebServerList()
	if len(gmctWebServerList) == 0 {
		glog.Error("none gmct http server found")
	}
	serverURL := gmctWebServerList[0]
	if len(gmctWebServerList) > 1 {
		selectIdx := []string{}
		for idx := range gmctWebServerList {
			selectIdx = append(selectIdx, fmt.Sprintf("%d", idx+1))
		}
		var qs = []*survey.Question{{
			Name: "index",
			Prompt: &survey.Select{
				Message: "which server do you want to select?",
				Options: selectIdx,
				Description: func(value string, index int) string {
					a, _ := URL.Parse(gmctWebServerList[index])
					return a.Hostname()
				},
			},
			Validate: survey.Required,
		},
		}
		answers := struct {
			Index int
		}{}
		e := survey.Ask(qs, &answers)
		if e != nil {
			glog.Error(e.Error())
		}
		serverURL = gmctWebServerList[answers.Index]
	}
	return serverURL
}
func (s *Tool) getFoundFile(serverURL string) string {
	var files []string
	s.listFiles(serverURL, "", &files)
	if len(files) == 0 {
		glog.Error("no files found on the specify server")
	}
	foundFiles := []string{}
	for _, f := range files {
		if strings.Contains(filepath.Base(f), *s.args.Download.File) {
			foundFiles = append(foundFiles, f)
		}
	}
	if len(foundFiles) == 0 {
		glog.Error("no matched file found on the specify server")
	}
	foundFile := foundFiles[0]
	if len(foundFiles) > 1 {
		selectIdx := []string{}
		for idx := range foundFiles {
			selectIdx = append(selectIdx, fmt.Sprintf("%d", idx+1))
		}
		var qs = []*survey.Question{{
			Name: "index",
			Prompt: &survey.Select{
				Message: "multiple matched files found, which file do you want to download?",
				Options: selectIdx,
				Description: func(value string, index int) string {
					a, _ := URL.Parse(foundFiles[index])
					return a.Path
				},
			},
			Validate: survey.Required,
		},
		}
		answers := struct {
			Index int
		}{}
		e := survey.Ask(qs, &answers)
		if e != nil {
			glog.Error(e.Error())
		}
		foundFile = foundFiles[answers.Index]
	}
	return foundFile
}
func (s *Tool) confirmOverwrite(basename string) bool {
	if gfile.Exists(basename) {
		var qs = []*survey.Question{{
			Name: "confirm",
			Prompt: &survey.Confirm{
				Message: "file [" + basename + "] already exists, overwrite it?",
				Default: false,
			},
			Validate: survey.Required,
		},
		}
		answers := struct {
			Confirm bool
		}{}
		e := survey.Ask(qs, &answers)
		if e != nil {
			glog.Error(e.Error())
			return false
		}
		return answers.Confirm
	}
	return true
}
func (s *Tool) downloadFile(basename, foundFile string) {
	fmt.Println("downloading: " + foundFile)
	req, _ := http.NewRequest("GET", foundFile, nil)
	if u, p, ok := s.getBasicAuth(); ok {
		req.SetBasicAuth(u, p)
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		glog.Errorf("download error: %s", e)
	}
	defer resp.Body.Close()
	tmpfile := basename + ".tmp"
	f, _ := os.OpenFile(tmpfile, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
	)
	_, e = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if e != nil {
		glog.Errorf("download error: %s", e)
	}
	if gfile.Exists(basename) {
		e = os.Remove(basename)
		if e != nil {
			glog.Errorf("remove old file error: %s", e)
		}
	}
	e = os.Rename(tmpfile, basename)
	if e != nil {
		os.Remove(tmpfile)
		glog.Errorf("rename file error: %s", e)
	}
	glog.Info("download SUCCESS")
}

func (s *Tool) listFiles(serverURL, path string, files *[]string) {
	body, _, _, e := s.args.Download.client.Get(serverURL+path, time.Second*3, nil)
	if e != nil {
		return
	}
	doc, e := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if e != nil {
		return
	}
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, _ := selection.Attr("href")
		if href == "" {
			return
		}
		if strings.HasSuffix(href, "/") {
			deep := strings.Count(path+href, "/")
			if *s.args.Download.MaxDeepLevel > 0 && deep > *s.args.Download.MaxDeepLevel {
				return
			}
			s.listFiles(serverURL, path+href, files)
		} else {
			*files = append(*files, serverURL+path+href)
		}
	})
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
		w.Header().Set("Powered-By", `GMCT`)
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
