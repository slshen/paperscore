package cmd

import (
	"fmt"
	"os"

	"github.com/slshen/sb/pkg/dataframe"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/spf13/cobra"
)

func statsCommand(statsType string) *cobra.Command {
	var (
		csv bool
		re  reArgs
	)
	mg := stats.NewGameStats(nil)
	c := &cobra.Command{
		Use:     fmt.Sprintf("%s-stats", statsType),
		Aliases: []string{statsType},
		Short:   "Print stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGameFiles(args)
			if err != nil {
				return err
			}
			mg.RE, err = re.getRunExpectancy()
			if err != nil {
				return err
			}
			for _, g := range games {
				if err := mg.Read(g); err != nil {
					return err
				}
			}
			var data *dataframe.Data
			if statsType == "batting" {
				data = mg.GetBattingData()
			} else {
				data = mg.GetPitchingData()
			}
			if csv {
				return data.RenderCSV(os.Stdout)
			} else {
				fmt.Println(data)
			}
			return nil
		},
	}
	re.registerFlags(c.Flags())
	c.Flags().BoolVar(&csv, "csv", false, "Print in CSV format")
	return c
}
