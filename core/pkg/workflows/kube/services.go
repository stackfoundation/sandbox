package kube

import (
	"strings"

	workflowsv1 "github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	log "github.com/stackfoundation/sandbox/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
)

func extractProtocol(port string) (v1.Protocol, string) {
	protocol := v1.ProtocolTCP

	protocolSeparator := strings.Index(port, "/")
	if protocolSeparator > -1 {
		if "udp" == port[protocolSeparator+1:] {
			protocol = v1.ProtocolUDP
		}

		port = port[:protocolSeparator]
	}

	return protocol, port
}

func createServicePort(port workflowsv1.Port) v1.ServicePort {
	protocol := v1.ProtocolTCP

	if strings.ToLower(port.Protocol) == "udp" {
		protocol = v1.ProtocolUDP
	}

	containerPort := parseInt(port.Container, 0)
	internalPort := parseInt(port.Internal, 0)
	if internalPort == 0 {
		internalPort = containerPort
	}

	servicePort := v1.ServicePort{
		Protocol: protocol,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: containerPort,
		},
		Port: internalPort,
	}

	if len(port.External) > 0 {
		servicePort.NodePort = parseInt(port.External, 0)
	}

	return servicePort
}

func createService(context *podContext, port workflowsv1.Port, labels map[string]string) (*v1.Service, error) {
	servicePort := createServicePort(port)

	serviceType := v1.ServiceTypeClusterIP
	if len(port.External) > 0 {
		serviceType = v1.ServiceTypeNodePort
	}

	serviceName := port.Name
	if len(serviceName) == 0 {
		serviceName = workflowsv1.GenerateServiceName()
	}

	log.Debugf("Creating service %v", serviceName)
	return context.serviceClient.Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: v1.ServiceSpec{
			Type:     serviceType,
			Selector: labels,
			Ports:    []v1.ServicePort{servicePort},
		},
	})
}

func createServices(context *podContext, labels map[string]string) ([]*v1.Service, error) {
	creationSpec := context.creationSpec

	services := make([]*v1.Service, 0, len(creationSpec.Ports))
	for _, port := range creationSpec.Ports {
		service, err := createService(context, port, labels)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}
