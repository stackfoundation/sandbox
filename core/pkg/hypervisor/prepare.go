package hypervisor

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/stackfoundation/sandbox/log"
	"github.com/stackfoundation/sandbox/net/download"
	"github.com/stackfoundation/sandbox/process"

	"github.com/stackfoundation/sandbox/core/pkg/io"
	"github.com/stackfoundation/sandbox/install"
	"github.com/stackfoundation/sandbox/metadata"
)

func downloadVirtualBoxIfNecessary() error {
	installPath, err := install.GetInstallPath()
	if err != nil {
		return err
	}

	pkg, md5 := platformVirtualBoxPackage()
	extension := path.Ext(pkg)
	virtualBoxInstaller := filepath.Join(installPath, "VirtualBoxInstall"+extension)

	if !io.MD5SumEquals(virtualBoxInstaller, md5) {
		err = download.WithProgress("Downloading VirtualBox", pkg, virtualBoxInstaller, "VirtualBoxDownload")
		if err != nil {
			return err
		}
	}

	return nil
}

func relaunchForInstall(command string) error {
	if log.IsDebug() {
		return process.CombineStdStreams(os.Args[0], command, "--debug")
	}

	return process.CombineStdStreams(os.Args[0], command)
}

// SelectAndPrepareHypervisor Select the preferred hypervisor for the platform, and prepare it if necessary
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
			err := downloadVirtualBoxIfNecessary()
			if err != nil {
				fmt.Println("Error downloading VirtualBox: " + err.Error())
				fmt.Println("Continuing but further failures might occur")
			} else {
				fmt.Println("Installing Virtualbox (this may take some time)")
				err = relaunchForInstall("virtualbox")

				vboxManageCmd, found := DetectVBoxManageCmd()
				if found {
					vbox = vboxManageCmd
				}

				if err != nil {
					fmt.Println("Error installing VirtualBox: " + err.Error())
					fmt.Println("Continuing but further failures might occur")
				} else {
					fmt.Println("VirtualBox installed!")
				}
			}
		} else {
			vbox = vboxManageCmd
		}
	} else if preferred == "xhyve" {
		if !isxhyveInstalled() {
			fmt.Println("Installing xyhve driver")
			relaunchForInstall("xhyve")
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
