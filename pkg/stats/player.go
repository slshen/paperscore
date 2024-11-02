package stats

import "github.com/slshen/paperscore/pkg/game"

type PlayerData struct {
	Name            string
	Team            string
	Number          string
	Games           int
	GameAppearances map[string]bool
	Inactive        bool
}

func NewPlayerData(team string, player *game.Player) PlayerData {
	return PlayerData{
		Name:            player.NameOrNumber(),
		Team:            team,
		Number:          player.Number,
		GameAppearances: map[string]bool{},
		Inactive:        player.Inactive,
	}
}

func (pd *PlayerData) Update() {
	pd.Games = len(pd.GameAppearances)
}
