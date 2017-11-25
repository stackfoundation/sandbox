package hypervisor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stackfoundation/sandbox/log"

	"golang.org/x/sys/windows/registry"
)

func DetectVBoxManageCmd() (string, bool) {
	cmd := "VBoxManage"
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path, true
		}
	}

	if p := os.Getenv("VBOX_MSI_INSTALL_PATH"); p != "" {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path, true
		}
	}

	// Look in default installation path for VirtualBox version > 5
	if path, err := exec.LookPath(filepath.Join("C:\\Program Files\\Oracle\\VirtualBox", cmd)); err == nil {
		return path, true
	}

	// Look in windows registry
	if p, err := findVBoxInstallDirInRegistry(); err == nil {
		if path, err := exec.LookPath(filepath.Join(p, cmd)); err == nil {
			return path, true
		}
	}

	if path, err := exec.LookPath(cmd); err == nil {
		return path, true
	}
	return cmd, false
}

func findVBoxInstallDirInRegistry() (string, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Oracle\VirtualBox`, registry.QUERY_VALUE)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find VirtualBox registry entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", errors.New(errorMessage)
	}

	defer registryKey.Close()

	installDir, _, err := registryKey.GetStringValue("InstallDir")
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find InstallDir registry key within VirtualBox registries entries, is VirtualBox really installed properly? %s", err)
		log.Debugf(errorMessage)
		return "", errors.New(errorMessage)
	}

	return installDir, nil
}
