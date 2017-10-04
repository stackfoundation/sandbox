// +build linux darwin

package install

import (
	"os/exec"
)

func getStackFoundationRoot() (string, error) {
	return "/usr/local/sf", nil
}

func ElevatedExecute(binary, parameters string) error {
	cmd := exec.Command("/bin/sh", "-c", "sudo "+binary+" "+parameters)
	_, err := cmd.CombinedOutput()

	return err
}
