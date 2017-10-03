// +build linux darwin

package install

import (
	"os/exec"
	"os/user"
	"path/filepath"
)

func getStackFoundationRoot() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(currentUser.HomeDir, ".sf"), nil
}

func ElevatedExecute(binary, parameters string) error {
	cmd := exec.Command("/bin/sh", "-c", "sudo "+binary+" "+parameters)
	_, err := cmd.CombinedOutput()

	return err
}
