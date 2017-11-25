package v1

// AppendChange Add a new change to this workflow
func (w *Workflow) AppendChange(c *Change) *Change {
	w.Spec.State.Changes = append(w.Spec.State.Changes, *c)
	return &w.Spec.State.Changes[len(w.Spec.State.Changes)-1]
}

// MarkHandled Mark the specified change in the workflow as handled
func (w *Workflow) MarkHandled(c *Change) {
	for i, change := range w.Spec.State.Changes {
		if change.ID == c.ID {
			w.Spec.State.Changes[i].Handled = true
			break
		}
	}
}

// NextUnhandled Get the next unhandled change in workflow
func (w *Workflow) NextUnhandled() *Change {
	for i := len(w.Spec.State.Changes) - 1; i >= 0; i-- {
		c := &w.Spec.State.Changes[i]
		if !c.Handled {
			return c
		}
	}

	return nil
}

// IsCompoundStepComplete Is the compound step complete?
func (s *WorkflowStep) IsCompoundStepComplete() bool {
	for _, step := range s.Compound.Steps {
		if step.Service != nil {
			if !step.State.Ready || !step.State.Done {
				return false
			}
		} else if step.Run != nil && step.Run.Parallel == "true" {
			if !step.State.Done {
				return false
			}
		} else if step.Generator != nil && step.Generator.Parallel == "true" {
			if !step.State.Done {
				return false
			}
		} else if step.External != nil && step.External.Parallel == "true" {
			if !step.State.Done {
				return false
			}
		}
	}

	return true
}
