/*
Adapted from: https://github.com/kubernetes/minikube/blob/master/cmd/minikube/cmd/start.go
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stackfoundation/core/pkg/hypervisor"

	"github.com/stackfoundation/net/proxy"

	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stackfoundation/core/pkg/minikube/cluster"
	cfg "github.com/stackfoundation/core/pkg/minikube/config"
	"github.com/stackfoundation/core/pkg/minikube/constants"
	"github.com/stackfoundation/core/pkg/minikube/kubernetes_versions"
	"github.com/stackfoundation/core/pkg/minikube/machine"
	"github.com/stackfoundation/core/pkg/util"
	pkgutil "github.com/stackfoundation/core/pkg/util"
	"github.com/stackfoundation/core/pkg/util/kubeconfig"
)

const (
	memory                = "memory"
	cpus                  = "cpus"
	humanReadableDiskSize = "disk-size"
	vmDriver              = "vm-driver"
	xhyveDiskDriver       = "xhyve-disk-driver"
	hostOnlyCIDR          = "host-only-cidr"
	containerRuntime      = "container-runtime"
	networkPlugin         = "network-plugin"
	hypervVirtualSwitch   = "hyperv-virtual-switch"
	kvmNetwork            = "kvm-network"
	createMount           = "mount"
	featureGates          = "feature-gates"
	apiServerName         = "apiserver-name"
	dnsDomain             = "dns-domain"
	mountString           = "mount-string"
	disableDriverMounts   = "disable-driver-mounts"
)

var (
	registryMirror    []string
	dockerEnv         []string
	dockerOpt         []string
	insecureRegistry  []string
	extraOptions      util.ExtraOptionSlice
	kubernetesVersion = constants.DefaultKubernetesVersion
)

func dockerEnvSet(variable string) bool {
	for _, env := range dockerEnv {
		lower := strings.ToLower(env)
		if strings.HasPrefix(lower, variable) {
			return true
		}
	}

	return false
}

func appendProxy() {
	httpProxy, overrides := proxy.SystemSettings()

	if len(httpProxy) > 0 {
		if !dockerEnvSet("http_proxy=") {
			dockerEnv = append(dockerEnv, "http_proxy="+httpProxy)
		}
	}

	if len(overrides) > 0 {
		if !dockerEnvSet("no_proxy=") {
			if proxy.ProxyOverrideSeparator == ";" {
				overrides = strings.Replace(overrides, ";", ",", -1)
			}

			dockerEnv = append(dockerEnv, "no_proxy="+overrides)
		}
	}
}

func startKube() {
	driver := hypervisor.SelectAndPrepareHypervisor(viper.GetString(vmDriver))

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

	if ms == state.Running.String() || ms == state.Starting.String() || ms == state.Stopping.String() {
		return
	}

	appendProxy()

	diskSize := viper.GetString(humanReadableDiskSize)
	diskSizeMB := calculateDiskSizeInMB(diskSize)

	if diskSizeMB < constants.MinimumDiskSizeMB {
		err := fmt.Errorf("Disk Size %dMB (%s) is too small, the minimum disk size is %dMB", diskSizeMB, diskSize, constants.MinimumDiskSizeMB)
		glog.Errorln("Error parsing disk size:", err)
		os.Exit(1)
	}

	if dv := kubernetesVersion; dv != constants.DefaultKubernetesVersion {
		validateK8sVersion(dv)
	}

	config := cluster.MachineConfig{
		MinikubeISO:         constants.DefaultIsoUrl,
		Memory:              viper.GetInt(memory),
		CPUs:                viper.GetInt(cpus),
		DiskSize:            diskSizeMB,
		VMDriver:            driver,
		XhyveDiskDriver:     viper.GetString(xhyveDiskDriver),
		DockerEnv:           dockerEnv,
		DockerOpt:           dockerOpt,
		InsecureRegistry:    insecureRegistry,
		RegistryMirror:      registryMirror,
		HostOnlyCIDR:        viper.GetString(hostOnlyCIDR),
		HypervVirtualSwitch: viper.GetString(hypervVirtualSwitch),
		KvmNetwork:          viper.GetString(kvmNetwork),
		Downloader:          pkgutil.DefaultDownloader{},
		DisableDriverMounts: viper.GetBool(disableDriverMounts),
	}

	fmt.Printf("Setting up and starting a local Kubernetes %s cluster...\n", kubernetesVersion)
	fmt.Println("Starting Sandbox VM...")
	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(api, config)
		if err != nil {
			glog.Errorf("Error starting host: %s.\n\n Retrying.\n", err)
		}
		return err
	}
	err = util.RetryAfter(5, start, 2*time.Second)
	if err != nil {
		glog.Errorln("Error starting host: ", err)
		MaybeReportErrorAndExit(err)
	}

	fmt.Println("Getting Sandbox VM IP address...")
	ip, err := host.Driver.GetIP()
	if err != nil {
		glog.Errorln("Error getting Sandbox VM IP address: ", err)
		MaybeReportErrorAndExit(err)
	}
	kubernetesConfig := cluster.KubernetesConfig{
		KubernetesVersion: kubernetesVersion,
		NodeIP:            ip,
		APIServerName:     viper.GetString(apiServerName),
		DNSDomain:         viper.GetString(dnsDomain),
		FeatureGates:      viper.GetString(featureGates),
		ContainerRuntime:  viper.GetString(containerRuntime),
		NetworkPlugin:     viper.GetString(networkPlugin),
		ExtraOptions:      extraOptions,
	}

	fmt.Println("Moving files into single-node Kubernetes cluster...")
	if err := cluster.UpdateCluster(host.Driver, kubernetesConfig); err != nil {
		glog.Errorln("Error updating cluster: ", err)
		MaybeReportErrorAndExit(err)
	}

	fmt.Println("Setting up certificates...")
	if err := cluster.SetupCerts(host.Driver, kubernetesConfig.APIServerName, kubernetesConfig.DNSDomain); err != nil {
		glog.Errorln("Error configuring authentication: ", err)
		MaybeReportErrorAndExit(err)
	}

	fmt.Println("Starting single-node Kubernetes cluster components...")

	if err := cluster.StartCluster(api, kubernetesConfig); err != nil {
		glog.Errorln("Error starting cluster: ", err)
		MaybeReportErrorAndExit(err)
	}

	fmt.Println("Connecting to single-node Kubernetes cluster...")
	kubeHost, err := host.Driver.GetURL()
	if err != nil {
		glog.Errorln("Error connecting to cluster: ", err)
	}
	kubeHost = strings.Replace(kubeHost, "tcp://", "https://", -1)
	kubeHost = strings.Replace(kubeHost, ":2376", ":"+strconv.Itoa(pkgutil.APIServerPort), -1)

	fmt.Println("Setting up kubeconfig...")
	// setup kubeconfig

	kubeConfigEnv := GetKubeConfigPath()
	var kubeConfigFile string
	if kubeConfigEnv == "" {
		kubeConfigFile = constants.KubeconfigPath
	} else {
		kubeConfigFile = filepath.SplitList(kubeConfigEnv)[0]
	}

	kubeCfgSetup := &kubeconfig.KubeConfigSetup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: kubeHost,
		ClientCertificate:    constants.MakeMiniPath("apiserver.crt"),
		ClientKey:            constants.MakeMiniPath("apiserver.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          true,
	}
	kubeCfgSetup.SetKubeConfigFile(kubeConfigFile)

	if err := kubeconfig.SetupKubeConfig(kubeCfgSetup); err != nil {
		glog.Errorln("Error setting up kubeconfig: ", err)
		MaybeReportErrorAndExit(err)
	}

	// start 9p server mount
	if viper.GetBool(createMount) {
		fmt.Printf("Setting up hostmount on %s...\n", viper.GetString(mountString))

		path := os.Args[0]
		mountDebugVal := 0
		if glog.V(8) {
			mountDebugVal = 1
		}
		mountCmd := exec.Command(path, "mount", fmt.Sprintf("--v=%d", mountDebugVal), viper.GetString(mountString))
		mountCmd.Env = append(os.Environ(), constants.IsMinikubeChildProcess+"=true")
		if glog.V(8) {
			mountCmd.Stdout = os.Stdout
			mountCmd.Stderr = os.Stderr
		}
		err = mountCmd.Start()
		if err != nil {
			glog.Errorf("Error running command Sandbox mount %s", err)
			MaybeReportErrorAndExit(err)
		}
		err = ioutil.WriteFile(filepath.Join(constants.GetMinipath(), constants.MountProcessFileName), []byte(strconv.Itoa(mountCmd.Process.Pid)), 0644)
		if err != nil {
			glog.Errorf("Error writing mount process pid to file: %s", err)
			MaybeReportErrorAndExit(err)
		}
	}

	fmt.Printf("A local Kubernetes cluster has been started. If you are familiar with Kubernetes and use kubectl, "+
		"note that the kubectl context has not been altered. If you want to use kubectl with the cluster that "+
		"was just started, kubectl will require \"--context=%s\".\n",
		kubeCfgSetup.ClusterName)

	if config.VMDriver == "none" {
		fmt.Println(`===================
WARNING: IT IS RECOMMENDED NOT TO RUN THE NONE DRIVER ON PERSONAL WORKSTATIONS
	The 'none' driver will run an insecure kubernetes apiserver as root that may leave the host vulnerable to CSRF attacks
`)

		if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
			fmt.Println(`When using the none driver, the kubectl config and credentials generated will be root owned and will appear in the root home directory.
You will need to move the files to the appropriate location and then set the correct permissions.  An example of this is below:
	sudo mv /root/.kube $HOME/.kube # this will overwrite any config you have.  You may have to append the file contents manually
	sudo chown -R $USER $HOME/.kube
	sudo chgrp -R $USER $HOME/.kube
	
    sudo mv /root/.minikube $HOME/.minikube # this will overwrite any config you have.  You may have to append the file contents manually
	sudo chown -R $USER $HOME/.minikube
	sudo chgrp -R $USER $HOME/.minikube 
This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true`)
		}
		if err := util.MaybeChownDirRecursiveToMinikubeUser(constants.GetMinipath()); err != nil {
			glog.Errorf("Error recursively changing ownership of directory %s: %s",
				constants.GetMinipath(), err)
			MaybeReportErrorAndExit(err)
		}
	}
}

func validateK8sVersion(version string) {
	validVersion, err := kubernetes_versions.IsValidLocalkubeVersion(version, constants.KubernetesVersionGCSURL)
	if err != nil {
		glog.Errorln("Error getting valid kubernetes versions", err)
		os.Exit(1)
	}
	if !validVersion {
		fmt.Println("Invalid Kubernetes version.")
		kubernetes_versions.PrintKubernetesVersionsFromGCS(os.Stdout)
		os.Exit(1)
	}
}

func calculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		glog.Errorf("Invalid disk size: %s", err)
	}
	return int(diskSize / units.MB)
}

func configureKubeStartingCommandFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(createMount, false, "This will start the mount daemon and automatically mount files into Sandbox")
	cmd.Flags().String(mountString, constants.DefaultMountDir+":"+constants.DefaultMountEndpoint, "The argument to pass the Sandbox mount command on start")
	cmd.Flags().Bool(disableDriverMounts, false, "Disables the filesystem mounts provided by the hypervisors (vboxfs, xhyve-9p)")
	cmd.Flags().String(vmDriver, "auto", fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	cmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the Sandbox VM")
	cmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the Sandbox VM")
	cmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the Sandbox VM (format: <number>[<unit>], where unit = b, k, m or g)")
	cmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the Sandbox VM (only supported with Virtualbox driver)")
	cmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)")
	cmd.Flags().String(kvmNetwork, "default", "The KVM network name. (only supported with KVM driver)")
	cmd.Flags().String(xhyveDiskDriver, "ahci-hd", "The disk driver to use [ahci-hd|virtio-blk] (only supported with xhyve driver)")
	cmd.Flags().StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	cmd.Flags().StringArrayVar(&dockerOpt, "docker-opt", nil, "Specify arbitrary flags to pass to the Docker daemon. (format: key=value)")
	cmd.Flags().String(apiServerName, constants.APIServerName, "The apiserver name which is used in the generated certificate for localkube/kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	cmd.Flags().String(dnsDomain, constants.ClusterDNSDomain, "The cluster dns domain name used in the kubernetes cluster")
	cmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", []string{pkgutil.DefaultInsecureRegistry}, "Insecure Docker registries to pass to the Docker daemon")
	cmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	cmd.Flags().String(containerRuntime, "", "The container runtime to be used")
	cmd.Flags().String(networkPlugin, "", "The name of the network plugin")
	cmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	cmd.Flags().Var(&extraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
                The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
                Valid components are: kubelet, apiserver, controller-manager, etcd, proxy, scheduler.`)
	viper.BindPFlags(cmd.Flags())
}
