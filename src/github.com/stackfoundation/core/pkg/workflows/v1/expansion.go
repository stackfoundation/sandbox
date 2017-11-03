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

func expandHeaders(headers []HTTPHeader, variables *properties.Properties) ([]HTTPHeader, error) {
	expandedHeaders := headers[:0]
	composite := errors.NewCompositeError()

	for _, header := range headers {
		name, err := variables.Expand(header.Name)
		composite.Append(err)

		value, err := variables.Expand(header.Value)
		composite.Append(err)

		expandedHeaders = append(expandedHeaders, HTTPHeader{
			Name:  name,
			Value: value,
		})
	}

	return expandedHeaders, composite.OrNilIfEmpty()
}

func expandHealthCheck(check *HealthCheck, variables *properties.Properties) error {
	if check != nil {
		composite := errors.NewCompositeError()

		grace, err := variables.Expand(check.Grace)
		check.Grace = grace
		composite.Append(err)

		headers, err := expandHeaders(check.Headers, variables)
		check.Headers = headers
		composite.Append(err)

		interval, err := variables.Expand(check.Interval)
		check.Interval = interval
		composite.Append(err)

		path, err := variables.Expand(check.Path)
		check.Path = path
		composite.Append(err)

		port, err := variables.Expand(check.Port)
		check.Port = port
		composite.Append(err)

		retries, err := variables.Expand(check.Retries)
		check.Retries = retries
		composite.Append(err)

		script, err := variables.Expand(check.Script)
		check.Script = script
		composite.Append(err)

		skipWait, err := variables.Expand(check.SkipWait)
		check.SkipWait = skipWait
		composite.Append(err)

		timeout, err := variables.Expand(check.Timeout)
		check.Timeout = timeout
		composite.Append(err)

		checkType, err := variables.Expand(string(check.Type))
		check.Type = HealthCheckType(checkType)
		composite.Append(err)

		return composite.OrNilIfEmpty()
	}

	return nil
}

func expandPorts(ports []Port, variables *properties.Properties) ([]Port, error) {
	expandedPorts := ports[:0]
	composite := errors.NewCompositeError()

	for _, port := range ports {
		name, err := variables.Expand(port.Name)
		composite.Append(err)

		containerPort, err := variables.Expand(port.ContainerPort)
		composite.Append(err)

		internalPort, err := variables.Expand(port.InternalPort)
		composite.Append(err)

		externalPort, err := variables.Expand(port.ExternalPort)
		composite.Append(err)

		protocol, err := variables.Expand(port.Protocol)
		composite.Append(err)

		expandedPorts = append(expandedPorts, Port{
			Name:          name,
			InternalPort:  internalPort,
			ExternalPort:  externalPort,
			ContainerPort: containerPort,
			Protocol:      protocol,
		})
	}

	return expandedPorts, composite.OrNilIfEmpty()
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

// ExpandStep Expand any placeholders in this step that haven't been expanded yet
func ExpandStep(step *WorkflowStep, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	dockerfile, err := variables.Expand(step.Dockerfile)
	step.Dockerfile = dockerfile
	composite.Append(err)

	cache, err := variables.Expand(step.Cache)
	step.Cache = cache
	composite.Append(err)

	dockerignore, err := variables.Expand(step.Dockerignore)
	step.Dockerignore = dockerignore
	composite.Append(err)

	environment, err := expandEnvironment(step.Environment, variables)
	step.Environment = environment
	composite.Append(err)

	generator, err := variables.Expand(step.Generator)
	step.Generator = generator
	composite.Append(err)

	composite.Append(expandHealthCheck(step.Health, variables))

	excludeVariables, err := expandStringSlice(step.ExcludeVariables, variables)
	step.ExcludeVariables = excludeVariables
	composite.Append(err)

	includeVariables, err := expandStringSlice(step.IncludeVariables, variables)
	step.IncludeVariables = includeVariables
	composite.Append(err)

	image, err := variables.Expand(step.Image)
	step.Image = image
	composite.Append(err)

	imageSource, err := variables.Expand(string(step.ImageSource))
	step.ImageSource = ImageSource(imageSource)
	composite.Append(err)

	name, err := variables.Expand(step.Name)
	step.Name = name
	composite.Append(err)

	omitSource, err := variables.Expand(step.OmitSource)
	step.OmitSource = omitSource
	composite.Append(err)

	sourceIncludes, err := expandStringSlice(step.SourceIncludes, variables)
	step.SourceIncludes = sourceIncludes
	composite.Append(err)

	sourceExcludes, err := expandStringSlice(step.SourceExcludes, variables)
	step.SourceExcludes = sourceExcludes
	composite.Append(err)

	ports, err := expandPorts(step.Ports, variables)
	step.Ports = ports
	composite.Append(err)

	composite.Append(expandHealthCheck(step.Readiness, variables))

	script, err := variables.Expand(step.Script)
	step.Script = script
	composite.Append(err)

	sourceLocation, err := variables.Expand(step.SourceLocation)
	step.SourceLocation = sourceLocation
	composite.Append(err)

	target, err := variables.Expand(step.Target)
	step.Target = target
	composite.Append(err)

	terminationGrace, err := variables.Expand(step.TerminationGrace)
	step.TerminationGrace = terminationGrace
	composite.Append(err)

	stepType, err := variables.Expand(step.Type)
	step.Type = stepType
	composite.Append(err)

	volumes, err := expandVolumes(step.Volumes, variables)
	step.Volumes = volumes
	composite.Append(err)

	return composite.OrNilIfEmpty()
}
