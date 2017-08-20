package workflows

import "bytes"

func writeFromInstruction(dockerfile *bytes.Buffer, step *WorkflowStep) {
	if step.ImageSource == SourceCatalog || step.ImageSource == SourceManual {
		dockerfile.WriteString("FROM ")
		dockerfile.WriteString(step.Image)
		if len(step.Tag) > 0 {
			dockerfile.WriteString(":")
			dockerfile.WriteString(step.Tag)
		}
	} else if step.ImageSource == SourceStep {
		// Use previous step image
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

func buildDockerfile(step *WorkflowStep) string {
	var dockerfile bytes.Buffer

	writeFromInstruction(&dockerfile, step)
	writeVariables(&dockerfile, step)
	writePorts(&dockerfile, step)
	dockerfile.WriteString("COPY . /app/")

	return dockerfile.String()
}
