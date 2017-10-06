// +build linux darwin

package install

import (
	"github.com/stackfoundation/process"
)

func getStackFoundationRoot() (string, error) {
	return "/usr/local/sf", nil
}

func ElevatedExecute(binary, parameters string) error {
	err := process.CombineStdStreams("/bin/sh", "-c", "sudo "+binary+" "+parameters)

	return err
}
