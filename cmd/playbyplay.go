package cmd

import (
	"os"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/playbyplay"
	"github.com/spf13/cobra"
)

func playByPlayCommand() *cobra.Command {
	pbp := playbyplay.Generator{}
	c := &cobra.Command{
		Use:   "plays",
		Short: "Generate play by play",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGameFiles(args)
			if err != nil {
				return err
			}
			for _, g := range games {
				pbp.Game = g
				if err := pbp.Generate(os.Stdout); err != nil {
					return err
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&pbp.ScoringOnly, "scoring", false, "Only show scoring plays")
	return c
}
