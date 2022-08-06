package cmd

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/spf13/cobra"
)

func pitchingTimesSeenLineupCommand() *cobra.Command {
	var (
		us      string
		team    bool
		include string
	)
	c := &cobra.Command{
		Use: "pitching-times-seen",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGames(args)
			b := stats.NewPitcherTimesLineup()
			if err != nil {
				return err
			}
			for _, g := range games {
				b.Record(g)
			}
			fmt.Println(b.GetData())
			return nil
		},
	}
	flags := c.Flags()
	flags.StringVar(&us, "us", "", "Show only `us` batters")
	flags.BoolVar(&team, "team", false, "Show team data")
	flags.StringVar(&include, "include", "", "Only include these columns")
	return c
}
