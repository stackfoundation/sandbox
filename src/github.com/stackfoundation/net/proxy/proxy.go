package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/stackfoundation/log"
)

// ProxyCapableClient An HTTP client that uses the current system proxy settings
var ProxyCapableClient = &http.Client{
	Transport: &http.Transport{
		Proxy: cachedSystemProxy,
	},
}

var cachedProxy *url.URL
var proxyWasCached bool

func cachedSystemProxy(request *http.Request) (*url.URL, error) {
	if proxyWasCached {
		return cachedProxy, nil
	}

	cachedProxy, err := systemProxy(request)
	proxyWasCached = true

	if err != nil {
		log.Debug("ProxyError", "Error while detecting system proxy settings: %v", err.Error())
	}

	if cachedProxy == nil {
		log.Debug("NoProxy", "No proxy is set")
	} else {
		log.Debug("Proxy", "Using %v as the proxy server", cachedProxy.String())
	}

	return cachedProxy, err
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
