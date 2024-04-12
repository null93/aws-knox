package color

import (
	"fmt"
	"io"
	"os"
)

const (
	ResetStyle = "\033[0m"
)

var (
	Writer io.Writer = os.Stdout
)

type style struct {
	attributes map[uint8]bool
	fgColor    colorful
	bgColor    colorful
}

func NewStyle() *style {
	return &style{
		attributes: map[uint8]bool{},
		fgColor:    nil,
		bgColor:    nil,
	}
}

func (s *style) WithForeground(c colorful) *style {
	s.fgColor = c
	return s
}

func (s *style) WithBackground(c colorful) *style {
	s.bgColor = c
	return s
}

func (s *style) WithBold(value bool) *style {
	if value {
		s.attributes[1] = true
	} else {
		delete(s.attributes, 1)
	}
	return s
}

func (s *style) WithUnderline(value bool) *style {
	if value {
		s.attributes[4] = true
	} else {
		delete(s.attributes, 4)
	}
	return s
}

func (s *style) WithBlink(value bool) *style {
	if value {
		s.attributes[5] = true
	} else {
		delete(s.attributes, 5)
	}
	return s
}

func (s *style) WithInverse(value bool) *style {
	if value {
		s.attributes[7] = true
	} else {
		delete(s.attributes, 7)
	}
	return s
}

func (s *style) Decorator() func(string, ...interface{}) string {
	return func(text string, args ...interface{}) string {
		return s.Sprintf(text, args...)
	}
}

func (s *style) Sprintf(format string, args ...interface{}) string {
	styled := ""
	attributes := []uint8{}
	for k, _ := range s.attributes {
		attributes = append(attributes, k)
	}
	if s.fgColor != nil {
		attributes = append(attributes, uint8(38))
		attributes = append(attributes, s.fgColor.attributes()...)
	} else {
		attributes = append(attributes, uint8(39))
	}
	if s.bgColor != nil {
		attributes = append(attributes, uint8(48))
		attributes = append(attributes, s.bgColor.attributes()...)
	} else {
		attributes = append(attributes, uint8(49))
	}
	styled += fmt.Sprintf("\033[%sm", serializeAttributes(attributes))
	return fmt.Sprintf(styled+format+ResetStyle, args...)
}

func (s *style) Sprintfln(format string, args ...interface{}) string {
	return s.Sprintf(format, args...) + "\n"
}

func (s *style) Fprintf(writer io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(writer, s.Sprintf(format, args...))
}

func (s *style) Fprintfln(writer io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(writer, s.Sprintf(format, args...)+"\n")
}

func (s *style) Printf(format string, args ...interface{}) {
	fmt.Fprintf(Writer, s.Sprintf(format, args...))
}

func (s *style) Printfln(format string, args ...interface{}) {
	fmt.Fprintf(Writer, s.Sprintf(format, args...)+"\n")
}
