package gprofile

import (
	"fmt"
	gbytes "github.com/snail007/gmc/util/bytes"
	gcast "github.com/snail007/gmc/util/cast"
	gcond "github.com/snail007/gmc/util/cond"
	gexec "github.com/snail007/gmc/util/exec"
	gset "github.com/snail007/gmc/util/set"
	gvalue "github.com/snail007/gmc/util/value"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"regexp"
	"sort"
	"strings"
	"time"
)

var numberPrinter = message.NewPrinter(language.English)

type PprofPBFile struct {
	ProfileFilePath string
	BinFilename     string
	Metadata        *Metadata
	Stacks          map[string]*StackItem
	RawStacks       []*CallItem
	Option          *ParseOption
	ImportPkgList   []string
	GoPath          string
	GoRoot          string
}

func NewPprofPBFile(opt *ParseOption, file string) (pf *PprofPBFile, err error) {
	heapType := getHeapType(opt.ProfileType)
	heapType = gcond.Cond(heapType != "", "-"+heapType, "").String()
	c := "go tool pprof " + heapType + " -traces " + file
	output, err := gexec.NewCommand(c).Exec()
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`[-]+\+[-]+\n`)
	a := reg.Split(output, -1)
	info := NewMetadata()
	info.Parse(a[0])
	stacks := map[string]*StackItem{}
	var rawStacks []*CallItem
	pf = &PprofPBFile{}
	for _, v := range a[1:] {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		bytesN := uint64(0)
		b := strings.Split(v, "\n")
		switch info.Type {
		case "inuse_space", "alloc_space", "inuse_objects", "alloc_objects":
			bytesN, _ = gbytes.ParseSize(strings.Fields(b[0])[1])
			b = b[1:]
		}
		h := strings.Fields(b[0])
		if h[0][0] == '-' {
			continue
		}
		path := h[1]
		stack := b[1:]
		for k := range b[1:] {
			stack[k] = strings.TrimSpace(stack[k])
		}

		count := int64(0)
		switch info.Type {
		case "cpu", "delay":
			count = gcast.ToDuration(h[0]).Nanoseconds()
		case "inuse_space", "alloc_space":
			count = int64(gvalue.Must(gbytes.ParseSize(strings.ToUpper(h[0]))).Uint64())
			count += int64(bytesN)
		default:
			count = gcast.ToInt64(h[0])
		}
		callInfo := &CallItem{
			Count:    count,
			Percent:  GetPercent(info.TotalSamples, count),
			Path:     path,
			Stack:    stack,
			Metadata: info,
			pf:       pf,
		}
		rawStacks = append(rawStacks, callInfo)
		if v1, ok := stacks[h[1]]; ok {
			v1.Total += count
			v1.CallInfo = append(v1.CallInfo, callInfo)
		} else {
			item := &StackItem{
				Total:    count,
				Path:     h[1],
				CallInfo: []*CallItem{callInfo},
				Metadata: info,
			}
			stacks[h[1]] = item
		}
	}
	for i := range stacks {
		sort.Sort(ByCount{Items: stacks[i].CallInfo, Ascending: false})
	}
	//当采集goroutine没有seconds参数当时候，结果文件里面没有Duration字段，TotalSamples就会是0，这里兼容一下。
	if opt.ProfileType == ProfileTypeGoroutine && info.TotalSamples == 0 {
		total := int64(0)
		for _, v := range rawStacks {
			total += v.Count
		}
		info.TotalSamples = total
		info.Duration = 1
	}

	pf.BinFilename = info.File
	pf.Metadata = info
	pf.Stacks = stacks
	pf.Option = opt
	pf.RawStacks = rawStacks
	pf.ProfileFilePath = file

	pinfo, err := GetProfileInfo([]string{file})
	if err != nil {
		return nil, err
	}
	pf.GoRoot = pinfo.GoRoot
	pf.GoPath = pinfo.GoPath
	for _, pkg := range pinfo.ImportLibraryList {
		pf.ImportPkgList = append(pf.ImportPkgList, pkg.ModFullPath())
	}
	return
}

// ByTotal 定义排序的结构体
type ByTotal struct {
	Items     []*StackItem
	Ascending bool
}

func (a ByTotal) Len() int      { return len(a.Items) }
func (a ByTotal) Swap(i, j int) { a.Items[i], a.Items[j] = a.Items[j], a.Items[i] }
func (a ByTotal) Less(i, j int) bool {
	if a.Ascending {
		return a.Items[i].Total < a.Items[j].Total
	}
	return a.Items[i].Total > a.Items[j].Total
}

type ByCount struct {
	Items     []*CallItem
	Ascending bool
}

func (a ByCount) Len() int      { return len(a.Items) }
func (a ByCount) Swap(i, j int) { a.Items[i], a.Items[j] = a.Items[j], a.Items[i] }
func (a ByCount) Less(i, j int) bool {
	if a.Ascending {
		return a.Items[i].Count < a.Items[j].Count
	}
	return a.Items[i].Count > a.Items[j].Count
}

// SortStacks 对 PprofPBFile 的 Stacks 进行排序的方法
func (p *PprofPBFile) SortStacks(ascending bool) []*StackItem {
	var stackItems []*StackItem

	// 将 map 中的 StackItem 放入切片中
	for _, item := range p.Stacks {
		stackItems = append(stackItems, item)
	}

	// 使用排序规则对切片进行排序
	sort.Sort(ByTotal{Items: stackItems, Ascending: ascending})

	return stackItems
}

func ReadableCount(count int64) string {
	return numberPrinter.Sprintf("%d", count)
}

type StackItem struct {
	Total    int64
	Path     string
	CallInfo []*CallItem
	Metadata *Metadata
}

func (s *StackItem) TotalString() string {
	return CountToString(s.Metadata, s.Total)
}

func (s *StackItem) TotalPercent() string {
	return GetPercentString(s.Metadata.TotalSamples, s.Total)
}

type CallItem struct {
	Percent  float64
	Count    int64
	Path     string
	Stack    []string
	Metadata *Metadata
	pf       *PprofPBFile
}

func (s *CallItem) CountString() string {
	return CountToString(s.Metadata, s.Count)
}

func (s *CallItem) CountPercent() string {
	return GetPercentString(s.Metadata.TotalSamples, s.Count)
}

func (s *CallItem) getVendorPkgIdx() int {
	vendorPkgIdx := -1
	for idx, v1 := range s.Stack {
		if vendorPkgIdx < 0 && s.pf.Option.VendorPkgChecker(v1) {
			vendorPkgIdx = idx
			break
		}
	}
	return vendorPkgIdx
}

func (s *CallItem) StartFunc() string {
	if len(s.Stack) == 0 {
		return ""
	}
	return s.Stack[len(s.Stack)-1]
}
func (s *CallItem) MustStartFunc() string {
	f := s.StartFunc()
	if f == "" {
		f = s.Path
	}
	return f
}

func (s *CallItem) MustVendorFunc() string {
	f := s.VendorFunc()
	if f == "" {
		f = s.StartFunc()
	}
	if f == "" {
		return s.Path
	}
	return f
}
func (s *CallItem) VendorFunc() string {
	vendorPkgIdx := s.getVendorPkgIdx()
	if vendorPkgIdx < 0 {
		return ""
	}
	return s.Stack[vendorPkgIdx]
}
func (s *CallItem) NextFunc() string {
	if len(s.Stack) <= 1 {
		return ""
	}
	vendorPkgIdx := s.getVendorPkgIdx()
	return gcond.CondFn(vendorPkgIdx > 0, func() interface{} {
		return s.Stack[vendorPkgIdx-1]
	}, func() interface{} {
		return ""
	}).String()
}
func (s *CallItem) EndFunc() string {
	return s.Path
}

func CountToString(metadata *Metadata, count int64) string {
	switch metadata.Type {
	case "cpu", "delay":
		return time.Duration(count).String()
	case "inuse_space", "alloc_space":
		return gvalue.Must(gbytes.SizeStr(uint64(count))).String()
	}
	return ReadableCount(count)
}

type Metadata struct {
	File         string
	BuildID      string
	Type         string
	Time         time.Time
	Duration     time.Duration
	TotalSamples int64
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func (s *Metadata) TotalString() string {
	return CountToString(s, s.TotalSamples)
}

func (s *Metadata) Parse(header string) {
	//File: convoy-weibo-mesh
	//Build ID: 35389ef031ef81a0d9acda0fd2e29f90084516de
	//Type: cpu
	//Time: Nov 8, 2023 at 4:53pm (CST)
	//Duration: 300s, Total samples = 760.29s (253.43%)

	//File: convoy-weibo-mesh
	//	Build ID: 35389ef031ef81a0d9acda0fd2e29f90084516de
	//Type: goroutine
	//Time: Nov 8, 2023 at 4:58pm (CST)
	//Duration: 300s, Total samples = 18

	//heap
	//File: convoy-weibo-mesh
	//Build ID: 35389ef031ef81a0d9acda0fd2e29f90084516de
	//Type: inuse_space
	//Time: Nov 8, 2023 at 4:58pm (CST)
	//Duration: 300s, Total samples = 30.89MB

	//allocs
	//File: convoy-weibo-mesh
	//Build ID: 35389ef031ef81a0d9acda0fd2e29f90084516de
	//Type: inuse_space
	//Time: Nov 8, 2023 at 4:58pm (CST)
	//Duration: 300.01s, Total samples = 30.89MB

	//mutex & block
	//File: convoy-weibo-mesh
	//	Build ID: ee7fa290171182dd1542473d59deb357b8275fbc
	//Type: delay
	//Time: Nov 15, 2023 at 3:15pm (CST)
	//Duration: 300.01s, Total samples = 5.44hrs (6531.88%)

	for _, v := range strings.Split(header, "\n") {
		a := strings.SplitN(v, ":", 2)
		if len(a) < 2 {
			continue
		}
		a[1] = strings.TrimSpace(a[1])
		switch a[0] {
		case "File":
			s.File = a[1]
		case "Build ID":
			s.BuildID = a[1]
		case "Type":
			s.Type = a[1]
		case "Time":
			layout := "Jan 2, 2006 at 3:04pm (MST)"
			s.Time, _ = time.ParseInLocation(layout, a[1], time.Local)
		case "Duration":
			//Duration: 300s, Total samples = 760.29s (253.43%)
			lineArr := strings.Split(a[1], ",")
			s.Duration = gcast.ToDuration(lineArr[0])
			switch s.Type {
			case "cpu", "delay":
				r := strings.NewReplacer("hrs", "h")
				t := r.Replace(strings.Fields(lineArr[1])[3])
				s.TotalSamples = gcast.ToDuration(t).Nanoseconds()
			case "inuse_space", "alloc_space":
				s.TotalSamples = int64(gvalue.Must(
					gbytes.ParseSize(strings.ToUpper(strings.Fields(lineArr[1])[3]))).Uint64())
			default:
				s.TotalSamples = gcast.ToInt64(strings.Fields(lineArr[1])[3])
			}

		}
	}
}

func IsGoSrcPkg(p string) bool {
	a := strings.SplitN(p, "/", 2)[0]
	return !(strings.Contains(a, ".") && strings.Contains(p, "/"))
}

type SingleInfo struct {
	Path       string
	Count      string
	CountVal   int64
	Percent    string
	PercentVal float64
}

func (s SingleInfo) String() string {
	return fmt.Sprintf("%s %s %s", s.Path, s.Count, s.Percent)
}

type CallInfo struct {
	StartFunc  string
	VendorFunc string
	NextFunc   string
	EndFunc    string
	Count      string
	CountVal   int64
	Percent    string
	PercentVal float64
	//
	pbf *PprofPBFile
}

func (s CallInfo) GetPBF() *PprofPBFile {
	return s.pbf
}

func (s CallInfo) String() string {
	return fmt.Sprintf("%s -> ... -> %s -> %s -> ... -> [%s] %s %s",
		s.StartFunc, s.VendorFunc, s.NextFunc, s.EndFunc, s.Count, s.Percent)
}

type ParseOption struct {
	//profile pb.gz 格式文件路径（必须参数）
	File string
	//用来判断一个方法路径是go源码的还是项目的，用来定位调用关系（必须参数）
	VendorPkgChecker func(path string) bool
	// profile 文件类型（必须参数）
	ProfileType ProfileType
}

func getHeapType(profileType ProfileType) string {
	switch profileType {
	case ProfileTypeHeapInuseObjects:
		return "inuse_objects"
	case ProfileTypeHeapInuseSpace:
		return "inuse_space"
	case ProfileTypeHeapAllocObjects:
		return "alloc_objects"
	case ProfileTypeHeapAllocSpace:
		return "alloc_space"
	}
	return ""
}

func Parse(f string, opt *ParseOption) (pb *PprofPBFile, singleItems []*SingleInfo, callItems []*CallInfo, err error) {
	pb, err = NewPprofPBFile(opt, f)
	if err != nil {
		return nil, nil, nil, err
	}
	r := pb.SortStacks(false)
	for _, v := range r {
		item := &SingleInfo{
			Path:       v.Path,
			Count:      v.TotalString(),
			CountVal:   v.Total,
			Percent:    v.TotalPercent(),
			PercentVal: GetPercent(v.Metadata.TotalSamples, v.Total),
		}
		singleItems = append(singleItems, item)
	}

	mark := gset.New()
	for _, v := range r {
		for _, v0 := range v.CallInfo {
			if v0.Count == 0 {
				continue
			}
			vendorPkgIdx := -1
			for idx, v1 := range v0.Stack {
				if vendorPkgIdx < 0 && opt.VendorPkgChecker(v1) {
					vendorPkgIdx = idx
					break
				}
			}
			if vendorPkgIdx < 0 {
				continue
			}
			callStartFunc := v0.Stack[len(v0.Stack)-1]
			callVendorFunc := v0.Stack[vendorPkgIdx]
			callNextFunc := gcond.CondFn(vendorPkgIdx > 0, func() interface{} {
				return v0.Stack[vendorPkgIdx-1]
			}, func() interface{} {
				return ""
			}).String()
			callEndFunc := v.Path

			key := fmt.Sprintf("%s-%s-%s", callVendorFunc, callNextFunc, callEndFunc)
			if mark.Contains(key) {
				continue
			}
			mark.Add(key)

			item := &CallInfo{
				StartFunc:  callStartFunc,
				VendorFunc: callVendorFunc,
				NextFunc:   callNextFunc,
				EndFunc:    callEndFunc,
				Count:      v0.CountString(),
				CountVal:   v0.Count,
				Percent:    v0.CountPercent(),
				PercentVal: v0.Percent,
				pbf:        pb,
			}
			callItems = append(callItems, item)
		}
	}
	return
}
