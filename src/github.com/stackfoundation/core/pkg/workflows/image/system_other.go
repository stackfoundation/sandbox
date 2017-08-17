// +build !windows

package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/system/filesys.go
// Licensed under the Apache License Version 2.0

import "os"

// OpenSequential opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func OpenSequential(name string) (*os.File, error) {
        return os.Open(name)
}

// OpenFileSequential is the generalized open call; most users will use Open
// or Create instead. It opens the named file with specified flag
// (O_RDONLY etc.) and perm, (0666 etc.) if applicable. If successful,
// methods on the returned File can be used for I/O.
// If there is an error, it will be of type *PathError.
func OpenFileSequential(name string, flag int, perm os.FileMode) (*os.File, error) {
        return os.OpenFile(name, flag, perm)
}