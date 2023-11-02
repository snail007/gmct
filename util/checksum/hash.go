package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"time"

	"golang.org/x/crypto/blake2s"
)

const bufferSize = 65536

// MD5sumReader returns MD5 checksum of content in reader
func MD5sumReader(reader io.ReadCloser) (string, error) {
	return sumReader(md5.New(), reader)
}

// SHA256sumReader returns SHA256 checksum of content in reader
func SHA256sumReader(reader io.ReadCloser) (string, error) {
	return sumReader(sha256.New(), reader)
}

// SHA1sumReader returns SHA1 checksum of content in reader
func SHA1sumReader(reader io.ReadCloser) (string, error) {
	return sumReader(sha1.New(), reader)
}

// Blake2s256Reader returns SHA1 checksum of content in reader
func Blake2s256Reader(reader io.ReadCloser) (string, error) {
	hash, _ := blake2s.New256([]byte{})
	return sumReader(hash, reader)
}

// CRCReader returns CRC-32-IEEE checksum of content in reader
func CRCReader(reader io.Reader) (string, error) {
	table := crc32.MakeTable(crc32.IEEE)
	checksum := crc32.Checksum([]byte(""), table)
	buf := make([]byte, bufferSize)
	for {
		switch n, err := reader.Read(buf); err {
		case nil:
			checksum = crc32.Update(checksum, table, buf[:n])
		case io.EOF:
			return fmt.Sprintf("%08x", checksum), nil
		default:
			return "", err
		}
	}
}

// sumReader calculates the hash based on a provided hash provider
func sumReader(hashAlgorithm hash.Hash, reader io.ReadCloser) (string, error) {
	buf := make([]byte, bufferSize)
	t := time.AfterFunc(time.Second*30, func() {
		reader.Close()
	})
	defer t.Stop()
	for {
		t.Reset(time.Second * 30)
		switch n, err := reader.Read(buf); err {
		case nil:
			hashAlgorithm.Write(buf[:n])
		case io.EOF:
			return fmt.Sprintf("%x", hashAlgorithm.Sum(nil)), nil
		default:
			return "", err
		}
	}
}
