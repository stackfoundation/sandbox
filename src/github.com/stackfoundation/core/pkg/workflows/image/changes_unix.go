// +build !windows

package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/changes_unix.go
// Licensed under the Apache License Version 2.0

import (
        "os"
        "syscall"
)

func hasHardlinks(fi os.FileInfo) bool {
        return fi.Sys().(*syscall.Stat_t).Nlink > 1
}

