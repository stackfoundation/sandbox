package progress

import (
	"fmt"
	"io"
	"time"

	"github.com/stackfoundation/sandbox/log"
)

const barWidth = 52

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type progressAwareReader struct {
	io.Reader
	title        string
	code         string
	total        int64
	current      int64
	lastProgress int64
}

func (reader *progressAwareReader) Read(p []byte) (int, error) {
	n, err := reader.Reader.Read(p)
	if n > 0 {
		reader.current += int64(n)
	}

	now := timestamp()
	if reader.current == reader.total || (n > 0 && (now-reader.lastProgress) > 100) {
		reader.lastProgress = now

		if log.JSONOutput {
			fmt.Printf("{\"code\":\"%v\",\"message\":\"%v\",\"current\":%v,\"total\":%v}",
				reader.code, reader.title, reader.current, reader.total)
			fmt.Println()
		} else {
			fmt.Printf("%v [", reader.title)
			position := int(barWidth * (float64(reader.current) / float64(reader.total)))
			for i := 0; i < barWidth; i++ {
				if i < position {
					fmt.Printf("=")
				} else if i == position {
					fmt.Printf(">")
				} else {
					fmt.Printf(" ")
				}
			}

			fmt.Printf("] %.2f / %.2f MB\r", float32(reader.current)/1048576.0, float32(reader.total)/1048576.0)
		}
	}

	return n, err
}

// NewProgressAwareReader Create a new reader which shows progress as the read proceeds
func NewProgressAwareReader(reader io.Reader, title string, code string, total int64) io.Reader {
	return &progressAwareReader{
		Reader:       reader,
		title:        title,
		code:         code,
		total:        total,
		lastProgress: timestamp(),
	}
}
