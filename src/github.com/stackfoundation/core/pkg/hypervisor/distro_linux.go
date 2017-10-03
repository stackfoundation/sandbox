package hypervisor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/stackfoundation/install"
	"github.com/stackfoundation/process"
)

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
	VER=` + "`cat /etc/SUSE-release | tr \"\n\" ' '`" + `
elif [ -f /etc/redhat-release ]; then
    # Older Red Hat, CentOS, etc.
	OS='RedHat'
	VER=` + "`cat /etc/redhat-release | tr \"\n\" ' '`" + `
else
    # Fall back to uname, e.g. "Linux <version>", also works for BSD, etc.
    OS=$(uname -s)
    VER=$(uname -r)
fi

echo "${OS} ${VER} ${CODE}"`

func distroCode() string {
	installPath, err := install.GetInstallPath()
	if err != nil {
		return "generic"
	}

	os.MkdirAll(installPath, 0777)
	distroScriptPath := filepath.Join(installPath, "distro.sh")
	err = ioutil.WriteFile(distroScriptPath, []byte(distroScript), 4555)
	if err != nil {
		return "generic"
	}

	out, err := process.CommandOut(distroScriptPath)

	out = strings.ToLower(out)

	if strings.Contains(out, "jessie") {
		return "jessie"
	} else if strings.Contains(out, "stretch") {
		return "stretch"
	} else if strings.Contains(out, "wheezy") {
		return "wheezy"
	} else if strings.Contains(out, "precise") {
		return "precise"
	} else if strings.Contains(out, "trusty") {
		return "trusty"
	} else if strings.Contains(out, "wily") {
		return "wily"
	} else if strings.Contains(out, "xenial") {
		return "xenial"
	} else if strings.Contains(out, "yakkety") {
		return "yakkety"
	} else if strings.Contains(out, "zesty") {
		return "zesty"
	} else if strings.Contains(out, "redhat") {
		if strings.Contains(out, "5.") {
			return "el5"
		} else if strings.Contains(out, "6.") {
			return "el5"
		} else if strings.Contains(out, "7") {
			return "el7"
		}
	} else if strings.Contains(out, "fedora") {
		if strings.Contains(out, "26") {
			return "fedora26"
		} else if strings.Contains(out, "25") {
			return "fedora25"
		} else if strings.Contains(out, "24") {
			return "fedora24"
		} else if strings.Contains(out, "23") {
			return "fedora23"
		} else if strings.Contains(out, "22") {
			return "fedora22"
		}
	}

	return "generic"
}
