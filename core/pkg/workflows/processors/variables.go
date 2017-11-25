package processors

import (
	"bytes"
	"io"
	"strings"

	coreio "github.com/stackfoundation/sandbox/core/pkg/io"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/buffer"
)

const maxVariableNameLength = 256
const maxVariableValueLength = 1024
const maxVariableDeclarationLineSize = maxVariableNameLength + maxVariableValueLength + 256 // Lots of flexibility for spacing

const varKeyword = "var"

func isVariableKeyword(rawKeyword string) bool {
	keyword := strings.TrimSpace(rawKeyword)
	return keyword == varKeyword
}

func cleanVariableName(rawName string) string {
	variableName := strings.TrimSpace(rawName)
	if len(variableName) > maxVariableNameLength {
		return variableName[:maxVariableNameLength]
	}

	return variableName
}

func cleanVariableValue(rawValue string) string {
	if len(rawValue) > maxVariableValueLength {
		return rawValue[:maxVariableValueLength]
	}

	return rawValue
}

func extractVariable(line []byte) (string, string) {
	separator := bytes.IndexByte(line, '=')
	if separator > 0 {
		rawDeclaration := line[:separator]

		variableDeclaration := strings.Split(string(rawDeclaration), " ")
		if len(variableDeclaration) == 2 {
			if isVariableKeyword(variableDeclaration[0]) {
				variableName := cleanVariableName(variableDeclaration[1])
				variableValue := cleanVariableValue(string(line[separator+1:]))

				return variableName, variableValue
			}
		}
	}
	return "", ""
}

func variableDeclarationLineLimitProducer() int {
	return maxVariableDeclarationLineSize
}

// NewVariableDetector Create a new log processor that detects variables declared in lines
func NewVariableDetector(reader io.ReadCloser, receiver func(string, string)) io.ReadCloser {
	lineReceiver := func([]byte) {}
	if receiver != nil {
		lineReceiver = func(line []byte) {
			name, value := extractVariable(line)
			if len(name) > 0 {
				receiver(name, value)
			}
		}
	}

	lineBuffer := buffer.NewLineBuffer(lineReceiver, variableDeclarationLineLimitProducer)

	return coreio.TeeReadCloser(reader, lineBuffer)
}
