package picker

import (
	"sort"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/null93/aws-knox/pkg/ansi"
	"github.com/null93/aws-knox/pkg/color"
	. "github.com/null93/aws-knox/sdk/style"
)

type picker struct {
	actions        []action
	options        []option
	filtered       []*option
	initialIndex   int
	selectedIndex  int
	term           string
	filterStrategy string
	title          string
	longestCols    []int
	emptyMessage   string
	maxHeight      int
	windowStart    int
	windowEnd      int
	headers        []string
}

type option struct {
	Columns []string
	Value   interface{}
}

type action struct {
	key         keys.KeyCode
	name        string
	description string
}

func NewPicker() *picker {
	p := picker{
		actions:        []action{},
		options:        []option{},
		filtered:       []*option{},
		initialIndex:   0,
		selectedIndex:  0,
		title:          "Please Pick One",
		term:           "",
		filterStrategy: "fuzzy",
		longestCols:    []int{},
		emptyMessage:   "Nothing Found",
		maxHeight:      5,
		windowStart:    0,
		windowEnd:      5,
		headers:        []string{},
	}
	return &p
}

func (p *picker) WithFilterStrategy(strategy string) {
	p.filterStrategy = strategy
}

func (p *picker) WithMaxHeight(maxHeight int) {
	p.maxHeight = maxHeight
	p.windowStart = 0
	p.windowEnd = maxHeight
}

func (p *picker) WithEmptyMessage(emptyMessage string) {
	p.emptyMessage = emptyMessage
}

func (p *picker) WithTitle(title string) {
	p.title = title
}

func (p *picker) WithInitialIndex(index int) {
	p.initialIndex = index
}

func (p *picker) WithHeaders(headers ...string) {
	p.headers = headers
	for i, header := range headers {
		if len(p.longestCols) <= i {
			p.longestCols = append(p.longestCols, 0)
		}
		if len(header) > p.longestCols[i] {
			p.longestCols[i] = len(header)
		}
	}
}

func (p *picker) AddOption(value interface{}, cols ...string) {
	o := option{
		Value:   value,
		Columns: cols,
	}
	for i, label := range cols {
		if len(p.longestCols) <= i {
			p.longestCols = append(p.longestCols, 0)
		}
		if len(label) > p.longestCols[i] {
			p.longestCols[i] = len(label)
		}
	}
	p.term = ""
	p.selectedIndex = 0
	p.options = append(p.options, o)
	p.filtered = append(p.filtered, &o)
}

func (p *picker) AddAction(key keys.KeyCode, name string, description string) {
	p.actions = append(p.actions, action{key, name, description})
}

func (p *picker) filter() {
	p.filtered = []*option{}
	p.selectedIndex = 0
	p.windowStart = 0
	p.windowEnd = p.maxHeight

	if p.filterStrategy == "fuzzy" {
		optionsMap := map[string]*option{}
		fullValues := []string{}
		for i, option := range p.options {
			if p.term == "" {
				p.filtered = append(p.filtered, &p.options[i])
				continue
			}
			fullValue := strings.Join(option.Columns, " ")
			fullValues = append(fullValues, fullValue)
			optionsMap[fullValue] = &p.options[i]
		}
		if p.term != "" {
			ranks := fuzzy.RankFindFold(p.term, fullValues)
			sort.Sort(ranks)
			for _, rank := range ranks {
				p.filtered = append(p.filtered, optionsMap[rank.Target])
			}
		}
	} else {
		for i, option := range p.options {
			if p.term == "" {
				p.filtered = append(p.filtered, &p.options[i])
				continue
			}
			for _, col := range option.Columns {
				if strings.Contains(strings.ToLower(col), strings.ToLower(p.term)) {
					p.filtered = append(p.filtered, &p.options[i])
					break
				}
			}
		}
	}

	if len(p.filtered) < 1 {
		p.selectedIndex = -1
	}
}

func (p *picker) render() {
	ansi.ClearDown()
	lightGray := color.ToForeground(LightGrayColor).Decorator()
	darkGray := color.ToForeground(DarkGrayColor).Decorator()
	DefaultStyle.Printfln("")
	TitleStyle.Printf(" %s", p.title)
	DefaultStyle.Printfln("")
	DefaultStyle.Printf(lightGray(" filter: %s", SearchTermStyle.Sprintf(p.term)))
	CursorStyle.Printfln("█")
	if p.windowStart > 0 {
		DefaultStyle.Printfln(" " + darkGray("…"))
	} else {
		DefaultStyle.Printfln("")
	}
	if len(p.headers) > 0 && len(p.filtered) > 0 {
		for i, header := range p.headers {
			HeaderStyle.Printf(" %-*s ", p.longestCols[i], header)
		}
		HeaderStyle.Printfln("")
	}
	if len(p.filtered) < 1 {
		SubTitleStyle.Printfln(" %s ", p.emptyMessage)
	}
	for index, option := range p.filtered {
		if index < p.windowStart || index >= p.windowEnd {
			continue
		}
		rowStyle := OptionStyle
		if index == p.selectedIndex {
			rowStyle = HighlightOptionStyle
		}
		for i, col := range option.Columns {
			rowStyle.Printf(" %-*s ", p.longestCols[i], col)
		}
		rowStyle.Printfln("")
	}
	if p.windowEnd < len(p.filtered) {
		DefaultStyle.Printfln(" " + darkGray("…"))
	} else {
		DefaultStyle.Printfln("")
	}
	helpMenu := darkGray(" %d/%d items •", len(p.filtered), len(p.options))
	helpMenu += lightGray(" ↑ ") + darkGray("up •")
	helpMenu += lightGray(" ↓ ") + darkGray("down •")
	helpMenu += lightGray(" enter ") + darkGray("choose •")
	for _, action := range p.actions {
		helpMenu += lightGray(" %s ", action.name) + darkGray("%s •", action.description)
	}
	helpMenu += lightGray(" ctl+c ") + darkGray("quit") + color.ResetStyle
	DefaultStyle.Printf(helpMenu)
	DefaultStyle.Printfln("")
	lines := len(p.filtered)
	if p.maxHeight < lines {
		lines = p.maxHeight
	}
	if len(p.headers) > 0 {
		lines++
	}
	ansi.MoveCursorUp(6 + lines)
}

func (p *picker) Pick(initialFilter string) (*option, *keys.KeyCode) {
	p.term = initialFilter
	p.filter()
	if len(p.filtered) > p.initialIndex {
		p.selectedIndex = p.initialIndex
	} else {
		p.selectedIndex = 0
	}
	ansi.HideCursor()
	defer ansi.ClearDown()
	defer ansi.ShowCursor()
	p.render()
	var firedActionKeyCode *keys.KeyCode
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		if key.Code == keys.CtrlC {
			p.selectedIndex = -1
			return true, nil
		}
		if key.Code == keys.Up {
			if p.selectedIndex > 0 {
				p.selectedIndex--
				if p.selectedIndex < p.windowStart {
					p.windowStart--
					p.windowEnd--
				}
				p.render()
			}
		}
		if key.Code == keys.Down {
			if p.selectedIndex < len(p.filtered)-1 {
				p.selectedIndex++
				if p.selectedIndex >= p.windowEnd {
					p.windowStart++
					p.windowEnd++
				}
				p.render()
			}
		}
		if key.Code == keys.Enter {
			if p.selectedIndex > -1 {
				return true, nil
			}
		}
		if key.Code == keys.RuneKey || key.Code == keys.Space {
			if p.term == "" && key.Code == keys.Space {
				return false, nil
			}
			p.term += string(key.Runes)
			p.filter()
			p.render()
		}
		if key.Code == keys.Backspace {
			if len(p.term) > 0 {
				p.term = p.term[:len(p.term)-1]
				p.filter()
				p.render()
			}
		}
		for _, action := range p.actions {
			if key.Code == action.key {
				firedActionKeyCode = &action.key
				return true, nil
			}
		}
		return false, nil
	})
	if p.selectedIndex < 0 {
		return nil, firedActionKeyCode
	}
	if p.selectedIndex >= len(p.filtered) {
		return nil, firedActionKeyCode
	}
	return p.filtered[p.selectedIndex], firedActionKeyCode
}
