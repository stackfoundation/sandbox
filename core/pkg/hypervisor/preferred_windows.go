package hypervisor

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/stackfoundation/sandbox/process"
)

func parseLines(stdout string) []string {
	resp := []string{}

	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		resp = append(resp, s.Text())
	}

	return resp
}

func hypervAvailable() bool {
	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		return false
	}

	out, err := process.CommandOut(powershell, "-NoProfile", "-NonInteractive", "@(Get-Command Get-VM).ModuleName")
	if err != nil {
		return false
	}

	resp := parseLines(out)
	if len(resp) < 1 || resp[0] != "Hyper-V" {
		return false
	}

	return true
}

func platformPreferred() string {
	if hypervAvailable() {
		return "hyperv"
	}

	return "virtualbox"
}
