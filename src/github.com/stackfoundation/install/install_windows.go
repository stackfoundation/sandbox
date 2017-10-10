package install

import (
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	csidlAppData          = 26
	errorFileNotFound     = 2
	errorPathNotFound     = 3
	errorBadFormat        = 11
	seErrAccessDenied     = 5
	seErrOOM              = 8
	seErrDLLNotFound      = 32
	seErrShare            = 26
	seErrAssocIncomplete  = 27
	seErrDDETimeout       = 28
	seErrDDEFail          = 29
	seErrDDEBusy          = 30
	seErrNoAssoc          = 31
	seeMaskNoAsync        = 0x00000100
	seeMaskNoCloseProcess = 0x00000040
)

var shell32 = syscall.NewLazyDLL("shell32.dll")
var shellExecuteEx = shell32.NewProc("ShellExecuteExW")
var getFolderPath = shell32.NewProc("SHGetFolderPathW")

type shellExecuteInfo struct {
	cbSize       uintptr
	fMask        uint32
	hwnd         uintptr
	lpVerb       uintptr
	lpFile       uintptr
	lpParameters uintptr
	lpDirectory  uintptr
	nShow        uintptr
	hInstApp     uintptr
	lpIDList     uintptr
	lpClass      uintptr
	hkeyClass    uintptr
	dwHotKey     uint32
	hIcon        uintptr
	hProcess     uintptr
}

func getRoamingAppDataDir() (string, error) {
	appDataPath := make([]uint16, syscall.MAX_PATH)

	result, _, err := getFolderPath.Call(0, csidlAppData, 0, 0, uintptr(unsafe.Pointer(&appDataPath[0])))
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

// ElevatedExecute Executes a shell command, requesting elevated privileges
func ElevatedExecute(binary, parameters string) error {
	var info shellExecuteInfo
	var errorCode int

	info.cbSize = unsafe.Sizeof(info)
	info.fMask = seeMaskNoCloseProcess
	info.lpVerb = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("runas")))
	info.lpFile = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(binary)))
	if len(parameters) != 0 {
		info.lpParameters = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(parameters)))
	}
	info.hInstApp = uintptr(unsafe.Pointer(&errorCode))

	_, _, _ = shellExecuteEx.Call(uintptr(unsafe.Pointer(&info)))
	s, e := syscall.WaitForSingleObject(syscall.Handle(info.hProcess), syscall.INFINITE)

	errorMsg := ""
	if errorCode != 0 && errorCode <= 32 {
		switch int(errorCode) {
		case errorFileNotFound:
			errorMsg = "The specified file was not found."
		case errorPathNotFound:
			errorMsg = "The specified path was not found."
		case errorBadFormat:
			errorMsg = "The .exe file is invalid (non-Win32 .exe or error in .exe image)."
		case seErrAccessDenied:
			errorMsg = "The operating system denied access to the specified file."
		case seErrAssocIncomplete:
			errorMsg = "The file name association is incomplete or invalid."
		case seErrDDEBusy:
			errorMsg = "The DDE transaction could not be completed because other DDE transactions were being processed."
		case seErrDDEFail:
			errorMsg = "The DDE transaction failed."
		case seErrDDETimeout:
			errorMsg = "The DDE transaction could not be completed because the request timed out."
		case seErrDLLNotFound:
			errorMsg = "The specified DLL was not found."
		case seErrNoAssoc:
			errorMsg = "There is no application associated with the given file name extension. This error will also be returned if you attempt to print a file that is not printable."
		case seErrOOM:
			errorMsg = "There was not enough memory to complete the operation."
		case seErrShare:
			errorMsg = "A sharing violation occurred."
		default:
			errorMsg = fmt.Sprintf("Unknown error occurred with error code %v", errorCode)
		}

		return errors.New(errorMsg)
	}

	if s == syscall.WAIT_FAILED {
		return e
	} else if s != syscall.WAIT_OBJECT_0 {
		return errors.New("Unexpected result while waiting for process to finish")
	}

	return nil
}
