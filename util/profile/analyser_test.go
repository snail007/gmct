package gprofile

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	glog "github.com/snail007/gmc/module/log"
	"strings"
	"testing"
)

func TestPercentString(t *testing.T) {
	a := 0.1234
	assert.Equal(t, PercentString(a, 1), "12.3%")
}

func TestPercentFloat(t *testing.T) {
	a := "12.34%"
	assert.Equal(t, PercentFloat(a), 0.1234)
}

func TestAnalyser_BeginnerFuncs(t *testing.T) {
	var checker = func(pkgPath string) bool {
		return strings.Contains(pkgPath, "github.com/weibocom/motan-go")
	}
	parser, err := NewAnalyserParser(
		&ParseOption{
			VendorPkgChecker: checker,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/cpu.pb.gz",
			ProfileType: ProfileTypeCPU,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/heap.pb.gz",
			ProfileType: ProfileTypeHeapInuseSpace,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/heap.pb.gz",
			ProfileType: ProfileTypeHeapInuseObjects,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/heap.pb.gz",
			ProfileType: ProfileTypeHeapAllocSpace,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/heap.pb.gz",
			ProfileType: ProfileTypeHeapAllocObjects,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/allocs.pb.gz",
			ProfileType: ProfileTypeAllocs,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/goroutine.pb.gz",
			ProfileType: ProfileTypeGoroutine,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/mutex.pb.gz",
			ProfileType: ProfileTypeMutex,
		},
		&ParseOption{
			File:        "/Users/user/Desktop/tmp/block.pb.gz",
			ProfileType: ProfileTypeBlock,
		},
	)
	if err != nil {
		glog.Fatal(err)
		return
	}
	fmt.Println("")
	DumpStringSlice(parser.ImportPackageList())
	//parser.GetAnalyser(ProfileTypeCPU).Top("1%").Dump()
	//parser.GetAnalyser(ProfileTypeBlock).Top("1ms").Dump()
	//parser.GetAnalyser(ProfileTypeMutex).Top("1ms").Dump()
	//parser.GetAnalyser(ProfileTypeAllocs).Top("1K").Dump()
	//parser.GetAnalyser(ProfileTypeGoroutine).Top(1).Dump()
	//
	//parser.GetAnalyser(ProfileTypeHeapInuseSpace).Top("512B").Dump()
	//parser.GetAnalyser(ProfileTypeHeapInuseObjects).Top(10).Dump()
	//parser.GetAnalyser(ProfileTypeHeapAllocSpace).Top("100M").Dump()
	//parser.GetAnalyser(ProfileTypeHeapAllocObjects).Top(100).Dump()

	parser.TopMultiple(map[ProfileType]interface{}{
		ProfileTypeCPU:            "5%",
		ProfileTypeHeapAllocSpace: "1%",
	}).Dump()

	//parser.GetAnalyser(ProfileTypeHeapAllocSpace).SumVendor("500M").Dump()
	//parser.GetAnalyser(ProfileTypeHeapInuseSpace).SumVendor("512").Dump()

	//inuse := parser.GetAnalyser(ProfileTypeHeapInuseSpace).SumVendor("1%")
	//parser.GetAnalyser(ProfileTypeHeapAllocSpace).SumVendor("1%").Diff(inuse).Dump()

	//DumpStringSlice(parser.GetAnalyser(ProfileTypeHeapInuseSpace).BeginnerFuncs())

	//list := parser.GetAnalyser(ProfileTypeCPU).Suggest()
	//list.Dump()
}
