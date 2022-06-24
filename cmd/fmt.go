package cmd

import (
	"github.com/slshen/sb/pkg/game"
	"github.com/spf13/cobra"
)

func fmtCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "fmt",
		Short: "Format a game file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := game.ReadGameFile(args[0])
			if err != nil {
				return err
			}
			g.File.Write(cmd.OutOrStdout())
			return nil
		},
	}
	return c
}
