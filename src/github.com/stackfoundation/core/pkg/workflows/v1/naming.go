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

// GeneratePodName Generates a name for a pod
func GeneratePodName() string {
	uuid := uuid.NewUUID()
	return "pod-" + uuid.String()[:8]
}

// GenerateVolumeName Generates a name for a step volume
func GenerateVolumeName() string {
	uuid := uuid.NewUUID()
	return "vol-" + uuid.String()
}

// GenerateWorkflowID Generates an ID for a workflow
func GenerateWorkflowID() string {
	uuid := uuid.NewUUID()
	return "wflow-" + uuid.String()
}

// GenerateChangeID Generates an ID for a change
func GenerateChangeID() string {
	uuid := uuid.NewUUID()
	return "c-" + uuid.String()
}

// StepName Returns the name for a step
func (s *WorkflowStep) StepName(selector []int) string {
	var stepName string

	if len(s.Name) > 0 {
		stepName = s.Name
	} else {
		var nameBuilder bytes.Buffer

		for i, segment := range selector {
			if i > 0 {
				nameBuilder.WriteString(".")
			}

			nameBuilder.WriteString(strconv.Itoa(segment))
		}

		stepName = nameBuilder.String()
	}

	return stepName
}
