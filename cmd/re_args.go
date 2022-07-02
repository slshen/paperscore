package cmd

import (
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/stats"
	"github.com/spf13/pflag"
)

type reArgs struct {
	gamesDir   string
	matrixFile string
}

func (re *reArgs) registerFlags(flags *pflag.FlagSet) {
	flags.StringVar(&re.gamesDir, "re-games", "", "Use an observed RE matrix from games in `dir`")
	flags.StringVar(&re.matrixFile, "re-matrix", "", "Use RE from a CSV `file`")
}

func (re *reArgs) getRunExpectancy() (stats.RunExpectancy, error) {
	if re.gamesDir != "" {
		games, err := game.ReadGames([]string{re.gamesDir})
		if err != nil {
			return nil, err
		}
		re := &stats.ObservedRunExpectancy{}
		for _, g := range games {
			if err := re.Read(g); err != nil {
				return nil, err
			}
		}
		return re, nil
	}
	if re.matrixFile != "" {
		return stats.ReadREMatrix(re.matrixFile)
	}
	return nil, nil
}
