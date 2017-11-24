package io

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Unzip Unzips all files in the specified zip file to the specified destination
func Unzip(src io.ReaderAt, size int64, dest string) error {
	zipReader, err := zip.NewReader(src, size)
	if err != nil {
		return err
	}

	for _, entry := range zipReader.File {
		entryReader, err := entry.Open()
		if err != nil {
			return err
		}
		defer entryReader.Close()

		entryPath := filepath.Join(dest, entry.Name)

		if entry.FileInfo().IsDir() {
			os.MkdirAll(entryPath, os.ModePerm)
		} else {
			entryDirectory := filepath.Dir(entryPath)

			err = os.MkdirAll(entryDirectory, os.ModePerm)
			if err != nil {
				return err
			}

			file, err := os.OpenFile(entryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode())
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, entryReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
