package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"

	"github.com/slshen/sb/pkg/boxscore"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/playbyplay"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func Root() *cobra.Command {
	root := &cobra.Command{}
	root.SilenceUsage = true
	root.AddCommand(readCommand(), boxCommand(), playByPlayCommand())
	return root
}

func readCommand() *cobra.Command {
	var home, visitor bool
	c := &cobra.Command{
		Use:   "read",
		Short: "Read and print a score file",
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args
			sort.Strings(files)
			for _, path := range files {
				g, err := game.ReadGameFile(path)
				if err != nil {
					return err
				}
				states, err := g.GetStates()
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
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&home, "home", false, "Print only home plays")
	c.Flags().BoolVar(&visitor, "visitor", false, "Print only visitor plays")
	return c
}

func boxCommand() *cobra.Command {
	var (
		yamlFormat   bool
		pdfFormat    bool
		scoringPlays bool
		plays        bool
	)
	c := &cobra.Command{
		Use:   "box",
		Short: "Generate a box score",
		RunE: func(cmd *cobra.Command, args []string) error {
			var out io.Writer
			if pdfFormat {
				paps := exec.Command("paps", "--format=pdf", "--font=Andale Mono 10",
					"--left-margin=18", "--right-margin=18", "--top-margin=18", "--bottom-margin=18")
				w, err := paps.StdinPipe()
				paps.Stdout = os.Stdout
				paps.Stderr = os.Stderr
				if err != nil {
					return err
				}
				defer w.Close()
				out = w
				if err := paps.Start(); err != nil {
					return err
				}
			} else {
				out = os.Stdout
			}
			files := args
			sort.Strings(files)
			for i, path := range files {
				g, err := game.ReadGameFile(path)
				if err != nil {
					return err
				}
				box, err := boxscore.NewBoxScore(g)
				if err != nil {
					return err
				}
				box.IncludeScoringPlays = scoringPlays
				box.IncludePlays = plays
				if yamlFormat {
					dat, err := yaml.Marshal(box)
					if err != nil {
						return err
					}
					if _, err := out.Write(dat); err != nil {
						return err
					}
				} else if err := box.Render(out); err != nil {
					return err
				}
				if i != len(files)-1 {
					if _, err := out.Write([]byte{'\f'}); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&yamlFormat, "yaml", false, "")
	c.Flags().BoolVar(&pdfFormat, "pdf", false, "Run paps to convert output to pdf")
	c.Flags().BoolVar(&scoringPlays, "scoring", false, "Include scoring plays in box")
	c.Flags().BoolVar(&plays, "plays", false, "Include play by play in box")
	return c
}

func playByPlayCommand() *cobra.Command {
	pbp := playbyplay.Generator{}
	c := &cobra.Command{
		Use:   "plays",
		Short: "Generate play by play",
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args
			sort.Strings(files)
			for _, path := range files {
				g, err := game.ReadGameFile(path)
				if err != nil {
					return err
				}
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
