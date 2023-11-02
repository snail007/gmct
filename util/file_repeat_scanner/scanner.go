package file_repeat_scanner

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	gbytes "github.com/snail007/gmc/util/bytes"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmc/util/gpool"
	glist "github.com/snail007/gmc/util/list"
	gmap "github.com/snail007/gmc/util/map"
	gset "github.com/snail007/gmc/util/set"
	"github.com/snail007/gmct/util/checksum"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var (
	defaultExclude = []string{`.git`}
)

type ResultItem struct {
	Info     fs.FileInfo `json:"-"`
	Filepath string      `json:"filepath"`
	MD5      string      `json:"md5"`
}

type RepeatResultItem struct {
	MD5         string       `json:"md5"`
	FileItem    ResultItem   `json:"raw_file"`
	RepeatFiles []ResultItem `json:"repeat_files"`
}

type RepeatFileScanner struct {
	dir          string
	scannedCount int64
	allFiles     []ResultItem
	resultMap    *gmap.Map
	repeatCount  int64
	repeatSpace  int64
	logfile      string
	debug        bool
	buf          *gbytes.BytesBuilder
	workersCount int
	doDelete     bool
	deleteOld    bool
	outputJSON   bool
	include      []string
	exclude      []string
	startTime    time.Time
	endTime      time.Time
}

func (s *RepeatFileScanner) OutputJSON(outputJSON bool) *RepeatFileScanner {
	s.outputJSON = outputJSON
	return s
}

func (s *RepeatFileScanner) DeleteOld(deleteOld bool) *RepeatFileScanner {
	s.deleteOld = deleteOld
	return s
}

func (s *RepeatFileScanner) Delete(doDelete bool) *RepeatFileScanner {
	s.doDelete = doDelete
	return s
}

func (s *RepeatFileScanner) WorkersCount(workersCount int) *RepeatFileScanner {
	s.workersCount = workersCount
	return s
}

func (s *RepeatFileScanner) Debug(debug bool) *RepeatFileScanner {
	s.debug = debug
	return s
}

func (s *RepeatFileScanner) Logfile(logfile string) *RepeatFileScanner {
	s.logfile = logfile
	return s
}

func (s *RepeatFileScanner) checkSkip(path string, fileInfo os.FileInfo) bool {
	name := filepath.Base(path)
	for _, v := range append(defaultExclude, s.exclude...) {
		if ok, err := filepath.Match(v, name); ok {
			if err != nil {
				fmt.Printf("error exclude rule: %s\n", v)
			}
			return true
		}
	}
	if len(s.include) > 0 {
		for _, v := range s.include {
			if ok, err := filepath.Match(v, name); ok {
				if err != nil {
					fmt.Printf("error include rule: %s\n", v)
				}
				return false
			}
		}
		return true
	}
	return false
}

func NewRepeatFileScanner(dir string) *RepeatFileScanner {
	return &RepeatFileScanner{
		dir:          dir,
		resultMap:    gmap.New(),
		workersCount: 8,
		buf:          gbytes.NewBytesBuilder(),
	}
}
func (s *RepeatFileScanner) DeleteRepeat() {
	logfile := s.logfile
	if logfile == "" {
		logfile = "repeat_result_" + time.Now().Format("2006-01-02-15-04-05") + "_" + fmt.Sprintf("%d", time.Now().UnixMicro()) + ".txt"
	}
	var onDeleted = func(path string) {
		s.buf.WriteStrLn("%s deleted", path)
	}
	var onRepeated = func(path string) {
		s.buf.WriteStrLn("%s repeated", path)
	}
	var onViewItem = func(item RepeatResultItem) {
		s.buf.WriteStrLn("%s -> %d", item.MD5, len(item.RepeatFiles)+1)
		s.buf.WriteStrLn(item.FileItem.Filepath)
	}
	var onWriteResult = func(logPath string) {
		gfile.Write(logfile, s.buf.Bytes(), false)
	}
	result := s.RepeatResult()
	size, _ := gbytes.SizeStr(uint64(s.repeatSpace))
	if s.outputJSON {
		onDeleted = func(path string) {}
		onRepeated = func(path string) {}
		onViewItem = func(item RepeatResultItem) {}
		onWriteResult1 := onWriteResult
		onWriteResult = func(logPath string) {
			r := gmap.M{
				"total":     len(s.allFiles),
				"repeat":    s.repeatCount,
				"size":      size,
				"data":      result,
				"start_at":  s.startTime.Format(time.DateTime),
				"end_at":    s.endTime.Format(time.DateTime),
				"cost_time": s.endTime.Sub(s.startTime).Round(time.Second).String(),
			}
			s.buf = gbytes.NewBytesBuilder()
			j, _ := json.Marshal(r)
			s.buf.Write(j)
			onWriteResult1(logPath)
		}
	}
	s.buf.WriteStrLn("total: %d, repeat: %d, size: %s", len(s.allFiles), s.repeatCount, size)
	for _, item := range result {
		onViewItem(item)
		for _, f := range item.RepeatFiles {
			if s.doDelete {
				os.Remove(f.Filepath)
				gfile.WriteString(f.Filepath+".txt",
					fmt.Sprintf("重复文件：%s\n原始文件：%s",
						strings.TrimPrefix(f.Filepath, s.dir),
						strings.TrimPrefix(item.FileItem.Filepath, s.dir)), false)
				onDeleted(f.Filepath)
			} else {
				onRepeated(f.Filepath)
			}
		}
	}
	onWriteResult(logfile)
	fmt.Println(fmt.Sprintf("\nresult in file: [%s]", logfile))
}
func (s *RepeatFileScanner) Scan() *RepeatFileScanner {
	s.dir = gfile.Abs(s.dir)
	filepath.Walk(s.dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if s.checkSkip(path, info) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		s.allFiles = append(s.allFiles, ResultItem{
			Filepath: path,
			Info:     info,
		})
		return nil
	})
	bar := progressbar.NewOptions(len(s.allFiles),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	pool := gpool.New(s.workersCount)
	start := time.Now()
	s.startTime = start
	processedCount := new(int64)
	filelist := gset.New()
	for _, v := range s.allFiles {
		item := v
		path := item.Filepath
		pool.Submit(func() {
			defer func() {
				filelist.Delete(path)
				atomic.AddInt64(processedCount, 1)
				err := bar.Add64(1)
				if err != nil {
					fmt.Println("fail to add processed count")
				}
			}()
			filelist.Add(path)
			file, _ := os.Stat(path)
			size, _ := gbytes.SizeStr(uint64(file.Size()))
			if file.Size() == 0 {
				return
			}
			start := time.Now()
			hash, e := checksum.MD5sum(path)
			if e != nil {
				fmt.Println(path, e)
				return
			}
			item.MD5 = hash
			if s.debug {
				str := fmt.Sprintf("path:%s, hash:%s, size:%s, time:%s", path, hash, size, time.Since(start).Round(time.Second).String())
				fmt.Println(str)
			}
			l, _ := s.resultMap.LoadAndStoreFunc(hash, func(oldValue interface{}, loaded bool) (newValue interface{}) {
				if !loaded {
					return glist.New()
				}
				return oldValue
			})
			list := l.(*glist.List)
			list.Add(item)
		})
	}
	pool.WaitDone()
	s.endTime = time.Now()
	s.buf.WriteStrLn("scan start at: %s, cost time: %s",
		start.Format("2006-01-02 15:04:05"), time.Since(start).Round(time.Second).String())

	//sum
	for _, v := range s.RepeatResult() {
		s.repeatCount += int64(len(v.RepeatFiles))
		for _, v1 := range v.RepeatFiles {
			s.repeatSpace += v1.Info.Size()
		}
	}
	return s
}

func (s *RepeatFileScanner) RepeatResult() (result []RepeatResultItem) {
	s.resultMap.RangeFast(func(h, v interface{}) bool {
		hash := h.(string)
		item := v.(*glist.List)
		if item.Len() == 1 {
			return true
		}
		var files []ResultItem
		item.RangeFast(func(_ int, v interface{}) bool {
			files = append(files, v.(ResultItem))
			return true
		})
		var latestItem = files[0]
		for _, f := range files[1:] {
			if s.deleteOld {
				if f.Info.ModTime().After(latestItem.Info.ModTime()) {
					latestItem = f
				}
			} else {
				if f.Info.ModTime().Before(latestItem.Info.ModTime()) {
					latestItem = f
				}
			}
		}
		var repeatFiles []ResultItem
		for _, f := range files {
			if f.Filepath != latestItem.Filepath {
				repeatFiles = append(repeatFiles, f)
			}
		}
		result = append(result, RepeatResultItem{
			MD5:         hash,
			FileItem:    latestItem,
			RepeatFiles: repeatFiles,
		})
		return true
	})
	return
}
