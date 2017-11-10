package v1

import (
	"github.com/stackfoundation/core/pkg/workflows/errors"
	"github.com/stackfoundation/core/pkg/workflows/properties"
)

func expandStringSlice(slice []string, variables *properties.Properties) ([]string, error) {
	expandedSlice := slice[:0]
	composite := errors.NewCompositeError()

	for _, element := range slice {
		expanded, err := variables.Expand(element)
		composite.Append(err)

		expandedSlice = append(expandedSlice, expanded)
	}

	return expandedSlice, composite.OrNilIfEmpty()
}

func expandVariableOptions(variableOptions *VariableOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	exclude, err := expandStringSlice(variableOptions.Exclude, variables)
	variableOptions.Exclude = exclude
	composite.Append(err)

	include, err := expandStringSlice(variableOptions.Include, variables)
	variableOptions.Include = include
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandGeneratorStepOptions(generator *GeneratorStepOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandScriptStepOptions(&generator.ScriptStepOptions, variables))

	parallel, err := variables.Expand(generator.Parallel)
	generator.Parallel = parallel
	composite.Append(err)

	composite.Append(expandVariableOptions(&generator.Variables, variables))

	return composite.OrNilIfEmpty()
}

func expandExternalStepOptions(external *ExternalStepOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandStepOptions(&external.StepOptions, variables))

	parallel, err := variables.Expand(external.Parallel)
	external.Parallel = parallel
	composite.Append(err)

	composite.Append(expandVariableOptions(&external.Variables, variables))

	workflow, err := variables.Expand(external.Workflow)
	external.Workflow = workflow
	composite.Append(err)

	return composite.OrNilIfEmpty()
}
