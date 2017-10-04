package io

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func md5sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// MD5SumEquals Check if the MD5 checksum of a file equals the specified value
func MD5SumEquals(file string, md5 string) bool {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	fileMD5, err := md5sum(file)
	if err != nil {
		return false
	}

	return fileMD5 == md5
}
