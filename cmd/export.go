package cmd

import (
	"fmt"
	"strings"

	"github.com/slshen/sb/pkg/config"
	"github.com/slshen/sb/pkg/game"
	"github.com/slshen/sb/pkg/sheetexport"
	"github.com/spf13/cobra"
)

func exportCommand() *cobra.Command {
	var (
		us            string
		league        string
		spreadsheetID string
		dryRun        bool
		re            reArgs
	)
	c := &cobra.Command{
		Use:   "export",
		Short: "Export games and stats to Google sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if us == "" {
				return fmt.Errorf("--us is required")
			}
			config, err := config.NewConfig()
			if err != nil {
				return err
			}
			if spreadsheetID != "" {
				config.SpreadsheetID = spreadsheetID
			}
			sheets, err := sheetexport.NewSheetExport(config)
			if err != nil {
				return err
			}
			re, err := re.getRunExpectancy()
			if err != nil {
				return err
			}
			export, err := sheetexport.NewExport(sheets, re)
			export.DryRun = dryRun
			if err != nil {
				return err
			}
			export.Us = us
			export.League = strings.ToLower(league)
			games, err := game.ReadGameFiles(args)
			if err != nil {
				return err
			}
			return export.Export(games)
		},
	}
	re.registerFlags(c.Flags())
	c.Flags().StringVar(&us, "us", "", "Our `team`")
	c.Flags().StringVar(&league, "league", "", "Include only games in `league`")
	c.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "The spreadsheet to use")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "Print the sheets instead of uploading them")
	return c
}
