// +build !windows
// +build !darwin

package proxy

import (
	"net/http"
	"net/url"
)

const ProxyOverrideSeparator = ","

func platformProxy(request *http.Request) (*url.URL, error) {
	return http.ProxyFromEnvironment(request)
}
