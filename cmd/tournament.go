package cmd

import (
	"fmt"
	"strings"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/tournament"
	"github.com/spf13/cobra"
)

func tournamentCommand() *cobra.Command {
	var (
		us             string
		plays          int
		re             reArgs
		tournamentName string
		playsOnly      bool
	)
	c := &cobra.Command{
		Use:   "tournament",
		Short: "Report on tournament results",
		RunE: func(cmd *cobra.Command, args []string) error {
			if us == "" {
				return fmt.Errorf("--us is required")
			}
			re, err := re.getRunExpectancy()
			if err != nil {
				return err
			}
			games, err := game.ReadGames(args)
			if err != nil {
				return err
			}
			for _, gr := range tournament.GroupByTournament(games) {
				if tournamentName != "" && !strings.Contains(strings.ToLower(gr.Name), tournamentName) {
					continue
				}
				rep, err := tournament.NewReport(us, re, gr)
				if err != nil {
					return err
				}
				if !playsOnly {
					fmt.Println(rep.GetBattingData())
				}
				topPlays := rep.GetBestAndWorstRE24(plays)
				if playsOnly {
					topPlays = topPlays.Select(
						dataframe.Col("Rnk"),
						dataframe.Col("Game"),
						dataframe.Col("ID"),
						dataframe.Col("O"),
						dataframe.Col("Rnr"),
						dataframe.Col("Play"),
						dataframe.Col("RE24"),
					)
				}
				fmt.Println(topPlays)
			}
			return nil
		},
	}
	re.registerFlags(c.Flags())
	c.Flags().StringVar(&us, "us", "", "Our `team`")
	c.Flags().IntVar(&plays, "plays", 15, "Show the top and bottom `n` plays by RE24")
	c.Flags().StringVar(&tournamentName, "tournament", "", "Show only `tournament`")
	c.Flags().BoolVar(&playsOnly, "plays-only", false, "Only list top plays")
	return c
}
