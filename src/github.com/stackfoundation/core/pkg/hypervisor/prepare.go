package hypervisor

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/stackfoundation/net/download"

	"github.com/stackfoundation/core/pkg/io"
	"github.com/stackfoundation/install"
	"github.com/stackfoundation/metadata"
)

func downloadVirtualBoxIfNecessary() (string, error) {
	installPath, err := install.GetInstallPath()
	if err != nil {
		return "", err
	}

	pkg, md5 := platformVirtualBoxPackage()
	extension := path.Ext(pkg)
	virtualBoxInstaller := filepath.Join(installPath, "VirtualBoxInstall"+extension)

	if !io.MD5SumEquals(virtualBoxInstaller, md5) {
		err = download.WithProgress("Downloading VirtualBox", pkg, virtualBoxInstaller, "VirtualBoxDownload")
		if err != nil {
			return "", err
		}
	}

	return virtualBoxInstaller, nil
}

func SelectAndPrepareHypervisor(preferred string) string {
	var m *metadata.Metadata
	var err error
	var vbox string

	if preferred == "auto" {
		m, err = metadata.GetMetadata()
		if err == nil && m != nil {
			preferred = m.Driver
		}
	}

	if preferred == "auto" || len(preferred) < 1 {
		preferred = platformPreferred()
	}

	if preferred == "virtualbox" {
		vboxManageCmd, found := DetectVBoxManageCmd()
		if !found {
			installer, err := downloadVirtualBoxIfNecessary()
			if err != nil {
				fmt.Println("Error downloading: " + err.Error())
			}

			fmt.Println("Installing Virtualbox")
			installVirtualBox(installer)
		} else {
			vbox = vboxManageCmd
		}
	}

	if m != nil &&
		(m.Driver != preferred ||
			(preferred == "virtualbox" && m.VirtualBox != vbox)) {
		m.Driver = preferred
		m.VirtualBox = vbox

		metadata.SaveMetadata(m)
	}

	return preferred
}
