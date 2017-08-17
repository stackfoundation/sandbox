// +build windows

package image

import (
        "os"
        "syscall"
        "golang.org/x/sys/windows"
        "unsafe"
)

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/system/filesys_windows.go
// Licensed under the Apache License Version 2.0


// OpenSequential opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func OpenSequential(name string) (*os.File, error) {
        return OpenFileSequential(name, os.O_RDONLY, 0)
}

// OpenFileSequential is the generalized open call; most users will use Open
// or Create instead.
// If there is an error, it will be of type *PathError.
func OpenFileSequential(name string, flag int, _ os.FileMode) (*os.File, error) {
        if name == "" {
                return nil, &os.PathError{Op: "open", Path: name, Err: syscall.ENOENT}
        }
        r, errf := windowsOpenFileSequential(name, flag, 0)
        if errf == nil {
                return r, nil
        }
        return nil, &os.PathError{Op: "open", Path: name, Err: errf}
}

func windowsOpenFileSequential(name string, flag int, _ os.FileMode) (file *os.File, err error) {
        r, e := windowsOpenSequential(name, flag|windows.O_CLOEXEC, 0)
        if e != nil {
                return nil, e
        }
        return os.NewFile(uintptr(r), name), nil
}

func makeInheritSa() *windows.SecurityAttributes {
        var sa windows.SecurityAttributes
        sa.Length = uint32(unsafe.Sizeof(sa))
        sa.InheritHandle = 1
        return &sa
}

func windowsOpenSequential(path string, mode int, _ uint32) (fd windows.Handle, err error) {
        if len(path) == 0 {
                return windows.InvalidHandle, windows.ERROR_FILE_NOT_FOUND
        }
        pathp, err := windows.UTF16PtrFromString(path)
        if err != nil {
                return windows.InvalidHandle, err
        }
        var access uint32
        switch mode & (windows.O_RDONLY | windows.O_WRONLY | windows.O_RDWR) {
        case windows.O_RDONLY:
                access = windows.GENERIC_READ
        case windows.O_WRONLY:
                access = windows.GENERIC_WRITE
        case windows.O_RDWR:
                access = windows.GENERIC_READ | windows.GENERIC_WRITE
        }
        if mode&windows.O_CREAT != 0 {
                access |= windows.GENERIC_WRITE
        }
        if mode&windows.O_APPEND != 0 {
                access &^= windows.GENERIC_WRITE
                access |= windows.FILE_APPEND_DATA
        }
        sharemode := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE)
        var sa *windows.SecurityAttributes
        if mode&windows.O_CLOEXEC == 0 {
                sa = makeInheritSa()
        }
        var createmode uint32
        switch {
        case mode&(windows.O_CREAT|windows.O_EXCL) == (windows.O_CREAT | windows.O_EXCL):
                createmode = windows.CREATE_NEW
        case mode&(windows.O_CREAT|windows.O_TRUNC) == (windows.O_CREAT | windows.O_TRUNC):
                createmode = windows.CREATE_ALWAYS
        case mode&windows.O_CREAT == windows.O_CREAT:
                createmode = windows.OPEN_ALWAYS
        case mode&windows.O_TRUNC == windows.O_TRUNC:
                createmode = windows.TRUNCATE_EXISTING
        default:
                createmode = windows.OPEN_EXISTING
        }
        // Use FILE_FLAG_SEQUENTIAL_SCAN rather than FILE_ATTRIBUTE_NORMAL as implemented in golang.
        //https://msdn.microsoft.com/en-us/library/windows/desktop/aa363858(v=vs.85).aspx
        const fileFlagSequentialScan = 0x08000000 // FILE_FLAG_SEQUENTIAL_SCAN
        h, e := windows.CreateFile(pathp, access, sharemode, sa, createmode, fileFlagSequentialScan, 0)
        return h, e
}
