package main

//import "k8s.io/minikube/pkg/minikube/machine"
//import "k8s.io/minikube/cmd/minikube/cmd"
//import "k8s.io/minikube/pkg/minikube/config"
//import "github.com/spf13/viper"
//import "time"
import "./cmd"

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
        cmd.Execute()
}
