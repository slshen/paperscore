package game

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type TeamID string

type Team struct {
	ID      TeamID
	Name    string `yaml:"name"`
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

func GetTeam(dir, name, id string) (*Team, error) {
	team := &Team{
		Name:      name,
		ID:        TeamID(id),
		Players:   make(map[PlayerID]*Player),
		playerIDs: make(map[string]PlayerID),
	}
	if id == "" {
		team.ID = TeamID(strings.ReplaceAll(team.Name, " ", "-"))
	} else {
		if dir == "" {
			return nil, fmt.Errorf("team %s cannot be loaded without a directory", id)
		}
		if err := team.readFile(dir, id); err != nil {
			return nil, err
		}
	}
	return team, nil
}

func (team *Team) IsUs(us string) bool {
	return strings.HasPrefix(strings.ToLower((team.Name)), us)
}

func (team *Team) readFile(dir, id string) error {
	path := filepath.Join(dir, fmt.Sprintf("%s.yaml", id))
	dat, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := yaml.Unmarshal(dat, team); err != nil {
		return err
	}
	for playerID, player := range team.Players {
		player.PlayerID = playerID
		if player.Number == "" {
			player.Number = getDefaultPlayerNumber(playerID)
		}
	}
	return nil
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
