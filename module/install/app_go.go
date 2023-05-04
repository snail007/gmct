package installtool

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
	glog "github.com/snail007/gmc/module/log"
	gcast "github.com/snail007/gmc/util/cast"
	gexec "github.com/snail007/gmc/util/exec"
	ghttp "github.com/snail007/gmc/util/http"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	gzURLTemplate  = "https://golang.google.cn/dl/go%s.%s-%s.tar.gz"
	indexURL       = "https://golang.google.cn/dl/"
	targetRootPath = "/usr/local"
)

func init() {
	AddInstaller(GoAPPName, NewGoInstaller())
}

type GoInstaller struct {
}

func (s *GoInstaller) NeedRoot() bool {
	return true
}

func NewGoInstaller() *GoInstaller {
	return &GoInstaller{}
}

func (s *GoInstaller) Install(version string) error {
	tr, err := ghttp.NewTriableRequestByURL(nil, http.MethodGet, indexURL, 2, time.Second*10, nil, nil)
	if err != nil {
		return err
	}
	resp := tr.Execute()
	if resp.Err() != nil {
		return resp.Err()
	}
	htmlTxt := resp.Body()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlTxt))
	if err != nil {
		glog.Fatalf("parse api info fail, error: %s,\ncontent: %s", err, string(htmlTxt))
	}
	sel := doc.Find(".toggleButton")
	if len(sel.Nodes) == 0 {
		return fmt.Errorf("parse info fail, url: %s, body: %s", indexURL, string(htmlTxt))
	}
	lastVersionMap := map[string]string{}
	sel.Each(func(i int, el *goquery.Selection) {
		ver := strings.Trim(el.Text(), " \r\n\t")
		if ok, _ := regexp.MatchString(`^go\d+.\d+(.\d+)?$`, ver); !ok {
			return
		}
		ver = strings.TrimPrefix(ver, "go")
		verInfo := strings.Split(ver, ".")
		mainVer := strings.Join(verInfo[:2], ".")
		if _, ok := lastVersionMap[mainVer]; !ok {
			lastVersionMap[mainVer] = ver
		}
	})
	finalVersion, found := lastVersionMap[version]
	if !found {
		return fmt.Errorf("last version of %s not found", version)
	}
	URL := fmt.Sprintf(gzURLTemplate, finalVersion, runtime.GOOS, runtime.GOARCH)
	tr, err = ghttp.NewTriableRequestByURL(nil, http.MethodHead, URL, 2, time.Second*10, nil, nil)
	if err != nil {
		return err
	}
	tr.Close()
	resp = tr.Execute()
	if resp.Err() != nil {
		return resp.Err()
	}
	length := gcast.ToInt(resp.Header.Get("Content-Length"))
	if length == 0 {
		return fmt.Errorf("get file length error, length: 0, url: %s", URL)
	}
	filename := filepath.Base(URL)
	// create bars
	bar := progressbar.NewOptions(length,
		progressbar.OptionSetDescription(filename),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	tfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer func() {
		tfile.Close()
		os.Remove(filename)
		os.RemoveAll("go.tmp")
	}()
	_, err = ghttp.DownloadToWriter(URL, time.Hour, nil, nil, io.MultiWriter(tfile, bar))
	if err != nil {
		return err
	}
	tfile.Close()
	targetDir := filepath.Join(targetRootPath, "go"+finalVersion)
	cmd := gexec.NewCommand(`
set -ex
rm -rf go.tmp
mkdir go.tmp
cd  go.tmp
tar zfx ../` + filename + ` >/dev/null 2>&1
rm -rf ` + targetDir + `
mv go  ` + targetDir + `
`)
	_, err = cmd.Exec()
	if err != nil {
		return err
	}
	fmt.Println("\n" + targetDir + " installed, switch exec: chgo " + version)
	return nil
}

func (s *GoInstaller) InstallForce(version string) error {
	return nil
}

func (s *GoInstaller) Uninstall(version string) error {
	os.RemoveAll(filepath.Join(targetRootPath, "go"+version))
	return nil
}

func (s *GoInstaller) Exists(version string) bool {
	return false
}
