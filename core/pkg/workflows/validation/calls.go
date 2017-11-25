package validation

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/errors"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

func validateExternalStep(external *v1.ExternalStepOptions, selector []int, ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()

	if len(external.Workflow) < 1 {
		composite.Append(newValidationError("A workflow must be specified for external " +
			external.StepName(selector)))
	}

	composite.Append(validateFlag(&external.StepOptions, external.Parallel, "Parallel", selector, ignorePlaceholders))

	return composite.OrNilIfEmpty()
}

func validateGeneratorStep(generator *v1.GeneratorStepOptions, selector []int, ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()

	composite.Append(validateScriptStep(&generator.ScriptStepOptions, selector, ignorePlaceholders))
	composite.Append(validateFlag(&generator.StepOptions, generator.Parallel, "Parallel", selector, ignorePlaceholders))

	return composite.OrNilIfEmpty()
}
