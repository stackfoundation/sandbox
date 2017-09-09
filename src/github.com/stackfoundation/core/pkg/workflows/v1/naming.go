package v1

import (
	"bytes"
	"strconv"

	"github.com/pborman/uuid"
)

// GenerateContainerName Generates a name for a step container
func GenerateContainerName() string {
	uuid := uuid.NewUUID()
	return "sbox-" + uuid.String()
}

// GenerateImageName Generates a name for a step image
func GenerateImageName() string {
	uuid := uuid.NewUUID()
	return "step:" + uuid.String()
}

// GenerateScriptName Generates a name for a step script
func GenerateScriptName() string {
	uuid := uuid.NewUUID()
	return "script-" + uuid.String()[:8] + ".sh"
}

// GenerateVolumeName Generates a name for a step volume
func GenerateVolumeName() string {
	uuid := uuid.NewUUID()
	return "vol-" + uuid.String()
}

// StepName Returns the name for a step
func StepName(step *WorkflowStep, stepSelector []int) string {
	var stepName string

	if len(step.Name) > 0 {
		stepName = `"` + step.Name + `"`
	} else {
		var nameBuilder bytes.Buffer

		nameBuilder.WriteString("step ")
		for i, segment := range stepSelector {
			if i > 0 {
				nameBuilder.WriteString(".")
			}

			nameBuilder.WriteString(strconv.Itoa(segment))
		}

		stepName = nameBuilder.String()
	}

	return stepName
}
