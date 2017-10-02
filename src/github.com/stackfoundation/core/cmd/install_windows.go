package cmd

import (
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
var sendMessageTimeout = user32.NewProc("SendMessageTimeoutW")

const HWND_BROADCAST = 0xffff
const WM_SETTINGCHANGE = 0x001A
const SMTO_ABORTIFHUNG = 0x0002

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
