package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	var dir string
	root := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if dir != "" {
				return os.Chdir(dir)
			}
			return nil
		},
	}
	root.SilenceUsage = true
	root.PersistentFlags().StringVar(&dir, "working-dir", "", "Change working directory to `dir`")
	root.AddCommand(readCommand(), boxCommand(), playByPlayCommand(),
		statsCommand("batting"), statsCommand("pitching"), reCommand(),
		tournamentCommand(), reAnalysisCommand(),
		fmtCommand(), altCommand(), webdataCommand(), newGameCommand(),
		battingCountCommand(), battingTimesSeenPitcherCommand(),
		pitchingTimesSeenLineupCommand(), simCommand(),
	)
	return root
}
