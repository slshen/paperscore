package cmd

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/spf13/cobra"
)

func battingCountCommand() *cobra.Command {
	var (
		us    string
		notus string
	)
	c := &cobra.Command{
		Use:   "batting-count",
		Short: "Display the batting stats by count",
		RunE: func(cmd *cobra.Command, args []string) error {
			bc := stats.NewBattingByCount()
			bc.Us = us
			bc.NotUs = notus
			gs, err := game.ReadGames(args)
			if err != nil {
				return err
			}
			for _, gm := range gs {
				bc.Read(gm)
			}
			fmt.Println(bc.GetData())
			return nil
		},
	}
	c.Flags().StringVar(&us, "us", "", "Limit at bats to team ID's that contain `us`")
	c.Flags().StringVar(&notus, "not-us", "", "Limit at bats to team ID's that do not contain `us`")
	return c
}
