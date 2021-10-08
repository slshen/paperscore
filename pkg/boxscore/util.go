package boxscore

import (
	"bufio"
	"fmt"
	"strings"
	"text/template"
)

var ordinalDictionary = map[int]string{
	0: "th",
	1: "st",
	2: "nd",
	3: "rd",
	4: "th",
	5: "th",
	6: "th",
	7: "th",
	8: "th",
	9: "th",
}

func firstWord(s string, w int) string {
	out := &strings.Builder{}
	for s != "" {
		space := strings.IndexRune(s, ' ')
		if space > 0 {
			if out.Len()+space < w {
				out.WriteString(s[0:space])
				s = s[space+1:]
				continue
			}
		}
		if out.Len()+len(s) < w {
			out.WriteString(s)
		}
		break
	}

	return out.String()
}

func paste(c1, c2 string, widths ...int) string {
	sepWidth := 2
	var leftLen int
	if len(widths) > 0 {
		sepWidth = widths[0]
		if len(widths) > 1 {
			leftLen = widths[1]
		}
	}
	if leftLen == 0 {
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

func ordinal(i int) string {
	return fmt.Sprintf("%d%s", i, ordinalDictionary[i%10])
}
