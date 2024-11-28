package dataexport

import (
	"fmt"
	"time"

	"github.com/slshen/paperscore/pkg/tournament"
)

type Tournament struct {
	Name         string
	TournamentID string
	Date         string
	Wins         int
	Losses       int
	Ties         int

	group *tournament.Group
}

func newTournament(group *tournament.Group) *Tournament {
	return &Tournament{
		TournamentID: ToID(fmt.Sprintf("%s-%s", group.Date.Format("2006-01-02"), group.Name)),
		Name:         group.Games[0].Tournament,
		Date:         group.Date.Format(time.RFC3339),
		group:        group,
	}
}
