package expansion

import (
	"github.com/stackfoundation/core/pkg/workflows/errors"
	"github.com/stackfoundation/core/pkg/workflows/properties"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func expandPorts(ports []v1.Port, variables *properties.Properties) ([]v1.Port, error) {
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

		expandedPorts = append(expandedPorts, v1.Port{
			Name:      name,
			Internal:  internal,
			External:  external,
			Container: container,
			Protocol:  protocol,
		})
	}

	return expandedPorts, composite.OrNilIfEmpty()
}

func expandServiceStepOptions(service *v1.ServiceStepOptions, variables *properties.Properties) error {
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
