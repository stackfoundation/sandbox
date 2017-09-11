package util

import (
	"bytes"
	"io"
	"strings"
)

const maxVariableNameLength = 256
const maxVariableValueLength = 1024
const maxLineSize = maxVariableNameLength + maxVariableValueLength + 256 // Lots of flexibility for spacing

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

type detector struct {
	io.Reader
	io.Closer
}

type detectionBuffer struct {
	lineBuffer bytes.Buffer
	receiver   func(string, string)
}

func (p *detector) Close() error {
	return p.Closer.Close()
}

func (p *detector) Read(dest []byte) (int, error) {
	return p.Reader.Read(dest)
}

func (b *detectionBuffer) copyToLineBuffer(src []byte) (int, error) {
	if b.lineBuffer.Len() < maxLineSize {
		return b.lineBuffer.Write(src[:min(len(src), maxLineSize-b.lineBuffer.Len())])
	}

	return 0, nil
}

func (b *detectionBuffer) detectVariable() {
	lineBytes := b.lineBuffer.Bytes()

	separator := bytes.IndexByte(lineBytes, '=')
	if separator > 0 {
		rawName := lineBytes[:separator]

		variableName := strings.TrimSpace(string(rawName))
		if len(variableName) > maxVariableNameLength {
			variableName = variableName[:maxVariableNameLength]
		}

		variableValue := string(lineBytes[separator+1:])
		if len(variableValue) > maxVariableValueLength {
			variableValue = variableValue[:maxVariableValueLength]
		}

		b.receiver(variableName, variableValue)
	}
}

func (b *detectionBuffer) Write(src []byte) (int, error) {
	srcLength := len(src)
	if srcLength > 0 {
		srcPos := 0
		for srcPos < srcLength {
			lineBreak := bytes.IndexByte(src[srcPos:], '\n')
			if lineBreak > -1 {
				lineEnd := srcPos + lineBreak

				_, err := b.copyToLineBuffer(src[srcPos:lineEnd])
				if err != nil {
					return 0, err
				}

				b.detectVariable()
				b.lineBuffer.Reset()

				srcPos = lineEnd + 1
			} else {
				n, err := b.copyToLineBuffer(src[srcPos:])
				if err != nil {
					return 0, err
				}

				srcPos += n
			}
		}
	}

	return srcLength, nil
}

// NewDetector Create a new log processor that detects variables declared in lines
func NewDetector(reader io.ReadCloser, receiver func(string, string)) io.ReadCloser {
	if receiver == nil {
		receiver = func(string, string) {}
	}

	buffer := &detectionBuffer{
		receiver: receiver,
	}
	buffer.lineBuffer.Grow(maxLineSize)

	return &detector{
		Reader: io.TeeReader(reader, buffer),
		Closer: reader,
	}
}
