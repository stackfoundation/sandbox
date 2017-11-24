package wrapper

import (
	"bytes"
	"os"

	"github.com/stackfoundation/core/pkg/io"
	"github.com/stackfoundation/core/pkg/minikube/assets"
)

// ExtractWrappers Extract the CLI wrappers to the specified directory
func ExtractWrappers(path string) error {
	err := os.MkdirAll(path, os.ModePerm)

	if err == nil {
		data, err := assets.Asset("out/cli.zip")

		if err != nil {
			return err
		}

		dataReader := bytes.NewReader(data)
		return io.Unzip(dataReader, int64(len(data)), path)
	}

	return err
}
