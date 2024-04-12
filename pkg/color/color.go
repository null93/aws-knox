package color

type colorful interface {
	attributes() []uint8
}

func FromRGB(r, g, b uint8) *color24 {
	return &color24{red: r, green: g, blue: b}
}

func FromHex(hex int) *color24 {
	r := uint8(hex & 0x00FF0000 >> 16)
	g := uint8(hex & 0x0000FF00 >> 8)
	b := uint8(hex & 0x000000FF)
	return &color24{red: r, green: g, blue: b}
}

func FromIndex(index uint8) *color8 {
	index = index & 0xFF
	return &color8{index}
}

func ToForeground(c colorful) *style {
	return &style{fgColor: c}
}

func ToBackground(c colorful) *style {
	return &style{bgColor: c}
}
