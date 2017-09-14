package v1

// Select Select a workflow step
func (w *Workflow) Select(selector []int) *WorkflowStep {
	var step *WorkflowStep

	for _, segment := range selector {
		if step == nil {
			step = &w.Spec.Steps[segment]
		} else {
			step = &step.Steps[segment]
		}
	}

	return step
}

// Parent Get parent step of specified step
func (w *Workflow) Parent(selector []int) *WorkflowStep {
	var parent *WorkflowStep
	var step *WorkflowStep

	for _, segment := range selector {
		if step == nil {
			step = &w.Spec.Steps[segment]
		} else {
			parent = step
			step = &step.Steps[segment]
		}
	}

	return parent
}

// IncrementStepSelector Increment the given step selector, taking into account compound steps
func (w *Workflow) IncrementStepSelector(selector []int) []int {
	if len(selector) == 0 {
		return []int{0}
	}

	numSegments := len(selector)

	newSelector := make([]int, numSegments)

	steps := w.Spec.Steps
	stepCounts := make([]int, numSegments)
	for i, segment := range selector {
		stepCounts[i] = len(steps)
		steps = steps[segment].Steps
	}

	segment := numSegments - 1
	for segment >= 0 {
		newSelector[segment] = selector[segment] + 1

		if newSelector[segment] < stepCounts[segment] {
			break
		}

		newSelector = newSelector[:segment]
		segment--
	}

	if len(newSelector) > 0 {
		for {
			step := w.Select(newSelector)
			if step.Type == StepCompound {
				newSelector = append(newSelector, 0)
			} else {
				break
			}
		}
	}

	return newSelector
}
