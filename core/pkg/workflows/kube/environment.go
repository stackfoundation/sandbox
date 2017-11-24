package kube

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/properties"
	"k8s.io/client-go/pkg/api/v1"
)

func createEnvironment(environment *properties.Properties) []v1.EnvVar {
	if environment != nil {
		props := environment.Map()
		numVariables := len(props)
		if numVariables > 0 {
			variables := make([]v1.EnvVar, 0, numVariables)
			for k, v := range props {
				variables = append(variables, v1.EnvVar{
					Name:  k,
					Value: v,
				})
			}

			return variables
		}
	}

	return nil
}
