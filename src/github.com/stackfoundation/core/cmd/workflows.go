package cmd

import (
        "path/filepath"
        "os"
        "io/ioutil"
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

func DeleteWorkflow(workflow string) (bool, error) {
        workflowsDirectory, err := getWorkflowsDirectory()
        if err != nil {
                return false, err
        }

        workflowsDirectoryExists, err := directoryExists(workflowsDirectory)
        if err != nil || !workflowsDirectoryExists {
                return false, err
        }

        workflowFile := filepath.Join(workflowsDirectory, workflow + workflowExtension)

        workflowFileExists, err := fileExists(workflowFile)
        if err != nil || !workflowFileExists {
                return false, err
        }

        err = os.Remove(workflowFile)
        if err != nil {
                return false, err
        }

        return true, nil
}

func ListWorkflows() ([]string, error) {
        workflowsDirectory, err := getWorkflowsDirectory()
        if err != nil {
                return []string{}, nil
        }

        workflows, err := ioutil.ReadDir(workflowsDirectory)
        if err != nil {
                return []string{}, nil
        }

        var workflowNames []string
        for _, workflow := range workflows {
                if !workflow.IsDir() && filepath.Ext(workflow.Name()) == workflowExtension {
                        workflowNames = append(workflowNames, strings.TrimSuffix(workflow.Name(), workflowExtension))
                }
        }

        return workflowNames, nil
}
