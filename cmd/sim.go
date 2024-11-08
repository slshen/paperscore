package cmd

import (
	"fmt"

	"github.com/slshen/paperscore/pkg/markov"
	"github.com/slshen/paperscore/pkg/markov/expr"
	"github.com/spf13/cobra"
)

func simCommand() *cobra.Command {
	var (
		model   string
		innings int
		games   int
		trace   bool
	)
	c := &cobra.Command{
		Use: "sim",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := expr.ParseFile(model)
			if err != nil {
				return err
			}
			m, diags := expr.NewModel(f)
			if err := diags.ErrorOrNil(); err != nil {
				return err
			}
			sim := markov.Simulation{
				Model: m,
			}
			var runs float64
			for i := 0; i < games; i++ {
				sim.Runs = 0
				for j := 0; j < innings; j++ {
					if trace {
						sim.Trace = make([]markov.Event, 0)
					}
					if err := sim.RunInning(); err != nil {
						return err
					}
					if trace {
						var r float64
						for _, t := range sim.Trace {
							fmt.Printf("%s %3s %.0f -> %s\n", t.State, t.Event, t.Runs, t.Next)
							r += t.Runs
							if t.Next == markov.EndState {
								fmt.Printf("%0.f runs\n", r)
								r = 0
							}
						}
					}
				}
				runs += sim.Runs
			}
			fmt.Printf("%d games %f runs %f runs/game\n", games, runs, runs/float64(games))
			re := sim.GetExpectedRuns()
			fmt.Printf("Rnr  0out  1out  2out\n")
			for i := 0; i < 8; i++ {
				out0 := markov.BaseOutState(0 + i)
				out1 := markov.BaseOutState(8 + i)
				out2 := markov.BaseOutState(16 + i)
				fmt.Printf("%s  %4.2f  %4.2f  %4.2f\n", out0.String()[1:4], re[out0], re[out1], re[out2])
			}
			return nil
		},
	}
	flags := c.Flags()
	flags.StringVar(&model, "model", "", "The model file")
	_ = c.MarkFlagRequired("model")
	flags.IntVar(&innings, "innings", 5, "The number of innings per game")
	flags.IntVarP(&games, "games", "n", 100, "The number of games to simulate")
	flags.BoolVar(&trace, "trace", false, "Generate game traces")
	return c
}
