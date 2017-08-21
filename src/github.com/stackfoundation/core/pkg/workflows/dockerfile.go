package workflows

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

func writeVariables(dockerfile *bytes.Buffer, step *WorkflowStep) {
	if step.Variables != nil && len(step.Variables) > 0 {
		dockerfile.WriteString("ENV")
		for _, variable := range step.Variables {
			dockerfile.WriteString(" ")
			dockerfile.WriteString(variable.Name)
			dockerfile.WriteString("=\"")
			dockerfile.WriteString(variable.Value)
			dockerfile.WriteString("\"")
		}
		dockerfile.WriteString("\n")
	}
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
	}
}

func buildDockerfile(step *WorkflowStep) string {
	var dockerfile bytes.Buffer

	writeFromInstruction(&dockerfile, step)
	writeSourceMount(&dockerfile, step)
	writeVariables(&dockerfile, step)
	writePorts(&dockerfile, step)

	return dockerfile.String()
}
