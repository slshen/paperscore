package webdata

import (
	"fmt"
	"strings"
	"time"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/gamefile"
	"github.com/slshen/sb/pkg/tournament"
)

type Tournament struct {
	Name              string
	Date              time.Time
	Wins              int
	Losses            int
	Ties              int
	Batting           *dataframe.Data
	BestAndWorstPlays *dataframe.Data
	Games             []*Game

	group *tournament.Group
}

func newTournament(wdat *WebData, group *tournament.Group) (*Tournament, error) {
	tourney := &Tournament{
		group: group,
	}
	report, err := tournament.NewReport(wdat.us, wdat.re, group)
	if err != nil {
		return nil, err
	}
	tourney.Batting = report.GetBattingData()
	tourney.Batting.Name = "Batting"
	tourney.BestAndWorstPlays = report.GetBestAndWorstRE24(100)
	tourney.BestAndWorstPlays.Name = "Plays by RE24"

	// tourney.Wins = ?

	for _, g := range group.Games {
		us, _ := g.GetUsAndThem(wdat.us)
		if us == g.Home {
			switch {
			case g.Final.Home > g.Final.Visitor:
				tourney.Wins++
			case g.Final.Home == g.Final.Visitor:
				tourney.Ties++
			default:
				tourney.Losses++
			}
		} else {
			switch {
			case g.Final.Visitor > g.Final.Home:
				tourney.Wins++
			case g.Final.Visitor == g.Final.Home:
				tourney.Ties++
			default:
				tourney.Losses++
			}
		}
		game, err := newGame(wdat, g)
		if err != nil {
			return nil, err
		}
		tourney.Games = append(tourney.Games, game)
	}

	return tourney, nil
}

func (t *Tournament) GetPage() *Page {
	p := &Page{
		ID: t.group.Tournament,
		Pages: []HasPage{
			newTable(t.BestAndWorstPlays),
			newTable(t.Batting),
		},
		Content: t.GetContent,
	}
	for _, g := range t.Games {
		p.Pages = append(p.Pages, g)
	}
	return p.SetFrontStruct(t).
		Set("Title", t.group.Tournament).
		Set("Games", len(t.Games)).
		Set("StartDate", t.group.Date.Format(gamefile.GameDateFormat)).
		Set("EndDate", t.group.Games[len(t.group.Games)-1].GetDate().Format(gamefile.GameDateFormat))
}

func (t *Tournament) GetContent() string {
	s := &strings.Builder{}
	fmt.Fprintln(s, "{{% tournament %}}")
	for _, g := range t.Games {
		_ = g.box.InningScoreTable().RenderMarkdown(s)
		fmt.Fprintln(s)
	}
	return s.String()
}
