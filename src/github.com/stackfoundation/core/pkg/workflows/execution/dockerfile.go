package execution

import "bytes"

func writeFromInstruction(dockerfile *bytes.Buffer, step *WorkflowStep) {
	if step.ImageSource == SourceStep {
		// Use previous step image
	} else {
		dockerfile.WriteString("FROM ")
		dockerfile.WriteString(step.Image)
		if len(step.Tag) > 0 {
			dockerfile.WriteString(":")
			dockerfile.WriteString(step.Tag)
		}
	}

	dockerfile.WriteString("\n")
}

func writePorts(dockerfile *bytes.Buffer, step *WorkflowStep) {
	if step.Ports != nil && len(step.Ports) > 0 {
		dockerfile.WriteString("EXPOSE")
		for _, port := range step.Ports {
			dockerfile.WriteString(" ")
			dockerfile.WriteString(port)
		}
		dockerfile.WriteString("\n")
	}
}

func writeSourceMount(dockerfile *bytes.Buffer, step *WorkflowStep) {
	if !step.OmitSource {
		sourceLocation := "/app/"
		if len(step.SourceLocation) > 0 {
			sourceLocation = step.SourceLocation
		}
		dockerfile.WriteString("COPY . ")
		dockerfile.WriteString(sourceLocation)
		dockerfile.WriteString("\n")
	}

	if len(step.StepScript) > 0 {
		dockerfile.WriteString("COPY ")
		dockerfile.WriteString(step.StepScript)
		dockerfile.WriteString(" /")
		dockerfile.WriteString(step.StepScript)
		dockerfile.WriteString("\n")
	}
}

func buildDockerfile(step *WorkflowStep) string {
	var dockerfile bytes.Buffer

	writeFromInstruction(&dockerfile, step)
	writeSourceMount(&dockerfile, step)
	writePorts(&dockerfile, step)

	return dockerfile.String()
}
