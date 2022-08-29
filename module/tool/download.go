package tool

import (
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/gobwas/glob"
	"github.com/schollz/progressbar/v3"
	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmc/util/gpool"
	ghttp "github.com/snail007/gmc/util/http"
	gset "github.com/snail007/gmc/util/set"
	"github.com/spf13/viper"
	"io"
	"net"
	"net/http"
	URL "net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	defaultConfigName = ".gmct_download"
)

type DownloadArgs struct {
	Port         *[]string
	Net          *[]string
	Name         *string
	File         *string
	MaxDeepLevel *int
	Host         *[]string
	Auth         *string
	ServerID     *string
	DownloadAll  *bool
	cfg          *viper.Viper
}

type serverFileItem struct {
	url    *URL.URL
	server *serverItem
}

type serverItem struct {
	id      string
	version string
	url     *URL.URL
	auth    []string
}

func (s *Tool) initDownload() {
	if *s.args.Download.File == "" {
		glog.Error("download file name required, use option: -f xxx")
		return
	}
	f := filepath.Join(gfile.HomeDir(), defaultConfigName)
	if gfile.Exists(f) {
		cfg := viper.New()
		cfg.SetConfigType("toml")
		cfg.SetConfigFile(f)
		e := cfg.ReadInConfig()
		if e != nil {
			glog.Warnf("read config error: %s, file:%s", e, f)
		}
		s.args.Download.cfg = cfg
	} else {
		gfile.WriteString(f, `# default networks to scan
# example: net=["192.168.1.0","192.168.2.0"]
net=[]

# default hosts to connect
# example: host=["192.168.1.2","192.168.1.3"]
# you can specify auth info in url, example: host=["foo_user:foo_pass@192.168.1.2"]
# you can specify port in url, example: host=["192.168.1.2:9966"]
# if host is specified, net will ignored.
host=[]

# default auth info be used when connect to server.
# example: auth="foo_user:foo_password", username and password seperated by a colon.
# if --auth option is specified, the auth here will ignored.
auth=""
`, false)
	}
}
func (s *Tool) download() {
	basename := strings.TrimPrefix(*s.args.Download.Name, "/")
	if basename != "" {
		if !s.confirmOverwrite(basename) {
			return
		}
		if strings.Contains(basename, "/") {
			dir, _ := filepath.Abs(filepath.Dir(basename))
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				glog.Errorf("create directory [%s] fail, error: %s", dir, err)
			}
		}
	}

	foundFiles := s.getFoundFiles(s.getServerURL())
	if len(foundFiles) == 1 {
		foundFile := foundFiles[0]
		if basename == "" {
			basename = filepath.Base(foundFile.url.String())
			if !s.confirmOverwrite(basename) {
				return
			}
		}
		s.downloadFile(1, 1, basename, foundFile, "")
	} else {
		total := len(foundFiles)
		for i, foundFile := range foundFiles {
			basename = filepath.Base(foundFile.url.Path)
			dir, _ := filepath.Abs(filepath.Join("download_files", strings.TrimPrefix(filepath.Dir(foundFile.url.Path), "/")))
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				glog.Errorf("create directory [%s] fail, error: %s", dir, err)
			}
			s.downloadFile(i+1, total, basename, foundFile, dir)
		}
	}
}

// 1
func (s *Tool) getServerURL() *serverItem {
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
					item := gmctWebServerList[index]
					desc := ""
					if item.id != "" {
						desc = fmt.Sprintf("[id: %s, version: %s]", item.id, item.version)
					}
					return item.url.Hostname() + " " + desc
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

// 2
func (s *Tool) getWebServerList() []*serverItem {
	scanURLArr := s.getScanURLs()
	length := len(scanURLArr)
	pool := gpool.New(length)
	g := sync.WaitGroup{}
	g.Add(length)
	gmctWebServerList := []*serverItem{}
	for _, u1 := range scanURLArr {
		scanURL := u1
		pool.Submit(func() {
			defer g.Done()
			url, _ := URL.Parse(scanURL)
			user, pass, client := s.getDownloadHTTPClient(nil, url)
			_, _, resp, e := client.Get(scanURL, time.Second*3, nil)
			if e != nil {
				return
			}
			if strings.EqualFold(resp.Header.Get(headerPoweredByKey), headerPoweredByValue) {
				serverID := resp.Header.Get(headerServerIDKey)
				if *s.args.Download.ServerID != "" && serverID != *s.args.Download.ServerID {
					return
				}
				item := &serverItem{
					url:     url,
					version: resp.Header.Get(headerVersionKey),
					id:      serverID,
				}
				if user != "" {
					item.auth = []string{user, pass}
					url.User = nil
				}
				gmctWebServerList = append(gmctWebServerList, item)
			}
		})
	}
	g.Wait()
	return gmctWebServerList
}

// 3
func (s *Tool) getScanURLs() []string {
	serverURLs := []string{}
	var getHostURL = func(host string) string {
		u, _ := URL.Parse(fmt.Sprintf("http://%s/", host))
		if _, _, err := net.SplitHostPort(u.Host); err != nil {
			u.Host = net.JoinHostPort(u.Host, DefaultPort)
		}
		return u.String()
	}

	//  from command line
	for _, host := range *s.args.Download.Host {
		serverURLs = append(serverURLs, getHostURL(host))
	}
	if len(serverURLs) > 0 {
		return serverURLs
	}

	//  from config file
	if len(serverURLs) == 0 {
		hostArr := s.getDownloadConfig("host")
		if hostArr != nil {
			for _, h := range hostArr.([]interface{}) {
				serverURLs = append(serverURLs, getHostURL(h.(string)))
			}
		}
	}
	if len(serverURLs) > 0 {
		return serverURLs
	}

	//  from auto scan
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

// 4
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
		net := s.getDownloadConfig("net")
		if net != nil {
			for _, v := range net.([]interface{}) {
				subnetList.Add(checkNet(v.(string)))
			}
		}
	}
	if subnetList.Len() == 0 {
		for _, v := range getLocalIP() {
			subnetList.Add(checkNet(v))
		}
	}
	return subnetList.ToStringSlice()
}

// 1
func (s *Tool) getFoundFiles(serverItem *serverItem) (foundFiles []*serverFileItem) {
	var files []*serverFileItem
	s.listFiles(serverItem, "", &files)
	if len(files) == 0 {
		glog.Error("no files found on the specify server")
	}
	// filter files
	toMatch := *s.args.Download.File
	g, err := glob.Compile(toMatch)
	if err != nil {
		glog.Errorf("parse file name error: %s", err)
	}
	for _, f := range files {
		basename := filepath.Base(f.url.Path)
		if strings.Contains(toMatch, "/") {
			//contains path
			basename = f.url.Path
		}
		if strings.Contains(basename, toMatch) || g.Match(basename) {
			foundFiles = append(foundFiles, f)
		}
	}
	if len(foundFiles) == 0 {
		glog.Error("no matched file found on the specify server")
	}

	if len(foundFiles) > 1 {
		if *s.args.Download.DownloadAll {
			//download all
			return foundFiles
		}
		foundFiles = append([]*serverFileItem{nil}, foundFiles...)
		selectIdx := []string{}
		for idx := range foundFiles {
			selectIdx = append(selectIdx, fmt.Sprintf("%d", idx))
		}
		var qs = []*survey.Question{{
			Name: "index",
			Prompt: &survey.Select{
				Message: "multiple matched files(" + fmt.Sprintf("%d", len(foundFiles)-1) + ") found, which file do you want to download?",
				Options: selectIdx,
				Description: func(value string, index int) string {
					if index == 0 {
						return "download all files"
					}
					return foundFiles[index].url.Path
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
		if answers.Index == 0 {
			//download all
			return foundFiles[1:]
		} else {
			//select a file
			return []*serverFileItem{foundFiles[answers.Index]}
		}
	} else {
		return foundFiles
	}
}

// 2
func (s *Tool) listFiles(server *serverItem, path string, files *[]*serverFileItem) {
	_, _, client := s.getDownloadHTTPClient(server.auth, nil)
	body, _, resp, e := client.Get(server.url.String()+path, time.Second*3, nil)
	if e != nil {
		glog.Warnf("fetch [%s] error: %s", server.url, e)
		return
	}
	if resp.Header.Get(headerPoweredByKey) != headerPoweredByValue {
		glog.Warnf("[%s] is not a gmct http server", server.url.Host)
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
			s.listFiles(server, path+href, files)
		} else {
			u, _ := URL.Parse(server.url.String() + path + href)
			*files = append(*files, &serverFileItem{
				url:    u,
				server: server,
			})
		}
	})
}

func (s *Tool) getBasicAuth(url *URL.URL) (username, password string, isSet bool) {
	authInfo := ""
	// form url
	if url != nil && url.User.Username() != "" {
		p, _ := url.User.Password()
		return url.User.Username(), p, true
	}
	//form command line
	if authInfo == "" {
		if *s.args.Download.Auth != "" {
			authInfo = *s.args.Download.Auth
		}
	}
	// from config
	if authInfo == "" {
		a := s.getDownloadConfig("auth")
		if a != nil {
			authInfo = a.(string)
		}
	}
	if authInfo == "" {
		return "", "", false
	}
	a := strings.Split(authInfo, ":")
	return a[0], a[1], true
}
func (s *Tool) getDownloadConfig(key string) interface{} {
	if s.args.Download.cfg == nil {
		return nil
	}
	return s.args.Download.cfg.Get(key)
}
func (s *Tool) getDownloadHTTPClient(auth []string, scanURL *URL.URL) (user, pass string, client *ghttp.HTTPClient) {
	client = ghttp.NewHTTPClient()
	client.SetProxyFromEnv(true)
	if len(auth) == 2 {
		user, pass = auth[0], auth[1]
	}
	if user == "" {
		user, pass, _ = s.getBasicAuth(scanURL)
	}
	if user != "" {
		client.SetBasicAuth(user, pass)
	}
	return user, pass, client
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
func (s *Tool) downloadFile(i, total int, basename string, foundFile *serverFileItem, dir string) {
	fmt.Println("downloading: " + foundFile.url.String())
	req, _ := http.NewRequest("GET", foundFile.url.String(), nil)
	if req != nil && foundFile.server.auth != nil {
		req.SetBasicAuth(foundFile.server.auth[0], foundFile.server.auth[1])
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		glog.Errorf("download error: %s", e)
	}
	defer resp.Body.Close()
	tmpfile := basename + ".tmp"
	f, _ := os.OpenFile(tmpfile, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription(fmt.Sprintf("(%d/%d)", i, total)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	bar.RenderBlank()
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
	e = os.Rename(tmpfile, filepath.Join(dir, basename))
	if e != nil {
		os.Remove(tmpfile)
		glog.Errorf("rename file error: %s", e)
	}
	glog.Info("download SUCCESS")
}
