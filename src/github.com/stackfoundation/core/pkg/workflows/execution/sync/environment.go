package sync

import (
	"github.com/magiconair/properties"

	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func collectStepEnvironment(environment []v1.EnvironmentSource) *properties.Properties {
	numSources := len(environment)

	if numSources > 0 {
		props := properties.NewProperties()

		for _, variable := range environment {
			if len(variable.File) > 0 {
				fileProperties, err := properties.LoadFile(variable.File, properties.UTF8)
				if err != nil || fileProperties == nil {
					log.Debugf("Error loading properties from file %v", variable.File)
					continue
				}

				props.Merge(fileProperties)
			} else {
				props.Set(variable.Name, variable.Value)
			}
		}

		return props
	}

	return nil
}
