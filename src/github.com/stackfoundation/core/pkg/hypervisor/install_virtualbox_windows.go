package hypervisor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/stackfoundation/install"
)

func selectInstaller(installers []os.FileInfo, use64Bit bool) string {
	for _, installer := range installers {
		name := installer.Name()
		if !installer.IsDir() && strings.Contains(name, ".msi") {
			if use64Bit {
				if strings.Contains(name, "64") {
					return name
				}
			} else {
				return name
			}
		}
	}

	return ""
}

func selectTargetArchInstaller(installersFolder string) (string, error) {
	installers, err := ioutil.ReadDir(installersFolder)
	if err != nil {
		return "", err
	}

	use64Bit := false
	if runtime.GOARCH == "amd64" {
		use64Bit = true
	}

	installer := selectInstaller(installers, use64Bit)
	return filepath.Join(installersFolder, installer), nil
}

func installVirtualBoxWithInstaller(installer string) error {
	installPath, err := install.GetInstallPath()
	if err != nil {
		return err
	}

	installersFolder := filepath.Join(installPath, "VirtualBoxInstallers")

	cmd := exec.Command(installer, "-silent", "-extract", "-path", installersFolder)
	err = cmd.Run()
	if err != nil {
		return err
	}

	targetArchInstaller, err := selectTargetArchInstaller(installersFolder)
	if err != nil {
		return err
	}

	installLog := filepath.Join(installPath, "VirtualBoxInstall.log")

	cmd = exec.Command("msiexec", "/i", targetArchInstaller, "/qn", "/Lime!", installLog)
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()

	return err
}
