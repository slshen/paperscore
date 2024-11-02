package game

import (
	"fmt"
	"strings"

	"github.com/slshen/paperscore/pkg/gamefile"
)

type FieldingError struct {
	Fielder int
	Modifiers
}

var NoError = FieldingError{}

func parseFieldingError(play gamefile.Play, s string) (FieldingError, error) {
	if len(s) < 2 || s[0] != 'E' || (len(s) > 2 && s[2] != '/') {
		return FieldingError{}, NewError("illegal error code %s", play.GetPos(), s)
	}
	if s[1] < '1' || s[1] > '9' {
		return FieldingError{}, NewError("illegal fielder %c in error code %s", play.GetPos(), s[1], s)
	}
	fe := FieldingError{
		Fielder: int(s[1] - '0'),
	}
	if len(s) > 2 {
		fe.Modifiers = Modifiers(strings.Split(s[4:len(s)-1], "/"))
	}
	return fe, nil
}

func (fe FieldingError) IsFieldingError() bool {
	return fe.Fielder != 0
}

func (fe FieldingError) String() string {
	if !fe.IsFieldingError() {
		return ""
	}
	s := strings.Builder{}
	fmt.Fprintf(&s, "E%d", fe.Fielder)
	for _, mod := range fe.Modifiers {
		fmt.Fprintf(&s, "/%s", mod)
	}
	return s.String()
}
