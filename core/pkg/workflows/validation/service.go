package validation

import (
	"strconv"

	"github.com/stackfoundation/core/pkg/workflows/errors"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func validatePositiveInt(step *v1.StepOptions, value string, propertyName string, checkName string, selector []int, ignorePlaceholders bool) error {
	if len(value) > 0 {
		if ignorePlaceholders && containsPlaceholders(value) {
			return nil
		}

		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v < 1 {
			return newValidationError(propertyName + " must be a positive integer in the " + checkName + " check step " +
				step.StepName(selector))
		}
	}

	return nil
}

func validateHealthCheckOptions(
	service *v1.ServiceStepOptions,
	checkName string,
	check *v1.HealthCheckOptions,
	selector []int,
	ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()

	composite.Append(validateFlag(&service.StepOptions, check.SkipWait, "skip wait", selector, ignorePlaceholders))
	composite.Append(validatePositiveInt(&service.StepOptions, check.Grace, "Grace", checkName, selector, ignorePlaceholders))
	composite.Append(validatePositiveInt(&service.StepOptions, check.Interval, "Interval", checkName, selector, ignorePlaceholders))
	composite.Append(validatePositiveInt(&service.StepOptions, check.Retries, "Retries", checkName, selector, ignorePlaceholders))
	composite.Append(validatePositiveInt(&service.StepOptions, check.Timeout, "Timeout", checkName, selector, ignorePlaceholders))

	return composite.OrNilIfEmpty()
}

func validateHealthCheck(service *v1.ServiceStepOptions, checkName string, check *v1.HealthCheck, selector []int, ignorePlaceholders bool) error {
	types := 0

	if check.HTTP != nil {
		types++
	} else if check.HTTPS != nil {
		types++
	} else if check.Script != nil {
		types++
	} else if check.TCP != nil {
		types++
	}

	if types > 1 {
		return newValidationError("Only one type of " + checkName + " check can be specified for " +
			service.StepName(selector))
	}

	if check.HTTP != nil {
		return validateHealthCheckOptions(service, checkName, &check.HTTP.HealthCheckOptions, selector, ignorePlaceholders)
	} else if check.HTTPS != nil {
		return validateHealthCheckOptions(service, checkName, &check.HTTPS.HealthCheckOptions, selector, ignorePlaceholders)
	} else if check.Script != nil {
		return validateHealthCheckOptions(service, checkName, &check.Script.HealthCheckOptions, selector, ignorePlaceholders)
	} else if check.TCP != nil {
		return validateHealthCheckOptions(service, checkName, &check.TCP.HealthCheckOptions, selector, ignorePlaceholders)
	}

	return nil
}

func validateServiceStep(service *v1.ServiceStepOptions, selector []int, ignorePlaceholders bool) error {
	composite := errors.NewCompositeError()

	composite.Append(validateScriptStep(&service.ScriptStepOptions, selector, ignorePlaceholders))

	if service.Readiness != nil {
		composite.Append(validateHealthCheck(service, "readiness", service.Readiness, selector, ignorePlaceholders))
	}

	if service.Health != nil {
		composite.Append(validateHealthCheck(service, "health", service.Health, selector, ignorePlaceholders))
	}

	return composite.OrNilIfEmpty()
}
