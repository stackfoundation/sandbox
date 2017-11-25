package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/copy.go
// Licensed under the Apache License Version 2.0

import "path/filepath"

// specifiesCurrentDir returns whether the given path specifies
// a "current directory", i.e., the last path segment is `.`.
func specifiesCurrentDir(path string) bool {
        return filepath.Base(path) == "."
}

// SplitPathDirEntry splits the given path between its directory name and its
// basename by first cleaning the path but preserves a trailing "." if the
// original path specified the current directory.
func splitPathDirEntry(path string) (dir, base string) {
        cleanedPath := filepath.Clean(normalizePath(path))

        if specifiesCurrentDir(path) {
                cleanedPath += string(filepath.Separator) + "."
        }

        return filepath.Dir(cleanedPath), filepath.Base(cleanedPath)
}
