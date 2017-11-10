package v1

import (
	"github.com/stackfoundation/core/pkg/workflows/errors"
	"github.com/stackfoundation/core/pkg/workflows/properties"
)

func expandEnvironment(environment []VariableSource, variables *properties.Properties) ([]VariableSource, error) {
	expandedEnvironment := environment[:0]
	composite := errors.NewCompositeError()

	for _, variable := range environment {
		name, err := variables.Expand(variable.Name)
		composite.Append(err)

		value, err := variables.Expand(variable.Value)
		composite.Append(err)

		file, err := variables.Expand(variable.File)
		composite.Append(err)

		expandedEnvironment = append(expandedEnvironment, VariableSource{
			Name:  name,
			Value: value,
			File:  file,
		})
	}

	return expandedEnvironment, composite.OrNilIfEmpty()
}

func expandScriptStepOptions(scriptOptions *ScriptStepOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandStepOptions(&scriptOptions.StepOptions, variables))

	dockerfile, err := variables.Expand(scriptOptions.Dockerfile)
	scriptOptions.Dockerfile = dockerfile
	composite.Append(err)

	environment, err := expandEnvironment(scriptOptions.Environment, variables)
	scriptOptions.Environment = environment
	composite.Append(err)

	image, err := variables.Expand(scriptOptions.Image)
	scriptOptions.Image = image
	composite.Append(err)

	script, err := variables.Expand(scriptOptions.Script)
	scriptOptions.Script = script
	composite.Append(err)

	composite.Append(expandSourceOptions(&scriptOptions.Source, variables))

	previousStep, err := variables.Expand(string(scriptOptions.Step))
	scriptOptions.Step = previousStep
	composite.Append(err)

	volumes, err := expandVolumes(scriptOptions.Volumes, variables)
	scriptOptions.Volumes = volumes
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandSourceOptions(sourceOptions *SourceOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	dockerignore, err := variables.Expand(sourceOptions.Dockerignore)
	sourceOptions.Dockerignore = dockerignore
	composite.Append(err)

	omit, err := variables.Expand(sourceOptions.Omit)
	sourceOptions.Omit = omit
	composite.Append(err)

	include, err := expandStringSlice(sourceOptions.Include, variables)
	sourceOptions.Include = include
	composite.Append(err)

	exclude, err := expandStringSlice(sourceOptions.Exclude, variables)
	sourceOptions.Exclude = exclude
	composite.Append(err)

	location, err := variables.Expand(sourceOptions.Location)
	sourceOptions.Location = location
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandVolumes(volumes []Volume, variables *properties.Properties) ([]Volume, error) {
	expandedVolumes := volumes[:0]
	composite := errors.NewCompositeError()

	for _, volume := range volumes {
		name, err := variables.Expand(volume.Name)
		composite.Append(err)

		hostPath, err := variables.Expand(volume.HostPath)
		composite.Append(err)

		mountPath, err := variables.Expand(volume.MountPath)
		composite.Append(err)

		expandedVolumes = append(expandedVolumes, Volume{
			Name:      name,
			HostPath:  hostPath,
			MountPath: mountPath,
		})
	}

	return expandedVolumes, composite.OrNilIfEmpty()
}
