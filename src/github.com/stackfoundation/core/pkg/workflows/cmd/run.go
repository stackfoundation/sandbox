package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/stackfoundation/core/pkg/workflows/execution/controller"
	"github.com/stackfoundation/core/pkg/workflows/files"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

// Run Run a workflow in the current project
func Run(workflowName string, args []string) error {
	workflow, err := files.ReadWorkflow(workflowName)
	if err != nil {
		return err
	}

	err = v1.Validate(&workflow.Spec)
	if err != nil {
		return err
	}

	for i, arg := range args {
		workflow.Spec.State.Variables.Set("arg"+strconv.Itoa(i), arg)
	}

	combinedArgs := strings.Join(args, " ")
	workflow.Spec.State.Variables.Set("args", combinedArgs)

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
