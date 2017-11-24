// +build linux darwin

package install

import (
	"fmt"

	"github.com/stackfoundation/process"
)

func getStackFoundationRoot() (string, error) {
	return "/usr/local/sf", nil
}

func ElevatedExecute(binary, parameters string) error {
	fmt.Println("Root privileges are required for this step, you may be prompted for your password")
	err := process.CombineStdStreams("/bin/sh", "-c", "sudo "+binary+" "+parameters)

	return err
}
