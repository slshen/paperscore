package game

import (
	"regexp"
	"strings"
)

type playCodeParser struct {
	playCode    string
	playMatches []string
	modifiers   Modifiers
}

var playCodeRegexps = map[string]*regexp.Regexp{}

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

func (p *playCodeParser) getFielders(eventFields ...int) []int {
	res := make([]int, len(eventFields))
	for i := range res {
		res[i] = p.getFielder(eventFields[i])
	}
	return res
}

func (p *playCodeParser) getFielder(eventField int) int {
	return fielderNumber[p.playMatches[eventField]]
}

func (p *playCodeParser) parsePlayCode(code string) {
	parts := strings.Split(code, "/")
	p.playCode = parts[0]
	if len(parts) > 1 {
		p.modifiers = parts[1:]
	} else {
		p.modifiers = nil
	}
	p.playMatches = nil
}

func (p *playCodeParser) playIs(pattern string) bool {
	re := playCodeRegexps[pattern]
	if re == nil {
		str := strings.ReplaceAll(pattern, "(", `\(`)
		str = strings.ReplaceAll(str, "+", `\+`)
		str = strings.ReplaceAll(str, ")", `\)`)
		str = strings.ReplaceAll(strings.ReplaceAll(str, "$", "([0123456789])"),
			"%", "([B123H])")
		re = regexp.MustCompile("^" + str + "$")
		playCodeRegexps[pattern] = re
	}
	m := re.FindStringSubmatch(p.playCode)
	if len(m) > 1 {
		p.playMatches = m[1:]
	} else {
		p.playMatches = nil
	}
	return m != nil
}
