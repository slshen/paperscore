package cmd

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/spf13/cobra"
)

func altCommand() *cobra.Command {
	var (
		re reArgs
	)
	c := &cobra.Command{
		Use:   "alt",
		Short: "Display the cost of errors/misplays/good plays",
		RunE: func(cmd *cobra.Command, args []string) error {
			re, err := re.getRunExpectancy()
			if err != nil {
				return err
			}
			if re == nil {
				return fmt.Errorf("run expectancy required")
			}
			games, err := game.ReadGameFiles(args)
			if err != nil {
				return err
			}
			for _, g := range games {
				gs := stats.NewGameStats(re)
				if err := gs.Read(g); err != nil {
					return err
				}
				alt := gs.GetAltData()
				alt.Name = fmt.Sprintf("%s game %s %s at %s Alt Plays", g.Date, g.Number, g.Visitor, g.Home)
				alt.RemoveColumn("Game")
				fmt.Println(alt)
			}
			return nil
		},
	}
	re.registerFlags(c.Flags())
	return c
}
