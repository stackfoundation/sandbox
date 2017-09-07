package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const workflowExtension = ".wflow"

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

func getSandboxDirectory() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, ".sbox"), nil
}

func getAlternativeWorkflowsDirectory() (string, error) {
	sboxDirectory, err := getSandboxDirectory()
	if err != nil {
		return "", err
	}

	sboxDirectoryExists, err := directoryExists(sboxDirectory)
	if err != nil || !sboxDirectoryExists {
		return "", err
	}

	alternativeWorkflowsConfigFile := filepath.Join(sboxDirectory, "workflows")
	alternativeWorkflowsConfigFileExists, err := fileExists(alternativeWorkflowsConfigFile)
	if err != nil || !alternativeWorkflowsConfigFileExists {
		return "", err
	}

	alternativeWorkflowsDirectory, err := ioutil.ReadFile(alternativeWorkflowsConfigFile)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(alternativeWorkflowsDirectory)), nil
}

func getWorkflowsDirectory() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	alternativeWorkflowsDirectory, err := getAlternativeWorkflowsDirectory()
	if err == nil && len(alternativeWorkflowsDirectory) > 0 {
		return filepath.Join(path, alternativeWorkflowsDirectory), nil
	}

	return filepath.Join(path, "workflows"), nil
}
