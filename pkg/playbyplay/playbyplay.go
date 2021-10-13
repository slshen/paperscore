package playbyplay

import (
	"fmt"
	"io"
	"strings"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/text"
)

type Generator struct {
	Game        *game.Game
	ScoringOnly bool

	scoringInning                             int
	scoringHalf                               game.Half
	scoringVisitorPitcher, scoringHomePitcher game.PlayerID
	visitorPitcher, homePitcher               game.PlayerID
	lastState                                 *game.State
}

func (gen *Generator) Generate(w io.Writer) error {
	states, err := gen.Game.GetStates()
	if err != nil {
		return err
	}
	for _, state := range states {
		if !gen.ScoringOnly && (gen.lastState == nil || gen.lastState.Half != state.Half ||
			gen.lastState.InningNumber != state.InningNumber) {
			fmt.Fprintf(w, "%s of %s\n", state.Half, text.Ordinal(state.InningNumber))
		}
		var battingTeam, fieldingTeam *game.Team
		if state.Half == game.Top {
			battingTeam = gen.Game.VisitorTeam
			fieldingTeam = gen.Game.HomeTeam
		} else {
			battingTeam = gen.Game.HomeTeam
			fieldingTeam = gen.Game.VisitorTeam
		}
		pitchingChange := false
		if state.Half == game.Top && gen.homePitcher != state.Pitcher {
			pitchingChange = true
			gen.homePitcher = state.Pitcher
		} else if state.Half == game.Bottom && gen.visitorPitcher != state.Pitcher {
			pitchingChange = true
			gen.visitorPitcher = state.Pitcher
		}
		if !gen.ScoringOnly && pitchingChange {
			pitcher := fieldingTeam.GetPlayer(state.Pitcher)
			fmt.Fprintf(w, "  %s is now pitching for %s\n\n", pitcher.NameOrNumber(), fieldingTeam.Name)
		}
		line := &strings.Builder{}
		if batterPlay := batterPlayDescription(state); batterPlay != "" {
			batter := battingTeam.GetPlayer(state.Batter)
			fmt.Fprintf(line, "%s %s %s", batter.NameOrNumber(), countDescription(state.Pitches), batterPlay)
		} else if runnerPlay := runningPlayDescription(battingTeam, state, gen.lastState); runnerPlay != "" {
			fmt.Fprint(line, runnerPlay)
		}
		if len(state.Advances) > 0 {
			i := 0
			for _, advance := range state.Advances {
				if advance.Implied {
					continue
				}
				if i == 0 && line.Len() > 0 {
					line.WriteString(", ")
				}
				var runnerID game.PlayerID
				if advance.From == "B" {
					runnerID = state.Batter
				} else {
					runnerID = gen.lastState.Runners[game.BaseNumber[advance.From]]
				}
				runner := battingTeam.GetPlayer(runnerID)
				if i > 0 {
					fmt.Fprint(line, ", ")
				}
				i++
				if advance.Out {
					fmt.Fprintf(line, "%s is out advancing to %s", runner.NameOrNumber(), advance.To)
				} else {
					if advance.To == "H" {
						fmt.Fprintf(line, "%s scores", runner.NameOrNumber())
					} else {
						fmt.Fprintf(line, "%s advances to %s", runner.NameOrNumber(), advance.To)
					}
					if advance.FieldingError != nil {
						fmt.Fprintf(line, " on an E%d", advance.FieldingError.Fielder)
					}
				}
			}
			if len(state.ScoringRunners) > 0 {
				fmt.Fprintf(line, ". %d %s, %d %s", state.Score.Visitor, gen.Game.Visitor, state.Score.Home, gen.Game.Home)
			}
		}
		if state.Comment != "" {
			fmt.Fprintf(line, " (%s)", state.Comment)
		}
		if line.Len() > 0 {
			if gen.ScoringOnly && len(state.ScoringRunners) > 0 {
				if gen.scoringHalf != state.Half || gen.scoringInning != state.InningNumber {
					fmt.Fprintf(w, "%s of %s\n", state.Half, text.Ordinal(state.InningNumber))
					gen.scoringHalf = state.Half
					gen.scoringInning = state.InningNumber
				}
				var scoringPitcherID game.PlayerID
				if state.Half == game.Top && state.Pitcher != gen.scoringHomePitcher {
					scoringPitcherID = state.Pitcher
					gen.scoringHomePitcher = scoringPitcherID
				} else if state.Half == game.Bottom && state.Pitcher != gen.scoringVisitorPitcher {
					scoringPitcherID = state.Pitcher
					gen.scoringVisitorPitcher = scoringPitcherID
				}
				if scoringPitcherID != "" {
					scoringPitcher := fieldingTeam.GetPlayer(scoringPitcherID)
					fmt.Fprintf(w, "  With %s pitching for %s\n\n", scoringPitcher.NameOrNumber(), fieldingTeam.Name)
				}
			}
			if !gen.ScoringOnly || len(state.ScoringRunners) > 0 {
				fmt.Fprintf(line, ". %s", state.EventCode)
				fmt.Fprint(w, text.WrapIndent(line.String(), 80, "  "))
				fmt.Fprintln(w)
				fmt.Fprintln(w)
			}
		}
		gen.lastState = state
	}
	return nil
}

func countDescription(pitches game.Pitches) string {
	if pitches == "X" {
		return "on the first pitch"
	}
	return fmt.Sprintf("with the count %s", pitches.Count())
}

func positionName(fielder int) string {
	switch fielder {
	case 1:
		return "pitcher"
	case 2:
		return "catcher"
	case 3:
		return "first base"
	case 4:
		return "second base"
	case 5:
		return "third base"
	case 6:
		return "shortstop"
	case 7:
		return "left fielder"
	case 8:
		return "center fielder"
	case 9:
		return "right fielder"
	}
	return fmt.Sprintf("unknwn fielder %d", fielder)
}

func hitTrajectory(state *game.State, hit string, fielders []int) string {
	s := &strings.Builder{}
	s.WriteString(hit)
	trajectory := trajectoryDescription(state.Modifiers.Trajectory())
	if trajectory != "" {
		fmt.Fprintf(s, " on a %s", trajectory)
	}
	if len(fielders) > 0 {
		for _, fielder := range fielders {
			fmt.Fprintf(s, " to %s", positionName(fielder))
		}
	}
	return s.String()
}

func trajectoryDescription(trajectory game.Trajectory) string {
	switch trajectory {
	case game.LineDrive:
		return "line drive"
	case game.FlyBall:
		return "fly ball"
	case game.PopUp:
		return "pop fly"
	case game.Bunt:
		return "bunt"
	case game.GroundBall:
		return "ground ball"
	default:
		return ""
	}
}

func batterPlayDescription(state *game.State) string {
	play := state.Play
	switch play.Type {
	case game.Single:
		return hitTrajectory(state, "singles", play.Fielders)
	case game.Double:
		return hitTrajectory(state, "doubles", play.Fielders)
	case game.Triple:
		return hitTrajectory(state, "triples", play.Fielders)
	case game.Walk:
		return "walks"
	case game.HomeRun:
		return hitTrajectory(state, "hits a home run", nil)
	case game.HitByPitch:
		return "is hit by pitch"
	case game.CatcherInterference:
		return "reaches on catcher's interference"
	case game.ReachedOnError:
		throwing := "an"
		if play.FieldingError.Modifiers.Contains("TH") {
			throwing = "a throwing "
		}
		return fmt.Sprintf("reaches on %s error by %s", throwing, positionName(play.FieldingError.Fielder))
	case game.FieldersChoice:
		return "reaches on a fielder's choice"
	case game.StrikeOutWildPitch:
		return "reaches on a strikeout wild pitch"
	case game.StrikeOutPassedBall:
		return "reaches on a striekout passed ball"
	case game.StrikeOut:
		return "strikes out"
	case game.GroundOut:
		verb := "is out"
		if state.Modifiers.Trajectory() == game.Bunt {
			verb = "bunts out"
		}
		return hitTrajectory(state, verb, play.Fielders)
	case game.FlyOut:
		return hitTrajectory(state, "is out", play.Fielders)
	case game.DoublePlay:
		verb := "grounds"
		if state.Modifiers.Trajectory() == game.LineDrive {
			verb = "lines"
		}
		return fmt.Sprintf("%s into a double play", verb)
	default:
		return ""
	}
}

func runningPlayDescription(team *game.Team, state, lastState *game.State) string {
	play := state.Play
	switch state.Play.Type {
	case game.StolenBase:
		var sb []string
		for _, base := range play.StolenBases {
			var runner game.PlayerID
			switch base {
			case "2":
				runner = lastState.Runners[0]
				base = "second"
			case "3":
				runner = lastState.Runners[1]
				base = "third"
			case "H":
				runner = lastState.Runners[2]
				base = "home"
			}
			sb = append(sb,
				fmt.Sprintf("%s steals %s", team.GetPlayer(runner).NameOrNumber(), base))
		}
		return strings.Join(sb, ", ")
	case game.CaughtStealing:
		return fmt.Sprintf("%s is caught stealing %s", team.GetPlayer(play.Runners[0]).NameOrNumber(), play.Base)
	case game.WildPitch:
		return "On a wild pitch"
	case game.PassedBall:
		return "On a passed ball"
	default:
		return ""
	}
}
