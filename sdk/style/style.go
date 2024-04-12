package style

import (
	"github.com/null93/aws-knox/pkg/color"
)

var (
	YellowColor    = color.FromHex(0xFF9C0A)
	LightGrayColor = color.FromHex(0x909090)
	DarkGrayColor  = color.FromHex(0x606060)
	BlackColor     = color.FromHex(0x000000)
)

var (
	DefaultStyle         = color.NewStyle()
	TitleStyle           = color.NewStyle().WithBold(true)
	SubTitleStyle        = color.NewStyle().WithForeground(LightGrayColor)
	HeaderStyle          = color.NewStyle().WithBold(true)
	OptionStyle          = color.NewStyle()
	HighlightOptionStyle = color.NewStyle().WithForeground(BlackColor).WithBackground(YellowColor).WithBold(true)
	SearchTermStyle      = color.NewStyle()
	CursorStyle          = color.NewStyle().WithForeground(YellowColor).WithBlink(true)
)
