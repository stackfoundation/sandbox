package v1

import (
	"github.com/stackfoundation/core/pkg/workflows/properties"
	"github.com/stackfoundation/log"
)

// CollectVariables Collect all the variables from the specified sources
func CollectVariables(variables []VariableSource) *properties.Properties {
	numSources := len(variables)
	props := properties.NewProperties()

	if numSources > 0 {
		for _, variable := range variables {
			if len(variable.File) > 0 {
				fileProperties := properties.NewProperties()
				err := fileProperties.Load(variable.File)
				if err != nil {
					log.Debugf("Error loading properties from file %v", variable.File)
					continue
				}

				props.Merge(fileProperties)
			} else {
				props.Set(variable.Name, variable.Value)
			}
		}
	}

	return props
}
