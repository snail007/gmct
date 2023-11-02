// Package checksum computes checksums, like MD5 or SHA256, for large files
package checksum

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	gsync "github.com/snail007/gmc/util/sync"
	"hash"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/blake2s"
)

// MD5sum returns MD5 checksum of filename
func MD5sum(filename string) (string, error) {
	return sum(md5.New(), filename)
}

// SHA256sum returns SHA256 checksum of filename
func SHA256sum(filename string) (string, error) {
	return sum(sha256.New(), filename)
}

// SHA1sum returns SHA1 checksum of filename
func SHA1sum(filename string) (string, error) {
	return sum(sha1.New(), filename)
}

// Blake2s256 returns BLAKE2s-256 checksum of filename
func Blake2s256(filename string) (string, error) {
	hash, _ := blake2s.New256([]byte{})
	return sum(hash, filename)
}

// CRC32 returns CRC-32-IEEE checksum of filename
func CRC32(filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil || info.IsDir() {
		return "", err
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	return CRCReader(bufio.NewReader(file))
}

// sum calculates the hash based on a provided hash provider
func sum(hashAlgorithm hash.Hash, filename string) (string, error) {
	if info, err := os.Stat(filename); err != nil || info.IsDir() {
		return "", err
	}
	var file *os.File
	var err error
	g := sync.WaitGroup{}
	g.Add(1)
	go func() {
		defer g.Done()
		file, err = os.Open(filename)
	}()
	select {
	case <-gsync.Wait(&g):
	case <-time.After(time.Second * 3):
		return "", fmt.Errorf("open file timeout, %s", filename)
	}
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	return sumReader(hashAlgorithm, newBufReaderCloser(file))
}

type bufReaderCloser struct {
	*bufio.Reader
	f *os.File
}

func (s *bufReaderCloser) Close() error {
	if s.f != nil {
		return s.f.Close()
	}
	return nil
}

func newBufReaderCloser(f *os.File) *bufReaderCloser {
	return &bufReaderCloser{
		f:      f,
		Reader: bufio.NewReaderSize(f, 1024*8),
	}
}
