package progressbar

import (
	"fmt"
	"io"
	"math"
	"strings"
)

func draw(w io.Writer, i, total float64) {
	barWidth := 40
	percentage := i / total
	progress := int(math.Floor(percentage * float64(barWidth)))

	fmt.Fprintf(w, "\r[%s%s] %.1f%%", strings.Repeat("=", progress), strings.Repeat(" ", barWidth-progress), percentage*100.0)
}
