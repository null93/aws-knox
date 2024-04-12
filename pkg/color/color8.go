package color

import (
	"fmt"
)

// @see https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit

type color8 struct {
	index uint8
}

func (c *color8) String() string {
	return fmt.Sprintf("8-bit-color(%d)", c.index)
}

func (c *color8) attributes() []uint8 {
	return []uint8{5, c.index}
}
