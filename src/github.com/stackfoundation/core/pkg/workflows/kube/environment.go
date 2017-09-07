package kube

import (
	"github.com/magiconair/properties"
	"k8s.io/client-go/pkg/api/v1"
)

func createEnvironment(environment *properties.Properties) []v1.EnvVar {
	if environment != nil {
		numVariables := environment.Len()
		if numVariables > 0 {
			variables := make([]v1.EnvVar, 0, numVariables)
			keys := environment.Keys()
			for _, name := range keys {
				value, ok := environment.Get(name)
				if ok {
					variables = append(variables, v1.EnvVar{
						Name:  name,
						Value: value,
					})
				}
			}

			return variables
		}
	}

	return nil
}
