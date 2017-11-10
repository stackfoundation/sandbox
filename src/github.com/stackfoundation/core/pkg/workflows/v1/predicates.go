package v1

import "strconv"

// HasScript Does step have a dockerfile?
func (s *WorkflowStep) HasDockerfile() bool {
	return len(s.Dockerfile()) > 0
}

// HasScript Does step have a script to run?
func (s *WorkflowStep) HasScript() bool {
	return len(s.Script()) > 0
}

// UsesPreviousStep Does step use a previous step as it's base?
func (s *WorkflowStep) UsesPreviousStep() bool {
	return len(s.Step()) > 0
}

// Cached Is the step cached?
func (s *WorkflowStep) Cached() bool {
	if s.Run != nil {
		cache, _ := strconv.ParseBool(s.Run.Cache)
		return cache
	}

	return false
}

// IgnoreFailure Is ignore failure enabled for this step?
func (s *WorkflowStep) IgnoreFailure() *bool {
	options := s.StepOptions()
	if options != nil {
		return options.IgnoreFailure
	}

	return nil
}

// IgnoreMissing Is ignore missing enabled for this step?
func (s *WorkflowStep) IgnoreMissing() *bool {
	options := s.StepOptions()
	if options != nil {
		return options.IgnoreMissing
	}

	return nil
}

// IgnoreValidation Is ignore validation enabled for this step?
func (s *WorkflowStep) IgnoreValidation() *bool {
	options := s.StepOptions()
	if options != nil {
		return options.IgnoreValidation
	}

	return nil
}

// IsAsync Is this an async step (a paralell or service step that skips wait)?
func (s *WorkflowStep) IsAsync() bool {
	return (s.Service != nil && s.Service.Readiness != nil && s.Service.Readiness.SkipWait == "true") ||
		(s.Run != nil && s.Run.Parallel == "true") ||
		(s.External != nil && s.External.Parallel == "true") ||
		(s.Generator != nil && s.Generator.Parallel == "true")
}

// IsGenerator Is this a script generating step?
func (s *WorkflowStep) IsGenerator() bool {
	return s.Generator != nil
}

// IsServiceWithWait Is this a service step that waits for readiness?
func (s *WorkflowStep) IsServiceWithWait() bool {
	return s.Service != nil && s.Service.Readiness != nil && s.Service.Readiness.SkipWait != "true"
}

// ScriptlessImageBuild Does the step require an image to be built but doesn't have a script?
func (s *WorkflowStep) ScriptlessImageBuild() bool {
	return s.RequiresBuild()
}

// RequiresBuild Does the step require an image to be built (all except call steps)?
func (s *WorkflowStep) RequiresBuild() bool {
	return s.HasScript()
}

// OmitsSource Does the specified source options omit source?
func (s *SourceOptions) OmitsSource() bool {
	if s != nil {
		omitSource, _ := strconv.ParseBool(s.Omit)
		return omitSource
	}

	return false
}
