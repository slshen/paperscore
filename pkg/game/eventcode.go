package game

import (
	"regexp"
	"strings"
)

type eventCodeParser struct {
	playCode     string
	eventCode    string
	eventMatches []string
	advancesCode string
	modifiers    Modifiers
}

var eventCodeRegexps = map[string]*regexp.Regexp{}
var playRegexp = regexp.MustCompile(`([^./]+)`)
var fielderNumber = map[string]int{
	"1": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
	"6": 6,
	"7": 7,
	"8": 8,
	"9": 9,
}

func (p *eventCodeParser) getFielders(eventFields ...int) []int {
	res := make([]int, len(eventFields))
	for i := range res {
		res[i] = p.getFielder(eventFields[i])
	}
	return res
}

func (p *eventCodeParser) getFielder(eventField int) int {
	return fielderNumber[p.eventMatches[eventField]]
}

func (p *eventCodeParser) parseEvent(code string) {
	m := playRegexp.FindStringSubmatch(code)
	p.playCode = m[1]
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
	p.eventMatches = nil
}

func (p *eventCodeParser) eventIs(pattern string) bool {
	re := eventCodeRegexps[pattern]
	if re == nil {
		str := strings.ReplaceAll(pattern, "(", `\(`)
		str = strings.ReplaceAll(str, "+", `\+`)
		str = strings.ReplaceAll(str, ")", `\)`)
		str = strings.ReplaceAll(strings.ReplaceAll(str, "$", "([0123456789])"),
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
