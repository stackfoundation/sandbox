/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/constants"
	"github.com/stackfoundation/sandbox/core/pkg/version"
	"golang.org/x/crypto/ssh/terminal"
)

type ServiceContext struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

type Message struct {
	Message        string `json:"message"`
	ServiceContext `json:"serviceContext"`
}

type LookPath func(filename string) (string, error)

var lookPath LookPath

func init() {
	lookPath = exec.LookPath
}

func ReportError(err error, url string) error {
	errMsg, err := FormatError(err)
	if err != nil {
		return errors.Wrap(err, "Error formatting error message")
	}
	jsonErrorMsg, err := MarshallError(errMsg, "default", version.GetVersion())
	if err != nil {
		return errors.Wrap(err, "Error marshalling error message to JSON")
	}
	err = UploadError(jsonErrorMsg, url)
	if err != nil {
		return errors.Wrap(err, "Error uploading error message")
	}
	return nil
}

func FormatError(err error) (string, error) {
	if err == nil {
		return "", errors.New("Error: ReportError was called with nil error value")
	}

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	errOutput := []string{}
	errOutput = append(errOutput, err.Error())

	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			errOutput = append(errOutput, fmt.Sprintf("\tat %n(%v)", f, f))
		}
	} else {
		return "", errors.New("Error msg with no stack trace cannot be reported")
	}
	return strings.Join(errOutput, "\n"), nil
}

func MarshallError(errMsg, service, version string) ([]byte, error) {
	m := Message{errMsg, ServiceContext{service, version}}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return b, nil
}

func UploadError(b []byte, url string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "")
	} else if resp.StatusCode != 200 {
		return errors.Errorf("Error sending error report to %s, got response code %d", url, resp.StatusCode)
	}
	return nil
}

func MaybeReportErrorAndExit(errToReport error) {
	os.Exit(1)
}

func getInput(input chan string, r io.Reader) {
	reader := bufio.NewReader(r)
	fmt.Print("Please enter your response [Y/n]: \n")
	response, err := reader.ReadString('\n')
	if err != nil {
		glog.Errorf(err.Error())
	}
	input <- response
}

func PromptUserForAccept(r io.Reader) bool {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	input := make(chan string, 1)
	go getInput(input, r)
	select {
	case response := <-input:
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		} else {
			fmt.Println("Invalid response, error reporting remains disabled. Must be in form [Y/n]")
			return false
		}
	case <-time.After(30 * time.Second):
		return false
	}
}

func GetKubeConfigPath() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	return filepath.SplitList(kubeConfigEnv)[0]
}

// Ask the kernel for a free open port that is ready to use
func GetPort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", errors.Errorf("Error accessing port %d", addr.Port)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

func KillMountProcess() error {
	out, err := ioutil.ReadFile(filepath.Join(constants.GetMinipath(), constants.MountProcessFileName))
	if err != nil {
		return nil // no mount process to kill
	}
	pid, err := strconv.Atoi(string(out))
	if err != nil {
		return errors.Wrap(err, "error converting mount string to pid")
	}
	mountProc, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "error converting mount string to pid")
	}
	return mountProc.Kill()
}
