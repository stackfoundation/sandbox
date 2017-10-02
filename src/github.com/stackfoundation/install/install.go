package install

import (
	"path/filepath"
)

// GetInstallPath Get the install path for CLI components
func GetInstallPath() (string, error) {
	path, err := getStackFoundationRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, "cli"), nil
}
