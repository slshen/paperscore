package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/slshen/sb/pkg/gamefile"
	"github.com/spf13/cobra"
)

func writeNewGame(gm *gamefile.File, nextDay bool) (*gamefile.File, error) {
	ng := &gamefile.File{}
	date, err := gm.GetGameDate()
	if err != nil {
		return nil, fmt.Errorf("%s does not have a game date - %w", gm.Path, err)
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
		return nil, err
	}
	file := filepath.Join(filepath.Dir(gm.Path), fmt.Sprintf("%s-%s.gm", date.Format("20060102"), numberString))
	fmt.Printf("Creating new game %s\n", file)
	ng.Path = file
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY | os.O_EXCL
	f, err := os.OpenFile(file, flags, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ng.Write(f)
	return ng, nil
}

func newGameCommand() *cobra.Command {
	var (
		nextDay bool
		count   int
	)
	c := &cobra.Command{
		Use:  "new-game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gm, err := gamefile.ParseFile(args[0])
			if err != nil {
				return err
			}
			for i := 0; i < count; i++ {
				ng, err := writeNewGame(gm, nextDay)
				if err != nil {
					return err
				}
				nextDay = false
				gm = ng
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&nextDay, "next-day", "n", false, "Create a game for the next day")
	c.Flags().IntVar(&count, "count", 1, "Create `N` new games")
	return c
}
