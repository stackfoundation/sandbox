package controller

import (
	"context"
	"os"
	"os/signal"

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
		coordinator: coordinator,
	}, nil
}

// Execute Execute the specified workflow
func (c *executionController) Execute(context context.Context, workflow *v1.Workflow) {
	c.processTransitions()

	err := Execute(e, workflow)
	if err != nil {
		e.abort(err)
	}

	for _ = range e.change {
		err := Execute(e, e.workflow)
		if err != nil {
			e.abort(err)
		}
	}

	log.Debugf("Performing cleanup...")
	e.cleanupWaitGroup.Wait()
	log.Debugf("Finished cleanup")
}
}
