package install

import (
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/stackfoundation/sandbox/log"
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
var shellExecute = shell32.NewProc("ShellExecuteW")
var shellExecuteEx = shell32.NewProc("ShellExecuteExW")
var getFolderPath = shell32.NewProc("SHGetFolderPathW")

/*
typedef struct _SHELLEXECUTEINFO {
  DWORD     cbSize;
  ULONG     fMask;
  HWND      hwnd;
  LPCTSTR   lpVerb;
  LPCTSTR   lpFile;
  LPCTSTR   lpParameters;
  LPCTSTR   lpDirectory;
  int       nShow;
  HINSTANCE hInstApp;
  LPVOID    lpIDList;
  LPCTSTR   lpClass;
  HKEY      hkeyClass;
  DWORD     dwHotKey;
  union {
    HANDLE hIcon;
    HANDLE hMonitor;
  } DUMMYUNIONNAME;
  HANDLE    hProcess;
} SHELLEXECUTEINFO, *LPSHELLEXECUTEINFO;
*/
type shellExecuteInfo struct {
	cbSize       uint32
	fMask        uint32
	hwnd         unsafe.Pointer
	lpVerb       unsafe.Pointer
	lpFile       unsafe.Pointer
	lpParameters unsafe.Pointer
	lpDirectory  unsafe.Pointer
	nShow        int32
	hInstApp     unsafe.Pointer
	lpIDList     unsafe.Pointer
	lpClass      *uint16
	hkeyClass    unsafe.Pointer
	dwHotKey     uint32
	hIcon        unsafe.Pointer
	hProcess     unsafe.Pointer
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

func initializeExecuteInfo(info *shellExecuteInfo, binary, parameters string, errorCode *int) {
	info.fMask = seeMaskNoCloseProcess
	info.lpVerb = unsafe.Pointer(syscall.StringToUTF16Ptr("runas"))
	info.lpFile = unsafe.Pointer(syscall.StringToUTF16Ptr(binary))
	if len(parameters) != 0 {
		info.lpParameters = unsafe.Pointer(syscall.StringToUTF16Ptr(parameters))
	}

	info.hInstApp = unsafe.Pointer(errorCode)
	info.cbSize = uint32(unsafe.Sizeof(*info))
}

// ElevatedExecute Executes a shell command, requesting elevated privileges
func ElevatedExecute(binary, parameters string) error {
	var info shellExecuteInfo
	var errorCode int

	initializeExecuteInfo(&info, binary, parameters, &errorCode)

	log.Debugf("Executing %v %v with elevation", binary, parameters)
	r, _, err := shellExecuteEx.Call(uintptr(unsafe.Pointer(&info)))
	s, e := syscall.WaitForSingleObject(syscall.Handle(info.hProcess), syscall.INFINITE)

	if r == 0 && err != nil {
		return err
	}

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
