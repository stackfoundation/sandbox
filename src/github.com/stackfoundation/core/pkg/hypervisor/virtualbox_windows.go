package hypervisor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/stackfoundation/core/pkg/io"
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

func InstallVirtualBox() error {
	installPath, err := install.GetInstallPath()
	if err != nil {
		return err
	}

	virtualBoxInstaller := filepath.Join(installPath, "VirtualBoxInstall.exe")
	if !io.MD5SumEquals(virtualBoxInstaller, md5Windows) {
		// err = download.DownloadFromURL(packageWindows, virtualBoxInstaller, "VirtualBoxDownload")
		// if err != nil {
		// 	return err
		// }
	}

	installersFolder := filepath.Join(installPath, "VirtualBoxInstallers")

	cmd := exec.Command(virtualBoxInstaller, "-silent", "-extract", "-path", installersFolder)
	err = cmd.Run()
	if err != nil {
		return err
	}

	installer, err := selectTargetArchInstaller(installersFolder)
	if err != nil {
		return err
	}

	installLog := filepath.Join(installPath, "VirtualBoxInstall.log")

	cmd = exec.Command("msiexec", "/i", installer, "/qn", "/Lime!", installLog)
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()

	return err
}

const md5OSX = "620b3bdf96b7afb9de56e2742d373568"
const md5Windows = "935f8590faac3f60c8b61abd4f27d0c7"
