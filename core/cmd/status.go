package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/cluster"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/config"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/machine"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/service"
	"github.com/stackfoundation/sandbox/core/pkg/util/kubeconfig"
)

type retriableError struct {
	Err error
}

func (r retriableError) Error() string { return "Temporary Error: " + r.Err.Error() }

func retryAfter(attempts int, callback func() error, d time.Duration) (err error) {
	m := multiError{}
	for i := 0; i < attempts; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		m.Collect(err)
		if _, ok := err.(*retriableError); !ok {
			return m.ToError()
		}
		time.Sleep(d)
	}
	return m.ToError()
}

type multiError struct {
	Errors []error
}

func (m *multiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

func (m multiError) ToError() error {
	if len(m.Errors) == 0 {
		return nil
	}

	errStrings := []string{}
	for _, err := range m.Errors {
		errStrings = append(errStrings, err.Error())
	}
	return errors.New(strings.Join(errStrings, "\n"))
}

const defaultServiceFormatTemplate = "http://{{.IP}}:{{.Port}}"

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

		machineStatus, err := cluster.GetHostStatus(api)
		if err != nil {
			glog.Errorln("Error getting VM status:", err)
			MaybeReportErrorAndExit(err)
		}

		localKubeStatus := state.None.String()
		kubeRunning := false
		if machineStatus == state.Running.String() {
			localKubeStatus, err = cluster.GetLocalkubeStatus(api)
			if err != nil {
				glog.Errorln("Error getting VM status:", err)
				MaybeReportErrorAndExit(err)
			}
			ip, err := cluster.GetHostDriverIP(api)
			if err != nil {
				glog.Errorln("Error getting VM IP:", err)
				MaybeReportErrorAndExit(err)
			}
			kubeRunning, err = kubeconfig.GetKubeConfigStatus(ip, GetKubeConfigPath(), config.GetMachineName())
			if err != nil {
				glog.Errorln("Error kubeconfig status:", err)
				MaybeReportErrorAndExit(err)
			}
		}

		if kubeRunning && localKubeStatus == state.Running.String() {
			fmt.Println("The Sandbox VM and Kubernetes cluster are up and running, ready to run workflows")

			namespace := "kube-system"
			svc := "kubernetes-dashboard"

			err = retryAfter(20,
				func() error {
					return service.CheckService(namespace, svc)
				}, 6*time.Second)
			if err == nil {
				urls, err := service.GetServiceURLsForService(
					api, namespace, svc, template.Must(template.New("dashboardServiceFormat").Parse(defaultServiceFormatTemplate)))
				if err == nil {
					if len(urls) > 0 {
						fmt.Printf("The Kubernetes Dashboard is available at %v", urls[0])
					}
				}
			}

			if err != nil {
				fmt.Printf("Kubernetes Dashboard is not available: %v\n", err.Error())
			}
		} else {
			fmt.Println("No Sandbox VM is running. A VM will be started when you run a command or workflow")
			fmt.Println()
			fmt.Println(`Try running "sbox run --help" for more help`)
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
