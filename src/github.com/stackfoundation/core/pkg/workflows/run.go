package workflows

import (
        "bytes"

        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
        "k8s.io/client-go/pkg/api/v1"
        "fmt"
)

func writeFromInstruction(dockerfile *bytes.Buffer, step *WorkflowStep) {
        dockerfile.WriteString("FROM ")
        dockerfile.WriteString(step.Image)
        if len(step.Tag) > 0 {
                dockerfile.WriteString(":")
                dockerfile.WriteString(step.Tag)
        }
        dockerfile.WriteString("\n")
}

func writeVariables(dockerfile *bytes.Buffer, step *WorkflowStep) {
        if step.Variables != nil && len(step.Variables) > 0 {
                dockerfile.WriteString("ENV")
                for _, variable := range step.Variables {
                        dockerfile.WriteString(" ")
                        dockerfile.WriteString(variable.Name)
                        dockerfile.WriteString("=\"")
                        dockerfile.WriteString(variable.Value)
                        dockerfile.WriteString("\"")
                }
                dockerfile.WriteString("\n")
        }
}

func writePorts(dockerfile *bytes.Buffer, step *WorkflowStep) {
        if step.Ports != nil && len(step.Ports) > 0 {
                dockerfile.WriteString("EXPOSE")
                for _, port := range step.Ports {
                        dockerfile.WriteString(" ")
                        dockerfile.WriteString(port)
                }
                dockerfile.WriteString("\n")
        }
}

func buildDockerfile(step *WorkflowStep) string {
        var dockerfile bytes.Buffer

        writeFromInstruction(&dockerfile, step)
        writeVariables(&dockerfile, step)
        writePorts(&dockerfile, step)

        return dockerfile.String()
}

func uploadWorkflow(workflow *Workflow) error {
        workflowResource := make(map[string]interface{})
        workflowResource["apiVersion"] = "stack.foundation/v1"
        workflowResource["kind"] = "Workflow"
        workflowResource["spec"] = workflow.Spec
        workflowResource["metadata"] = workflow.ObjectMeta

        data := unstructured.Unstructured{
                Object: workflowResource,
        }

        client, err := createDynamicClient()
        if err != nil {
                return err
        }

        _, err = client.Resource(&metav1.APIResource{Name: "workflows", Namespaced: true}, v1.NamespaceDefault).
                Create(&data)

        if err != nil {
                return err
        }

        return nil
}

func RunCommand(workflowName string) error {
        workflow, err := readWorkflow(workflowName)
        if err != nil {
                return err
        }

        clientSet, err := createExtensionsClient()
        if err != nil {
                return err
        }

        fmt.Println("Creating workflow resource type definition")
        createWorkflowResourceIfRequired(clientSet.CustomResourceDefinitions())
        fmt.Println("Uploading workflow")
        err = uploadWorkflow(workflow)
        if err != nil {
                panic(err)
        }

        fmt.Println("Running workflow controller")
        RunController()

        return nil
}
