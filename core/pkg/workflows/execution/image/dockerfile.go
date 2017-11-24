package image

import (
	"bytes"
	"strconv"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

func writeCherryPickCopies(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	picks := step.State.Picks
	if picks != nil {
		for i, pick := range picks {
			for _, copy := range pick.Copies {
				dockerfile.WriteString("COPY ")

				if len(pick.GeneratedBaseImage) > 0 {
					dockerfile.WriteString("--from=")
					dockerfile.WriteString(strconv.Itoa(i))
					dockerfile.WriteString(" ")
				}

				dockerfile.WriteString(copy)
				dockerfile.WriteString("\n")
			}
		}
	}
}

func writeCherryPickSources(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	picks := step.State.Picks
	if picks != nil {
		for _, pick := range picks {
			if len(pick.GeneratedBaseImage) > 0 {
				dockerfile.WriteString("FROM ")
				dockerfile.WriteString(pick.GeneratedBaseImage)
				dockerfile.WriteString("\n")
			}
		}
	}
}

func writeFromInstruction(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	dockerfile.WriteString("FROM ")

	if step.UsesPreviousStep() {
		dockerfile.WriteString(step.State.GeneratedBaseImage)
	} else {
		dockerfile.WriteString(step.Image())
	}

	dockerfile.WriteString("\n")
}

func writePorts(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	if step.Service != nil && step.Service.Ports != nil && len(step.Service.Ports) > 0 {
		dockerfile.WriteString("EXPOSE")
		for _, port := range step.Service.Ports {
			dockerfile.WriteString(" ")
			dockerfile.WriteString(port.Container)
		}
		dockerfile.WriteString("\n")
	}
}

func writeSourceMount(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	source := step.Source()
	if !source.OmitsSource() {
		sourceLocation := "/app/"
		if source != nil && len(source.Location) > 0 {
			sourceLocation = source.Location
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

func writeRunStepScriptInstruction(dockerfile *bytes.Buffer, step *v1.WorkflowStep) {
	if len(step.State.GeneratedScript) > 0 {
		dockerfile.WriteString("RUN [\"/bin/sh\", \"")
		dockerfile.WriteString(step.State.GeneratedScript)
		dockerfile.WriteString("\"]\n")
	}
}

func buildDockerfile(step *v1.WorkflowStep) string {
	var dockerfile bytes.Buffer

	writeCherryPickSources(&dockerfile, step)
	writeFromInstruction(&dockerfile, step)
	writeCherryPickCopies(&dockerfile, step)
	writeSourceMount(&dockerfile, step)
	writePorts(&dockerfile, step)

	if step.Cached() {
		writeRunStepScriptInstruction(&dockerfile, step)
	}

	return dockerfile.String()
}
