package hypervisor

import "os/exec"

func DetectVBoxManageCmd() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}

const distroScript = `#!/bin/sh
if [ -f /etc/os-release ]; then
    # freedesktop.org and systemd
    . /etc/os-release
    OS=$NAME
    VER="$VERSION_ID $VERSION"
	CODE=$VERSION_CODENAME
elif type lsb_release >/dev/null 2>&1; then
    # linuxbase.org
    OS=$(lsb_release -si)
    VER=$(lsb_release -sr)
	CODE=$(lsb_release -sc)
elif [ -f /etc/lsb-release ]; then
    # For some versions of Debian/Ubuntu without lsb_release command
    . /etc/lsb-release
    OS=$DISTRIB_ID
    VER=$DISTRIB_RELEASE
	CODE=$DISTRIB_CODENAME
elif [ -f /etc/debian_version ]; then
    # Older Debian/Ubuntu/etc.
    OS=Debian
    VER=$(cat /etc/debian_version | tr "\n" ' ')
elif [ -f /etc/SuSe-release ]; then
    # Older SuSE/etc.
	OS='SUSE'
	VER=``cat /etc/SUSE-release | tr "\n" ' '``
elif [ -f /etc/redhat-release ]; then
    # Older Red Hat, CentOS, etc.
	OS='RedHat'
	VER=``cat /etc/redhat-release | tr "\n" ' '``
else
    # Fall back to uname, e.g. "Linux <version>", also works for BSD, etc.
    OS=$(uname -s)
    VER=$(uname -r)
fi

echo "${OS} ${VER} ${CODE}"`