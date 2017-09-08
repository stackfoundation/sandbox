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
			text.WriteString(err.Error())
			text.WriteString("\n")
		}

		return text.String()
	}
	return ""
}

func newValidationError(text string) error {
	return &validationError{text}
}

func newCompositeError() *compositeError {
	return &compositeError{
		errors: make([]error, 0, 2),
	}
}

func (e *compositeError) append(err error) {
	e.errors = append(e.errors, err)
}

func (e *compositeError) orNilIfEmpty() error {
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
			return newValidationError("Invalid step type specified for " + StepName(step, stepSelector))
		}
	}

	return nil
}

func validateStepSource(step *WorkflowStep, stepSelector []int) error {
	if step.OmitSource && len(step.SourceLocation) > 0 {
		return newValidationError("Source location cannot be specified when source is omitted for " +
			StepName(step, stepSelector))
	}

	return nil
}

func validateStepImage(step *WorkflowStep, stepSelector []int) error {
	if len(step.ImageSource) > 0 &&
		step.ImageSource != SourceStep &&
		step.ImageSource != SourceImage {
		return newValidationError("Invalid image source specified for " +
			StepName(step, stepSelector))
	}

	if len(step.Image) < 1 && len(step.Dockerfile) < 1 {
		return newValidationError("An image must be specified for " +
			StepName(step, stepSelector))
	}

	if len(step.Image) > 1 && len(step.Dockerfile) > 1 {
		return newValidationError("Both a Dockerfile and an image cannot be specified for " +
			StepName(step, stepSelector))
	}

	return nil
}

func validateStep(step *WorkflowStep, stepSelector []int) error {
	errors := newCompositeError()

	err := validateStepType(step, stepSelector)
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