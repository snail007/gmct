package update

import (
	"bufio"
	"encoding/json"
	"fmt"
	gcore "github.com/snail007/gmc/core"
	glog "github.com/snail007/gmc/module/log"
	gcompress "github.com/snail007/gmc/util/compress"
	"github.com/snail007/gmc/util/gos"
	ghttp "github.com/snail007/gmc/util/http"
	"github.com/snail007/gmct/tool"
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
	return &Update{
	}
}

func (s *Update) init(args0 interface{}) (err error) {
	s.args = args0.(UpdateArgs)
	return
}

func (s *Update) Start(args interface{}) (err error) {
	glog.SetFlag(gcore.LFLAG_NORMAL)
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

	if newVersion == currentVersion {
		if !*s.args.Force {
			return fmt.Errorf("already installed newest version, you can using -f to force update")
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
	tmpFile := gos.TempFile("gmct-update-", ".tar.gz")
	defer func() {
		os.Remove(tmpFile)
	}()
	ext := ""
	gzURL := ""
	switch runtime.GOOS {
	case "windows":
		gzURL = fmt.Sprintf(downloadURL, newVersion, "windows")
		ext = ".exe"
	case "darwin":
		gzURL = fmt.Sprintf(downloadURL, newVersion, "mac")
	default:
		gzURL = fmt.Sprintf(downloadURL, newVersion, "linux")
	}
	glog.Info("downloading ...")
	err = ghttp.DownloadToFile(gzURL, time.Minute*5, nil, tmpFile)
	if err != nil {
		return
	}
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
	glog.Info("decompress ...")
	// uncompress
	gzfile, err := os.Open(tmpFile)
	if err != nil {
		return err
	}
	defer gzfile.Close()
	_, err = gcompress.Unpack(gzfile, os.TempDir())
	if err != nil {
		return err
	}
	gzfile.Close()
	newFile := filepath.Join(os.TempDir(), "gmct"+ext)
	os.Chmod(newFile, 0755)
	err = os.Rename(newFile, binPath)
	glog.Info("done!")
	return
}

func (s *Update) Stop() {
	return
}

type APIResponseData struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
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
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string      `json:"url"`
		ID       int         `json:"id"`
		NodeID   string      `json:"node_id"`
		Name     string      `json:"name"`
		Label    interface{} `json:"label"`
		Uploader struct {
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
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}
