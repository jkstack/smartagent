package utils

import (
	"crypto/md5"
	"io"
	"os"
)

func MD5Checksum(dir string) ([md5.Size]byte, error) {
	var ret [md5.Size]byte
	f, err := os.Open(dir)
	if err != nil {
		return ret, err
	}
	defer f.Close()
	enc := md5.New()
	_, err = io.Copy(enc, f)
	if err != nil {
		return ret, err
	}
	copy(ret[:], enc.Sum(nil))
	return ret, nil
}
