package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/hypervisor"
)

var fail bool

var virtualBoxCmd = &cobra.Command{
	Use:    "virtualbox",
	Hidden: true,
	Short:  "Install VirtualBox",
	Long:   `An internal command used to install VirtualBox on the current system`,
	Run: func(command *cobra.Command, args []string) {
		// err := hypervisor.InstallVirtualBox()
		// if err != nil {
		// 	if fail {
		// 		fmt.Println(err.Error())
		// 	} else {
		// 		install.ElevatedExecute(os.Args[0], "virtualbox --fail")
		// 	}
		// }
		hypervisor.SelectAndPrepareHypervisor("auto")
	},
}

func init() {
	virtualBoxCmd.Flags().BoolVar(&fail, "fail", false, "Fail on error, instead of retrying with elevation")
	RootCmd.AddCommand(virtualBoxCmd)
}
