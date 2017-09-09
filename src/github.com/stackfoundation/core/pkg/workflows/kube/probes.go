package kube

import (
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
)

func createTCPProbe(health *workflowsv1.HealthCheck) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: health.Port,
				},
			},
		},
	}
}

func createHTTPHeaders(health *workflowsv1.HealthCheck) []v1.HTTPHeader {
	numHeaders := len(health.Headers)
	if numHeaders > 0 {
		headers := make([]v1.HTTPHeader, 0, numHeaders)
		for _, header := range health.Headers {
			headers = append(headers, v1.HTTPHeader{
				Name:  header.Name,
				Value: header.Value,
			})
		}

		return headers
	}

	return nil
}

func createHTTPGetProbe(health *workflowsv1.HealthCheck) *v1.Probe {
	scheme := v1.URISchemeHTTP
	if health.Type == workflowsv1.HTTPSCheck {
		scheme = v1.URISchemeHTTPS
	}

	headers := createHTTPHeaders(health)

	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: health.Port,
				},
				Path:        health.Path,
				Scheme:      scheme,
				HTTPHeaders: headers,
			},
		},
	}
}

func createExecProbe(health *workflowsv1.HealthCheck) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{"/bin/sh", health.Script},
			},
		},
	}
}

func setupProbeOptions(probe *v1.Probe, health *workflowsv1.HealthCheck) {
	if health.Grace != nil {
		probe.InitialDelaySeconds = *health.Grace
	} else {
		probe.InitialDelaySeconds = workflowsv1.DefaultGrace
	}

	if health.Interval != nil {
		probe.PeriodSeconds = *health.Interval
	} else {
		probe.PeriodSeconds = workflowsv1.DefaultInterval
	}

	if health.Timeout != nil {
		probe.TimeoutSeconds = *health.Timeout
	} else {
		probe.TimeoutSeconds = workflowsv1.DefaultTimeout
	}

	if health.Retries != nil {
		probe.FailureThreshold = *health.Retries
	} else {
		probe.FailureThreshold = workflowsv1.DefaultRetries
	}
}

func createReadinessProbe(health *workflowsv1.HealthCheck) *v1.Probe {
	if health != nil {
		var probe *v1.Probe
		switch health.Type {
		case workflowsv1.TCPCheck:
			probe = createTCPProbe(health)
		case workflowsv1.HTTPSCheck:
			fallthrough
		case workflowsv1.HTTPCheck:
			probe = createHTTPGetProbe(health)
		case workflowsv1.ScriptCheck:
			probe = createExecProbe(health)
		default:
			probe = nil
		}

		if probe != nil {
			setupProbeOptions(probe, health)
		}

		return probe
	}

	return nil
}
