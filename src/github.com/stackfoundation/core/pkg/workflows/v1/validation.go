package v1

import (
	"regexp"
	"strconv"

	"github.com/stackfoundation/core/pkg/workflows/errors"
)

var placeholderMatcher = regexp.MustCompile("\\$\\{[\\w]+\\}")

type validationError struct {
	text string
}

func (e *validationError) Error() string {
	return e.text
}

func newValidationError(text string) error {
	return &validationError{text}
}

func containsPlaceholders(text string) bool {
	return placeholderMatcher.MatchString(text)
}

func validateStepType(step *WorkflowStep, stepSelector []int) error {
	if len(step.Type) > 0 {
		if containsPlaceholders(step.Type) {
			return nil
		}

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
	if len(step.OmitSource) > 0 {
		if containsPlaceholders(step.OmitSource) {
			return nil
		}

		omitSource, err := strconv.ParseBool(step.OmitSource)
		if err != nil {
			return newValidationError("Omit source flag must be a boolean (true or false) in step " +
				step.StepName(stepSelector))
		}

		if omitSource && len(step.SourceLocation) > 0 {
			return newValidationError("Source location cannot be specified when source is omitted for " +
				step.StepName(stepSelector))
		}
	}

	return nil
}

func validateStepImage(step *WorkflowStep, stepSelector []int) error {
	if len(step.ImageSource) > 0 {
		if containsPlaceholders(string(step.ImageSource)) {
			return nil
		}

		if step.ImageSource != SourceStep &&
			step.ImageSource != SourceImage {
			return newValidationError("Invalid image source specified for " +
				step.StepName(stepSelector))
		}
	}

	if len(step.Image) < 1 && len(step.Dockerfile) < 1 && len(step.Target) < 1 {
		return newValidationError("An image must be specified for " +
			step.StepName(stepSelector))
	}

	if len(step.Image) > 1 && len(step.Dockerfile) > 1 {
		return newValidationError("Both a Dockerfile and an image cannot be specified for " +
			step.StepName(stepSelector))
	}

	if len(step.Image) > 1 && len(step.Target) > 1 {
		return newValidationError("Both a target workflow and an image cannot be specified for " +
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
		composite := errors.NewCompositeError()

		for stepNumber, subStep := range step.Steps {
			subStepSelector := append(stepSelector, stepNumber)
			composite.Append(ValidateStep(&subStep, subStepSelector))
		}

		return composite.OrNilIfEmpty()
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

// ValidateStep Validate the specified workflow step
func ValidateStep(step *WorkflowStep, stepSelector []int) error {
	composite := errors.NewCompositeError()

	composite.Append(validateStepType(step, stepSelector))
	composite.Append(validateStepSubType(step, stepSelector))
	composite.Append(validateStepSource(step, stepSelector))
	composite.Append(validateStepImage(step, stepSelector))
	composite.Append(validateStepHealth(step, stepSelector))
	composite.Append(validateCompoundStep(step, stepSelector))

	return composite.OrNilIfEmpty()
}

// Validate Validate the specified workflow
func Validate(workflowSpec *WorkflowSpec) error {
	if len(workflowSpec.Steps) == 0 {
		return newValidationError("Workflow must contain at least 1 step!")
	}

	stepSelector := make([]int, 1, 2)
	for stepNumber, step := range workflowSpec.Steps {
		stepSelector[0] = stepNumber

		err := ValidateStep(&step, stepSelector)
		if err != nil {
			return err
		}
	}

	return nil
}
