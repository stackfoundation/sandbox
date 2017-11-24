package hypervisor

import (
	"os"
	"path"
	"path/filepath"

	"github.com/stackfoundation/sandbox/install"
	"github.com/stackfoundation/sandbox/log"
)

func virtualBoxInstallerPath() (string, error) {
	installPath, err := install.GetInstallPath()
	if err != nil {
		return "", err
	}

	pkg, _ := platformVirtualBoxPackage()
	extension := path.Ext(pkg)
	return filepath.Join(installPath, "VirtualBoxInstall"+extension), err
}

// InstallVirtualBox Perform a VirtualBox install
func InstallVirtualBox(fail bool) error {
	installerPath, err := virtualBoxInstallerPath()
	if err != nil {
		return err
	}

	err = installVirtualBoxWithInstaller(installerPath)
	if err != nil {
		if fail {
			return err
		}

		command := "virtualbox --fail"
		if log.IsDebug() {
			command += " --debug"
		}

		err = install.ElevatedExecute(os.Args[0], command)
	}

	return err
}
