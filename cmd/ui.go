package cmd

import (
	"log"
	"os"

	"github.com/slshen/sb/pkg/ui"
	"github.com/spf13/cobra"
)

func uiCommand() *cobra.Command {
	var (
		debugOut string
		re       reArgs
	)
	c := &cobra.Command{
		Use:  "ui",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var path string
			if len(args) == 1 {
				path = args[0]
			}
			ui := ui.New()
			re, err := re.getRunExpectancy()
			if err != nil {
				return err
			}
			ui.RE = re
			if debugOut != "" {
				f, err := os.Create(debugOut)
				if err != nil {
					return err
				}
				defer f.Close()
				ui.Logger = log.New(f, "", log.LstdFlags)
			}
			return ui.Run(path)
		},
	}
	c.Flags().StringVar(&debugOut, "debug-out", "", "Write debug logs to FILE")
	re.registerFlags(c.Flags())
	return c
}
