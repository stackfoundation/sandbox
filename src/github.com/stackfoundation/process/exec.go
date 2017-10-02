package process

import "os/exec"

// CommandOut Execute a process, wait till completion, and return the stdout
func CommandOut(command string, arg ...string) (string, error) {
	cmd := exec.Command(command, arg...)
	out, err := cmd.Output()

	if out == nil {
		return "", err
	}

	return string(out), err
}
