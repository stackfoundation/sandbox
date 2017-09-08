package controller

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/execution"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
)

type workflowController struct {
	cancel         context.CancelFunc
	context        context.Context
	cleanup        sync.WaitGroup
	cloner         *conversion.Cloner
	podsClient     *kubernetes.Clientset
	workflowClient *rest.RESTClient
}

func (controller *workflowController) run() error {
	_, err := controller.watchWorkflows()
	if err != nil {
		return err
	}

	clientSet, err := createKubeClient()
	if err != nil {
		return err
	}
	controller.podsClient = clientSet

	<-controller.context.Done()
	controller.cleanup.Wait()
	return controller.context.Err()
}

func (controller *workflowController) watchWorkflows() (cache.Controller, error) {
	client, err := createRestClient()
	if err != nil {
		return nil, err
	}

	controller.workflowClient = client

	source := cache.NewListWatchFromClient(
		client, WorkflowsPluralName, v1.NamespaceDefault, fields.Everything())

	_, cacheController := cache.NewInformer(
		source,
		&Workflow{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.workflowAdded,
			UpdateFunc: controller.workflowUpdated,
		})

	go cacheController.Run(controller.context.Done())
	return cacheController, nil
}

func (controller *workflowController) processWorkflow(workflow *workflowsv1.Workflow) {
	copyObj, err := controller.cloner.DeepCopy(workflow)
	if err != nil {
		return
	}

	workflowCopy := copyObj.(*Workflow)
	execution.ExecuteNextStep(controller, workflowCopy)
}

func (controller *workflowController) workflowAdded(obj interface{}) {
	controller.processWorkflow(obj.(*Workflow))
}

func (controller *workflowController) workflowUpdated(oldObj, newObj interface{}) {
	controller.processWorkflow(newObj.(*Workflow))
}

// RunWorkflowController Start and run the workflow controller
func RunWorkflowController() error {
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
