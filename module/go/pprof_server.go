package gotool

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"github.com/snail007/gmc"
	gcore "github.com/snail007/gmc/core"
	gtemplate "github.com/snail007/gmc/http/template"
	gview "github.com/snail007/gmc/http/view"
	gctx "github.com/snail007/gmc/module/ctx"
	gcond "github.com/snail007/gmc/util/cond"
	gfile "github.com/snail007/gmc/util/file"
	gmap "github.com/snail007/gmc/util/map"
	"github.com/snail007/gmct/module/go/static"
	"github.com/snail007/gmct/module/go/template"
	gprofile "github.com/snail007/gmct/util/profile"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type PprofFiles struct {
	CpuFile       string
	MemFile       string
	AllocsFile    string
	GoroutineFile string
	MutexFile     string
	BlockFile     string
}

func (s *PprofFiles) IsExistsCpuFile() bool {
	return s.CpuFile != "" && gfile.IsFile(s.CpuFile)
}

func (s *PprofFiles) IsExistsMemFile() bool {
	return s.MemFile != "" && gfile.IsFile(s.MemFile)
}

func (s *PprofFiles) IsExistsGoroutineFile() bool {
	return s.GoroutineFile != "" && gfile.IsFile(s.GoroutineFile)
}

func (s *PprofFiles) IsExistsAllocsFile() bool {
	return s.AllocsFile != "" && gfile.IsFile(s.AllocsFile)
}

func (s *PprofFiles) IsExistsMutexFile() bool {
	return s.MutexFile != "" && gfile.IsFile(s.MutexFile)
}

func (s *PprofFiles) IsExistsBlockFile() bool {
	return s.BlockFile != "" && gfile.IsFile(s.BlockFile)
}

type PprofServer struct {
	server      gcore.APIServer
	files       *PprofFiles
	staticFiles embed.FS
	tplFiles    embed.FS
	tpl         *gtemplate.Template
}

func NewPprofServer(addr string, files *PprofFiles) (s *PprofServer, err error) {
	if addr == "" {
		addr = ":"
	}
	server, err := gmc.New.APIServer(gctx.NewCtx(), addr)
	if err != nil {
		return
	}
	s = &PprofServer{
		files:       files,
		server:      server,
		staticFiles: static.Files,
		tplFiles:    template.Files,
	}
	err = s.init()
	if err != nil {
		return nil, err
	}
	return
}

func (s *PprofServer) getParser(pkg string) (parser *gprofile.AnalyserParser, err error) {
	checker := func(str string) bool {
		return strings.Contains(str, pkg)
	}
	var parseOptions []*gprofile.ParseOption
	if !s.files.IsExistsMemFile() || !s.files.IsExistsCpuFile() {
		return nil, fmt.Errorf("cpu or memory pprof file is missing")
	}

	if s.files.IsExistsCpuFile() {
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.CpuFile,
			ProfileType: gprofile.ProfileTypeCPU,
		})
	}

	if s.files.IsExistsMemFile() {
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.MemFile,
			ProfileType: gprofile.ProfileTypeHeapInuseSpace,
		})
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.MemFile,
			ProfileType: gprofile.ProfileTypeHeapInuseObjects,
		})
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.MemFile,
			ProfileType: gprofile.ProfileTypeHeapAllocSpace,
		})
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.MemFile,
			ProfileType: gprofile.ProfileTypeHeapAllocObjects,
		})
	}
	if s.files.IsExistsAllocsFile() {
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.AllocsFile,
			ProfileType: gprofile.ProfileTypeAllocs,
		})
	}
	if s.files.IsExistsGoroutineFile() {
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.GoroutineFile,
			ProfileType: gprofile.ProfileTypeGoroutine,
		})
	}
	if s.files.IsExistsMutexFile() {
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.MutexFile,
			ProfileType: gprofile.ProfileTypeMutex,
		})
	}
	if s.files.IsExistsBlockFile() {
		parseOptions = append(parseOptions, &gprofile.ParseOption{
			File:        s.files.BlockFile,
			ProfileType: gprofile.ProfileTypeBlock,
		})
	}
	parser, err = gprofile.NewAnalyserParser(
		&gprofile.ParseOption{
			VendorPkgChecker: checker,
		}, parseOptions...)
	return
}
func (s *PprofServer) init() (err error) {
	//init template
	s.tpl, err = gtemplate.NewTemplate(gctx.NewCtx(), "")
	if err != nil {
		return
	}
	s.tpl.DisableLoadDefaultBinData()
	parseTemplateFromEmbedFS(s.tpl, s.tplFiles, "tpl/")
	s.tpl.DdisableLogging()
	err = s.tpl.Parse()
	if err != nil {
		return
	}
	//init static files
	serveEmbedFS(s.server.Router(), s.staticFiles, "/static/")

	//init analyser ui
	s.server.Router().HandlerFuncAny("/", func(w http.ResponseWriter, r *http.Request) {
		s.getView(w).Render("profile/index")
	})

	s.server.Router().HandlerFuncAny("/analysis", func(w http.ResponseWriter, r *http.Request) {
		ctx := gctx.NewCtxWithHTTP(w, r)
		cpuMin := ctx.GET("min_cpu")
		memoryMin := ctx.GET("min_memory_size")
		objectsMin := ctx.GET("min_memory_objects")
		mutexMin := ctx.GET("min_mutex")
		blockMin := ctx.GET("min_block")
		goroutineMin := ctx.GET("min_goroutine")
		cpuSuggestMin := ctx.GET("min_cpu")
		pkg := ctx.GET("pkg")
		parser, err := s.getParser(pkg)
		if err != nil {
			ctx.Status(500)
			ctx.Write(err.Error())
			return
		}
		// cpu top
		cpuAnalyser := parser.GetAnalyser(gprofile.ProfileTypeCPU)
		cpuData := cpuAnalyser.Top(cpuMin)
		// 历史内存分配 top
		memoryAnalyser := parser.GetAnalyser(gprofile.ProfileTypeHeapAllocSpace)
		memoryData := memoryAnalyser.Top(memoryMin)
		// 历史内存创建对象 top
		objectsAnalyser := parser.GetAnalyser(gprofile.ProfileTypeHeapAllocObjects)
		objectsData := objectsAnalyser.Top(objectsMin)
		// cpu和历史内存分配 top 交叉
		cpuMemoryData := parser.TopMultiple(map[gprofile.ProfileType]interface{}{
			gprofile.ProfileTypeCPU:              cpuMin,
			gprofile.ProfileTypeHeapAllocSpace:   memoryMin,
			gprofile.ProfileTypeHeapAllocObjects: objectsMin,
		})

		// 内存使用 top
		memoryInuseAnalyser := parser.GetAnalyser(gprofile.ProfileTypeHeapInuseSpace)
		memorInuseyData := memoryInuseAnalyser.Top(memoryMin)
		// 历史内存创建对象 top
		objectsInuseAnalyser := parser.GetAnalyser(gprofile.ProfileTypeHeapInuseObjects)
		objectsInuseData := objectsInuseAnalyser.Top(objectsMin)
		// cpu和历史内存分配 top 交叉
		cpuMemoryInuseData := parser.TopMultiple(map[gprofile.ProfileType]interface{}{
			gprofile.ProfileTypeCPU:              cpuMin,
			gprofile.ProfileTypeHeapInuseSpace:   memoryMin,
			gprofile.ProfileTypeHeapInuseObjects: objectsMin,
		})

		// mutex top
		mutexAnalyser := parser.GetAnalyser(gprofile.ProfileTypeMutex)
		mutexData := getTopViewData(mutexAnalyser, mutexMin)

		// block top
		blockAnalyser := parser.GetAnalyser(gprofile.ProfileTypeBlock)
		blockData := getTopViewData(blockAnalyser, blockMin)

		// goroutine top
		goroutineAnalyser := parser.GetAnalyser(gprofile.ProfileTypeGoroutine)
		goroutineData := getTopViewData(goroutineAnalyser, goroutineMin)

		//内存分配与正在使用内存差值
		inuse := parser.GetAnalyser(gprofile.ProfileTypeHeapInuseSpace).SumVendor(memoryMin)
		allocDiffInuseData := parser.GetAnalyser(gprofile.ProfileTypeHeapAllocSpace).SumVendor(memoryMin).Diff(inuse)
		//累积创建对象数量与正在使用对象数量差值
		inuseObj := parser.GetAnalyser(gprofile.ProfileTypeHeapInuseObjects).SumVendor(objectsMin)
		allocObjDiffInuseData := parser.GetAnalyser(gprofile.ProfileTypeHeapAllocObjects).SumVendor(objectsMin).Diff(inuseObj)

		//汇总vendor的“转折点方法”占用CPU情况
		sumVendorCPUData := parser.GetAnalyser(gprofile.ProfileTypeCPU).SumVendor(cpuMin)
		//汇总vendor的“转折点方法”占用内存分配情况
		sumVendorMemoryData := parser.GetAnalyser(gprofile.ProfileTypeHeapAllocSpace).SumVendor(memoryMin)
		//汇总vendor的“转折点方法”占用内存对象创建情况
		sumVendorObjectsData := parser.GetAnalyser(gprofile.ProfileTypeHeapAllocObjects).SumVendor(objectsMin)

		// 通用CPU分析
		cpuSuggestData := parser.GetAnalyser(gprofile.ProfileTypeCPU).Suggest(cpuSuggestMin)

		d := gmap.M{
			"cpuData":               getTopItemListInfo(cpuAnalyser, cpuData),
			"memoryData":            getTopItemListInfo(memoryAnalyser, memoryData),
			"objectsData":           getTopItemListInfo(objectsAnalyser, objectsData),
			"cpuMemoryData":         parseCpuMemoryData(cpuMemoryData, false),
			"memoryInuseData":       getTopItemListInfo(memoryInuseAnalyser, memorInuseyData),
			"objectsInuseData":      getTopItemListInfo(objectsInuseAnalyser, objectsInuseData),
			"cpuMemoryInuseData":    parseCpuMemoryData(cpuMemoryInuseData, true),
			"allocDiffInuseData":    allocDiffInuseData,
			"allocObjDiffInuseData": allocObjDiffInuseData,
			"sumVendorCPUData":      sumVendorCPUData,
			"sumVendorMemoryData":   sumVendorMemoryData,
			"sumVendorObjectsData":  sumVendorObjectsData,
			"cpuSuggestData": map[string]interface{}{
				"Info": getDefaultSuggestConfigMap(),
				"Data": cpuSuggestData,
			},
			"mutexData":     getTopItemListInfo(mutexAnalyser, mutexData),
			"blockData":     getTopItemListInfo(blockAnalyser, blockData),
			"goroutineData": getTopItemListInfo(goroutineAnalyser, goroutineData),
		}
		s.getView(w).SetMap(d).Render("profile/analysis")
	})
	return
}

func (s *PprofServer) getView(w http.ResponseWriter) gcore.View {
	return gview.New(w, s.tpl).Layout("layout/page")
}

func (s *PprofServer) Start() (err error) {
	return s.server.Run()
}

func parseTemplateFromEmbedFS(tpl *gtemplate.Template, tplFiles embed.FS, trimPrefix string) {
	bindData := map[string]string{}
	fs.WalkDir(tplFiles, "tpl", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(d.Name()) == ".html" {
			b, _ := tplFiles.ReadFile(path)
			key := gcond.Cond(trimPrefix == "", path, strings.TrimPrefix(path, "tpl/")).String()
			key = strings.TrimSuffix(key, ".html")
			bindData[key] = base64.StdEncoding.EncodeToString(b)
		}
		return nil
	})
	tpl.SetBinBase64(bindData)
}

func serveEmbedFS(router gcore.HTTPRouter, staticFiles embed.FS, pathPrefix string) {
	pathPrefix = strings.TrimSuffix(pathPrefix, "/")
	bindPath := pathPrefix + "/*filepath"
	router.HandlerFuncAny(bindPath, func(w http.ResponseWriter, r *http.Request) {
		ctx := gctx.NewCtxWithHTTP(w, r)
		path := strings.TrimPrefix(ctx.Request().URL.Path, pathPrefix+"/")
		b, err := staticFiles.ReadFile(path)
		if err != nil {
			ctx.Status(500)
			ctx.Write(err)
			return
		}
		http.ServeContent(w, r, filepath.Base(path), time.Time{}, bytes.NewReader(b))
	})
}

type TopItemListInfo struct {
	Total   string
	Percent string
	Data    gprofile.TopItemList
}

func getTopViewData(a *gprofile.Analyser, min interface{}) gprofile.TopItemList {
	if a != nil {
		return a.Top(min)
	}
	return nil
}
func getTopItemListInfo(a *gprofile.Analyser, list gprofile.TopItemList) TopItemListInfo {
	if a == nil {
		return TopItemListInfo{}
	}
	info := TopItemListInfo{Data: list}
	total := int64(0)
	percent := float64(0)
	for _, v := range list {
		total += v.Info.CountVal
		percent += v.Info.PercentVal
	}
	info.Percent = gprofile.PercentString(percent, 4)
	info.Total = gprofile.CountToString(a.GetPBF().Metadata, total)
	return info
}

type CPUMemoryDataRow struct {
	Path              string
	OrderTotalPercent string
	CPUInfo           map[string]interface{}
	MemoryInfo        map[string]interface{}
	ObjectsInfo       map[string]interface{}
	Data              []map[string]string
}

func getDefaultSuggestConfigMap() map[string]*gprofile.SuggestConfigItem {
	ret := map[string]*gprofile.SuggestConfigItem{}
	for _, v := range gprofile.DefaultSuggestConfig {
		ret[v.Path] = v
	}
	return ret
}
func parseCpuMemoryData(list gprofile.TopMultipleItemList, isInuse bool) []CPUMemoryDataRow {
	cpuType := gprofile.ProfileTypeCPU
	memType := gprofile.ProfileTypeHeapAllocSpace
	objType := gprofile.ProfileTypeHeapAllocObjects
	if isInuse {
		memType = gprofile.ProfileTypeHeapInuseSpace
		objType = gprofile.ProfileTypeHeapInuseObjects
	}
	var d []CPUMemoryDataRow
	for _, v := range list {
		item := CPUMemoryDataRow{
			OrderTotalPercent: gprofile.PercentString(v.OrderTotalPercent, 4),
			Path:              v.Path,
			CPUInfo: map[string]interface{}{
				"Count":      v.TotalMap[cpuType],
				"Percent":    v.TotalPercentMap[cpuType],
				"CountVal":   v.TotalValMap[cpuType],
				"PercentVal": v.TotalPercentValMap[cpuType],
			},
			MemoryInfo: map[string]interface{}{
				"Count":      v.TotalMap[memType],
				"Percent":    v.TotalPercentMap[memType],
				"CountVal":   v.TotalValMap[memType],
				"PercentVal": v.TotalPercentValMap[memType],
			},
			ObjectsInfo: map[string]interface{}{
				"Count":      v.TotalMap[objType],
				"Percent":    v.TotalPercentMap[objType],
				"CountVal":   v.TotalValMap[objType],
				"PercentVal": v.TotalPercentValMap[objType],
			},
		}
		for _, vv := range v.List {
			item0 := map[string]string{
				"NextFunc":    vv.NextFunc,
				"EndFunc":     vv.EndFunc,
				"Count":       vv.Count,
				"Percent":     vv.Percent,
				"ProfileType": vv.GetPBF().Option.ProfileType.String(),
			}
			item.Data = append(item.Data, item0)
		}

		d = append(d, item)
	}
	return d
}
