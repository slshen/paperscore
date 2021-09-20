package game

import "strings"

type Narrative struct {
	strings.Builder
	Separator string
}

func (n *Narrative) Write(s string) {
	if len(s) == 0 {
		return
	}
	if n.Len() > 0 {
		if len(n.Separator) == 0 {
			n.WriteRune(' ')
		} else {
			n.WriteString(n.Separator)
		}
	}
	n.WriteString(s)
}
