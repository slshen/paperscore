package markov

import (
	"fmt"
	"regexp"
	"strconv"
)

type BaseOutState int

var (
	EndState   = BaseOutState(24)
	StartState = BaseOutState(0)
)

var baseOutStateRegexp = regexp.MustCompile(`[0123][3x][2x][1x]`)

func FromOutsAndRunners(outs int, r1 bool, r2 bool, r3 bool) BaseOutState {
	s := outs * 8
	if r1 {
		s += 1
	}
	if r2 {
		s += 2
	}
	if r3 {
		s += 4
	}
	return BaseOutState(s)
}

func MustParseBaseOutState(s string) BaseOutState {
	state, err := ParseBaseOutState(s)
	if err != nil {
		panic(err)
	}
	return state
}

func ParseBaseOutState(s string) (state BaseOutState, err error) {
	if !baseOutStateRegexp.MatchString(s) {
		err = fmt.Errorf("not a base out state '%s'", s)
		return
	}
	var (
		outs       int
		r1, r2, r3 bool
	)
	outs, err = strconv.Atoi(s[0:1])
	r1 = s[3] == '1'
	r2 = s[2] == '2'
	r3 = s[1] == '3'
	state = FromOutsAndRunners(outs, r1, r2, r3)
	return
}

func (state BaseOutState) Outs() int {
	return int(state) / 24
}

func (state BaseOutState) R1() bool {
	return state&1 != 0
}

func (state BaseOutState) String() string {
	b := make([]byte, 4)
	b[0] = '0' + byte(state/8)
	if b[0] == '3' {
		b[1] = 'x'
		b[2] = 'x'
		b[3] = 'x'
	} else {
		if (state & 0x4) != 0 {
			b[1] = '3'
		} else {
			b[1] = 'x'
		}
		if (state & 0x2) != 0 {
			b[2] = '2'
		} else {
			b[2] = 'x'
		}
		if (state & 0x1) != 0 {
			b[3] = '1'
		} else {
			b[3] = 'x'
		}
	}
	return string(b)
}
