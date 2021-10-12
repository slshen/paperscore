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
			if gen.lastState != nil {
				fmt.Fprintln(w)
			}
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
			fmt.Fprintf(line, "%s %s", batter.NameOrNumber(), batterPlay)
		} else if runnerPlay := runningPlayDescription(battingTeam, state, gen.lastState); runnerPlay != "" {
			fmt.Fprint(line, runnerPlay)
		}
		if len(state.Advances) > 0 {
			if line.Len() > 0 {
				line.WriteString(". ")
			}
			i := 0
			for _, advance := range state.Advances {
				if advance.To == "H" && !advance.Out {
					// will be noted in scoring runners
					continue
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
					fmt.Fprintf(line, "%s advances to %s", runner.NameOrNumber(), advance.To)
					if advance.FieldingError != nil {
						fmt.Fprintf(line, " on an E%d", advance.FieldingError.Fielder)
					}
				}
			}
		}
		if len(state.ScoringRunners) > 0 {
			if line.Len() > 0 {
				line.WriteString(". ")
			}
			for i, runnerID := range state.ScoringRunners {
				if i > 0 {
					fmt.Fprint(line, ", ")
				}
				runner := battingTeam.GetPlayer(runnerID)
				fmt.Fprintf(line, "%s scores", runner.NameOrNumber())
			}
			fmt.Fprintf(line, ". %d %s, %d %s.", state.Score.Visitor, gen.Game.Visitor,
				state.Score.Home, gen.Game.Home)
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
				fmt.Fprint(w, text.WrapIndent(line.String(), 75, "  "))
				fmt.Fprintln(w)
				fmt.Fprintln(w)
			}
		}
		gen.lastState = state
	}
	return nil
}

func positionName(fielder int) string {
	switch fielder {
	case 1:
		return "pitcher"
	case 2:
		return "catcher"
	case 3:
		return "1B infielder"
	case 4:
		return "2B infielder"
	case 5:
		return "3B infielder"
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

func hitTrajectory(state *game.State, hit string) string {
	trajectory := trajectoryDescription(state.Modifiers.Trajectory())
	if trajectory != "" {
		return fmt.Sprintf("%s on a %s", hit, trajectory)
	}
	return hit
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
	switch {
	case play.Single():
		return hitTrajectory(state, "singles")
	case play.Double():
		return hitTrajectory(state, "doubles")
	case play.Triple():
		return hitTrajectory(state, "triples")
	case play.Walk():
		return "walks"
	case play.HomeRun():
		return hitTrajectory(state, "hits a home run")
	case play.HitByPitch():
		return "is hit by pitch"
	case play.CatcherInterference():
		return "reaches on catcher's interference"
	case play.ReachedOnError():
		fe, err := play.FieldingError()
		if err != nil {
			panic(err)
		}
		var throwing string
		if fe.Contains("TH") {
			throwing = "a throwing "
		}
		return fmt.Sprintf("reaches on %serror by the %s", throwing, positionName(fe.Fielder))
	case play.FieldersChoice():
		return "reaches on a fielder's choice"
	case play.StrikeOut() && play.WildPitch():
		return "reaches on a strikeout wild pitch"
	case play.StrikeOut() && play.PassedBall():
		return "reaches on a striekout passed ball"
	case play.StrikeOut():
		return "strikes out"
	case play.GroundOut():
		verb := "grounds out"
		if state.Modifiers.Trajectory() == game.Bunt {
			verb = "bunts out"
		}
		return fmt.Sprintf("%s (%s)", verb, play)
	case play.FlyOut():
		verb := "flys out"
		switch state.Modifiers.Trajectory() {
		case game.PopUp:
			verb = "pops out"
		case game.LineDrive:
			verb = "lines out"
		}
		return fmt.Sprintf("%s (%s)", verb, play)
	default:
		return ""
	}
}

func runningPlayDescription(team *game.Team, state, lastState *game.State) string {
	play := state.Play
	switch {
	case play.StolenBase():
		var sb []string
		for _, base := range play.StolenBases() {
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
	case play.CaughtStealing():
		base := string(play)[2:3]
		runner := lastState.Runners[game.BaseNumber[game.PreviousBase[base]]]
		return fmt.Sprintf("%s is caught stealing %s", team.GetPlayer(runner).NameOrNumber(), base)
	default:
		return ""
	}
}
