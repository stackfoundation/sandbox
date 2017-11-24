package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/hypervisor"
)

var failVirtualBox bool

var virtualBoxCmd = &cobra.Command{
	Use:    "virtualbox",
	Hidden: true,
	Short:  "Install VirtualBox",
	Long:   `An internal command used to install VirtualBox on the current system`,
	Run: func(command *cobra.Command, args []string) {
		err := hypervisor.InstallVirtualBox(failVirtualBox)
		if err != nil {
			fmt.Println("Error while installing VirtualBox: " + err.Error())
		}
	},
}

func init() {
	virtualBoxCmd.Flags().BoolVar(&failVirtualBox, "fail", false, "Fail on error, instead of retrying with elevation")
	RootCmd.AddCommand(virtualBoxCmd)
}
