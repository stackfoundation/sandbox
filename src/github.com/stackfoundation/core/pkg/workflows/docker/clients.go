package docker

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/docker/engine-api/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/stackfoundation/core/pkg/minikube/cluster"
	"github.com/stackfoundation/core/pkg/minikube/constants"
	"github.com/stackfoundation/core/pkg/minikube/machine"
)

func getHostDockerEnv() (map[string]string, error) {
	machineClient, err := machine.NewAPIClient()
	if err != nil {
		return nil, err
	}
	defer machineClient.Close()

	return cluster.GetHostDockerEnv(machineClient)
}

func createDockerHTTPClient(hostDockerEnv map[string]string) (*http.Client, error) {
	if dockerCertPath := hostDockerEnv["DOCKER_CERT_PATH"]; dockerCertPath != "" {
		tlsConfigOptions := tlsconfig.Options{
			CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
			CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
			KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
			InsecureSkipVerify: hostDockerEnv["DOCKER_TLS_VERIFY"] == "",
		}

		tlsClientConfig, err := tlsconfig.Client(tlsConfigOptions)
		if err != nil {
			return nil, err
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsClientConfig,
			},
		}

		return httpClient, nil
	}

	return nil, errors.New("Unable to determine Docker configuration")
}

// CreateDockerClient Create a new docker client pointing to the Sandbox VM's Docker daemon
func CreateDockerClient() (*client.Client, error) {
	hostDockerEnv, err := getHostDockerEnv()
	if err != nil {
		return nil, err
	}

	httpClient, err := createDockerHTTPClient(hostDockerEnv)
	if err != nil {
		return nil, err
	}

	host := hostDockerEnv["DOCKER_HOST"]
	return client.NewClient(host, constants.DockerAPIVersion, httpClient, nil)
}
