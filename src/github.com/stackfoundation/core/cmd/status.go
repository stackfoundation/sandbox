package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
        "github.com/docker/machine/libmachine/state"
        "github.com/golang/glog"
        "os"
        "github.com/stackfoundation/core/pkg/minikube/cluster"
        "github.com/stackfoundation/core/pkg/minikube/config"
        "github.com/stackfoundation/core/pkg/minikube/constants"
        "github.com/stackfoundation/core/pkg/minikube/machine"
        "github.com/stackfoundation/core/pkg/util/kubeconfig"
        "text/template"
)

type Status struct {
        MinikubeStatus   string
        LocalkubeStatus  string
        KubeconfigStatus string
}

var statusCmd = &cobra.Command{
        Use:   "status",
        Short: "Show the status of any running workflows, as well as overall Sandbox status",
        Long:  `Show the status of any running workflows, as well as overall Sandbox status.`,
        Run: func(command *cobra.Command, args []string) {
                api, err := machine.NewAPIClient()
                if err != nil {
                        fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
                        os.Exit(1)
                }
                defer api.Close()

                ms, err := cluster.GetHostStatus(api)
                if err != nil {
                        glog.Errorln("Error getting machine status:", err)
                        MaybeReportErrorAndExit(err)
                }

                ls := state.None.String()
                ks := state.None.String()
                if ms == state.Running.String() {
                        ls, err = cluster.GetLocalkubeStatus(api)
                        if err != nil {
                                glog.Errorln("Error localkube status:", err)
                                MaybeReportErrorAndExit(err)
                        }
                        ip, err := cluster.GetHostDriverIP(api)
                        if err != nil {
                                glog.Errorln("Error host driver ip status:", err)
                                MaybeReportErrorAndExit(err)
                        }
                        kstatus, err := kubeconfig.GetKubeConfigStatus(ip, GetKubeConfigPath(), config.GetMachineName())
                        if err != nil {
                                glog.Errorln("Error kubeconfig status:", err)
                                MaybeReportErrorAndExit(err)
                        }
                        if kstatus {
                                ks = "Correctly Configured: pointing to minikube-vm at " + ip.String()
                        } else {
                                ks = "Misconfigured: pointing to stale minikube-vm." +
                                        "\nTo fix the kubectl context, run minikube update-context"
                        }
                }

                status := Status{ms, ls, ks}

                tmpl, err := template.New("status").Parse(constants.DefaultStatusFormat)
                if err != nil {
                        glog.Errorln("Error creating status template:", err)
                        os.Exit(1)
                }
                err = tmpl.Execute(os.Stdout, status)
                if err != nil {
                        glog.Errorln("Error executing status template:", err)
                        os.Exit(1)
                }
        },
}

func init() {
        RootCmd.AddCommand(statusCmd)
}
