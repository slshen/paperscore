package cmd

import (
	"fmt"
	"log"

	"github.com/slshen/paperscore/pkg/dataexport"
	"github.com/slshen/paperscore/pkg/dataframe/pkg"
	"github.com/slshen/paperscore/pkg/game"
	"github.com/spf13/cobra"
)

func dataExportCommand() *cobra.Command {
	var (
		re       reArgs
		id       string
		dir      string
		gameDirs []string
	)
	c := &cobra.Command{
		Use: "data-export",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			log.Printf("read %d games\n", len(games))
			exp := dataexport.NewDataExport(re)
			dp, err := exp.Read(games)
			if err != nil {
				return err
			}
			dp.ID = id
			dp.Title = "Softball data"
			dp.Licenses = []pkg.License{pkg.CopyrightAuthors}
			return dp.Write(dir)
		},
	}
	flags := c.Flags()
	re.registerFlags(flags)
	flags.StringVar(&id, "id", "", "The export ID")
	flags.StringVarP(&dir, "dir", "d", "", "Write web data to `dir`")
	flags.StringSliceVar(&gameDirs, "games", nil, "Read games from dir")
	return c
}
