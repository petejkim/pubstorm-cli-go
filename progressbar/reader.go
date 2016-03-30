package progressbar

import (
	"fmt"
	"io"
)

type Reader struct {
	io.Reader
	output     io.Writer
	totalBytes int64
	bytesRead  int64
}

func NewReader(reader io.Reader, output io.Writer, totalBytes int64) *Reader {
	return &Reader{Reader: reader, output: output, totalBytes: totalBytes}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.bytesRead += int64(n)

	if err == nil {
		draw(r.output, float64(r.bytesRead), float64(r.totalBytes))
	} else if err == io.EOF {
		fmt.Fprintln(r.output)
	}

	return n, err
}
