package print

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var (
	greenText    = color.New(color.FgGreen)
	yellowText   = color.New(color.FgYellow)
	cyanBoldText = color.New(color.FgCyan, color.Bold)
	cyanText     = color.New(color.FgCyan)
	redBoldText  = color.New(color.FgRed, color.Bold)
)

func SuccessStatusEvent(w io.Writer, message string) {
	fmt.Fprintf(w, "    ✅ %s\n", message)
}
func FailureStatusEvent(w io.Writer, message string) {
	fmt.Fprintf(w, "    ❌ %s\n", message)
}
