package processors

import (
	"bytes"
	"io"
	"strings"
	"unicode"

	coreio "github.com/stackfoundation/sandbox/core/pkg/io"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/buffer"
)

const workflowKeyword = "workflow"
const workflowDeclaration = workflowKeyword + "{"
const workflowDeclarationEnd = "}"
const maxWorkflowDeclarationLineSize = len(workflowKeyword) + 64 // Lots of flexibility for spacing

func removeWhitespace(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, text)
}

func isWorkflowDeclaration(line []byte) bool {
	rawDeclaration := removeWhitespace(string(line))
	return rawDeclaration == workflowDeclaration
}

func isWorkflowDeclarationEnd(line []byte) bool {
	rawDeclaration := removeWhitespace(string(line))
	return rawDeclaration == workflowDeclarationEnd
}

func workflowDeclarationLineLimitProducer() int {
	return maxWorkflowDeclarationLineSize
}

type workflowDetector struct {
	done           bool
	receiver       func(string)
	withinWorkflow bool
	workflowBuffer bytes.Buffer
}

func (d *workflowDetector) receive(line []byte) {
	if !d.done {
		if !d.withinWorkflow {
			if isWorkflowDeclaration(line) {
				d.withinWorkflow = true
			}
		} else {
			if isWorkflowDeclarationEnd(line) {
				d.withinWorkflow = false
				d.done = true

				d.receiver(d.workflowBuffer.String())

				d.workflowBuffer.Reset()
			} else {
				d.workflowBuffer.Write(line)
				d.workflowBuffer.WriteString("\n")
			}
		}
	}
}

func (d *workflowDetector) limit() int {
	if !d.done && d.withinWorkflow {
		return -1
	}

	return maxWorkflowDeclarationLineSize
}

// NewWorkflowDetector Create a new log processor that detects workflows declared in lines
func NewWorkflowDetector(reader io.ReadCloser, receiver func(string)) io.ReadCloser {
	lineReceiver := func([]byte) {}
	lineLimit := func() int { return maxWorkflowDeclarationLineSize }

	if receiver != nil {
		detector := &workflowDetector{
			receiver: receiver,
		}
		lineReceiver = detector.receive
		lineLimit = detector.limit
	}

	lineBuffer := buffer.NewLineBuffer(lineReceiver, lineLimit)

	return coreio.TeeReadCloser(reader, lineBuffer)
}
