// +build !windows

package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/copy_unix.go
// Licensed under the Apache License Version 2.0

import (
        "path/filepath"
)

func normalizePath(path string) string {
        return filepath.ToSlash(path)
}