package gprofile

import (
	"fmt"
	gbytes "github.com/snail007/gmc/util/bytes"
	gcast "github.com/snail007/gmc/util/cast"
	gcond "github.com/snail007/gmc/util/cond"
	gmap "github.com/snail007/gmc/util/map"
	gset "github.com/snail007/gmc/util/set"
	gvalue "github.com/snail007/gmc/util/value"
	"sort"
	"strings"
)

const (
	defaultPercentKeepCount = 4

	ProfileTypeCPU ProfileType = iota + 1
	ProfileTypeHeapInuseSpace
	ProfileTypeHeapInuseObjects
	ProfileTypeHeapAllocSpace
	ProfileTypeHeapAllocObjects
	ProfileTypeAllocs
	ProfileTypeBlock
	ProfileTypeMutex
	ProfileTypeGoroutine
)

var (
	profileTypeNameMap = map[ProfileType]string{
		ProfileTypeCPU:              "cpu",
		ProfileTypeHeapInuseSpace:   "heap_inuse_space",
		ProfileTypeHeapInuseObjects: "heap_inuse_objects",
		ProfileTypeHeapAllocSpace:   "heap_alloc_space",
		ProfileTypeHeapAllocObjects: "heap_alloc_objects",
		ProfileTypeAllocs:           "allocs",
		ProfileTypeBlock:            "block",
		ProfileTypeMutex:            "mutex",
		ProfileTypeGoroutine:        "goroutine",
	}
	DefaultSuggestConfigMin = "15%"
	DefaultSuggestConfig    = []*SuggestConfigItem{
		{
			Path: "runtime/internal/syscall.Syscall6",
			Desc: "系统调用占比大，请考虑减少系统调用",
		},
		{
			Path: "runtime.mallocgc",
			Desc: "内存分配占比大，请考虑减少对象分配或者优化内存逃逸",
		},
		{
			Path: "runtime.gcBgMarkWorker",
			Desc: "内存回收占比大，请考虑减少创建和销毁对象的频率",
		},
		{
			Path: "runtime.mcall",
			Desc: "协程调度占比大，请考虑减少直接或者间接触发锁竞争的操作",
		},
		{
			Path: "runtime.memmove",
			Desc: "内存拷贝占比大，请考虑减少变量的复制或者转换，比如：字符串到字节slice，字节slice到字符串的转换等",
		},
		{
			Path: "runtime.newstack",
			Desc: "系统调用占比大，请检查方法里面，可能分配了代码中根本没有调用关系的变量",
		},
		{
			Path: "runtime.futex",
			Desc: "锁竞争时间占比大，请考虑减少锁的使用，减少锁竞争的发生",
		},
	}
)

type MathType int

const (
	MathTypeEqual MathType = iota + 1
	MathTypePrefix
)

type SuggestConfigItem struct {
	Path     string
	Desc     string
	Min      interface{}
	MathType MathType
}

func (s *SuggestConfigItem) Match(str string) bool {
	switch s.MathType {
	case MathTypePrefix:
		return strings.HasPrefix(str, s.Path)
	case MathTypeEqual:
		fallthrough
	default:
		return s.Path == str
	}
}

type ProfileType int

// 可能档值是：inuse_space, inuse_objects, alloc_space, alloc_objects
func (s ProfileType) String() string {
	return profileTypeNameMap[s]
}

type TopItem struct {
	//只有调用了Suggest方法，此变量才有值
	SuggestConfig *SuggestConfigItem
	analyser      *Analyser
	Info          *SingleInfo
	Stacks        []*CallInfo
}

type TopItemList []*TopItem

func (s TopItemList) Dump() {
	fmt.Println(s.String())
}

func (s TopItemList) String() string {
	if len(s) == 0 {
		return ""
	}
	a := s[0].analyser
	return a.parser.DumpTop(s)
}

type SortSumItem struct {
	Items     []*SumByCheckerResultItem
	Ascending bool
}

func (a SortSumItem) Len() int      { return len(a.Items) }
func (a SortSumItem) Swap(i, j int) { a.Items[i], a.Items[j] = a.Items[j], a.Items[i] }
func (a SortSumItem) Less(i, j int) bool {
	if a.Ascending {
		return a.Items[i].TotalVal < a.Items[j].TotalVal
	}
	return a.Items[i].TotalVal > a.Items[j].TotalVal
}

type SortTopMultipleItem struct {
	Items     []*TopMultipleItem
	Ascending bool
}

func (a SortTopMultipleItem) Len() int      { return len(a.Items) }
func (a SortTopMultipleItem) Swap(i, j int) { a.Items[i], a.Items[j] = a.Items[j], a.Items[i] }
func (a SortTopMultipleItem) Less(i, j int) bool {
	if a.Ascending {
		return a.Items[i].OrderTotalPercent < a.Items[j].OrderTotalPercent
	}
	return a.Items[i].OrderTotalPercent > a.Items[j].OrderTotalPercent
}

type SumByCheckerResultItem struct {
	Total      string
	VendorPath string
	TotalVal   int64
	pf         *PprofPBFile
	//下面四个值只有Diff返回的时候才有
	RawTotalVal  int64
	DiffTotalVal int64
	RawTotal     string
	DiffTotal    string
}

func (s SumByCheckerResultItem) GetPBF() *PprofPBFile {
	return s.pf
}

func (s SumByCheckerResultItem) Percent() string {
	return GetPercentString(s.pf.Metadata.TotalSamples, s.TotalVal)
}

type SumByCheckerResultItemList []*SumByCheckerResultItem

func (s SumByCheckerResultItemList) Dump() {
	fmt.Println(s.String())
}

func (s SumByCheckerResultItemList) ToMap() *gmap.Map {
	m := gmap.New()
	for _, v := range s {
		m.Store(v.VendorPath, v)
	}
	return m
}

// TopN 返回排名前N个。结果是排序好的。
func (s SumByCheckerResultItemList) TopN(count int) []*SumByCheckerResultItem {
	if len(s) == 0 {
		return nil
	}
	maxIdx := len(s) - 1
	if count-1 < maxIdx {
		maxIdx = count - 1
	}
	return s[:maxIdx]
}

// Diff 遍历自身列表，减去参数list里面对应的元素方法的占用值。通常用来计算alloc内存和inuse内存差值情况。结果是排序好的。
func (s SumByCheckerResultItemList) Diff(list SumByCheckerResultItemList) SumByCheckerResultItemList {
	diffMap := list.ToMap()
	for _, v := range s {
		v.RawTotalVal = v.TotalVal
		v.RawTotal = v.Total
	}
	for _, v := range s {
		if _item, ok := diffMap.Load(v.VendorPath); ok {
			diffItem := _item.(*SumByCheckerResultItem)
			v.TotalVal -= diffItem.TotalVal
			v.Total = CountToString(v.pf.Metadata, v.TotalVal)
			v.DiffTotal = diffItem.Total
			v.DiffTotalVal = diffItem.TotalVal
		} else {
			v.DiffTotal = "0"
			v.DiffTotalVal = 0
		}
	}
	return s
}

func (s SumByCheckerResultItemList) String() string {
	buf := gbytes.NewBytesBuilder()
	for idx, v := range s {
		percentStr := GetPercentString(v.pf.Metadata.TotalSamples, v.TotalVal)
		if v.DiffTotal != "" {
			buf.WriteStrLn("%d. %s, %s - %s = %s %s", idx+1, v.VendorPath, v.RawTotal, v.DiffTotal, v.Total, percentStr)
		} else {
			buf.WriteStrLn("%d. %s, %s %s", idx+1, v.VendorPath, v.Total, percentStr)
		}
	}
	return buf.String()
}

type Analyser struct {
	pf         *PprofPBFile
	data       []*CallInfo
	singleData []*SingleInfo
	opt        *ParseOption
	parser     *AnalyserParser
}

func (s *Analyser) GetPBF() *PprofPBFile {
	return s.pf
}

// Top 用于获取大于等于这个值的“终点方法”调用，结果是排序好的，
//
//	minStr 可能的写法：
//	内存大小：10B, 100KB, 1M, 1G, 1%
//	数值类型：10, 1000, 1%
//	时间类型：1ms, 2s, 3m, 4h, 1.5h, 1%
func (s *Analyser) Top(_min interface{}) TopItemList {
	min := s.parseCount(_min)
	return top(s, min)
}

// TopN 用于获取占用靠前的N个“终点方法”调用，结果是排序好的
func (s *Analyser) TopN(count int) TopItemList {
	return topN(s, count)
}

func (s *Analyser) Suggest(defaultMin string, configItems ...*SuggestConfigItem) (ret TopItemList) {
	if len(configItems) == 0 {
		configItems = DefaultSuggestConfig
	}
	if defaultMin == "" {
		defaultMin = DefaultSuggestConfigMin
	}
	list := s.Top(1)
	if len(list) == 0 {
		return nil
	}
	for _, v := range list {
		for _, cfg := range configItems {
			min := gcond.Cond(gvalue.IsEmpty(cfg.Min), defaultMin, cfg.Min).String()
			if cfg.Match(v.Info.Path) && s.parseCount(min) <= v.Info.CountVal {
				v.SuggestConfig = cfg
				ret = append(ret, v)
			}
		}
	}
	return
}

// BeginnerFuncs 获取全部的"起始方法"
func (s *Analyser) BeginnerFuncs() []string {
	m1 := map[string]bool{}
	for _, v := range s.pf.RawStacks {
		m1[v.MustStartFunc()] = true
	}
	for _, v := range s.pf.RawStacks {
		maxIdx := len(v.Stack) - 1
		for i := maxIdx; i >= 0; i-- {
			if i == maxIdx {
				continue
			}
			p := v.Stack[i]
			if m1[p] {
				delete(m1, p)
			}
		}
	}
	ret := []string{}
	for k := range m1 {
		ret = append(ret, k)
	}
	return ret
}

// SumVendor 汇总vendor的“转折点方法”占用情况
//
//	逻辑：
//	1.遍历原始数据全部的调用链，把包含“转折点方法”调用，根据“转折点方法”纬度，进行汇总。
func (s *Analyser) SumVendor(_min interface{}) (ret SumByCheckerResultItemList) {
	min := s.parseCount(_min)
	all := SumByCheckerResultItemList{}
	m := map[string]*SumByCheckerResultItem{}
	for _, v := range s.pf.RawStacks {
		k := v.MustVendorFunc()
		if !s.pf.Option.VendorPkgChecker(k) {
			continue
		}
		if item, ok := m[k]; ok {
			item.TotalVal += v.Count
		} else {
			item = &SumByCheckerResultItem{
				VendorPath: k,
				TotalVal:   v.Count,
				pf:         s.pf,
			}
			m[k] = item
		}
	}
	for _, v := range m {
		all = append(all, v)
	}
	for _, v := range all {
		if v.TotalVal < min {
			continue
		}
		v.Total = CountToString(s.pf.Metadata, v.TotalVal)
		ret = append(ret, v)
	}
	// 使用排序规则对切片进行排序
	sort.Sort(SortSumItem{Items: ret, Ascending: false})
	return
}

func (s *Analyser) parseCount(count interface{}) int64 {
	cnt := int64(0)
	str := gcast.ToString(count)
	if strings.HasSuffix(str, "%") {
		return int64(gcast.ToFloat64(str[:len(str)-1]) * float64(s.pf.Metadata.TotalSamples) / 100)
	}
	v := gvalue.NewAny(count)
	switch s.opt.ProfileType {
	case ProfileTypeMutex, ProfileTypeBlock, ProfileTypeCPU:
		cnt = v.Duration().Nanoseconds()
	case ProfileTypeHeapInuseObjects, ProfileTypeHeapAllocObjects:
		cnt = v.Int64()
	case ProfileTypeHeapInuseSpace, ProfileTypeHeapAllocSpace:
		cnt = gvalue.MustAny(gbytes.ParseSize(v.String())).Int64()
	case ProfileTypeAllocs:
		cnt = gvalue.MustAny(gbytes.ParseSize(v.String())).Int64()
	case ProfileTypeGoroutine:
		cnt = v.Int64()
	}
	return cnt
}

type AnalyserParser struct {
	analyserMap map[string]*Analyser
}

// NewAnalyserParser 可以同时分析cpu，heap（或allocs），mutex（或block），goroutine
func NewAnalyserParser(defaultOpt *ParseOption, opts ...*ParseOption) (s *AnalyserParser, err error) {
	if len(opts) == 0 {
		return nil, fmt.Errorf("option required")
	}
	if defaultOpt == nil {
		defaultOpt = &ParseOption{}
	}
	var mergeDefaultOpt = func(opt *ParseOption) {
		if opt.VendorPkgChecker == nil {
			opt.VendorPkgChecker = defaultOpt.VendorPkgChecker
		}
	}
	s = &AnalyserParser{analyserMap: map[string]*Analyser{}}
	for _, opt := range opts {
		mergeDefaultOpt(opt)
		pf, single, data, e := Parse(opt.File, opt)
		if e != nil {
			return nil, e
		}
		k := opt.ProfileType.String()
		s.analyserMap[k] = &Analyser{
			pf:         pf,
			data:       data,
			singleData: single,
			opt:        opt,
			parser:     s,
		}
	}
	return s, nil
}

func (s *AnalyserParser) GetAnalyser(profileType ProfileType) *Analyser {
	return s.analyserMap[profileType.String()]
}

type TopMultipleItem struct {
	List               []*CallInfo
	Path               string
	TotalMap           map[ProfileType]string
	TotalPercentMap    map[ProfileType]string
	TotalValMap        map[ProfileType]int64
	TotalPercentValMap map[ProfileType]float64
	//对TotalPercentValMap全部值求和，用于排序
	OrderTotalPercent float64
}

type TopMultipleItemList []*TopMultipleItem

func (s TopMultipleItemList) Dump() {
	fmt.Println(s.String())
}

func (s TopMultipleItemList) String() string {
	buf := gbytes.NewBytesBuilder()
	keys := []int{}
	if len(s) > 0 {
		for typ, _ := range s[0].TotalMap {
			keys = append(keys, int(typ))
		}
	}
	sort.Ints(keys)
	for idx, v := range s {
		buf.WriteStr("%d. %s", idx+1, v.Path)
		for _, _typ := range keys {
			typ := ProfileType(_typ)
			buf.WriteStr(", %s %s %s", typ.String(), v.TotalMap[typ], v.TotalPercentMap[typ])
		}
		buf.WriteStrLn("")
		for _, vv := range v.List {
			buf.WriteStrLn("	%s -> ... -> %s %s %s (%s)",
				vv.NextFunc, vv.EndFunc,
				vv.Count, vv.Percent, vv.pbf.Option.ProfileType)
		}
	}
	return buf.String()
}

// TopMultiple 多维度top汇总。比如cpu和内存都指向的“转折点方法”
//
//	逻辑是：
//	1.Top出终点方法
//	2.在终点方法的全部调用路径里面筛选出“转折点方法”
//	3.取“转折点方法”交集
//	4.结果是“转折点方法”列表，占用情况是“转折点方法”对应的调用链路占用情况汇总，并不是参数里面的大于min。
func (s *AnalyserParser) TopMultiple(minMap map[ProfileType]interface{}) (ret TopMultipleItemList) {
	mergeData := map[string]int{}
	matchListData := TopItemList{}
	for _, a := range s.analyserMap {
		min, ok := minMap[a.opt.ProfileType]
		if !ok {
			continue
		}
		list := a.Top(min)
		matchListData = append(matchListData, list...)
		addedMap := map[string]bool{}
		for _, v := range list {
			for _, vv := range v.Stacks {
				funcName := gcond.Cond(vv.VendorFunc == "", vv.StartFunc, vv.VendorFunc).String()
				if addedMap[funcName] {
					continue
				}
				addedMap[funcName] = true
				mergeData[funcName] += 1
			}
		}
	}
	targetLen := len(minMap)
	for f, cnt := range mergeData {
		if cnt < targetLen || cnt <= 1 {
			continue
		}
		item := &TopMultipleItem{
			Path:               f,
			TotalPercentMap:    map[ProfileType]string{},
			TotalMap:           map[ProfileType]string{},
			TotalValMap:        map[ProfileType]int64{},
			TotalPercentValMap: map[ProfileType]float64{},
		}
		for _, v := range matchListData {
			profileType := v.analyser.pf.Option.ProfileType
			for _, vv := range v.Stacks {
				if !(vv.StartFunc == f || vv.VendorFunc == f) {
					continue
				}
				item.List = append(item.List, vv)
				if _, ok := item.TotalMap[profileType]; !ok {
					item.TotalValMap[profileType] = 0
					item.TotalPercentValMap[profileType] = 0
				}
				p := PercentFloat(vv.Percent)
				item.TotalValMap[profileType] += vv.CountVal
				item.TotalPercentValMap[profileType] += p
				item.OrderTotalPercent += p

				item.TotalMap[profileType] = CountToString(v.analyser.pf.Metadata, item.TotalValMap[profileType])
				item.TotalPercentMap[profileType] = PercentString(item.TotalPercentValMap[profileType], defaultPercentKeepCount)
			}
		}
		ret = append(ret, item)
	}
	// 使用排序规则对切片进行排序
	sort.Sort(SortTopMultipleItem{Items: ret, Ascending: false})
	return
}

func (s *AnalyserParser) DumpTop(list TopItemList) string {
	buf := gbytes.NewBytesBuilder()
	pf := list[0].analyser.pf
	buf.WriteStrLn("=======================================")
	buf.WriteStrLn("Type: %s", pf.Option.ProfileType.String())
	buf.WriteStrLn("Duration: %s", pf.Metadata.Duration)
	buf.WriteStrLn("Total Samples: %s", pf.Metadata.TotalString())
	buf.WriteStrLn("=======================================")
	for idx, v := range list {
		idx++
		info := v.Info
		buf.WriteStrLn("\n%d. %s %s %s", idx, info.Path, info.Count, info.Percent)
		for _, vv := range v.Stacks {
			buf.WriteStrLn("	%s -> %s %s %s", vv.VendorFunc, vv.NextFunc, vv.Count, vv.Percent)
		}
	}
	return buf.String()
}

// ImportPackageList 用于获取程序import外部的全部类库的“版本路径”列表，可以用于go get等。
func (s *AnalyserParser) ImportPackageList() []string {
	set := gset.New()
	for _, a := range s.analyserMap {
		set.MergeStringSlice(a.pf.ImportPkgList)
	}
	return set.ToStringSlice()
}

// GoRoot 遍历内部全部Analyser，返回第一个非空的GoRoot
func (s *AnalyserParser) GoRoot() string {
	for _, a := range s.analyserMap {
		if a.GetPBF().GoRoot != "" {
			return a.GetPBF().GoRoot
		}
	}
	return ""
}

// GoPath 遍历内部全部Analyser，返回第一个非空的GoPath
func (s *AnalyserParser) GoPath() string {
	for _, a := range s.analyserMap {
		if a.GetPBF().GoPath != "" {
			return a.GetPBF().GoPath
		}
	}
	return ""
}

func top(a *Analyser, min int64) (ret TopItemList) {
	singleMap := map[string]*SingleInfo{}
	for _, v := range a.singleData {
		if v.CountVal < min {
			continue
		}
		singleMap[v.Path] = v
	}
	return filterBy(a, singleMap)
}

func topN(a *Analyser, count int) (ret TopItemList) {
	singleMap := map[string]*SingleInfo{}
	for idx, v := range a.singleData {
		if idx+1 <= count {
			singleMap[v.Path] = v
		} else {
			break
		}
	}
	return filterBy(a, singleMap)
}

func filterBy(a *Analyser, singleMap map[string]*SingleInfo) (ret TopItemList) {
	topItemsMap := gmap.New()
	for _, stack := range a.data {
		if _, ok := singleMap[stack.EndFunc]; !ok {
			continue
		}
		if _item, exists := topItemsMap.Load(stack.EndFunc); !exists {
			topItemsMap.Store(stack.EndFunc, &TopItem{
				Info:     singleMap[stack.EndFunc],
				Stacks:   []*CallInfo{stack},
				analyser: a,
			})
		} else {
			item := _item.(*TopItem)
			item.Stacks = append(item.Stacks, stack)
		}
	}
	topItemsMap.RangeFast(func(key, value interface{}) bool {
		ret = append(ret, value.(*TopItem))
		return true
	})
	return
}

func DumpStringSlice(list []string) {
	buf := gbytes.NewBytesBuilder()
	for idx, v := range list {
		buf.WriteStrLn("%d. %s", idx+1, v)
	}
	fmt.Println(buf.String())
}

func PercentFloat(percent string) float64 {
	if percent == "" {
		return 0
	}
	if strings.HasSuffix(percent, "%") {
		percent = percent[:len(percent)-1]
	}
	return gcast.ToFloat64(percent) / 100
}

func PercentString(percent float64, keepCount int) string {
	return fmt.Sprintf("%."+gcast.ToString(keepCount)+"f%%", percent*100)
}

func GetPercentString(total, val int64) string {
	if total == 0 {
		return "0%"
	}
	return PercentString(float64(val)/float64(total), defaultPercentKeepCount)
}

func GetPercent(total, val int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(val) / float64(total)
}
