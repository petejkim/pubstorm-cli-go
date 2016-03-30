package progressbar

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/nitrous-io/rise-cli-go/tui"
)

func draw(w io.Writer, i, total float64) {
	barWidth := 40
	percentage := i / total
	progress := int(math.Floor(percentage * float64(barWidth)))

	fmt.Fprintf(w, "\r["+tui.Blu("%s")+"%s] "+tui.Cyn("%.1f%%"), strings.Repeat("=", progress), strings.Repeat(" ", barWidth-progress), percentage*100.0)
}
