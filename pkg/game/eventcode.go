package game

import (
	"regexp"
	"strings"
)

type eventCodeParser struct {
	eventCode    string
	eventMatches []string
	advancesCode string
	modifiers    Modifiers
}

var eventCodeRegexps = map[string]*regexp.Regexp{}
var playRegexp = regexp.MustCompile(`([^./]+)`)

func (p *eventCodeParser) parseEvent(code string) Play {
	m := playRegexp.FindStringSubmatch(code)
	play := Play(m[1])
	dot := strings.IndexRune(code, '.')
	if dot > 0 {
		p.advancesCode = code[dot+1:]
		code = code[0:dot]
	} else {
		p.advancesCode = ""
	}
	parts := strings.Split(code, "/")
	p.eventCode = parts[0]
	if len(parts) > 1 {
		p.modifiers = parts[1:]
	} else {
		p.modifiers = nil
	}
	return play
}

func (p *eventCodeParser) eventIs(pattern string) bool {
	re := eventCodeRegexps[pattern]
	if re == nil {
		str := strings.ReplaceAll(pattern, "(", `\(`)
		str = strings.ReplaceAll(str, "+", `\+`)
		str = strings.ReplaceAll(str, ")", `\)`)
		str = strings.ReplaceAll(strings.ReplaceAll(str, "$", "([123456789])"),
			"%", "([B123H])")
		re = regexp.MustCompile("^" + str + "$")
		eventCodeRegexps[pattern] = re
	}
	m := re.FindStringSubmatch(p.eventCode)
	if len(m) > 1 {
		p.eventMatches = m[1:]
	} else {
		p.eventMatches = nil
	}
	return m != nil
}
