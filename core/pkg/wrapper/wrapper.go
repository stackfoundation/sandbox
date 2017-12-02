package wrapper

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	coreio "github.com/stackfoundation/sandbox/core/pkg/io"
	"github.com/stackfoundation/sandbox/net/proxy"
)

var cliURL = "https://updates.stack.foundation/cli.zip"

const maxCLISize = 500 * 1024

func downloadWrappers() (io.ReadCloser, error) {
	response, err := proxy.ProxyCapableClient.Get(cliURL)
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

// ExtractWrappers Extract the CLI wrappers to the specified directory
func ExtractWrappers(path string) error {
	err := os.MkdirAll(path, os.ModePerm)

	if err == nil {
		wrappersReader, err := downloadWrappers()
		if err != nil {
			return err
		}
		defer wrappersReader.Close()

		data, err := ioutil.ReadAll(io.LimitReader(wrappersReader, maxCLISize))
		if err != nil {
			return err
		}

		dataReader := bytes.NewReader(data)
		return coreio.Unzip(dataReader, int64(len(data)), path)
	}

	return err
}
