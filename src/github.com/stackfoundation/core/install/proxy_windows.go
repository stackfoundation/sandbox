package bootstrap

import (
        "golang.org/x/sys/windows/registry"
)

func proxyFromRegistry() (string, error) {
        k, err := registry.OpenKey(
                registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
                registry.QUERY_VALUE)
        if err != nil {
                return "", err
        }
        defer k.Close()

        v, _, err := k.GetIntegerValue("ProxyEnable")
        if err != nil {
                return "", err
        }

        if (v == 0) {
                return "no proxy", nil
        } else {
                return "proxy present", nil
        }
}

func ProxyFromSystem() (string, error) {
        return proxyFromRegistry()
}
