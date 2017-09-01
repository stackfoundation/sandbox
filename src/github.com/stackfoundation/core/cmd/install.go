package cmd

const (
	ERROR_FILE_NOT_FOUND = 2
	ERROR_PATH_NOT_FOUND = 3
	ERROR_BAD_FORMAT     = 11
)

const HWND_BROADCAST = 0xffff
const WM_SETTINGCHANGE = 0x001A
const SMTO_ABORTIFHUNG = 0x0002
const (
	SE_ERR_ACCESSDENIED    = 5
	SE_ERR_OOM             = 8
	SE_ERR_DLLNOTFOUND     = 32
	SE_ERR_SHARE           = 26
	SE_ERR_ASSOCINCOMPLETE = 27
	SE_ERR_DDETIMEOUT      = 28
	SE_ERR_DDEFAIL         = 29
	SE_ERR_DDEBUSY         = 30
	SE_ERR_NOASSOC         = 31
)

func init() {
	RootCmd.AddCommand(installCmd)
}
