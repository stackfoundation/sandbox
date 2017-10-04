package proxy

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
)

const ProxyOverrideSeparator = ","

func proxyFromSystem() (string, string, error) {
	pathCmd := exec.Command("scutil --proxy | awk '/HTTPEnable/ { enabled = $3; } /HTTPProxy/ { server = $3; } /HTTPPort/ { port = $3; } END { if (enabled == \"1\") { print \"http://\" server \":\" port; } }'")
	pathOut, _ := pathCmd.StdoutPipe()
	pathCmd.Start()
	pathBytes, _ := ioutil.ReadAll(pathOut)
	pathCmd.Wait()

	overrides := ""

	return string(pathBytes), overrides, nil
}

func platformProxy(request *http.Request) (*url.URL, error) {
	return cachingProxy(request)
}
