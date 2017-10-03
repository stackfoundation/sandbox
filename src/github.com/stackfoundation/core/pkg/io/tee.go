package io

import "io"

type readCloser struct {
	io.Reader
	io.Closer
}

func (p *readCloser) Close() error {
	return p.Closer.Close()
}

func (p *readCloser) Read(dest []byte) (int, error) {
	return p.Reader.Read(dest)
}

// TeeReadCloser Create a io.ReadCloser that behaves like a io.TeeReader but also closes
func TeeReadCloser(original io.ReadCloser, writer io.Writer) io.ReadCloser {
	return &readCloser{
		Reader: io.TeeReader(original, writer),
		Closer: original,
	}
}
