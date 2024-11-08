package cmd

import (
	"github.com/slshen/paperscore/pkg/gamefile"
	"github.com/spf13/cobra"
)

func newGameCommand() *cobra.Command {
	var (
		nextDay bool
		count   int
	)
	c := &cobra.Command{
		Use:  "new-game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gm, err := gamefile.ParseFile(args[0])
			if err != nil {
				return err
			}
			for i := 0; i < count; i++ {
				ng, err := gm.WriteNewGame(nextDay)
				if err != nil {
					return err
				}
				nextDay = false
				gm = ng
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&nextDay, "next-day", "n", false, "Create a game for the next day")
	c.Flags().IntVar(&count, "count", 1, "Create `N` new games")
	return c
}
