package run

import (
	"github.com/stackfoundation/core/pkg/workflows/properties"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func collectVariables(variables []v1.VariableSource) *properties.Properties {
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

func filterVariables(inclusions []string, exclusions []string, variables *properties.Properties) *properties.Properties {
	includeAll := false
	if inclusions == nil || (len(inclusions) == 1 && inclusions[0] == "*") {
		includeAll = true
	}

	excludeNone := false
	if exclusions == nil {
		excludeNone = true
	}

	props := properties.NewProperties()

	for key, value := range variables.Map() {
		include := includeAll

		if !include {
			if globAll(inclusions, key) {
				include = true
			}
		}

		if include && !excludeNone {
			if globAll(exclusions, key) {
				include = false
			}
		}

		if include {
			props.Set(key, value)
		}
	}

	return props
}
