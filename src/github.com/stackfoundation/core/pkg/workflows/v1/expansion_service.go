package v1

import (
	"github.com/stackfoundation/core/pkg/workflows/errors"
	"github.com/stackfoundation/core/pkg/workflows/properties"
)

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

		container, err := variables.Expand(port.Container)
		composite.Append(err)

		internal, err := variables.Expand(port.Internal)
		composite.Append(err)

		external, err := variables.Expand(port.External)
		composite.Append(err)

		protocol, err := variables.Expand(port.Protocol)
		composite.Append(err)

		expandedPorts = append(expandedPorts, Port{
			Name:      name,
			Internal:  internal,
			External:  external,
			Container: container,
			Protocol:  protocol,
		})
	}

	return expandedPorts, composite.OrNilIfEmpty()
}

func expandServiceStepOptions(service *ServiceStepOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandScriptStepOptions(&service.ScriptStepOptions, variables))

	composite.Append(expandHealthCheck(service.Health, variables))

	ports, err := expandPorts(service.Ports, variables)
	service.Ports = ports
	composite.Append(err)

	composite.Append(expandHealthCheck(service.Readiness, variables))

	grace, err := variables.Expand(service.Grace)
	service.Grace = grace
	composite.Append(err)

	return composite.OrNilIfEmpty()
}
