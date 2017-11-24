// Portions copied from:
// (under LICENSE: https://github.com/golang/go/blob/d0c1888739ac0d5d0c9f82a4b86945c0351caef6/LICENSE)
// https://github.com/golang/go/blob/d0c1888739ac0d5d0c9f82a4b86945c0351caef6/src/net/http/transport.go
// https://github.com/golang/go/blob/d0c1888739ac0d5d0c9f82a4b86945c0351caef6/src/net/http/http.go

package proxy

import (
	"net"
	"net/url"
	"strings"
)

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func canonicalAddr(url *url.URL) string {
	addr := url.Host
	if !hasPort(addr) {
		return addr + ":" + portMap[url.Scheme]
	}
	return addr
}

func useProxy(proxyOverrides, addr string) bool {
	if len(addr) == 0 {
		return true
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	if proxyOverrides == "*" {
		return false
	}

	addr = strings.ToLower(strings.TrimSpace(addr))
	if hasPort(addr) {
		addr = addr[:strings.LastIndex(addr, ":")]
	}

	for _, p := range strings.Split(proxyOverrides, ProxyOverrideSeparator) {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) == 0 {
			continue
		}
		if hasPort(p) {
			p = p[:strings.LastIndex(p, ":")]
		}
		if addr == p {
			return false
		}
		if p[0] == '.' && (strings.HasSuffix(addr, p) || addr == p[1:]) {
			// proxyOverrides ".foo.com" matches "bar.foo.com" or "foo.com"
			return false
		}
		if p[0] != '.' && strings.HasSuffix(addr, p) && addr[len(addr)-len(p)-1] == '.' {
			// proxyOverrides "foo.com" matches "bar.foo.com"
			return false
		}
	}
	return true
}
