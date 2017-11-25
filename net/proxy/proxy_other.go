// +build !windows
// +build !darwin

package proxy

// Portions copied from:
// (under LICENSE: https://github.com/golang/go/blob/d0c1888739ac0d5d0c9f82a4b86945c0351caef6/LICENSE)
// https://github.com/golang/go/blob/d0c1888739ac0d5d0c9f82a4b86945c0351caef6/src/net/http/transport.go

import (
	"net/http"
	"net/url"
	"os"
	"sync"
)

var (
	httpProxyEnv = &envOnce{
		names: []string{"HTTP_PROXY", "http_proxy"},
	}
	httpsProxyEnv = &envOnce{
		names: []string{"HTTPS_PROXY", "https_proxy"},
	}
	noProxyEnv = &envOnce{
		names: []string{"NO_PROXY", "no_proxy"},
	}
)

type envOnce struct {
	names []string
	once  sync.Once
	val   string
}

func (e *envOnce) Get() string {
	e.once.Do(e.init)
	return e.val
}

func (e *envOnce) init() {
	for _, n := range e.names {
		e.val = os.Getenv(n)
		if e.val != "" {
			return
		}
	}
}

const ProxyOverrideSeparator = ","

func platformProxy(request *http.Request) (*url.URL, error) {
	if len(ProxyArg) > 0 {
		return parseProxyURL(ProxyArg)
	}

	return http.ProxyFromEnvironment(request)
}

func proxyFromSystem() (string, string, error) {
	return httpProxyEnv.Get(), noProxyEnv.Get(), nil
}
