package process

import (
	"os"
	"os/exec"
	"strings"

	"github.com/stackfoundation/sandbox/log"
)

// CommandOut Execute a process, wait till completion, and return the stdout
func CommandOut(command string, args ...string) (string, error) {
	log.Debugf("Executing %v %v", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	out, err := cmd.Output()

	if out == nil {
		return "", err
	}

	if err != nil {
		log.Debug("Error executing %v %v: %v", command, strings.Join(args, " "), err.Error())
	}

	return string(out), err
}

// CombineStdStreams Execute a process, wait till completion, and combine std streams to this process
func CombineStdStreams(command string, args ...string) error {
	log.Debugf("Executing %v %v", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		log.Debugf("Error executing %v %v: %v", command, strings.Join(args, " "), err.Error())
	}

	return err
}

// CommandWithoutOut Execute a process, wait till completion, and don't ouput process's streams
func CommandWithoutOut(command string, args ...string) error {
	log.Debugf("Executing %v %v", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	err := cmd.Run()

	if err != nil {
		log.Debugf("Error executing %v %v: %v", command, strings.Join(args, " "), err.Error())
	}

	return err
}
