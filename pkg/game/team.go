package game

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"unicode"

	"gopkg.in/yaml.v3"
)

type Team struct {
	Name    string
	Players map[PlayerID]*Player

	playerIDs map[string]PlayerID
}

type Player struct {
	PlayerID `yaml:"-"`
	Name     string
	Number   string
	Inactive bool
}

var playerNumberRegexp = regexp.MustCompile(`\d+`)

func NewTeam(name string) *Team {
	return &Team{
		Name:      name,
		Players:   make(map[PlayerID]*Player),
		playerIDs: make(map[string]PlayerID),
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

func (team *Team) parsePlayerID(s string) PlayerID {
	if unicode.IsDigit(rune(s[0])) {
		playerID := team.playerIDs[s]
		if playerID != "" {
			return playerID
		}
		for playerID, player := range team.Players {
			if player.Number == s {
				// fmt.Printf("Using %s for %s in game file\n", playerID, s)
				team.playerIDs[s] = playerID
				return playerID
			}
		}
	}
	return PlayerID(s)
}

func getDefaultPlayerNumber(player PlayerID) string {
	return playerNumberRegexp.FindString(string(player))
}

func (player *Player) NameOrNumber() string {
	if player.Name != "" {
		return player.Name
	}
	if player.Number != "" {
		return fmt.Sprintf("#%s", player.Number)
	}
	return "?"
}
