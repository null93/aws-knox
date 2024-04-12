package color

import (
	"fmt"
)

func serializeAttributes(attributes []uint8) string {
	result := ""
	delimiter := ""
	for _, attr := range attributes {
		result += fmt.Sprintf("%s%d", delimiter, attr)
		delimiter = ";"
	}
	return result
}
