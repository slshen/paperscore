package text

import "strings"

func Center(s string, w int) string {
	if len(s) >= w {
		return s
	}
	b := &strings.Builder{}
	n := w - len(s)
	b.WriteString(strings.Repeat(" ", n/2))
	b.WriteString(s)
	b.WriteString(strings.Repeat(" ", w-n/2-len(s)))
	return b.String()
}
