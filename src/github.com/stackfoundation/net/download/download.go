package download

import (
	"io"
	"os"
	"path/filepath"

	"github.com/stackfoundation/log"
	"github.com/stackfoundation/net/proxy"
	"github.com/stackfoundation/progress"
)

// DownloadFromURL Download from an external URL to a local location
func DownloadFromURL(url string, dest string, logCode string) error {
	log.Debug(logCode, "Downloading %v to %v", url, dest)

	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		return err
	}

	output, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := proxy.ProxyCapableClient.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output,
		progress.NewProgressAwareReader(response.Body, "Downloading", logCode, response.ContentLength))
	if err != nil {
		return err
	}

	return nil
}
