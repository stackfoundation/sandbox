package workflows

import (
	"context"
	"os"
	"os/signal"

	"k8s.io/apimachinery/pkg/conversion"

	log "github.com/stackfoundation/core/pkg/log"
)

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
