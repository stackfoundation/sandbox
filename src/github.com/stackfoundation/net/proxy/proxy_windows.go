package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func findHTTPProxy(proxies string) string {
	httpProxy := ""

	proxyList := strings.Split(proxies, ";")
	if len(proxyList) == 1 {
		return proxies
	}

	for _, proxy := range proxyList {
		typeSeparator := strings.Index(proxy, "=")
		if typeSeparator > 0 {
			if strings.TrimSpace(proxy[:typeSeparator]) == "http" {
				httpProxy = strings.TrimSpace(proxy[typeSeparator+1:])
			}
		}
	}

	return httpProxy
}

func areUpdatesOverridden(overrides string) bool {
	overridesList := strings.Split(overrides, ";")
	for _, override := range overridesList {
		pattern := strings.TrimSpace(override)
		if pattern == "stack.foundation" ||
			pattern == "updates.stack.foundation" ||
			pattern == "*.stack.foundation" {
			return true
		}
	}
	return false
}

func proxyFromRegistry() (string, error) {
	k, err := registry.OpenKey(
		registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	i, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil {
		return "", err
	}

	if i == 0 {
		return "", nil
	}

	s, _, err := k.GetStringValue("ProxyServer")
	if err != nil {
		return "", err
	}

	httpProxy := findHTTPProxy(s)
	if len(httpProxy) > 0 {
		s, _, err = k.GetStringValue("ProxyOverride")
		if err == nil {
			if areUpdatesOverridden(s) {
				return "", nil
			}
		}
	}

	return httpProxy, nil
}

func systemProxy(request *http.Request) (*url.URL, error) {
	httpProxy, err := proxyFromRegistry()
	if err != nil {
		return nil, err
	}

	return parseProxyURL(httpProxy)
}
