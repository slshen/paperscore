package game

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Team struct {
	Name    string
	Players map[PlayerID]*Player
}

type Player struct {
	PlayerID `yaml:"-"`
	Name     string
	Number   string
}

var playerNumberRegexp = regexp.MustCompile(`\d+`)

func NewTeam(name string) *Team {
	return &Team{
		Name:    name,
		Players: make(map[PlayerID]*Player),
	}
}

func ReadTeamFile(name, path string) (*Team, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	team := NewTeam(name)
	if err := yaml.Unmarshal(dat, team); err != nil {
		return nil, err
	}
	for playerID, player := range team.Players {
		player.PlayerID = playerID
		if player.Number == "" {
			player.Number = getDefaultPlayerNumber(playerID)
		}
	}
	return team, nil
}

func (team *Team) GetPlayer(id PlayerID) *Player {
	if team != nil {
		p := team.Players[id]
		if p != nil {
			return p
		}
	}
	player := &Player{
		PlayerID: id,
		Name:     string(id),
		Number:   getDefaultPlayerNumber(id),
	}
	if player.Name == player.Number {
		player.Name = ""
	}
	return player
}

func getDefaultPlayerNumber(player PlayerID) string {
	return playerNumberRegexp.FindString(string(player))
}

func (player *Player) NameOrNumber() string {
	if player.Name != "" {
		return player.Name
	}
	return fmt.Sprintf("#%s", player.Number)
}

func (player *Player) NameOrQ() string {
	if player.Name != "" {
		return player.Name
	}
	return "?"
}
