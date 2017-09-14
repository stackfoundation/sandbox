package v1

import (
	"bytes"
)

type validationError struct {
	text string
}

type compositeError struct {
	errors []error
}

func (e *validationError) Error() string {
	return e.text
}

func (e *compositeError) Error() string {
	if e.errors != nil {
		var text bytes.Buffer

		for _, err := range e.errors {
			if err != nil {
				text.WriteString(err.Error())
				text.WriteString("\n")
			}
		}

		return text.String()
	}
	return ""
}

func newValidationError(text string) error {
	return &validationError{text}
}

func newCompositeError() *compositeError {
	err := &compositeError{
		errors: make([]error, 0, 2),
	}

	return err
}

func (e *compositeError) append(err error) {
	e.errors = append(e.errors, err)
}

func (e *compositeError) orNilIfEmpty() *compositeError {
	if len(e.errors) > 0 {
		return e
	}

	return nil
}

func validateStepType(step *WorkflowStep, stepSelector []int) error {
	if len(step.Type) > 0 {
		if step.Type != StepSequential &&
			step.Type != StepCompound &&
			step.Type != StepService &&
			step.Type != StepParallel {
			return newValidationError("Invalid step type specified for " + step.StepName(stepSelector))
		}
	}

	return nil
}

func validateStepSource(step *WorkflowStep, stepSelector []int) error {
	if step.OmitSource && len(step.SourceLocation) > 0 {
		return newValidationError("Source location cannot be specified when source is omitted for " +
			step.StepName(stepSelector))
	}

	return nil
}

func validateStepImage(step *WorkflowStep, stepSelector []int) error {
	if len(step.ImageSource) > 0 &&
		step.ImageSource != SourceStep &&
		step.ImageSource != SourceImage {
		return newValidationError("Invalid image source specified for " +
			step.StepName(stepSelector))
	}

	if len(step.Image) < 1 && len(step.Dockerfile) < 1 {
		return newValidationError("An image must be specified for " +
			step.StepName(stepSelector))
	}

	if len(step.Image) > 1 && len(step.Dockerfile) > 1 {
		return newValidationError("Both a Dockerfile and an image cannot be specified for " +
			step.StepName(stepSelector))
	}

	return nil
}

func validateStepHealth(step *WorkflowStep, stepSelector []int) error {
	if step.Readiness != nil {
		if len(step.Type) > 0 && step.Type != StepService {
			return newValidationError("Readiness is only vaild for service steps, and cannot be specified for " +
				step.StepName(stepSelector))
		}
	}

	if step.Health != nil {
		if len(step.Type) > 0 && step.Type != StepService {
			return newValidationError("Health is only vaild for service steps, and cannot be specified for " +
				step.StepName(stepSelector))
		}
	}

	return nil
}

func validateCompoundStep(step *WorkflowStep, stepSelector []int) error {
	if len(step.Type) == 0 || step.Type == StepCompound {
		errors := newCompositeError()

		for stepNumber, subStep := range step.Steps {
			subStepSelector := append(stepSelector, stepNumber)
			err := validateStep(&subStep, subStepSelector)
			if err != nil {
				errors.append(err)
			}
		}

		err := errors.orNilIfEmpty()
		if err != nil {
			return err
		}
	} else if len(step.Steps) > 0 {
		return newValidationError("Sub-steps are only valid in compound steps, and cannot be specified for " +
			step.StepName(stepSelector))
	}

	return nil
}

func newMultipleSubTypesError(step *WorkflowStep, stepSelector []int, types string) error {
	return newValidationError("A " + types + " cannot be specified at the same time for " +
		step.StepName(stepSelector))
}

func validateStepSubType(step *WorkflowStep, stepSelector []int) error {
	scriptPresent := len(step.Script) > 0
	generatorPresent := len(step.Generator) > 0
	targetPresent := len(step.Target) > 0
	dockerfilePresent := len(step.Dockerfile) > 0

	if scriptPresent {
		if generatorPresent {
			return newMultipleSubTypesError(step, stepSelector, "script and a generator script")
		}

		if targetPresent {
			return newMultipleSubTypesError(step, stepSelector, "script and a call target")
		}

		if dockerfilePresent {
			return newMultipleSubTypesError(step, stepSelector, "script and a Dockerfile")
		}

		return nil
	}

	if generatorPresent {
		if targetPresent {
			return newMultipleSubTypesError(step, stepSelector, "generator script and a call target")
		}

		if dockerfilePresent {
			return newMultipleSubTypesError(step, stepSelector, "generator script and a Dockerfile")
		}

		return nil
	}

	if targetPresent {
		if dockerfilePresent {
			return newMultipleSubTypesError(step, stepSelector, "call target and a Dockerfile")
		}

		return nil
	}

	if dockerfilePresent {
		return nil
	}

	return nil
}

func validateStep(step *WorkflowStep, stepSelector []int) *compositeError {
	errors := newCompositeError()

	err := validateStepType(step, stepSelector)
	if err != nil {
		errors.append(err)
	}

	err = validateStepSubType(step, stepSelector)
	if err != nil {
		errors.append(err)
	}

	err = validateStepSource(step, stepSelector)
	if err != nil {
		errors.append(err)
	}

	err = validateStepImage(step, stepSelector)
	if err != nil {
		errors.append(err)
	}

	err = validateStepHealth(step, stepSelector)
	if err != nil {
		errors.append(err)
	}

	err = validateCompoundStep(step, stepSelector)
	if err != nil {
		errors.append(err)
	}

	return errors.orNilIfEmpty()
}

// Validate Validate the specified workflow
func Validate(workflowSpec *WorkflowSpec) error {
	if len(workflowSpec.Steps) == 0 {
		return newValidationError("Workflow must contain at least 1 step!")
	}

	stepSelector := make([]int, 1, 2)
	for stepNumber, step := range workflowSpec.Steps {
		stepSelector[0] = stepNumber

		err := validateStep(&step, stepSelector)
		if err != nil {
			return err
		}
	}

	return nil
}
