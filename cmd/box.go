package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/slshen/paperscore/pkg/boxscore"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/slshen/paperscore/pkg/stats"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type boxCmd struct {
	*cobra.Command
	yamlFormat   bool
	pdfFormat    bool
	scoringPlays bool
	plays        bool
	reArgs       reArgs
	outputDir    string
	re           stats.RunExpectancy
}

func (b *boxCmd) writeBox(g *game.Game, firstGame bool) error {
	if b.re == nil {
		re, err := b.reArgs.getRunExpectancy()
		if err != nil {
			return err
		}
		b.re = re
	}
	var out io.Writer
	if b.outputDir != "" {
		base := filepath.Base(g.File.Path)
		dot := strings.LastIndexByte(base, '.')
		if dot > 1 {
			base = base[:dot]
		}
		path := filepath.Join(b.outputDir, fmt.Sprintf("%s.pdf", base))
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	} else {
		out = os.Stdout
	}
	if b.pdfFormat {
		paps := exec.Command("paps", "--format=pdf", "--font=Courier New 10",
			"--left-margin=18", "--right-margin=18", "--top-margin=18", "--bottom-margin=18")
		w, err := paps.StdinPipe()
		paps.Stdout = out
		paps.Stderr = os.Stderr
		if err != nil {
			return err
		}
		defer w.Close()
		out = w
		if err := paps.Start(); err != nil {
			return err
		}
	}
	box, err := boxscore.NewBoxScore(g, b.re)
	if err != nil {
		return err
	}
	box.IncludeScoringPlays = b.scoringPlays
	box.IncludePlays = b.plays
	if b.yamlFormat {
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
	if b.pdfFormat && !firstGame && b.outputDir == "" {
		if _, err := out.Write([]byte{'\f'}); err != nil {
			return err
		}
	}
	return nil
}

func (b *boxCmd) init() *cobra.Command {
	b.Command = &cobra.Command{
		Use:   "box",
		Short: "Generate a box score",
		RunE: func(cmd *cobra.Command, args []string) error {
			games, err := game.ReadGames(args)
			if err != nil {
				return err
			}
			for i, g := range games {
				if err := b.writeBox(g, i == 0); err != nil {
					return err
				}
			}
			return nil
		},
	}
	flags := b.Flags()
	flags.BoolVar(&b.yamlFormat, "yaml", false, "")
	flags.BoolVar(&b.pdfFormat, "pdf", false, "Run paps to convert output to pdf")
	flags.BoolVar(&b.scoringPlays, "scoring", false, "Include scoring plays in box")
	flags.BoolVar(&b.plays, "plays", false, "Include play by play in box")
	flags.StringVar(&b.outputDir, "outdir", "", "Write individual box scores to this directory")
	b.reArgs.registerFlags(flags)
	return b.Command
}

func boxCommand() *cobra.Command {
	var b boxCmd
	return b.init()
}
