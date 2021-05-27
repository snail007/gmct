package newx

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	gmccompress "github.com/snail007/gmc/util/compress"
	gfile "github.com/snail007/gmc/util/file"
	ghttp "github.com/snail007/gmc/util/http"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	only = []string{".go"}
)

type NewxArgs struct {
	SubName   *string
	Web       WebArgs
	API       APIArgs
	SimpleAPI SimpleAPIArgs
	Admin     AdminArgs
}

type WebArgs struct {
	Package *string
}

type APIArgs struct {
	Package *string
}

type SimpleAPIArgs struct {
	Package *string
}

type AdminArgs struct {
	Package *string
	Update  *bool
}

func NewArgs() NewxArgs {
	return NewxArgs{
		SubName: new(string),
		Web: WebArgs{
			Package: new(string),
		},
		API: APIArgs{
			Package: new(string),
		},
		SimpleAPI: SimpleAPIArgs{
			Package: new(string),
		},
		Admin: AdminArgs{
			Package: new(string),
			Update:  new(bool),
		},
	}
}

type Newx struct {
	tool.GMCTool
	GOPATH string
	dest   string
	cfg    NewxArgs
}

func NewX() *Newx {
	return &Newx{}
}

func (s *Newx) init(args0 interface{}) (err error) {

	s.cfg = args0.(NewxArgs)
	s.GOPATH = strings.TrimSpace(os.Getenv("GOPATH"))
	if s.GOPATH == "" {
		return fmt.Errorf("GOPATH environment variable not found")
	}
	// GOPATH may be contains multiple paths
	if filepath.Separator == '\\' {
		// windows
		s.GOPATH = strings.Split(s.GOPATH, ";")[0]
	} else {
		// linux
		s.GOPATH = strings.Split(s.GOPATH, ":")[0]
	}
	s.GOPATH = filepath.Join(s.GOPATH, "src")
	return
}

func (s *Newx) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	err = s.decompress()
	if err != nil {
		return
	}
	s.replace()
	fmt.Printf("initialized at: %s\n", s.dest)
	return
}

func (s *Newx) Stop() {

	return
}
func (s *Newx) replace() {
	var oldStr, newStr string
	switch *s.cfg.SubName {
	case "web":
		oldStr = "mygmcweb"
		newStr = *s.cfg.Web.Package
	case "api":
		oldStr = "mygmcapi"
		newStr = *s.cfg.API.Package
	case "api-simple":
		oldStr = "mygmcapi"
		newStr = *s.cfg.SimpleAPI.Package
	case "admin":
		oldStr = "mygmcadmin"
		newStr = *s.cfg.Admin.Package
	}
	filepath.Walk(s.dest, func(path string, info os.FileInfo, err error) error {
		ok := false
		for _, v := range only {
			if util.ExistsFile(path) &&
				(strings.HasSuffix(info.Name(), v) || v == info.Name()) {
				ok = true
				break
			}
		}
		if !ok {
			return nil
		}
		b, _ := ioutil.ReadFile(path)
		newFileStr := strings.Replace(string(b), oldStr, newStr, 1)
		ioutil.WriteFile(path, []byte(newFileStr), 0755)
		return nil
	})
	modTpl := fmt.Sprintf(`module %s

go 1.12
`, newStr)

	ioutil.WriteFile(filepath.Join(s.dest, "go.mod"), []byte(modTpl), 0755)
}
func (s *Newx) decompress() (err error) {
	data := ""
	var d []byte
	s.dest = s.GOPATH
	switch *s.cfg.SubName {
	case "web":
		data = webData
		s.dest = filepath.Join(s.dest, *s.cfg.Web.Package)
	case "api":
		data = apiData
		s.dest = filepath.Join(s.dest, *s.cfg.API.Package)
	case "api-simple":
		data = simpleapiData
		s.dest = filepath.Join(s.dest, *s.cfg.SimpleAPI.Package)
	case "admin":
		data = ""
		d, err = s.getAdmData()
		if err != nil {
			return
		}
		s.dest = filepath.Join(s.dest, *s.cfg.Admin.Package)
	}
	s.dest, _ = filepath.Abs(s.dest)

	if !util.Exists(s.dest) {
		err = os.MkdirAll(s.dest, 0755)
		if err != nil {
			return
		}
	} else if !util.IsEmptyDir(s.dest) {
		err = fmt.Errorf("%s directory is not empty.", s.dest)
		if err != nil {
			return
		}
	}
	if d == nil {
		d, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return
		}
	}
	var b bytes.Buffer
	b.Write(d)

	_, err = gmccompress.Unpack(&b, s.dest)
	if err == nil && *s.cfg.SubName == "admin" {
		root := filepath.Dir(s.dest)
		dstDirname := filepath.Base(s.dest)
		fs, err := filepath.Glob(s.dest + "/*")
		if err != nil {
			return err
		}
		tarDirname := ""
		for _, v := range fs {
			if v != "." && v != ".." && gfile.IsDir(v) {
				tarDirname = filepath.Base(v)
				break
			}
		}
		if tarDirname == "" {
			return fmt.Errorf("error format sorce code archive")
		}
		tmp := filepath.Join(root, dstDirname+".tmp")
		os.Rename(s.dest, tmp)
		os.Rename(filepath.Join(tmp, tarDirname), s.dest)
		os.RemoveAll(tmp)
	}
	return
}

func (s *Newx) getAdmData() (data []byte, err error) {
	home, _ := os.UserHomeDir()
	filePath := filepath.Join(home, ".gmct", "mygmcadmin.zip")
	if *s.cfg.Admin.Update || !gfile.Exists(filePath) {
		os.Remove(filePath)
		os.MkdirAll(filepath.Dir(filePath), 0755)
		fmt.Println("starting to download the admin source code ...")
		//download
		infoURL := "https://mirrors.host900.com/https://api.github.com/repos/snail007/mygmcadmin/releases/latest"
		downloadURL := "https://mirrors.host900.com/https://github.com/snail007/mygmcadmin/archive/refs/tags/%s.tar.gz"
		client := ghttp.NewHTTPClient()
		b, _, _, err := client.Get(infoURL, time.Second*30, nil)
		if err != nil {
			return nil, err
		}
		var resp ReleaseResponse
		err = json.Unmarshal(b, &resp)
		if err != nil {
			return nil, err
		}
		tag := resp.TagName
		downloadURL = fmt.Sprintf(downloadURL, tag)
		b, _, _, err = client.Get(downloadURL, time.Minute*15, nil)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(filePath, b, 0755)
		if err != nil {
			return nil, err
		}
	}
	return ioutil.ReadFile(filePath)
}

type ReleaseResponse struct {
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
