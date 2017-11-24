package expansion

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/errors"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/properties"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
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

func expandVariableOptions(variableOptions *v1.VariableOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	exclude, err := expandStringSlice(variableOptions.Exclude, variables)
	variableOptions.Exclude = exclude
	composite.Append(err)

	include, err := expandStringSlice(variableOptions.Include, variables)
	variableOptions.Include = include
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandGeneratorStepOptions(generator *v1.GeneratorStepOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandScriptStepOptions(&generator.ScriptStepOptions, variables))

	parallel, err := variables.Expand(generator.Parallel)
	generator.Parallel = parallel
	composite.Append(err)

	composite.Append(expandVariableOptions(&generator.Variables, variables))

	return composite.OrNilIfEmpty()
}

func expandExternalStepOptions(external *v1.ExternalStepOptions, variables *properties.Properties) error {
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
