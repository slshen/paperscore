package gamefile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (f *File) WriteNewGame(nextDay bool) (*File, error) {
	ng := &File{}
	date, err := f.GetGameDate()
	if err != nil {
		return nil, fmt.Errorf("%s does not have a game date - %w", f.Path, err)
	}
	var numberString string
	if nextDay {
		date = date.AddDate(0, 0, 1)
		numberString = "1"
	} else {
		number, _ := strconv.Atoi(f.Properties["game"])
		numberString = fmt.Sprintf("%d", number+1)
	}
	dateString := date.Format(GameDateFormat)
	for _, prop := range f.PropertyList {
		switch {
		case prop.Key == "date":
			ng.PropertyList = append(ng.PropertyList,
				&Property{
					Key:   "date",
					Value: dateString,
				})
		case prop.Key == "game":
			ng.PropertyList = append(ng.PropertyList,
				&Property{
					Key:   "game",
					Value: numberString,
				})
		case prop.Key == "visitorid" || prop.Key == "homeid" ||
			prop.Key == "tournament" || prop.Key == "league" ||
			prop.Key == "timelimit":
			ng.PropertyList = append(ng.PropertyList, prop)
		default:
			ng.PropertyList = append(ng.PropertyList,
				&Property{
					Key: prop.Key,
				})
		}
	}
	ng.HomeEvents = []*Event{
		{Pitcher: "0"},
		{Play: &ActualPlay{
			PlateAppearance: "1",
			Batter:          "1",
			PitchSequence:   ".",
			Code:            "NP",
		}},
	}
	ng.VisitorEvents = []*Event{
		{Pitcher: "0"},
		{Play: &ActualPlay{
			PlateAppearance: "1",
			Batter:          "1",
			PitchSequence:   ".",
			Code:            "NP",
		}},
	}
	if err := ng.Validate(); err != nil {
		return nil, err
	}
	file := filepath.Join(filepath.Dir(f.Path), fmt.Sprintf("%s-%s.gm", date.Format("20060102"), numberString))
	ng.Path = file
	fd, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	ng.Write(fd)
	return ng, fd.Close()
}
