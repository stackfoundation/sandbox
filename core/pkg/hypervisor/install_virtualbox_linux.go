package hypervisor

import (
	"strings"

	"github.com/stackfoundation/log"
	"github.com/stackfoundation/process"
)

func installVirtualBoxWithInstaller(installer string) error {
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
		script = "dpkg -i " + installer + "; apt-get update; apt-get -f install"
	}

	var err error
	if log.IsDebug() {
		err = process.CommandWithoutOut("bin/sh", "-c", script)
	} else {
		err = process.CombineStdStreams("/bin/sh", "-c", script)
	}

	return err
}
