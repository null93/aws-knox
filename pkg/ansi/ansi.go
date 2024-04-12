package ansi

import (
	"fmt"
	"io"
	"os"
)

var (
	Writer io.Writer = os.Stdout
)

func HideCursor() {
	fmt.Fprint(Writer, "\033[?25l")
}

func ShowCursor() {
	fmt.Fprint(Writer, "\033[?25h")
}

func SaveCursor() {
	fmt.Fprint(Writer, "\033[s")
}

func RestoreCursor() {
	fmt.Fprint(Writer, "\033[u")
}

func ClearDown() {
	fmt.Fprint(Writer, "\033[J")
}

func ClearRight() {
	fmt.Fprint(Writer, "\033[K")
}

func AlternativeBuffer() {
	fmt.Fprint(Writer, "\033[?1048h")
}

func NormalBuffer() {
	fmt.Fprint(Writer, "\033[?1049l")
}
