package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/time_linux.go
// Licensed under the Apache License Version 2.0

import (
        "syscall"
        "time"
)

func timeToTimespec(time time.Time) (ts syscall.Timespec) {
        if time.IsZero() {
                // Return UTIME_OMIT special value
                ts.Sec = 0
                ts.Nsec = ((1 << 30) - 2)
                return
        }
        return syscall.NsecToTimespec(time.UnixNano())
}