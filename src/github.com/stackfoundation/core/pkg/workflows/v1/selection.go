package v1

// SelectStep Select the specified step in the workflow
func SelectStep(workflowSpec *WorkflowSpec, stepSelector []int) *WorkflowStep {
	var step *WorkflowStep

	for _, segment := range stepSelector {
		if step == nil {
			step = &workflowSpec.Steps[segment]
		} else {
			step = &step.Steps[segment]
		}
	}

	return step
}

// IncrementStepSelector Increment the given step selector, taking into account compound steps
func IncrementStepSelector(workflowSpec *WorkflowSpec, stepSelector []int) []int {
	numSegments := len(stepSelector)

	newSelector := make([]int, numSegments)

	steps := workflowSpec.Steps
	stepCounts := make([]int, numSegments)
	for i, segment := range stepSelector {
		stepCounts[i] = len(steps)
		steps = steps[segment].Steps
	}

	segment := numSegments - 1
	for segment >= 0 {
		newSelector[segment] = stepSelector[segment] + 1

		if newSelector[segment] < stepCounts[segment] {
			break
		}

		newSelector = newSelector[:segment]
		segment--
	}

	if len(newSelector) > 0 {
		for {
			step := SelectStep(workflowSpec, newSelector)
			if step.Type == StepCompound {
				newSelector = append(newSelector, 0)
			} else {
				break
			}
		}
	}

	return newSelector
}
