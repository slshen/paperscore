package gamefile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

type Numbers string

type File struct {
	Path         string
	PropertyList []*Property   `parser:"@@*"`
	TeamEvents   []*TeamEvents `parser:"@@*"`

	Properties map[string]string
}

type Property struct {
	Pos   lexer.Position
	Key   string `parser:"@Ident"`
	Value string `parser:"@Text (NL|EOF)"`
}

type TeamEvents struct {
	Pos    lexer.Position
	TeamID string   `parser:"'plays' (@Code | @Keyword) (NL|EOF)"`
	Events []*Event `parser:"@@*"`
}

type Event struct {
	Pos         lexer.Position
	Play        *Play        `parser:"@@ (NL|EOF)"`
	Alternative *Alternative `parser:"| 'alt' @@ (NL|EOF)"`
	Pitcher     string       `parser:"| ('pitcher'|'pitching') @Code (NL|EOF)"`
	RAdjRunner  string       `parser:"| 'radj' @Numbers"`
	RAdjBase    string       `parser:"      @Code (NL|EOF)"`
	Score       string       `parser:"| 'score' @Code (NL|EOF)"`
	Final       string       `parser:"| 'final' @Code (NL|EOF)"`
	Empty       bool         `parser:"| @NL"`
}

type Play struct {
	Pos                      lexer.Position
	PlateAppearance          Numbers  `parser:"((@Numbers"`
	Batter                   Numbers  `parser:"  @Numbers)"`
	ContinuedPlateAppearance bool     `parser:" | @Dots)"`
	PitchSequence            string   `parser:" @Code"`
	Code                     string   `parser:" @Code"`
	Advances                 []string `parser:" @Code*"`
	Comment                  string   `parser:" @Text?"`
}

type Alternative struct {
	Code     string    `parser:"@Code"`
	Advances []*string `parser:"@Code*"`
	Comment  string    `parser:" @Text?"`
}

func (n *Numbers) UnmarshalText(dat []byte) error {
	*n = Numbers(strings.TrimRight(string(dat), " \t"))
	return nil
}

func (n Numbers) String() string {
	return string(n)
}

func (n Numbers) Int() int {
	i, _ := strconv.Atoi(n.String())
	return i
}

func (f *File) GetVisitorEvents() *TeamEvents {
	if events := f.findTeamEvents("visitorid"); events != nil {
		return events
	}
	if len(f.TeamEvents) > 0 {
		return f.TeamEvents[0]
	}
	return nil
}

func (f *File) GetHomeEvents() *TeamEvents {
	if events := f.findTeamEvents("homeid"); events != nil {
		return events
	}
	if len(f.TeamEvents) > 1 {
		return f.TeamEvents[1]
	}
	return nil
}

func (f *File) findTeamEvents(key string) *TeamEvents {
	if id := f.Properties[key]; id != "" {
		for _, events := range f.TeamEvents {
			if events.TeamID == id {
				return events
			}
		}
	}
	return nil
}

func (f *File) validate() error {
	if len(f.TeamEvents) > 2 {
		return fmt.Errorf("%s has more than 2 play sections", f.Path)
	}
	f.Properties = make(map[string]string)
	for _, prop := range f.PropertyList {
		f.Properties[prop.Key] = prop.Value
	}
	return nil
}
