package game

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/gamefile"
)

type FieldingError struct {
	Fielder int
	Modifiers
}

func parseFieldingError(play gamefile.Play, s string) (*FieldingError, error) {
	if len(s) < 2 || s[0] != 'E' || (len(s) > 2 && s[2] != '/') {
		return nil, fmt.Errorf("%s: illegal error code %s", play.GetPos(), s)
	}
	if s[1] < '1' || s[1] > '9' {
		return nil, fmt.Errorf("%s: illegal fielder %c in error code %s", play.GetPos(), s[1], s)
	}
	fe := &FieldingError{
		Fielder: int(s[1] - '0'),
	}
	if len(s) > 2 {
		fe.Modifiers = Modifiers(strings.Split(s[4:len(s)-1], "/"))
	}
	return fe, nil
}
