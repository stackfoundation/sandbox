// +build !darwin

package cmd

import "github.com/stackfoundation/core/pkg/minikube/constants"

func DefaultVMDriver() string {
	return constants.DefaultVMDriver
}
