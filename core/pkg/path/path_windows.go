package path

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var shell32 = syscall.NewLazyDLL("shell32.dll")
var user32 = syscall.NewLazyDLL("user32.dll")
var sendMessageTimeout = user32.NewProc("SendMessageTimeoutW")

const hwndBroadcast = 0xffff
const wmSettingChange = 0x001A
const smtoAbortIfHung = 0x0002

func notifySettingChange() uintptr {
	var result uint32
	ret, _, _ := sendMessageTimeout.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))),
		uintptr(smtoAbortIfHung),
		uintptr(5000),
		uintptr(unsafe.Pointer(&result)))

	return ret
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		path = path[:len(path)-1]
	}

	return path
}

func AddSboxToSystemPath(installDirectory string) error {
	return AddToSystemPath(installDirectory)
}

func IsInSystemPath(node string) (bool, error) {
	environmentVariables, err := registry.OpenKey(
		registry.CURRENT_USER,
		"Environment",
		registry.ALL_ACCESS)
	defer environmentVariables.Close()
	if err != nil {
		return false, err
	}

	pathVariable, _, err := environmentVariables.GetStringValue("Path")
	if err != nil {
		return false, err
	}

	directoryAlreadyOnPath := false
	node = normalizePath(node)

	paths := strings.Split(pathVariable, ";")
	for _, path := range paths {
		if node == normalizePath(path) {
			return true, nil
		}
	}

	return false, nil
}

// AddToSystemPath Add the specified node to the system PATH variable
func AddToSystemPath(node string) error {
	environmentVariables, err := registry.OpenKey(
		registry.CURRENT_USER,
		"Environment",
		registry.ALL_ACCESS)
	defer environmentVariables.Close()
	if err != nil {
		return err
	}

	pathVariable, _, err := environmentVariables.GetStringValue("Path")
	if err != nil {
		return err
	}

	directoryAlreadyOnPath := false
	node = normalizePath(node)

	paths := strings.Split(pathVariable, ";")
	for _, path := range paths {
		if node == normalizePath(path) {
			directoryAlreadyOnPath = true
			break
		}
	}

	if !directoryAlreadyOnPath {
		pathVariable = strings.TrimSpace(pathVariable)
		if strings.HasSuffix(pathVariable, ";") {
			pathVariable = pathVariable + node
		} else {
			pathVariable = pathVariable + ";" + node
		}

		environmentVariables.SetStringValue("Path", pathVariable)
		notifySettingChange()
	}

	return nil
}
