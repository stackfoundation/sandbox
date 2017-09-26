package v1

// HasScript Does step have a script to run?
func (s *WorkflowStep) HasScript() bool {
	return len(s.Script) > 0 || s.IsGenerator()
}

// IsAsync Is this an async step (a paralell or service step that skips wait)?
func (s *WorkflowStep) IsAsync() bool {
	return s.Type == StepParallel ||
		(s.Type == StepService && s.Readiness != nil && s.Readiness.SkipWait == "true") ||
		(len(s.Type) == 0 && s.Readiness != nil && s.Readiness.SkipWait == "true")
}

// IsGenerator Is this a script generating step?
func (s *WorkflowStep) IsGenerator() bool {
	return len(s.Generator) > 0
}

// IsServiceWithWait Is this a service step that waits for readiness?
func (s *WorkflowStep) IsServiceWithWait() bool {
	return (s.Type == StepService && s.Readiness != nil && s.Readiness.SkipWait != "true") ||
		((len(s.Type) == 0) && s.Readiness != nil && s.Readiness.SkipWait != "true")
}

// ScriptlessImageBuild Does the step require an image to be built but doesn't have a script?
func (s *WorkflowStep) ScriptlessImageBuild() bool {
	return len(s.Image) > 0 &&
		(len(s.Script) == 0 && len(s.Dockerfile) == 0 && len(s.Generator) == 0 && len(s.Target) == 0)
}

// RequiresBuild Does the step require an image to be built (all except call steps)?
func (s *WorkflowStep) RequiresBuild() bool {
	return s.HasScript() || len(s.Dockerfile) > 0 || s.ScriptlessImageBuild()
}
