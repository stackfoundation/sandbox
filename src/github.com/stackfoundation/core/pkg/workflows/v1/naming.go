package v1

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
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

// GenerateCachedScriptName Generates a name for a cached step script
func GenerateCachedScriptName(content string) string {
	hash := md5.New()
	io.WriteString(hash, content)
	return "script-" + hex.EncodeToString(hash.Sum(nil)) + ".sh"
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

// GenerateServiceName Generates a service name
func GenerateServiceName() string {
	uuid := uuid.NewUUID()
	return "svc-" + uuid.String()[:8]
}

// GenerateServiceAssociation Generates an association key between a service and a pod
func GenerateServiceAssociation() string {
	uuid := uuid.NewUUID()
	return "assoc-" + uuid.String()[:8]
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
