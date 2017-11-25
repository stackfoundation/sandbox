package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const ProxyOverrideSeparator = ";"

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

func proxyFromSystem() (string, string, error) {
	k, err := registry.OpenKey(
		registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", "", err
	}
	defer k.Close()

	i, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil {
		return "", "", err
	}

	if i == 0 {
		return "", "", nil
	}

	proxy, _, err := k.GetStringValue("ProxyServer")
	if err != nil {
		return "", "", err
	}

	httpProxy := findHTTPProxy(proxy)
	overrides, _, err := k.GetStringValue("ProxyOverride")

	return httpProxy, overrides, nil
}

func platformProxy(request *http.Request) (*url.URL, error) {
	return cachingProxy(request)
}
