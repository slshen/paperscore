package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/slshen/sb/pkg/gamefile"
	"github.com/spf13/cobra"
)

func newGameCommand() *cobra.Command {
	var (
		nextDay bool
		force   bool
	)
	c := &cobra.Command{
		Use:  "new-game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gm, err := gamefile.ParseFile(args[0])
			if err != nil {
				return err
			}
			ng := &gamefile.File{}
			date, err := gm.GetGameDate()
			if err != nil {
				return fmt.Errorf("%s does not have a game date - %w", args[0], err)
			}
			var numberString string
			if nextDay {
				date = date.AddDate(0, 0, 1)
				numberString = "1"
			} else {
				number, _ := strconv.Atoi(gm.Properties["game"])
				numberString = fmt.Sprintf("%d", number+1)
			}
			dateString := date.Format(gamefile.GameDateFormat)
			for _, prop := range gm.PropertyList {
				switch {
				case prop.Key == "date":
					ng.PropertyList = append(ng.PropertyList,
						&gamefile.Property{
							Key:   "date",
							Value: dateString,
						})
				case prop.Key == "game":
					ng.PropertyList = append(ng.PropertyList,
						&gamefile.Property{
							Key:   "game",
							Value: numberString,
						})
				case prop.Key == "visitorid" || prop.Key == "homeid" ||
					prop.Key == "tournament" || prop.Key == "league" ||
					prop.Key == "timelimit":
					ng.PropertyList = append(ng.PropertyList, prop)
				default:
					ng.PropertyList = append(ng.PropertyList,
						&gamefile.Property{
							Key: prop.Key,
						})
				}
			}
			if err := ng.Validate(); err != nil {
				return err
			}
			file := filepath.Join(filepath.Dir(args[0]), fmt.Sprintf("%s-%s.gm", date.Format("20060102"), numberString))
			fmt.Printf("Creating new game %s\n", file)
			flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
			if !force {
				flags |= os.O_EXCL
			}
			f, err := os.OpenFile(file, flags, 0666)
			if err != nil {
				return err
			}
			defer f.Close()
			ng.Write(f)
			return nil
		},
	}
	c.Flags().BoolVarP(&nextDay, "next-day", "n", false, "Create a game for the next day")
	c.Flags().BoolVar(&force, "force", false, "Create the new game file even if it already exists")
	return c
}
