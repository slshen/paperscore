package boxscore

import (
	"bufio"
	"fmt"
	"strings"
	"text/template"

	"github.com/slshen/paperscore/pkg/text"
)

func paste(c1, c2 string, sepWidth, leftLen int) string {
	switch {
	case leftLen > 0:
		c1 = text.Wrap(c1, leftLen)
		c2 = text.Wrap(c2, leftLen)
	case leftLen < 0:
		leftLen = -leftLen
	default:
		leftLen = lineLength(c1)
	}
	scan1 := bufio.NewScanner(strings.NewReader(c1))
	scan2 := bufio.NewScanner(strings.NewReader(c2))
	s := &strings.Builder{}
	for {
		s1, l1 := scanAndText(scan1)
		s2, l2 := scanAndText(scan2)
		if !(s1 || s2) {
			break
		}
		fmt.Fprintf(s, "%-*s", leftLen, l1)
		if l2 != "" {
			fmt.Fprintf(s, "%*s%s", sepWidth, "", l2)
		}
		fmt.Fprintln(s)
	}
	return s.String()
}

func scanAndText(s *bufio.Scanner) (bool, string) {
	if s.Scan() {
		return true, s.Text()
	}
	return false, ""
}

func lineLength(s string) int {
	ln := 0
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > ln {
			ln = len(line)
		}
	}
	return ln
}

func executeFunc(tmpl *template.Template) func(string, interface{}) (string, error) {
	return func(name string, data interface{}) (string, error) {
		var s strings.Builder
		err := tmpl.ExecuteTemplate(&s, name, data)
		return s.String(), err
	}
}
