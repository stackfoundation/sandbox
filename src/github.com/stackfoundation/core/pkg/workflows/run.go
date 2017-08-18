package workflows

import (
	"bytes"
	"context"
	"os"
	"os/signal"

	"k8s.io/apimachinery/pkg/conversion"

	log "github.com/stackfoundation/core/pkg/log"
)

func writeFromInstruction(dockerfile *bytes.Buffer, step *WorkflowStep) {
	if step.ImageSource == SourceCatalog || step.ImageSource == SourceManual {
		dockerfile.WriteString("FROM ")
		dockerfile.WriteString(step.Image)
		if len(step.Tag) > 0 {
			dockerfile.WriteString(":")
			dockerfile.WriteString(step.Tag)
		}
	} else if step.ImageSource == SourceStep {
		// Use previous step image
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
	dockerfile.WriteString("COPY . /app/")

	return dockerfile.String()
}

func uploadWorkflow(workflow *Workflow) error {
	log.Debugf(`Uploading workflow "%v"`, workflow.ObjectMeta.Name)

	client, err := createRestClient()
	if err != nil {
		return err
	}

	err = client.Post().
		Name(workflow.ObjectMeta.Name).
		Namespace(workflow.ObjectMeta.Namespace).
		Resource(WorkflowsPluralName).
		Body(workflow).
		Do().
		Error()

	if err != nil {
		return err
	}

	return nil
}

func runWorkflowController() error {
	log.Debugf("Starting workflow controller")
	ctx, cancelFunc := context.WithCancel(context.Background())

	controller := workflowController{
		cloner:  conversion.NewCloner(),
		context: ctx,
		cancel:  cancelFunc,
	}

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		for _ = range interruptChannel {
			log.Debugf("An interrupt was requested, stopping controller!")
			cancelFunc()
		}
	}()

	return controller.run()
}

// RunCommand Run a workflow in the current project
func RunCommand(workflowName string) error {
	workflow, err := readWorkflow(workflowName)
	if err != nil {
		return err
	}

	clientSet, err := createExtensionsClient()
	if err != nil {
		return err
	}

	err = createWorkflowResourceIfRequired(clientSet.CustomResourceDefinitions())
	if err != nil {
		return err
	}

	err = uploadWorkflow(workflow)
	if err != nil {
		return err
	}

	return runWorkflowController()
}
