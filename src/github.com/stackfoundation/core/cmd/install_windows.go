package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/registry"
)

var shell32 = syscall.NewLazyDLL("shell32.dll")
var user32 = syscall.NewLazyDLL("user32.dll")
var shellExecute = shell32.NewProc("ShellExecuteW")
var sendMessageTimeout = user32.NewProc("SendMessageTimeoutW")

func ShellExecute(file, parameters string) error {
	var verb, param, directory uintptr
	verb = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("runas")))
	if len(parameters) != 0 {
		param = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(parameters)))
	}

	ret, _, _ := shellExecute.Call(
		uintptr(0),
		verb,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(file))),
		param,
		directory,
		uintptr(0))

	errorMsg := ""
	if ret != 0 && ret <= 32 {
		switch int(ret) {
		case ERROR_FILE_NOT_FOUND:
			errorMsg = "The specified file was not found."
		case ERROR_PATH_NOT_FOUND:
			errorMsg = "The specified path was not found."
		case ERROR_BAD_FORMAT:
			errorMsg = "The .exe file is invalid (non-Win32 .exe or error in .exe image)."
		case SE_ERR_ACCESSDENIED:
			errorMsg = "The operating system denied access to the specified file."
		case SE_ERR_ASSOCINCOMPLETE:
			errorMsg = "The file name association is incomplete or invalid."
		case SE_ERR_DDEBUSY:
			errorMsg = "The DDE transaction could not be completed because other DDE transactions were being processed."
		case SE_ERR_DDEFAIL:
			errorMsg = "The DDE transaction failed."
		case SE_ERR_DDETIMEOUT:
			errorMsg = "The DDE transaction could not be completed because the request timed out."
		case SE_ERR_DLLNOTFOUND:
			errorMsg = "The specified DLL was not found."
		case SE_ERR_NOASSOC:
			errorMsg = "There is no application associated with the given file name extension. This error will also be returned if you attempt to print a file that is not printable."
		case SE_ERR_OOM:
			errorMsg = "There was not enough memory to complete the operation."
		case SE_ERR_SHARE:
			errorMsg = "A sharing violation occurred."
		default:
			errorMsg = fmt.Sprintf("Unknown error occurred with error code %v", ret)
		}
	} else {
		return nil
	}

	return errors.New(errorMsg)
}

func NotifySettingChange() uintptr {
	ret, _, _ := sendMessageTimeout.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000))

	return ret
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Sandbox command-line globally",
	Long: `Install the Sandbox command-line, and make it available globally.

This adds to the system PATH variable so that the Sandbox command-line is available globally.`,
	Run: func(command *cobra.Command, args []string) {
		//environmentVariables, err := registry.OpenKey(
		//        registry.LOCAL_MACHINE,
		//        "System\\CurrentControlSet\\Control\\Session Manager\\Environment",
		//        registry.ALL_ACCESS)
		environmentVariables, err := registry.OpenKey(
			registry.CURRENT_USER,
			"Environment",
			registry.ALL_ACCESS)
		defer environmentVariables.Close()
		if err != nil {
			//ShellExecute(os.Args[0], "install")
			panic(err)
		}

		info, err := environmentVariables.Stat()
		if err != nil {
			panic(err)
		}

		fmt.Println(info.ValueCount)

		pathVariable, _, err := environmentVariables.GetStringValue("Path")
		if err != nil {
			panic(err)
		}

		coreDirectory := filepath.Dir(os.Args[0])
		coreDirectoryAlreadyOnPath := false

		paths := strings.Split(pathVariable, ";")
		for _, path := range paths {
			path = strings.TrimSpace(path)
			if coreDirectory == path {
				coreDirectoryAlreadyOnPath = true
				break
			}
		}

		if !coreDirectoryAlreadyOnPath {
			pathVariable = strings.TrimSpace(pathVariable)
			if strings.HasSuffix(pathVariable, ";") {
				pathVariable = pathVariable + coreDirectory
			} else {
				pathVariable = pathVariable + ";" + coreDirectory
			}

			environmentVariables.SetStringValue("Path", pathVariable)
			NotifySettingChange()
		}
	},
}
