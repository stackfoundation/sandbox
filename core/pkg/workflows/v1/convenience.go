package v1

import "strconv"

// CherryPick Get the cherry picked files for this step, if it has any
func (s *WorkflowStep) CherryPick() []CherryPick {
	// scriptOptions := s.scriptStepOptions()
	// if scriptOptions != nil {
	// 	return scriptOptions.CherryPick
	// }

	return nil
}

// Dockerfile Get the Dockerfile for this step, if it has one
func (s *WorkflowStep) Dockerfile() string {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return scriptOptions.Dockerfile
	}

	return ""
}

// Environment Get the environment variables for this step, if it has any
func (s *WorkflowStep) Environment() []VariableSource {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return scriptOptions.Environment
	}

	return nil
}

// Image Get the image for this step, if it has one
func (s *WorkflowStep) Image() string {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return scriptOptions.Image
	}

	return ""
}

// Name Get the name of the step, if it has one
func (s *WorkflowStep) Name() string {
	options := s.StepOptions()
	if options != nil {
		return options.Name
	}

	return ""
}

// Script Get the script for this step, if it has one
func (s *WorkflowStep) Script() string {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return scriptOptions.Script
	}

	return ""
}

func (s *WorkflowStep) scriptStepOptions() *ScriptStepOptions {
	if s.Run != nil {
		return &s.Run.ScriptStepOptions
	} else if s.Service != nil {
		return &s.Service.ScriptStepOptions
	} else if s.Generator != nil {
		return &s.Generator.ScriptStepOptions
	}

	return nil
}

// SetVolumes Set the volumes for this step
func (s *WorkflowStep) SetVolumes(volumes []Volume) {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		scriptOptions.Volumes = volumes
	}
}

// SkipWait Is skip waiting configured for this check?
func (c *HealthCheck) SkipWait() bool {
	if c.TCP != nil {
		skipWait, _ := strconv.ParseBool(c.TCP.SkipWait)
		return skipWait
	} else if c.HTTP != nil {
		skipWait, _ := strconv.ParseBool(c.HTTP.SkipWait)
		return skipWait
	} else if c.HTTPS != nil {
		skipWait, _ := strconv.ParseBool(c.HTTPS.SkipWait)
		return skipWait
	} else if c.Script != nil {
		skipWait, _ := strconv.ParseBool(c.Script.SkipWait)
		return skipWait
	}

	return false
}

// Source Get the source options for this step, if it has any
func (s *WorkflowStep) Source() *SourceOptions {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return &scriptOptions.Source
	}

	return nil
}

// Step Get the previous step for this step, if it has one
func (s *WorkflowStep) Step() string {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return scriptOptions.Step
	}

	return ""
}

// StepOptions Get the options for this step
func (s *WorkflowStep) StepOptions() *StepOptions {
	if s.Run != nil {
		return &s.Run.StepOptions
	} else if s.Service != nil {
		return &s.Service.StepOptions
	} else if s.External != nil {
		return &s.External.StepOptions
	} else if s.Compound != nil {
		return &s.Compound.StepOptions
	} else if s.Generator != nil {
		return &s.Generator.StepOptions
	}

	return nil
}

// Volumes Get the volumes for this step, if it has any
func (s *WorkflowStep) Volumes() []Volume {
	scriptOptions := s.scriptStepOptions()
	if scriptOptions != nil {
		return scriptOptions.Volumes
	}

	return nil
}
