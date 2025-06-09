package game

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/slshen/paperscore/pkg/text"
	"gopkg.in/yaml.v3"
)

type TeamID string

type Team struct {
	ID        TeamID
	Name      string `yaml:"name"`
	ShortName string `yaml:"short_name"`
	Us        bool   `yaml:"us"`
	Players   map[PlayerID]*Player

	playerIDs map[string]PlayerID
}

type Player struct {
	PlayerID `yaml:"-"`
	Team     *Team `yaml:"-"`
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
		for i := 0; i < 3; i++ {
			err := team.readFile(dir, id)
			if err == nil {
				log.Default().Printf("Loaded team %s from %s", id, dir)
				goto done
			}
			if errors.Is(err, os.ErrNotExist) {
				dir = filepath.Clean(filepath.Join(dir, ".."))
			} else {
				return nil, err
			}
		}
		return team, fmt.Errorf("cannot find team file for %s", id)
	}
done:
	if team.ShortName == "" {
		team.ShortName = text.Initialize(team.Name)
	}
	return team, nil
}

func (team *Team) readFile(dir, id string) error {
	path := filepath.Join(dir, fmt.Sprintf("%s.yaml", id))
	dat, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(dat, team); err != nil {
		return err
	}
	for playerID, player := range team.Players {
		player.Team = team
		player.PlayerID = playerID
		if player.Number == "" {
			player.Number = team.getDefaultPlayerNumber(playerID)
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
		Team:     team,
		Name:     string(id),
		Number:   team.getDefaultPlayerNumber(id),
	}
	if player.Name == player.Number {
		player.Name = ""
	}
	team.Players[id] = player
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

func (team *Team) getDefaultPlayerNumber(player PlayerID) string {
	n := playerNumberRegexp.FindString(string(player))
	if n != "" {
		return n
	}
	return fmt.Sprintf("00%d", len(team.Players)+1)
}

func (player *Player) GetShortName() string {
	return text.NameShorten(player.NameOrNumber())
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

func (player *Player) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		player.Name = value.Value
		return nil
	case yaml.MappingNode:
		var v any
		if err := value.Decode(&v); err != nil {
			return err
		}
		player.Name = v.(map[string]any)["name"].(string)
		return nil
	default:
		return fmt.Errorf("cannot unmarshal player")
	}
}
