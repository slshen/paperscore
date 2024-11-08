package cmd

import (
	"fmt"
	"os"

	"github.com/slshen/paperscore/pkg/dataframe"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/stats"
	"github.com/spf13/cobra"
)

func reCommand() *cobra.Command {
	var (
		csv        bool
		yamlFormat bool
		freq       bool
		pivot      bool
		raw        bool
		plain      bool
	)
	re24 := &stats.ObservedRunExpectancy{}
	c := &cobra.Command{
		Use:   "re",
		Short: "Determine the run expectancy matrix",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGames(args)
			if err != nil {
				return err
			}
			for _, g := range games {
				if err := re24.Read(g); err != nil {
					return err
				}
			}
			if yamlFormat {
				return re24.WriteYAML(os.Stdout)
			}
			var data *dataframe.Data
			switch {
			case raw:
				data = re24.GetRunData()
			case pivot || freq:
				rf := re24.GetRunExpectancyFrequency()
				if pivot {
					data = rf.Pivot()
				} else {
					data = &rf.Data
				}
			default:
				data = stats.GetRunExpectancyData(re24)
				if plain {
					data = data.Select(
						dataframe.Col("Runr"),
						dataframe.Col("0Out"),
						dataframe.Col("1Out"),
						dataframe.Col("2Out"),
					)
				}
			}
			if csv {
				return data.RenderCSV(os.Stdout, false)
			}
			fmt.Println(data)
			return nil
		},
	}
	c.Flags().BoolVar(&csv, "csv", false, "Print in CSV format")
	c.Flags().BoolVar(&yamlFormat, "yaml", false, "Print in YAML format")
	c.Flags().BoolVar(&freq, "freq", false, "Print the frequency of # runs scored per 24-base/out state")
	c.Flags().BoolVar(&pivot, "pivot", false, "Pivot the frequency data by runs")
	c.Flags().BoolVar(&raw, "raw", false, "Get the raw run data")
	c.Flags().BoolVar(&plain, "plain", false, "Write data without counts")
	return c
}

func reAnalysisCommand() *cobra.Command {
	var (
		re reArgs
	)
	c := &cobra.Command{
		Use:   "re-analysis",
		Short: "Analyze run expectancy",
		RunE: func(cmd *cobra.Command, args []string) error {
			re, err := re.getRunExpectancy()
			if err != nil {
				return err
			}
			if re == nil {
				return fmt.Errorf("no RE specified")
			}
			rea := stats.NewREAnalysis(re)
			fmt.Println(rea.Run())
			return nil
		},
	}
	re.registerFlags(c.Flags())
	return c
}
