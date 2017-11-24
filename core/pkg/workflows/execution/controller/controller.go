package controller

import (
	"context"
	"fmt"
	"sync"

	executioncontext "github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/execution/coordinator"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

// NewController Create a new workflow execution controller
func NewController() (Controller, error) {
	coordinator, err := coordinator.NewCoordinator()
	if err != nil {
		return nil, err
	}

	return &executionController{
		coordinator:        coordinator,
		pendingTransitions: make(chan pendingTransition),
	}, nil
}

func (c *executionController) processTransitionsAndChanges(wc *executioncontext.WorkflowContext) {
	log.Debugf("Running workflow \"%v\"", wc.Workflow.Name)
	for {
		c.processTransitions(wc)
		err := c.processNextChange(wc)
		if err != nil {
			fmt.Println(err.Error())
			wc.Cancel()
			return
		}

		select {
		case transition := <-c.pendingTransitions:
			transition.perform()
		case <-wc.Context.Done():
			return
		}
	}
}

// Execute Execute the specified workflow
func (c *executionController) Execute(ctx context.Context, workflow *v1.Workflow) {
	cleanup := &sync.WaitGroup{}

	completion, cancel := context.WithCancel(ctx)
	wc := executioncontext.NewWorkflowContext(completion, cancel, cleanup, workflow)

	c.processTransitionsAndChanges(wc)

	log.Debugf("Performing cleanup...")
	cleanup.Wait()
	log.Debugf("Finished cleanup")
}
