package export

import (
	"context"
	"fmt"
	"log"

	"github.com/slshen/sb/pkg/dataframe"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type SheetExport struct {
	SpreadsheetID string

	spreadsheet *sheets.Spreadsheet
	service     *sheets.Service
}

func NewSheetExport(config *Config) (*SheetExport, error) {
	ex := &SheetExport{
		SpreadsheetID: config.SpreadsheetID,
	}
	conf, err := google.JWTConfigFromJSON(config.jsonKey, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}
	conf.Subject = config.UserEmail
	client := conf.Client(context.Background())
	ex.service, err = sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	if err := ex.reload(); err != nil {
		return nil, err
	}
	log.Default().Printf("Loaded spreadsheet %s", config.SpreadsheetID)
	return ex, err
}

func (ex *SheetExport) reload() error {
	var err error
	ex.spreadsheet, err = ex.service.Spreadsheets.Get(ex.SpreadsheetID).
		IncludeGridData(true).Do()
	return err
}

func (ex *SheetExport) ExportData(data *dataframe.Data) error {
	_, err := ex.findOrCreateSheet(data.Name)
	if err != nil {
		return err
	}
	crange := fmt.Sprintf("%s!A2:%s1000", data.Name, columnLetters(len(data.Columns)-1))
	log.Default().Printf("Clearing %s of sheet %s", crange, data.Name)
	_, err = ex.service.Spreadsheets.Values.Clear(ex.SpreadsheetID,
		crange,
		&sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return err
	}
	headerRow := make([]interface{}, len(data.Columns))
	for i, col := range data.Columns {
		headerRow[i] = col.Name
	}
	values := [][]interface{}{headerRow}
	data.RApply(func(row int) {
		values = append(values, data.GetRow(row))
	})
	if data.HasSummary() {
		row := make([]interface{}, len(data.Columns))
		for i, col := range data.Columns {
			row[i] = col.GetSummary()
		}
		values = append(values, row)
	}
	vrange := fmt.Sprintf("%s!A1:%s%d", data.Name, columnLetters(len(data.Columns)-1),
		len(values)+1)
	log.Default().Printf("Updated values of %s in range %s", data.Name, vrange)
	_, err = ex.service.Spreadsheets.Values.Update(ex.SpreadsheetID, vrange, &sheets.ValueRange{
		Values: values,
	}).
		ValueInputOption("USER_ENTERED").Do()
	return err
}

func columnLetters(n int) string {
	var letters []rune
	for {
		letters = append(letters, rune('A'+(n%26)))
		n /= 26
		if n == 0 {
			break
		}
		n--
	}
	for i, j := 0, len(letters)-1; i < j; {
		letters[i], letters[j] = letters[j], letters[i]
		i++
		j--
	}
	return string(letters)
}

func (ex *SheetExport) findOrCreateSheet(sheetName string) (*sheets.Sheet, error) {
	once := false
	for {
		for i := range ex.spreadsheet.Sheets {
			sheet := ex.spreadsheet.Sheets[i]
			if sheet.Properties.Title == sheetName {
				log.Default().Printf("Found sheet %s index %d", sheetName, sheet.Properties.Index)
				return sheet, nil
			}
		}
		if once {
			return nil, fmt.Errorf("cannot find sheet after adding")
		}
		once = true
		_, err := ex.service.Spreadsheets.BatchUpdate(ex.SpreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					AddSheet: &sheets.AddSheetRequest{
						Properties: &sheets.SheetProperties{
							Title: sheetName,
						},
					},
				},
			},
		}).Do()
		if err != nil {
			return nil, err
		}
		log.Default().Printf("Created sheet %s", sheetName)
		if err := ex.reload(); err != nil {
			return nil, err
		}
	}
}
