package text

import (
	"fmt"
	"io"
	"strings"
)

type Column struct {
	Header string
	Width  int
	Number bool
	Left   bool
}

type Table struct {
	Columns []Column
	format  string
}

func (t *Table) Header() string {
	var s strings.Builder
	for i, col := range t.Columns {
		center(&s, col.Header, col.Width)
		if i < len(t.Columns)-1 {
			s.WriteRune(' ')
		} else {
			s.WriteRune('\n')
		}
	}
	return s.String()
}

func (t *Table) Format() string {
	if t.format == "" {
		var s strings.Builder
		for _, col := range t.Columns {
			s.WriteRune('%')
			if col.Left {
				s.WriteRune('-')
			}
			fmt.Fprintf(&s, "%d", col.Width)
			s.WriteRune('v')
			s.WriteRune(' ')
		}
		s.WriteRune('\n')
		t.format = s.String()
	}
	return t.format
}

func (t *Table) Fprint(w io.Writer, args ...interface{}) {
	for len(args) < len(t.Columns) {
		args = append(args, "")
	}
	fmt.Fprintf(w, t.Format(), args...)
}

func center(w io.Writer, s string, width int) {
	if len(s) >= width {
		fmt.Fprint(w, s[0:width])
		return
	}
	fmt.Fprintf(w, "%*s", -width, fmt.Sprintf("%*s", (width+len(s))/2, s))
}
