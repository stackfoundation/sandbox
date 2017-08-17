package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/ioutils/readers.go
// Licensed under the Apache License Version 2.0

import (
        "io"
)

type readCloserWrapper struct {
        io.Reader
        closer func() error
}

func (r *readCloserWrapper) Close() error {
        return r.closer()
}

// NewReadCloserWrapper returns a new io.ReadCloser.
func NewReadCloserWrapper(r io.Reader, closer func() error) io.ReadCloser {
        return &readCloserWrapper{
                Reader: r,
                closer: closer,
        }
}
