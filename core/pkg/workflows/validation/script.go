package validation

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/errors"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

func validateSource(script *v1.ScriptStepOptions, selector []int, ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()
	source := script.Source

	composite.Append(validateFlag(&script.StepOptions, source.Omit, "Omit source", selector, ignorePlaceholders))

	if len(source.Dockerignore) > 0 {
		if ignorePlaceholders && containsPlaceholders(source.Dockerignore) {
			return nil
		}

		if len(source.Include) > 0 {
			composite.Append(newValidationError("Source includes cannot be specified when dockerignore is also specified for " +
				script.StepName(selector)))
		}

		if len(source.Exclude) > 0 {
			composite.Append(newValidationError("Source excludes cannot be specified when dockerignore is also specified for " +
				script.StepName(selector)))
		}
	}

	return composite.OrNilIfEmpty()
}

func validateScriptStep(script *v1.ScriptStepOptions, selector []int, ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()

	if len(script.Image) < 1 && len(script.Dockerfile) < 1 && len(script.Step) < 1 {
		composite.Append(newValidationError("An image must be specified for " +
			script.StepName(selector)))
	}

	if len(script.Image) > 0 && len(script.Dockerfile) > 0 {
		composite.Append(newValidationError("Both a Dockerfile and an image cannot be specified for " +
			script.StepName(selector)))
	}

	if len(script.Image) > 0 && len(script.Step) > 0 {
		composite.Append(newValidationError("Both a previous step and an image cannot be specified for " +
			script.StepName(selector)))
	}

	if len(script.Step) > 0 && len(script.Dockerfile) > 0 {
		composite.Append(newValidationError("Both a Dockerfile and a previous step cannot be specified for " +
			script.StepName(selector)))
	}

	if len(script.Script) > 0 && len(script.Dockerfile) > 0 {
		composite.Append(newValidationError("A script and a Dockerfile cannot be specified for " +
			script.StepName(selector)))
	}

	composite.Append(validateSource(script, selector, ignorePlaceholders))

	return composite.OrNilIfEmpty()
}

func validateRunStep(run *v1.RunStepOptions, selector []int, ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()

	composite.Append(validateScriptStep(&run.ScriptStepOptions, selector, ignorePlaceholders))
	composite.Append(validateFlag(&run.StepOptions, run.Cache, "Cache", selector, ignorePlaceholders))
	composite.Append(validateFlag(&run.StepOptions, run.Parallel, "Parallel", selector, ignorePlaceholders))

	return composite.OrNilIfEmpty()
}
