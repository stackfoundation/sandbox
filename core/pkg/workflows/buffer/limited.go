package buffer

import "bytes"

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

type limitedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *limitedBuffer) append(src []byte) (int, error) {
	if b.limit > 0 {
		if b.buffer.Len() < b.limit {
			return b.buffer.Write(src[:min(len(src), b.limit-b.buffer.Len())])
		}
	} else {
		return b.buffer.Write(src)
	}

	return 0, nil
}

func (b *limitedBuffer) bytes() []byte {
	return b.buffer.Bytes()
}

func (b *limitedBuffer) reset() {
	b.buffer.Reset()
}

func (b *limitedBuffer) setLimit(limit int) {
	if limit > 0 {
		b.buffer.Grow(limit)
	}

	b.limit = limit
}
