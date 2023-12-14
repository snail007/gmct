package gprofile

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var checker = func(pkgPath string) bool {
	return strings.Contains(pkgPath, "github.com/weibocom/motan-go")
}

func TestCountToStr(t *testing.T) {
	fmt.Println(ReadableCount(1123894))
}

func TestNewPprofPBFile(t *testing.T) {
	opt := &ParseOption{
		VendorPkgChecker: checker,
		ProfileType:      ProfileTypeHeapAllocSpace,
	}
	pbf, err := NewPprofPBFile(opt, "/Users/user/go/src/test/f/heap.pb.gz")
	assert.Nil(t, err)
	fmt.Println(pbf.ImportPkgList, pbf.GoPath, pbf.GoRoot)
}
