package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/controller"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/files"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/validation"
	"github.com/stackfoundation/sandbox/log"
)

func addArgumentVariables(workflow *v1.Workflow, args []string) {
	if len(args) > 0 {
		for i, arg := range args {
			variable := "arg" + strconv.Itoa(i)
			workflow.Spec.State.Variables.Set(variable, arg)
		}

		combinedArgs := strings.Join(args, " ")
		workflow.Spec.State.Variables.Set("args", combinedArgs)
	}
}

// Run Run a workflow in the current project
func Run(workflowName string, args []string) error {
	workflow, err := files.ReadWorkflow(workflowName)
	if err != nil {
		return err
	}

	addArgumentVariables(workflow, args)

	err = validation.Validate(&workflow.Spec)
	if err != nil {
		return err
	}

	c, err := controller.NewController()
	if err != nil {
		return err
	}

	context, cancel := context.WithCancel(context.Background())

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		for _ = range interruptChannel {
			log.Debugf("An interrupt was requested, performing clean-up!")
			cancel()
		}
	}()

	c.Execute(context, workflow)
	return nil
}
