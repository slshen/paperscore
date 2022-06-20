package gamefile

import (
	"fmt"
	"io"
	"sort"
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
	Play        *ActualPlay  `parser:"@@ (NL|EOF)"`
	Alternative *Alternative `parser:"| 'alt' @@ (NL|EOF)"`
	Pitcher     string       `parser:"| ('pitcher'|'pitching') @Code (NL|EOF)"`
	RAdjRunner  Numbers      `parser:"| 'radj' @Numbers"`
	RAdjBase    string       `parser:"      @Code (NL|EOF)"`
	Score       string       `parser:"| 'score' @Code (NL|EOF)"`
	Final       string       `parser:"| 'final' @Code (NL|EOF)"`
	Empty       bool         `parser:"| @NL"`
}

type Play interface {
	GetPos() lexer.Position
	GetCode() string
	GetAdvances() []string
	GetComment() string
}

type ActualPlay struct {
	Pos                      lexer.Position
	PlateAppearance          Numbers  `parser:"((@Numbers"`
	Batter                   Numbers  `parser:"  @Numbers)"`
	ContinuedPlateAppearance bool     `parser:" | @Dots)"`
	PitchSequence            string   `parser:" @Code"`
	Code                     string   `parser:" @Code"`
	Advances                 []string `parser:" @Code*"`
	Comment                  string   `parser:" @Text?"`
}

var _ Play = (*ActualPlay)(nil)

type Alternative struct {
	Pos      lexer.Position
	Code     string   `parser:"@Code"`
	Advances []string `parser:"@Code*"`
	Comment  string   `parser:" @Text?"`
}

var _ Play = (*Alternative)(nil)

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

func (f *File) Write(w io.Writer) {
	printed := map[string]bool{}
	for _, name := range []string{"date", "game", "visitor", "visitorid", "home", "homeid", "start", "timelimit", "tournament", "league"} {
		val := f.Properties[name]
		printed[name] = true
		if val != "" {
			fmt.Fprintf(w, "%s: %s\n", name, val)
		}
	}
	var names []string
	for name := range f.Properties {
		if !printed[name] {
			printed[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		val := f.Properties[name]
		fmt.Fprintf(w, "%s: %s\n", name, val)
	}
	fmt.Fprintln(w, "---")
	f.writeEvents(w, f.GetVisitorEvents())
	f.writeEvents(w, f.GetHomeEvents())
}

func (f *File) writeEvents(w io.Writer, events *TeamEvents) {
	if events == nil {
		return
	}
	fmt.Fprintf(w, "plays %s\n", events.TeamID)
	var pa int
	for _, event := range events.Events {
		switch {
		case event.Play != nil:
			play := event.Play
			if !play.ContinuedPlateAppearance {
				if i := play.PlateAppearance.Int(); i != 0 {
					pa = i
				} else {
					pa += 1
				}
				fmt.Fprintf(w, "%d %s ", pa, play.Batter.String())
			} else {
				fmt.Fprintf(w, "  ... ")
			}
			fmt.Fprintf(w, "%s ", play.PitchSequence)
			f.writeCodeAdvancesComment(w, play.Code, play.Advances, play.Comment)
		case event.Alternative != nil:
			alt := event.Alternative
			fmt.Fprintf(w, "  alt")
			f.writeCodeAdvancesComment(w, alt.Code, alt.Advances, alt.Comment)
		case event.Pitcher != "":
			fmt.Fprintf(w, "pitching %s\n", event.Pitcher)
		case event.RAdjBase != "":
			fmt.Fprintf(w, "radj %s %s\n", event.RAdjRunner, event.RAdjBase)
		case event.Score != "":
			fmt.Fprintf(w, "score %s\n", event.Score)
		case event.Final != "":
			fmt.Fprintf(w, "final %s\n", event.Final)
		}
	}
	fmt.Fprintln(w)
}

func (f *File) writeCodeAdvancesComment(w io.Writer, code string, advances []string, comment string) {
	fmt.Fprintf(w, "%s", code)
	for _, adv := range advances {
		fmt.Fprintf(w, " %s", adv)
	}
	if comment != "" {
		fmt.Fprintf(w, " : %s", comment)
	}
	fmt.Fprintln(w)
}

func (p *ActualPlay) GetPos() lexer.Position {
	return p.Pos
}

func (p *ActualPlay) GetCode() string {
	return p.Code
}

func (p *ActualPlay) GetAdvances() []string {
	return p.Advances
}

func (p *ActualPlay) GetComment() string {
	return p.Comment
}

func (a *Alternative) GetPos() lexer.Position {
	return a.Pos
}

func (a *Alternative) GetCode() string {
	return a.Code
}

func (a *Alternative) GetAdvances() []string {
	return a.Advances
}

func (a *Alternative) GetComment() string {
	return a.Comment
}
