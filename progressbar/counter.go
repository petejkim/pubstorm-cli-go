package progressbar

import (
	"fmt"
	"io"
)

type Counter struct {
	output io.Writer
	total  int
	i      int
}

func NewCounter(output io.Writer, total int) *Counter {
	return &Counter{output: output, total: total}
}

func (c *Counter) Next() {
	if c.i == c.total {
		return
	}

	c.i++

	draw(c.output, float64(c.i), float64(c.total))

	if c.i == c.total {
		fmt.Fprintln(c.output)
	}

	return
}
