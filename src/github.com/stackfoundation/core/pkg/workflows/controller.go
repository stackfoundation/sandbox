package workflows

import (
        "context"
        "fmt"

        "k8s.io/client-go/tools/cache"
        "k8s.io/apimachinery/pkg/conversion"
        "k8s.io/apimachinery/pkg/fields"
        "k8s.io/client-go/pkg/api/v1"
        "k8s.io/client-go/rest"
        "time"
)

type WorkflowController struct {
        context context.Context
        cloner *conversion.Cloner
        client *rest.RESTClient
}

func (controller *WorkflowController) deleteWorkflow(workflow *Workflow) error {
        return controller.client.Delete().
                Name(workflow.ObjectMeta.Name).
                Namespace(workflow.ObjectMeta.Namespace).
                Resource(WorkflowsPluralName).
                Do().
                Error()
}

func (controller *WorkflowController) saveWorkflow(workflow *Workflow) error {
        return controller.client.Put().
                Name(workflow.ObjectMeta.Name).
                Namespace(workflow.ObjectMeta.Namespace).
                Resource(WorkflowsPluralName).
                Body(workflow).
                Do().
                Error()
}

func (controller *WorkflowController) buildImageForStep(workflowSpec *WorkflowSpec, stepNumber int) error {
        step := workflowSpec.Steps[stepNumber]

        fmt.Println("Image is from " + step.ImageSource)

        if step.ImageSource == SourceCatalog || step.ImageSource == SourceManual {
                dockerClient, err := createDockerClient()
                if err != nil {
                        return err
                }

                err = buildImage(controller.context, dockerClient, workflowSpec, &step)
                if err != nil {
                        return err
                }
        }

        workflowSpec.Status.Status = StatusStepImageBuilt
        return nil
}

func (controller *WorkflowController) proceedToNextStep(workflow *Workflow) error {
        if len(workflow.Spec.Status.Status) > 0 {
                if (StatusStepFinished == workflow.Spec.Status.Status) {
                        controller.buildImageForStep(&workflow.Spec, workflow.Spec.Status.Step)
                        return controller.saveWorkflow(workflow)
                } else if (StatusStepImageBuilt == workflow.Spec.Status.Status) {
                        if workflow.Spec.Status.Step < len(workflow.Spec.Steps) {
                                workflow.Spec.Status.Status = StatusFinished
                        } else {
                                workflow.Spec.Status.Step++
                        }

                        return controller.saveWorkflow(workflow)
                } else if (StatusFinished == workflow.Spec.Status.Status) {
                        // Delete workflow
                        controller.deleteWorkflow(workflow)
                }
        } else if len(workflow.Spec.Steps) > 0 {
                err := controller.buildImageForStep(&workflow.Spec, 0)
                if err != nil {
                        panic(err)
                }
                fmt.Println("Workflow save")
                fmt.Println(workflow)
                return controller.saveWorkflow(workflow)
        }

        return nil
}

func RunController() {
        controller := WorkflowController{
                cloner: conversion.NewCloner(),
        }

        ctx, cancelFunc := context.WithCancel(context.Background())
        defer cancelFunc()
        go controller.Run(ctx)

        time.Sleep(10 * time.Second)
}

func (c *WorkflowController) Run(ctx context.Context) error {
        fmt.Print("Watch workflow objects\n")

        _, err := c.watchWorkflows(ctx)
        if err != nil {
                fmt.Printf("Failed to register watch for workflow resource: %v\n", err)
                return err
        }

        <-ctx.Done()
        return ctx.Err()
}

func (c *WorkflowController) watchWorkflows(ctx context.Context) (cache.Controller, error) {
        client, err := createRestClient()
        if err != nil {
                panic(err)
        }

        c.client = client

        source := cache.NewListWatchFromClient(
                client, WorkflowsPluralName, v1.NamespaceDefault, fields.Everything())

        _, controller := cache.NewInformer(
                source,
                &Workflow{},
                0,
                cache.ResourceEventHandlerFuncs{
                        AddFunc:    c.onAdd,
                        UpdateFunc: c.onUpdate,
                        DeleteFunc: c.onDelete,
                })

        go controller.Run(ctx.Done())
        return controller, nil
}

func (c *WorkflowController) onAdd(obj interface{}) {
        workflow := obj.(*Workflow)
        fmt.Printf("[CONTROLLER] OnAdd %s\n", workflow.Spec)

        copyObj, err := c.cloner.DeepCopy(workflow)
        if err != nil {
                fmt.Printf("ERROR creating a deep copy of example object: %v\n", err)
                return
        }

        workflowCopy := copyObj.(*Workflow)
        c.proceedToNextStep(workflowCopy)
}

func (c *WorkflowController) onUpdate(oldObj, newObj interface{}) {
        newWorkflow := newObj.(*Workflow)

        copyObj, err := c.cloner.DeepCopy(newWorkflow)
        if err != nil {
                fmt.Printf("ERROR creating a deep copy of example object: %v\n", err)
                return
        }

        workflowCopy := copyObj.(*Workflow)
        c.proceedToNextStep(workflowCopy)
}

func (c *WorkflowController) onDelete(obj interface{}) {
        example := obj.(*Workflow)
        fmt.Printf("[CONTROLLER] OnDelete %s\n", example.Spec)
}
