package game

import (
	"fmt"

	"github.com/slshen/paperscore/pkg/gamefile"
)

type Error struct {
	Pos     gamefile.Position
	Message string
}

func NewError(template string, pos gamefile.Position, args ...any) Error {
	return Error{
		Pos:     pos,
		Message: fmt.Sprintf("%s: %s", pos, fmt.Sprintf(template, args...)),
	}
}

func (e Error) Error() string { return e.Message }

func (e Error) Position() gamefile.Position { return e.Pos }
