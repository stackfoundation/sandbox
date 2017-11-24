package kube

import (
	"strconv"

	workflowsv1 "github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
)

func setupProbeOptions(probe *v1.Probe, check *workflowsv1.HealthCheckOptions) {
	probe.InitialDelaySeconds = parseInt(check.Grace, workflowsv1.DefaultGrace)
	probe.PeriodSeconds = parseInt(check.Interval, workflowsv1.DefaultInterval)
	probe.TimeoutSeconds = parseInt(check.Timeout, workflowsv1.DefaultTimeout)
	probe.FailureThreshold = parseInt(check.Retries, workflowsv1.DefaultRetries)
}

func createTCPProbe(check *workflowsv1.TCPHealthCheckOptions) *v1.Probe {
	probe := &v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: parseInt(check.Port, 0),
				},
			},
		},
	}

	setupProbeOptions(probe, &check.HealthCheckOptions)
	return probe
}

func createHTTPHeaders(check *workflowsv1.HTTPHealthCheckOptions) []v1.HTTPHeader {
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

func createHTTPGetProbe(check *workflowsv1.HTTPHealthCheckOptions, https bool) *v1.Probe {
	scheme := v1.URISchemeHTTP
	if https {
		scheme = v1.URISchemeHTTPS
	}

	headers := createHTTPHeaders(check)

	probe := &v1.Probe{
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

	setupProbeOptions(probe, &check.HealthCheckOptions)
	return probe
}

func createExecProbe(check *workflowsv1.ScriptHealthCheckOptions) *v1.Probe {
	probe := &v1.Probe{
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{"/bin/sh", check.Path},
			},
		},
	}

	setupProbeOptions(probe, &check.HealthCheckOptions)
	return probe
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

func createProbe(check *workflowsv1.HealthCheck) *v1.Probe {
	if check != nil {
		if check.TCP != nil {
			return createTCPProbe(check.TCP)
		} else if check.HTTPS != nil {
			return createHTTPGetProbe(check.HTTPS, true)
		} else if check.HTTP != nil {
			return createHTTPGetProbe(check.HTTP, false)
		} else if check.Script != nil {
			return createExecProbe(check.Script)
		}
	}

	return nil
}
