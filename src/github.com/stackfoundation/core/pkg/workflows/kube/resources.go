package kube

import (
	"errors"

	log "github.com/stackfoundation/core/pkg/log"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kubeerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func createWorkflowResourceDefinition(definitions extensionsclient.CustomResourceDefinitionInterface) error {
	_, err := definitions.Create(&extensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: workflowsv1.WorkflowsCustomResource,
		},
		Spec: extensions.CustomResourceDefinitionSpec{
			Group:   workflowsv1.WorkflowsGroupName,
			Scope:   extensions.NamespaceScoped,
			Version: workflowsv1.WorkflowsGroupVersion,
			Names: extensions.CustomResourceDefinitionNames{
				Plural:   workflowsv1.WorkflowsPluralName,
				Singular: workflowsv1.WorkflowsSingularName,
				Kind:     workflowsv1.WorkflowsKind,
				ShortNames: []string{
					"wf",
					"wflow",
				},
			},
		},
	})

	return err
}

// CreateWorkflowResourceDefinitionIfRequired Create the workflows CRD if it doesn't already exist
func CreateWorkflowResourceDefinitionIfRequired(definitions extensionsclient.CustomResourceDefinitionInterface) error {
	_, err := definitions.Get(workflowsv1.WorkflowsCustomResource, metav1.GetOptions{})
	if err != nil {
		log.Debugf("Creating custom resource definition for workflows in Kubernetes")
		err = createWorkflowResourceDefinition(definitions)

		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteWorkflow Delete a workflow resource
func DeleteWorkflow(client *rest.RESTClient, workflow *workflowsv1.Workflow) error {
	log.Debugf(`Deleting workflow "%v"`, workflow.ObjectMeta.Name)
	return client.Delete().
		Name(workflow.ObjectMeta.Name).
		Namespace(workflow.ObjectMeta.Namespace).
		Resource(workflowsv1.WorkflowsPluralName).
		Do().
		Error()
}

// WorkflowUpdater A function which updates the given workflow in some way
type WorkflowUpdater func(*workflowsv1.Workflow)

// UpdateWorkflow Updates a workflow resource, re-syncing if necessary
func UpdateWorkflow(client *rest.RESTClient, workflow *workflowsv1.Workflow, updater WorkflowUpdater) error {
	updater(workflow)

	err := client.Put().
		Name(workflow.ObjectMeta.Name).
		Namespace(workflow.ObjectMeta.Namespace).
		Resource(workflowsv1.WorkflowsPluralName).
		Body(workflow).
		Do().
		Error()

	if kubeerr.IsConflict(err) {
		log.Debugf(`Workflow "%v" was changed, need to re-sync`, workflow.ObjectMeta.Name)
		result := client.Get().
			Name(workflow.ObjectMeta.Name).
			Namespace(workflow.ObjectMeta.Namespace).
			Resource(workflowsv1.WorkflowsPluralName).
			Do()

		err = result.Error()
		if err != nil {
			return err
		}

		updatedWorkflowRaw, err := result.Get()
		if err != nil {
			return err
		}

		updatedWorkflow, ok := updatedWorkflowRaw.(*workflowsv1.Workflow)
		if !ok {
			return errors.New(`Invalid workflow retrieved while trying to re-sync "` +
				workflow.ObjectMeta.Name + `"`)
		}

		updater(updatedWorkflow)

		err = client.Put().
			Name(workflow.ObjectMeta.Name).
			Namespace(workflow.ObjectMeta.Namespace).
			Resource(workflowsv1.WorkflowsPluralName).
			Body(workflow).
			Do().
			Error()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	log.Debugf(`Workflow "%v" updated`, workflow.ObjectMeta.Name)
	return nil
}

func uploadWorkflow(client *rest.RESTClient, workflow *workflowsv1.Workflow) error {
	return client.Post().
		Name(workflow.ObjectMeta.Name).
		Namespace(workflow.ObjectMeta.Namespace).
		Resource(workflowsv1.WorkflowsPluralName).
		Body(workflow).
		Do().
		Error()
}

// UploadWorkflow Upload a workflow resource, deleting an existing one, if present
func UploadWorkflow(workflow *workflowsv1.Workflow) error {
	log.Debugf(`Uploading workflow "%v"`, workflow.ObjectMeta.Name)

	client, err := CreateWorkflowsClient()
	if err != nil {
		return err
	}

	err = uploadWorkflow(client, workflow)

	if kubeerr.IsConflict(err) {
		log.Debugf(`A workflow with name "%v" already exists, deleting it`, workflow.ObjectMeta.Name)
		err = DeleteWorkflow(client, workflow)
		if err != nil {
			return err
		}

		err = uploadWorkflow(client, workflow)
	}

	if err != nil {
		return err
	}

	return nil
}
