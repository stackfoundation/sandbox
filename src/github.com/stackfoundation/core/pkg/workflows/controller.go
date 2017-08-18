package workflows

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/stackfoundation/core/pkg/log"
)

type workflowController struct {
	cancel         context.CancelFunc
	context        context.Context
	cloner         *conversion.Cloner
	podsClient     *kubernetes.Clientset
	workflowClient *rest.RESTClient
}

func (controller *workflowController) deleteWorkflow(workflow *Workflow) error {
	log.Debugf(`Deleting workflow "%v"`, workflow.ObjectMeta.Name)
	return controller.workflowClient.Delete().
		Name(workflow.ObjectMeta.Name).
		Namespace(workflow.ObjectMeta.Namespace).
		Resource(WorkflowsPluralName).
		Do().
		Error()
}

func (controller *workflowController) saveWorkflow(workflow *Workflow) error {
	log.Debugf(`Saving updated workflow "%v" - status is now %v`, workflow.ObjectMeta.Name, workflow.Spec.Status.Status)
	return controller.workflowClient.Put().
		Name(workflow.ObjectMeta.Name).
		Namespace(workflow.ObjectMeta.Namespace).
		Resource(WorkflowsPluralName).
		Body(workflow).
		Do().
		Error()
}

func workflowStep(workflowSpec *WorkflowSpec, stepNumber int) (*WorkflowStep, string) {
	step := &workflowSpec.Steps[stepNumber]

	var stepName string
	if len(step.Name) > 0 {
		stepName = `"` + step.Name + `"`
	} else {
		stepName = "step " + strconv.Itoa(stepNumber)
	}

	return step, stepName
}

func (controller *workflowController) buildImageForStep(workflowSpec *WorkflowSpec, stepNumber int) error {
	step, stepName := workflowStep(workflowSpec, stepNumber)
	fmt.Println("Building image for " + stepName + ":")

	uuid := uuid.NewUUID()
	step.StepImage = "step:" + uuid.String()

	dockerClient, err := createDockerClient()
	if err != nil {
		return err
	}

	err = buildImage(controller.context, dockerClient, workflowSpec, step)
	if err != nil {
		return err
	}

	log.Debugf(`Image %v was built for step "%v"`, step.StepImage, stepName)

	workflowSpec.Status.Status = StatusStepImageBuilt
	return nil
}

func (controller *workflowController) runStepContainer(workflowSpec *WorkflowSpec, stepNumber int) error {
	step, stepName := workflowStep(workflowSpec, stepNumber)
	fmt.Println("Running " + stepName + ":")

	pods := controller.podsClient.Pods("default")

	uuid := uuid.NewUUID()
	podName := "pod-" + uuid.String()

	pod, err := createPod(pods, podName, step.StepImage, []string{"/bin/sh", "/app/" + step.StepScript})
	if err != nil {
		return err
	}

	printLogsUntilPodFinished(pods, pod)

	workflowSpec.Status.Status = StatusStepFinished
	return nil
}

func (controller *workflowController) proceedToNextStep(workflow *Workflow) error {
	log.Debugf(`Proceeding to next step in workflow "%v"`, workflow.ObjectMeta.Name)

	if len(workflow.Spec.Status.Status) > 0 {
		if StatusStepFinished == workflow.Spec.Status.Status {
			controller.buildImageForStep(&workflow.Spec, workflow.Spec.Status.Step)
			return controller.saveWorkflow(workflow)
		} else if StatusStepImageBuilt == workflow.Spec.Status.Status {
			controller.runStepContainer(&workflow.Spec, workflow.Spec.Status.Step)

			workflow.Spec.Status.Step++
			if workflow.Spec.Status.Step >= len(workflow.Spec.Steps) {
				workflow.Spec.Status.Status = StatusFinished
			}

			return controller.saveWorkflow(workflow)
		} else if StatusFinished == workflow.Spec.Status.Status {
			// Delete workflow
			controller.deleteWorkflow(workflow)
			controller.cancel()
		}
	} else if len(workflow.Spec.Steps) > 0 {
		err := controller.buildImageForStep(&workflow.Spec, 0)
		if err != nil {
			panic(err)
		}

		return controller.saveWorkflow(workflow)
	}

	return nil
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

func (controller *workflowController) processWorkflow(workflow *Workflow) {
	copyObj, err := controller.cloner.DeepCopy(workflow)
	if err != nil {
		return
	}

	workflowCopy := copyObj.(*Workflow)
	controller.proceedToNextStep(workflowCopy)
}

func (controller *workflowController) workflowAdded(obj interface{}) {
	controller.processWorkflow(obj.(*Workflow))
}

func (controller *workflowController) workflowUpdated(oldObj, newObj interface{}) {
	controller.processWorkflow(newObj.(*Workflow))
}
