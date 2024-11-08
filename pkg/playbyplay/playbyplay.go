package playbyplay

import (
	"fmt"
	"io"
	"strings"

	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/text"
)

type Generator struct {
	Game        *game.Game
	ScoringOnly bool

	score struct {
		home, visitor int
	}
	scoringInning                             int
	scoringHalf                               game.Half
	scoringVisitorPitcher, scoringHomePitcher game.PlayerID
	visitorPitcher, homePitcher               game.PlayerID
	lastState                                 *game.State
}

func (gen *Generator) Generate(w io.Writer) error {
	states := gen.Game.GetStates()
	for _, state := range states {
		if !gen.ScoringOnly && (gen.lastState == nil || gen.lastState.Half != state.Half ||
			gen.lastState.InningNumber != state.InningNumber) {
			fmt.Fprintf(w, "%s of %s\n", state.Half, text.Ordinal(state.InningNumber))
		}
		var battingTeam, fieldingTeam *game.Team
		if state.Half == game.Top {
			battingTeam = gen.Game.Visitor
			fieldingTeam = gen.Game.Home
		} else {
			battingTeam = gen.Game.Home
			fieldingTeam = gen.Game.Visitor
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
					if state.NotOutOnPlay {
						fmt.Fprintf(line, "but ")
					}
					if advance.To == "H" {
						fmt.Fprintf(line, "%s scores", runner.NameOrNumber())
					} else {
						fmt.Fprintf(line, "%s advances to %s", runner.NameOrNumber(), advance.To)
					}
					if advance.IsFieldingError() {
						fmt.Fprintf(line, " on an E%d", advance.FieldingError.Fielder)
					}
				}
			}
			if len(state.ScoringRunners) > 0 {
				if state.Top() {
					gen.score.visitor += len(state.ScoringRunners)
				} else {
					gen.score.home += len(state.ScoringRunners)
				}
				fmt.Fprintf(line, ". %s %d, %s %d", gen.Game.Visitor.Name, gen.score.visitor,
					gen.Game.Home.Name, gen.score.home)
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
				fmt.Fprintf(line, ". %s", state.PlayCode)
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
	count, _, _ := pitches.Count()
	return fmt.Sprintf("with the count %s", count)
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
	return fmt.Sprintf("unknown fielder %d", fielder)
}

func locationName(fielder int) string {
	switch fielder {
	case 1:
		return "the circle"
	case 2:
		return "home plate"
	case 3:
		return "first base"
	case 4:
		return "second base"
	case 5:
		return "third base"
	case 6:
		return "5-6 hole"
	case 7:
		return "left field"
	case 8:
		return "center field"
	case 9:
		return "right field"
	}
	return fmt.Sprintf("unknown location %d", fielder)
}

func hitTrajectory(state *game.State, hit string, fielders []int) string {
	s := &strings.Builder{}
	s.WriteString(hit)
	trajectory := trajectoryDescription(state.Modifiers.Trajectory())
	if trajectory != "" {
		fmt.Fprintf(s, " on a %s", trajectory)
	}
	if len(fielders) > 0 {
		loc := state.Modifiers.Location()
		for i, fielder := range fielders {
			var adj string
			if i == 0 && loc != nil {
				adj = loc.Length
			}
			fmt.Fprintf(s, " to %s%s", adj, positionName(fielder))
		}
	} else {
		// H and DGR don't have fielders
		loc := state.Modifiers.Location()
		if loc != nil {
			var length string
			if loc.Length != "" {
				length = loc.Length + " "
			}
			fmt.Fprintf(s, " to %s%s", length, locationName(loc.Fielder))
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
	case game.BuntGrounder:
		return "bunt on the ground"
	case game.BuntPopup:
		return "popup bunt"
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
	case game.GroundRuleDouble:
		return hitTrajectory(state, "hits a ground rule double", play.Fielders)
	case game.Triple:
		return hitTrajectory(state, "triples", play.Fielders)
	case game.Walk:
		fallthrough
	case game.WalkPickedOff:
		return "walks"
	case game.WalkWildPitch:
		return "walks on wild pitch"
	case game.WalkPassedBall:
		return "walks on passed ball"
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
		if state.IsStrikeOut() {
			return "strikes out"
		}
		return ""
	}
}

func runningPlayDescription(team *game.Team, state, lastState *game.State) string {
	play := state.Play
	switch {
	case len(play.StolenBases) > 0:
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
	case state.Play.Type == game.CaughtStealing || state.Play.Type == game.StrikeOutCaughtStealing:
		return fmt.Sprintf("%s is caught stealing %s", team.GetPlayer(state.CaughtStealingRunner).NameOrNumber(), play.CaughtStealingBase)
	case state.Play.Type == game.WildPitch:
		return "On a wild pitch"
	case state.Play.Type == game.PassedBall:
		return "On a passed ball"
	default:
		return ""
	}
}
