package bootstrap

import (
	"io/ioutil"
	"os/exec"
)

func ProxyFromSystem() (string, error) {
	pathCmd := exec.Command("scutil --proxy | awk '/HTTPEnable/ { enabled = $3; } /HTTPProxy/ { server = $3; } /HTTPPort/ { port = $3; } END { if (enabled == \"1\") { print \"http://\" server \":\" port; } }'")
	pathOut, _ := pathCmd.StdoutPipe()
	pathCmd.Start()
	pathBytes, _ := ioutil.ReadAll(pathOut)
	pathCmd.Wait()

	return string(pathBytes), nil
}
