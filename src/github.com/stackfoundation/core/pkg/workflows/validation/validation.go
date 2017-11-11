package validation

import (
	"regexp"
	"strconv"

	"github.com/stackfoundation/core/pkg/workflows/errors"
	"github.com/stackfoundation/core/pkg/workflows/v1"
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

func validateFlag(step *v1.StepOptions, flag string, flagName string, selector []int, ignorePlaceholders bool) error {
	if len(flag) > 0 {
		if ignorePlaceholders && containsPlaceholders(flag) {
			return nil
		}

		_, err := strconv.ParseBool(flag)
		if err != nil {
			return newValidationError(flagName + " flag must be a boolean (true or false) in step " +
				step.StepName(selector))
		}
	}

	return nil
}

func validateStepType(step *v1.WorkflowStep, stepSelector []int, ignorePlaceholders bool) error {
	types := 0

	if step.Run != nil {
		types++
	}

	if step.Compound != nil {
		types++
	}

	if step.External != nil {
		types++
	}

	if step.Generator != nil {
		types++
	}

	if step.Service != nil {
		types++
	}

	if types > 1 {
		return newValidationError("Only a single type can be specified for " + step.StepName(stepSelector))
	} else if types < 1 {
		return newValidationError("Definition is incomplete for " + step.StepName(stepSelector))
	}

	return nil
}

func validateCompoundStep(compound *v1.CompoundStepOptions, selector []int) error {
	composite := errors.NewCompositeError()

	for stepNumber, subStep := range compound.Steps {
		subStepSelector := append(selector, stepNumber)
		composite.Append(validateStepInternal(&subStep, subStepSelector, false))
	}

	return composite.OrNilIfEmpty()
}

func validateStepInternal(step *v1.WorkflowStep, selector []int, ignorePlaceholders bool) error {
	err := validateStepType(step, selector, ignorePlaceholders)
	if err != nil {
		return err
	}

	if step.Run != nil {
		return validateRunStep(step.Run, selector, ignorePlaceholders)
	} else if step.Service != nil {
		return validateServiceStep(step.Service, selector, ignorePlaceholders)
	} else if step.External != nil {
		return validateExternalStep(step.External, selector, ignorePlaceholders)
	} else if step.Generator != nil {
		return validateGeneratorStep(step.Generator, selector, ignorePlaceholders)
	} else if step.Compound != nil {
		return validateCompoundStep(step.Compound, selector)
	}

	return nil
}

// ValidateStep Validate the specified workflow step
func ValidateStep(step *v1.WorkflowStep, stepSelector []int) error {
	return validateStepInternal(step, stepSelector, false)
}

// Validate Validate the specified workflow
func Validate(workflowSpec *v1.WorkflowSpec) error {
	if len(workflowSpec.Steps) == 0 {
		return newValidationError("Workflow must contain at least 1 step!")
	}

	stepSelector := make([]int, 1, 2)
	for stepNumber, step := range workflowSpec.Steps {
		stepSelector[0] = stepNumber

		err := validateStepInternal(&step, stepSelector, false)
		if err != nil {
			return err
		}
	}

	return nil
}
