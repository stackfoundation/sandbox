package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stackfoundation/sandbox/core/pkg/hypervisor"
)

var failXhyve bool

var xhyveCmd = &cobra.Command{
	Use:    "xhyve",
	Hidden: true,
	Short:  "Install xhyve driver",
	Long:   `An internal command used to install the xhyve driver on the current system`,
	Run: func(command *cobra.Command, args []string) {
		hypervisor.InstallXhyve(failXhyve)
	},
}

func init() {
	xhyveCmd.Flags().BoolVar(&failXhyve, "fail", false, "Fail on error, instead of retrying with elevation")
	RootCmd.AddCommand(xhyveCmd)
}
