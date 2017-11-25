package expansion

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/errors"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/properties"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

func expandHeaders(headers []v1.HTTPHeader, variables *properties.Properties) ([]v1.HTTPHeader, error) {
	expandedHeaders := headers[:0]
	composite := errors.NewCompositeError()

	for _, header := range headers {
		name, err := variables.Expand(header.Name)
		composite.Append(err)

		value, err := variables.Expand(header.Value)
		composite.Append(err)

		expandedHeaders = append(expandedHeaders, v1.HTTPHeader{
			Name:  name,
			Value: value,
		})
	}

	return expandedHeaders, composite.OrNilIfEmpty()
}

func expandHealthCheckOptions(check *v1.HealthCheckOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	grace, err := variables.Expand(check.Grace)
	check.Grace = grace
	composite.Append(err)

	interval, err := variables.Expand(check.Interval)
	check.Interval = interval
	composite.Append(err)

	retries, err := variables.Expand(check.Retries)
	check.Retries = retries
	composite.Append(err)

	skipWait, err := variables.Expand(check.SkipWait)
	check.SkipWait = skipWait
	composite.Append(err)

	timeout, err := variables.Expand(check.Timeout)
	check.Timeout = timeout
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandTCPHealthCheckOptions(check *v1.TCPHealthCheckOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandHealthCheckOptions(&check.HealthCheckOptions, variables))

	port, err := variables.Expand(check.Port)
	check.Port = port
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandScriptHealthCheckOptions(check *v1.ScriptHealthCheckOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandHealthCheckOptions(&check.HealthCheckOptions, variables))

	path, err := variables.Expand(check.Path)
	check.Path = path
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandHTTPHealthCheckOptions(check *v1.HTTPHealthCheckOptions, variables *properties.Properties) error {
	composite := errors.NewCompositeError()

	composite.Append(expandHealthCheckOptions(&check.HealthCheckOptions, variables))

	port, err := variables.Expand(check.Port)
	check.Port = port
	composite.Append(err)

	path, err := variables.Expand(check.Path)
	check.Path = path
	composite.Append(err)

	headers, err := expandHeaders(check.Headers, variables)
	check.Headers = headers
	composite.Append(err)

	return composite.OrNilIfEmpty()
}

func expandHealthCheck(check *v1.HealthCheck, variables *properties.Properties) error {
	if check != nil {
		if check.HTTP != nil {
			return expandHTTPHealthCheckOptions(check.HTTP, variables)
		} else if check.HTTPS != nil {
			return expandHTTPHealthCheckOptions(check.HTTPS, variables)
		} else if check.Script != nil {
			return expandScriptHealthCheckOptions(check.Script, variables)
		} else if check.TCP != nil {
			return expandTCPHealthCheckOptions(check.TCP, variables)
		}
	}

	return nil
}
