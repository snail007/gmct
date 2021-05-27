package update

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	glog "github.com/snail007/gmc/module/log"
	gcompress "github.com/snail007/gmc/util/compress"
	ghttp "github.com/snail007/gmc/util/http"
	grand "github.com/snail007/gmc/util/rand"
	"github.com/snail007/gmct/tool"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const updateAPIURL = "https://mirrors.host900.com/https://api.github.com/repos/snail007/gmct/releases/latest"
const downloadURL = "https://mirrors.host900.com/https://github.com/snail007/gmct/releases/download/v%s/gmct-%s-amd64.tar.gz"

type UpdateArgs struct {
	UpdateName *string
	SubName    *string
	Force      *bool
}

func NewUpdateArgs() UpdateArgs {
	return UpdateArgs{
		UpdateName: new(string),
		SubName:    new(string),
		Force:      new(bool),
	}
}

type Update struct {
	tool.GMCTool
	args UpdateArgs
}

func NewUpdate() *Update {
	return &Update{}
}

func (s *Update) init(args0 interface{}) (err error) {
	s.args = args0.(UpdateArgs)
	return
}

func (s *Update) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	currentVersion := tool.Version
	// check
	d, err := ghttp.Download(updateAPIURL, time.Second*30, nil)
	if err != nil {
		return
	}
	var versionInfo APIResponseData
	json.Unmarshal(d, &versionInfo)
	if versionInfo.TagName == "" {
		return fmt.Errorf("access update server fail")
	}
	newVersion := versionInfo.TagName[1:]
	newInfo := map[string]Assets{}
	for _, v := range versionInfo.Assets {
		arr := strings.Split(v.Name, "-")
		newInfo[arr[1]] = v
	}

	if newVersion == currentVersion {
		if !*s.args.Force {
			return fmt.Errorf("already installed newest version %s, you can using -f to force update", newVersion)
		}
	}

	// confirm
	if !*s.args.Force {
		fmt.Printf("Confirm update to v%s [y/N]:", newVersion)
		r := bufio.NewReader(os.Stdin)
		str, _ := r.ReadString('\n')
		if strings.ToLower(strings.Trim(str, " \n\t")) != "y" {
			return
		}
	}

	// start
	glog.Infof("ready update to v%s", newVersion)
	tmpFile := filepath.Join(os.TempDir(), grand.String(32), "gmct-update.tar.gz")
	tmpPath := filepath.Dir(tmpFile)
	err = os.Mkdir(tmpPath, 0755)
	if err != nil {
		return
	}
	tfile, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	defer func() {
		tfile.Close()
		os.Remove(tmpFile)
		os.RemoveAll(tmpPath)
	}()
	var newAsset Assets
	ext := ""
	gzURL := ""
	switch runtime.GOOS {
	case "windows":
		gzURL = fmt.Sprintf(downloadURL, newVersion, "windows")
		ext = ".exe"
		newAsset = newInfo["windows"]
	case "darwin":
		gzURL = fmt.Sprintf(downloadURL, newVersion, "mac")
		newAsset = newInfo["mac"]
	default:
		gzURL = fmt.Sprintf(downloadURL, newVersion, "linux")
		newAsset = newInfo["linux"]
	}
	glog.Info("downloading ...")
	// create bars
	bar := progressbar.NewOptions(newAsset.Size,
		progressbar.OptionSetDescription(newAsset.Name),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	err = ghttp.DownloadToWriter(gzURL, time.Minute*5, nil, io.MultiWriter(tfile, bar))
	if err != nil {
		return
	}
	tfile.Close()
	binPath, err := os.Executable()
	if err != nil {
		return err
	}
	old, err := os.Open(binPath)
	if err != nil {
		return err
	}
	defer old.Close()
	old.Close()
	fmt.Println("")
	glog.Info("decompress ...")
	// uncompress
	gzfile, err := os.Open(tmpFile)
	if err != nil {
		return fmt.Errorf("open update file fail, error: %s", err)
	}
	defer gzfile.Close()
	_, err = gcompress.Unpack(gzfile, tmpPath)
	if err != nil {
		return fmt.Errorf("decompress fail, %s", err)
	}
	gzfile.Close()
	newFile := filepath.Join(tmpPath, "gmct"+ext)

	fileNew, _ := os.Open(newFile)
	fileNewTmpPath := binPath + ".tmp"
	fileNewTmp, err := os.OpenFile(fileNewTmpPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("create temp target fail, %s", err)
	}
	defer func() {
		fileNew.Close()
		os.Remove(newFile)
		fileNewTmp.Close()
	}()
	// copy update file to bin path as temp file
	_, err = io.Copy(fileNewTmp, fileNew)
	if err != nil {
		return fmt.Errorf("wirte temp target fail, %s", err)
	}
	// replace old bin file with update file
	err = os.Rename(fileNewTmpPath, binPath)
	os.Chmod(binPath, 0755)
	glog.Info("done!")
	return
}

func (s *Update) Stop() {
	return
}

type APIResponseData struct {
	URL             string    `json:"url"`
	AssetsURL       string    `json:"assets_url"`
	UploadURL       string    `json:"upload_url"`
	HTMLURL         string    `json:"html_url"`
	ID              int       `json:"id"`
	Author          Author    `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []Assets  `json:"assets"`
	TarballURL      string    `json:"tarball_url"`
	ZipballURL      string    `json:"zipball_url"`
	Body            string    `json:"body"`
}
type Author struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type Uploader struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type Assets struct {
	URL                string      `json:"url"`
	ID                 int         `json:"id"`
	NodeID             string      `json:"node_id"`
	Name               string      `json:"name"`
	Label              interface{} `json:"label"`
	Uploader           Uploader    `json:"uploader"`
	ContentType        string      `json:"content_type"`
	State              string      `json:"state"`
	Size               int         `json:"size"`
	DownloadCount      int         `json:"download_count"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	BrowserDownloadURL string      `json:"browser_download_url"`
}
