package cmd

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func readCommand() *cobra.Command {
	var home, visitor bool
	c := &cobra.Command{
		Use:   "read",
		Short: "Read and print a score file",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGameFiles(args)
			for _, g := range games {
				states := g.GetStates()
				for _, state := range states {
					if home || visitor {
						if visitor && !state.Top() {
							continue
						}
						if home && state.Top() {
							continue
						}
					}
					d, _ := yaml.Marshal(state)
					fmt.Println(string(d))
				}
			}
			return err
		},
	}
	c.Flags().BoolVar(&home, "home", false, "Print only home plays")
	c.Flags().BoolVar(&visitor, "visitor", false, "Print only visitor plays")
	return c
}
