package cmd

import (
	"fmt"

	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/webdata"
	"github.com/spf13/cobra"
)

func webdataCommand() *cobra.Command {
	var (
		re       reArgs
		us       string
		dir      string
		gameDirs []string
	)
	c := &cobra.Command{
		Use: "webdata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if us == "" {
				return fmt.Errorf("--us is required")
			}
			if dir == "" {
				return fmt.Errorf("--dir is required")
			}
			re, err := re.getRunExpectancy()
			if err != nil {
				return err
			}
			if re == nil {
				return fmt.Errorf("run expectancy required")
			}
			games, err := game.ReadGames(gameDirs)
			if err != nil {
				return err
			}
			wdat, err := webdata.NewWebData(us, re, games)
			if err != nil {
				return err
			}
			return wdat.Write(dir)
		},
	}
	flags := c.Flags()
	re.registerFlags(flags)
	flags.StringVar(&us, "us", "", "The us team")
	flags.StringVarP(&dir, "dir", "d", "", "Write web data to `dir`")
	flags.StringSliceVar(&gameDirs, "games", nil, "Read games from dir")
	return c
}
