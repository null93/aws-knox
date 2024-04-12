package color

import (
	"fmt"
)

type color24 struct {
	red   uint8
	green uint8
	blue  uint8
}

func (c *color24) String() string {
	return fmt.Sprintf("rgb(%d,%d,%d)", c.red, c.green, c.blue)
}

func (c *color24) attributes() []uint8 {
	return []uint8{2, c.red, c.green, c.blue}
}
