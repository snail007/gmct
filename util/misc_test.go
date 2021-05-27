package util

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetPackageName(t *testing.T) {
	assert := assert.New(t)
	str := ` 
//  xxx
 package utils  

`
	str1 := ` 
//  xxx
 package utils_test  

`
	os.RemoveAll(".utils_test")
	os.MkdirAll(".utils_test/none", 0755)
	ioutil.WriteFile(".utils_test/str.go", []byte(str), 0755)
	ioutil.WriteFile(".utils_test/str_test.go", []byte(str1), 0755)
	ioutil.WriteFile(".utils_test/str_test.txt", []byte(str1), 0755)
	assert.Equal("utils", GetPackageName(".utils_test"))
	os.RemoveAll(".utils_test")
}
