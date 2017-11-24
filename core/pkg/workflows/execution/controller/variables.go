package controller

import "github.com/stackfoundation/core/pkg/workflows/properties"

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
