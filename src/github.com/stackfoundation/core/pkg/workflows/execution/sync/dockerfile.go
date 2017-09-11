package sync

import (
	"bytes"

	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func writeFromInstruction(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	if step.ImageSource == v1.SourceStep {
		// Use previous step image
	} else {
		dockerfile.WriteString("FROM ")
		dockerfile.WriteString(step.Image)
	}

	dockerfile.WriteString("\n")
}

func writePorts(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	if step.Ports != nil && len(step.Ports) > 0 {
		dockerfile.WriteString("EXPOSE")
		for _, port := range step.Ports {
			dockerfile.WriteString(" ")
			dockerfile.WriteString(port)
		}
		dockerfile.WriteString("\n")
	}
}

func writeSourceMount(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	if !step.OmitSource {
		sourceLocation := "/app/"
		if len(step.SourceLocation) > 0 {
			sourceLocation = step.SourceLocation
		}
		dockerfile.WriteString("COPY . ")
		dockerfile.WriteString(sourceLocation)
		dockerfile.WriteString("\n")
	}

	if len(step.State.GeneratedScript) > 0 {
		dockerfile.WriteString("COPY ")
		dockerfile.WriteString(step.State.GeneratedScript)
		dockerfile.WriteString(" /")
		dockerfile.WriteString(step.State.GeneratedScript)
		dockerfile.WriteString("\n")
	}
}

func buildDockerfile(step *v1.WorkflowStep) string {
	var dockerfile bytes.Buffer

	writeFromInstruction(&dockerfile, step)
	writeSourceMount(&dockerfile, step)
	writePorts(&dockerfile, step)

	return dockerfile.String()
}
