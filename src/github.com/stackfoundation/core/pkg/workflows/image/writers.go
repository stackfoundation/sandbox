package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/ioutils/writers.go
// Licensed under the Apache License Version 2.0

import "io"

type writeCloserWrapper struct {
        io.Writer
        closer func() error
}

func (r *writeCloserWrapper) Close() error {
        return r.closer()
}

// NewWriteCloserWrapper returns a new io.WriteCloser.
func NewWriteCloserWrapper(r io.Writer, closer func() error) io.WriteCloser {
        return &writeCloserWrapper{
                Writer: r,
                closer: closer,
        }
}
