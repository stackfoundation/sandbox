package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/stackfoundation/log"
)

// ProxyArg Proxy provided as an argument
var ProxyArg string

// BypassArg Proxy bypass list provided as an argument
var BypassArg string

// ProxyCapableClient An HTTP client that uses the current system proxy settings
var ProxyCapableClient = &http.Client{
	Transport: &http.Transport{
		Proxy: platformProxy,
	},
}

// SystemSettings Get system proxy settings
func SystemSettings() (string, string) {
	if len(ProxyArg) > 0 {
		return ProxyArg, BypassArg
	}

	proxy, overrides, _ := proxyFromSystem()
	return proxy, overrides
}

func parseProxyURL(proxyURL string) (*url.URL, error) {
	if len(proxyURL) == 0 {
		return nil, nil
	}

	if !strings.HasPrefix(proxyURL, "http://") {
		proxyURL = "http://" + proxyURL
	}

	return url.Parse(proxyURL)
}

var cachedProxy *url.URL
var cachedOverrides string
var proxyWasCached bool

func cachingProxy(request *http.Request) (*url.URL, error) {
	if !proxyWasCached {
		if len(ProxyArg) > 0 {
			cachedProxy, _ = parseProxyURL(ProxyArg)
			cachedOverrides = BypassArg
		} else {
			httpProxy, proxyOverrides, err := proxyFromSystem()
			if err != nil {
				log.Debug("ProxyError", "Error while detecting system proxy settings: %v", err.Error())
			} else {
				cachedProxy, err = parseProxyURL(httpProxy)
				cachedOverrides = proxyOverrides
			}
		}

		if cachedProxy == nil {
			log.Debug("NoProxy", "No proxy is set")
		} else {
			log.Debug("Proxy", "Using %v as the proxy server", cachedProxy.String())
		}

		proxyWasCached = true
	}

	if useProxy(cachedOverrides, canonicalAddr(request.URL)) {
		return cachedProxy, nil
	}

	return nil, nil
}
