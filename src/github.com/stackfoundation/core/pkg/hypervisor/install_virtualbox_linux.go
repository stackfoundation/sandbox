package hypervisor

import "strings"
import "github.com/stackfoundation/process"

func installVirtualBox(installer string) error {
	var script string
	distro := distroCode()
	if strings.Contains(distro, "fedora") {
		script = "dnf install " + installer
	} else if strings.Contains(distro, "el") {
		script = "yum install " + installer
	} else if distro == "suse" {
		script = "yast2 -i " + installer
	} else if distro == "linux" {
		script = installer
	} else {
		script = "dpkg -i " + installer + "; apt-get update; apt-get install -f ."
	}

	err := process.CombineStdStreams("/bin/sh", "-c", script)
	return err
}
