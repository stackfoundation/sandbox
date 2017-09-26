package kube

import (
	"strconv"

	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
)

func createTCPProbe(check *workflowsv1.HealthCheck) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: parseInt(check.Port, 0),
				},
			},
		},
	}
}

func createHTTPHeaders(check *workflowsv1.HealthCheck) []v1.HTTPHeader {
	numHeaders := len(check.Headers)
	if numHeaders > 0 {
		headers := make([]v1.HTTPHeader, 0, numHeaders)
		for _, header := range check.Headers {
			headers = append(headers, v1.HTTPHeader{
				Name:  header.Name,
				Value: header.Value,
			})
		}

		return headers
	}

	return nil
}

func createHTTPGetProbe(check *workflowsv1.HealthCheck) *v1.Probe {
	scheme := v1.URISchemeHTTP
	if check.Type == workflowsv1.HTTPSCheck {
		scheme = v1.URISchemeHTTPS
	}

	headers := createHTTPHeaders(check)

	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: parseInt(check.Port, 0),
				},
				Path:        check.Path,
				Scheme:      scheme,
				HTTPHeaders: headers,
			},
		},
	}
}

func createExecProbe(check *workflowsv1.HealthCheck) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{"/bin/sh", check.Script},
			},
		},
	}
}

func parseInt(value string, defaultValue int32) int32 {
	if len(value) > 0 {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return int32(parsed)
	}

	return defaultValue
}

func setupProbeOptions(probe *v1.Probe, check *workflowsv1.HealthCheck) {
	probe.InitialDelaySeconds = parseInt(check.Grace, workflowsv1.DefaultGrace)
	probe.PeriodSeconds = parseInt(check.Interval, workflowsv1.DefaultInterval)
	probe.TimeoutSeconds = parseInt(check.Timeout, workflowsv1.DefaultTimeout)
	probe.FailureThreshold = parseInt(check.Retries, workflowsv1.DefaultRetries)
}

func createProbe(check *workflowsv1.HealthCheck) *v1.Probe {
	if check != nil {
		var probe *v1.Probe
		switch check.Type {
		case workflowsv1.TCPCheck:
			probe = createTCPProbe(check)
		case workflowsv1.HTTPSCheck:
			fallthrough
		case workflowsv1.HTTPCheck:
			probe = createHTTPGetProbe(check)
		case workflowsv1.ScriptCheck:
			probe = createExecProbe(check)
		default:
			probe = nil
		}

		if probe != nil {
			setupProbeOptions(probe, check)
		}

		return probe
	}

	return nil
}
