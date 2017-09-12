package processors

import (
	"bytes"
	"io"
)

type prefixer struct {
	io.ReadCloser
	prefix         []byte
	buffer         []byte
	read           int
	prefixWritePos int
	eof            bool
}

func (p *prefixer) copyBufferPortion(dest []byte, destStart, srcEnd int) int {
	n := copy(dest[destStart:], p.buffer[p.read:srcEnd])
	destStart += n
	p.read += n

	return destStart
}

func (p *prefixer) copyNextPrefixedLine(dest []byte, destPos int) int {
	lineBreak := bytes.IndexByte(p.buffer[p.read:], '\n')
	if lineBreak > -1 {
		destPos = p.copyBufferPortion(dest, destPos, p.read+lineBreak+1)

		if destPos < len(dest) && p.read < len(p.buffer) {
			p.prefixWritePos = 0
			destPos += p.copyPrefix(dest[destPos:])
		} else {
			p.prefixWritePos = 0
		}
	} else {
		destPos = p.copyBufferPortion(dest, destPos, len(p.buffer))
	}

	return destPos
}

func (p *prefixer) copyPrefix(dest []byte) int {
	if len(p.prefix) > 0 {
		prefixWriteSize := copy(dest, p.prefix[p.prefixWritePos:])
		p.prefixWritePos = len(p.prefix) - prefixWriteSize

		if p.prefixWritePos == 0 {
			p.prefixWritePos = -1
		}

		return prefixWriteSize
	}

	return 0
}

func (p *prefixer) readIntoBuffer() error {
	p.buffer = p.buffer[:cap(p.buffer)]

	n, err := p.ReadCloser.Read(p.buffer)
	if n > 0 {
		p.read = 0
		p.buffer = p.buffer[:n]
	} else {
		p.buffer = p.buffer[:0]
	}

	return err
}

func (p *prefixer) Close() error {
	return p.ReadCloser.Close()
}

func (p *prefixer) Read(dest []byte) (int, error) {
	destPos := 0

	for {
		if p.prefixWritePos > -1 && p.read < len(p.buffer) {
			destPos += p.copyPrefix(dest)
		}

		for p.read < len(p.buffer) {
			destPos = p.copyNextPrefixedLine(dest, destPos)
			if destPos == len(dest) {
				return destPos, nil
			}
		}

		if destPos > 0 {
			return destPos, nil
		}

		if !p.eof {
			err := p.readIntoBuffer()
			if io.EOF == err {
				p.eof = true
			}
		} else {
			break
		}
	}

	return destPos, io.EOF
}

// NewPrefixer Create a new log processor that adds a prefix to each line
func NewPrefixer(reader io.ReadCloser, prefix string) io.ReadCloser {
	return &prefixer{
		ReadCloser:     reader,
		prefix:         []byte(prefix),
		prefixWritePos: 0,
		buffer:         make([]byte, 0, 65536),
	}
}
