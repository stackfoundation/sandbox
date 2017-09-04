package bootstrap

import (
        "syscall"
        "unsafe"
        "path/filepath"
)

const CSIDL_APPDATA = 26
var shell32 = syscall.NewLazyDLL("shell32.dll")
var getFolderPath = shell32.NewProc("SHGetFolderPathW")

func getRoamingAppDataDir() (string, error) {
        appDataPath := make([]uint16, syscall.MAX_PATH)

        result, _, err := getFolderPath.Call(0, CSIDL_APPDATA, 0, 0, uintptr(unsafe.Pointer(&appDataPath[0])))
        if uint32(result) != 0 {
                return "", err
        }

        return syscall.UTF16ToString(appDataPath), nil
}

func getStackFoundationRoot() (string, error) {
        path, err := getRoamingAppDataDir()
        if err != nil {
                return "", err
        }

        return filepath.Join(path, "sf"), nil
}
