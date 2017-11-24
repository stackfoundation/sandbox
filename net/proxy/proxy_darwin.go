package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/stackfoundation/process"
)

const ProxyOverrideSeparator = ","

func proxyFromSystem() (string, string, error) {
	cmd := `scutil --proxy | awk ' /HTTPEnable/ { enabled = $3; } /HTTPProxy/ { server = $3; } /HTTPPort/ { port = $3; } END {  if (enabled == "1") { print "http://" server ":" port; }  }'`

	pathOut, err := process.CommandOut("/bin/sh", "-c", cmd)
	overrides := ""

	if err == nil {
		cmd = `scutil --proxy | awk ' BEGIN {exceptionsList = 0; exceptionCount = 0; exceptionsStr= ""} /\}/ { exceptionsList = 0; } {if (exceptionsList) {exceptions[exceptionCount] = $3; ++exceptionCount;}} /ExceptionsList/ { exceptionsList = 1; } END {  for (i = 0; i < exceptionCount - 1; ++i) { exceptionsStr = exceptionsStr exceptions[i] ","; } exceptionsStr = exceptionsStr exceptions[exceptionCount - 1]; print exceptionsStr }'`

		overrides, err = process.CommandOut("/bin/sh", "-c", cmd)
	}

	return strings.TrimSpace(pathOut), strings.TrimSpace(overrides), err
}

func platformProxy(request *http.Request) (*url.URL, error) {
	return cachingProxy(request)
}
