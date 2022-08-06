package cmd

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/spf13/cobra"
)

func battingTimesSeenPitcherCommand() *cobra.Command {
	var (
		us      string
		team    bool
		include string
	)
	c := &cobra.Command{
		Use: "batting-times-seen",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGames(args)
			b := stats.NewBatterPitcherData()
			if err != nil {
				return err
			}
			for _, g := range games {
				b.Record(g)
			}
			switch {
			case team:
				fmt.Println(b.GetTeamData())
			case us != "":
				usdat := b.GetUsBatterData(us)
				removeBattingColumns(usdat, include)
				fmt.Println(usdat)
			default:
				fmt.Println(b.GetBatterData())
			}
			return nil
		},
	}
	flags := c.Flags()
	flags.StringVar(&us, "us", "", "Show only `us` batters")
	flags.BoolVar(&team, "team", false, "Show team data")
	flags.StringVar(&include, "include", "", "Only include these columns")
	return c
}

func removeBattingColumns(dat *dataframe.Data, include string) {
	if include != "" {
		for _, col := range dat.Columns {
			name := col.Name
			if name == "Batter" || strings.HasSuffix(name, "-AB") || strings.HasSuffix(name, "-PA") {
				continue
			}
			if strings.HasSuffix(name, "-"+include) {
				continue
			}
			dat.RemoveColumn(name)
		}
	}
}
