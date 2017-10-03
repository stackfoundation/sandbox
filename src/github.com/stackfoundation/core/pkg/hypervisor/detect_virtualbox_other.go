// +build !windows

package hypervisor

import "os/exec"

func DetectVBoxManageCmd() (string, bool) {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path, true
	}
	return cmd, false
}
