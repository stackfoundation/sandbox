package buffer

import (
	"bytes"
	"io"
)

type lineBuffer struct {
	buffer        *limitedBuffer
	lineReceiver  func([]byte)
	limitProducer func() int
}

func (b *lineBuffer) handleLineEnd() {
	b.lineReceiver(b.buffer.bytes())
	b.buffer.reset()
	b.buffer.setLimit(b.limitProducer())
}

func (b *lineBuffer) Write(src []byte) (int, error) {
	srcPos := 0
	srcLength := len(src)

	for srcPos < srcLength {
		lineBreak := bytes.IndexByte(src[srcPos:], '\n')
		if lineBreak > -1 {
			lineEnd := srcPos + lineBreak

			_, err := b.buffer.append(src[srcPos:lineEnd])
			if err != nil {
				return 0, err
			}

			b.handleLineEnd()

			srcPos = lineEnd + 1
		} else {
			n, err := b.buffer.append(src[srcPos:])
			if err != nil {
				return 0, err
			}

			srcPos += n
		}
	}

	return srcLength, nil
}

// NewLineBuffer Create line buffer which sends specified receiver of lines and uses the given limit producer
func NewLineBuffer(lineReceiver func([]byte), limitProducer func() int) io.Writer {
	if lineReceiver == nil {
		lineReceiver = func([]byte) {}
	}

	if limitProducer == nil {
		limitProducer = func() int { return -1 }
	}

	buffer := &limitedBuffer{}
	buffer.setLimit(limitProducer())

	return &lineBuffer{
		buffer:        buffer,
		lineReceiver:  lineReceiver,
		limitProducer: limitProducer,
	}
}
