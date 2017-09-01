package main

//import "github.com/stackfoundation/core/pkg/minikube/machine"
//import "k8s.io/minikube/cmd/minikube/cmd"
//import "github.com/stackfoundation/core/pkg/minikube/config"

//import "time"
import (
	"github.com/spf13/viper"
	"github.com/stackfoundation/core/cmd"
)

func initConfig() {
	//viper.Set(config.WantKubectlDownloadMsg, false)
	//viper.Set(config.WantUpdateNotification, false)
}

func main() {
	//machine.StartDriver()
	//cmd.RootCmd.SetArgs([]string{"start"})
	//cmd.RootCmd.Execute()
	//time.Sleep(10 * time.Second)
	//cmd.RootCmd.SetArgs([]string{"dashboard"})
	//cmd.RootCmd.Execute()
	viper.Set("vm-driver", "xhyve")
	cmd.Execute()
}
