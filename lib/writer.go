package lib

import (
	"context"
	"fmt"
	"github.com/gocarina/gocsv"
	"google.golang.org/api/sheets/v4"
	"strings"
)

type SheetWriter struct {
	srv           *sheets.Service
	spreadsheetId string
	sheetName     string
	//columnStart   string
	// columnEnd     string
	// rowStart      int

	// idx    int
	header bool

	ValueRenderOption    string
	DateTimeRenderOption string

	data [][]string
	e    error
}

var _ gocsv.CSVWriter = &SheetWriter{}

func NewWriter(srv *sheets.Service, spreadsheetId, sheetName string) *SheetWriter {
	return &SheetWriter{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
	}
}

func (w *SheetWriter) Write(row []string) error {
	out := make([]string, len(row))
	copy(out, row)
	w.data = append(w.data, out)
	return nil
}

func (w *SheetWriter) Flush() {
	err := w.ensureSheet(w.sheetName)
	if err != nil {
		w.e = err
		return
	}

	// read first column
	readRange := fmt.Sprintf("%s!A:A", w.sheetName)
	resp, err := w.srv.Spreadsheets.Values.Get(w.spreadsheetId, readRange).
		ValueRenderOption(w.ValueRenderOption).
		DateTimeRenderOption(w.DateTimeRenderOption).
		Do()
	if err != nil {
		w.e = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		return
	}

	var vals sheets.ValueRange

	if len(resp.Values) == 0 {
		vals = sheets.ValueRange{
			MajorDimension: "ROWS",
			Range:          fmt.Sprintf("%s!A%d", w.sheetName, 1),
			Values:         make([][]interface{}, len(w.data)),
		}
		for i := range w.data {
			vals.Values[i] = make([]interface{}, len(w.data[i]))
			for j := range w.data[i] {
				vals.Values[i][j] = w.data[i][j]
			}
		}
	} else {
		// A1:C1

		// read first row == header row
		readRange := fmt.Sprintf("%s!1:1", w.sheetName)
		headerResp, err := w.srv.Spreadsheets.Values.Get(w.spreadsheetId, readRange).
			ValueRenderOption(w.ValueRenderOption).
			DateTimeRenderOption(w.DateTimeRenderOption).
			Do()
		if err != nil {
			w.e = fmt.Errorf("unable to retrieve data from sheet: %v", err)
			return
		}




		type Index struct {
			Before int
			After  int
		}

		headerMap := map[string]*Index{}
		headerLength := 0
		for i, header := range headerResp.Values[0] {
			headerMap[header.(string)] = &Index{
				Before: i,
				After:  -1,
			}
			headerLength++
		}
		newHeaderStart := headerLength
		var newHeaders []interface{}

		for i, header := range w.data[0] {
			if _, ok := headerMap[header]; ok {
				headerMap[header].After = i
			} else {
				headerMap[header] = &Index{
					Before: headerLength,
					After:  i,
				}
				newHeaders = append(newHeaders, header)
				headerLength++
			}
		}
		// 1:1

		idmap := map[int]int{}
		for _, index := range headerMap {
			if index.After != -1 {
				idmap[index.After] = index.Before
			}
		}

		if len(newHeaders) > 0 {
			var sb strings.Builder
			sb.WriteRune(rune( 'A' + newHeaderStart))
			headerVals := sheets.ValueRange{
				MajorDimension: "ROWS",
				Range:          fmt.Sprintf("%s!%s%d", w.sheetName, sb.String(), 1),
				Values: [][]interface{}{
					newHeaders,
				},
			}
			_, err = w.srv.Spreadsheets.Values.Append(w.spreadsheetId, headerVals.Range, &headerVals).
				IncludeValuesInResponse(false).
				InsertDataOption("OVERWRITE").
				ValueInputOption("USER_ENTERED").
				Do()
			if err != nil {
				w.e = fmt.Errorf("unable to write new headers to sheet: %v", err)
				return
			}
		}

		vals = sheets.ValueRange{
			MajorDimension: "ROWS",
			Range:          fmt.Sprintf("%s!A%d", w.sheetName, 1+len(resp.Values)),
			Values:         make([][]interface{}, len(w.data)-1), // skip header
		}
		d22 := w.data[1:]
		for i := range d22 {
			vals.Values[i] = make([]interface{}, headerLength) // header length
			for j := range d22[i] {
				vals.Values[i][idmap[j]] = d22[i][j]
			}
		}
	}

	_, err = w.srv.Spreadsheets.Values.Append(w.spreadsheetId, vals.Range, &vals).
		IncludeValuesInResponse(false).
		InsertDataOption("INSERT_ROWS").
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		w.e = fmt.Errorf("unable to write data to sheet: %v", err)
		return
	}
}

func (w *SheetWriter) Error() error {
	return w.e
}

func (w *SheetWriter) getSheetId(name string) (int64, error) {
	resp, err := w.srv.Spreadsheets.Get(w.spreadsheetId).Do()
	if err != nil {
		return -1, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	var id int64
	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == name {
			id = sheet.Properties.SheetId
		}

	}

	return id, nil
}

func (w *SheetWriter) addNewSheet(name string) error {
	req := sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: name,
			},
		},
	}

	rbb := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&req},
	}

	_, err := w.srv.Spreadsheets.BatchUpdate(w.spreadsheetId, rbb).Context(context.Background()).Do()
	if err != nil {
		return err
	}

	return nil
}

func (w *SheetWriter) ensureSheet(name string) error {
	sheetId, err := w.getSheetId(name)
	if err != nil {
		return err
	}
	if sheetId != 0 {
		return nil
	}
	return w.addNewSheet(name)
}
