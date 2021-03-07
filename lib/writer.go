package lib

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"google.golang.org/api/sheets/v4"
	"io"
)

type SheetWriter struct {
	srv           *sheets.Service
	spreadsheetId string
	sheetName     string
	columnStart   string
	columnEnd     string
	rowStart      int

	idx    int
	header bool

	ValueRenderOption    string
	DateTimeRenderOption string

	vals *sheets.ValueRange
	e    error
}

var _ gocsv.CSVWriter = &SheetWriter{}

func NewWriter(srv *sheets.Service, spreadsheetId, sheetName, columnStart, columnEnd string, rowStart int) *SheetWriter {
	return &SheetWriter{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		columnStart:          columnStart,
		columnEnd:            columnEnd,
		rowStart:             rowStart,
		idx:                  rowStart,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",

		vals: &sheets.ValueRange{},
	}
}

func (w *SheetWriter) Write(row []string) error {
	out := make([]interface{}, len(row))
	for i := range row {
		out[i] = row[i]
	}
	w.vals.Values = append(w.vals.Values, out)
	return nil
}

func (w *SheetWriter) Flush() {
	// Add header if needed, else just append
	readRange := fmt.Sprintf("%s!A:A", w.sheetName)
	resp, err := w.srv.Spreadsheets.Values.Get(w.spreadsheetId, readRange).
		ValueRenderOption(w.ValueRenderOption).
		DateTimeRenderOption(w.DateTimeRenderOption).
		Do()
	if err != nil {
		w.e = fmt.Errorf("unable to retrieve data from sheet: %v", err)
		return
	}
	if resp.Values == nil {
		// TODO: WHEN THIS happens?
		// Not matching sheet?
		// matching sheet but empty
		w.e = io.EOF
		return
	}
	if len(resp.Values) == 0 {
		// add header
	}
	writeRange := fmt.Sprintf("%s!A:A", w.sheetName)
	_, err = w.srv.Spreadsheets.Values.Append(w.spreadsheetId, writeRange, w.vals).
		IncludeValuesInResponse(false).
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		w.e = fmt.Errorf("unable to write data to sheet: %v", err)
		return
	}
}

func (w *SheetWriter) Error() error {
	return w.e
}
