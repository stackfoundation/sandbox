// +build !linux
package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/time_unsupported.go
// Licensed under the Apache License Version 2.0

import (
        "syscall"
        "time"
)

func timeToTimespec(time time.Time) (ts syscall.Timespec) {
        nsec := int64(0)
        if !time.IsZero() {
                nsec = time.UnixNano()
        }
        return syscall.NsecToTimespec(nsec)
}