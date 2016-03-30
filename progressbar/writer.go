package progressbar

import (
	"fmt"
	"io"
)

type Writer struct {
	io.Writer
	output       io.Writer
	totalBytes   int64
	bytesWritten int64
}

func NewWriter(writer io.Writer, output io.Writer, totalBytes int64) *Writer {
	return &Writer{Writer: writer, output: output, totalBytes: totalBytes}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.bytesWritten += int64(n)

	if err == nil {
		draw(w.output, float64(w.bytesWritten), float64(w.totalBytes))
	} else if err == io.EOF {
		fmt.Fprintln(w.output)
	}

	return n, err
}
