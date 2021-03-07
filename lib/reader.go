package lib

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"google.golang.org/api/sheets/v4"
	"io"
	"strings"
)

type SheetReader struct {
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
}

var _ gocsv.CSVReader = &SheetReader{}

func NewReader(srv *sheets.Service, spreadsheetId, sheetName, columnStart string, rowStart int) *SheetReader {
	return &SheetReader{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		columnStart:          columnStart,
		rowStart:             rowStart,
		idx:                  rowStart,
		ValueRenderOption:    "FORMATTED_VALUE",
		DateTimeRenderOption: "SERIAL_NUMBER",
	}
}

// Read reads one record (a slice of fields) from r.
// If the record has an unexpected number of fields,
// Read returns the record along with the error ErrFieldCount.
// Except for that case, Read always returns either a non-nil
// record or a non-nil error, but not both.
// If there is no data left to be read, Read returns nil, io.EOF.
// If ReuseRecord is true, the returned slice may be shared
// between multiple calls to Read.
func (r *SheetReader) Read() (record []string, err error) {
	if !r.header && r.idx > 1 {
		record, err = r.read(1)
		if err != nil {
			return nil, err
		}
		r.header = true
		return record, err
	}

	record, err = r.read(r.idx)
	if err != nil {
		return nil, err
	}
	r.idx++
	return record, nil
}

func (r *SheetReader) read(idx int) (record []string, err error) {
	if r.columnEnd == "" {
		columnEnd, err := r.readHeader()
		if err != nil {
			return nil, err
		}
		r.columnEnd = columnEnd
	}

	readRange := fmt.Sprintf("%s!%s%d:%s%d", r.sheetName, r.columnStart, idx, r.columnEnd, idx)
	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetId, readRange).
		ValueRenderOption(r.ValueRenderOption).
		DateTimeRenderOption(r.DateTimeRenderOption).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if resp.Values == nil {
		return nil, io.EOF
	}
	if len(resp.Values) > 1 {
		return nil, fmt.Errorf("multiple rows returned")
	}

	record = make([]string, len(resp.Values[0]))
	for i := range resp.Values[0] {
		record[i] = fmt.Sprintf("%v", resp.Values[0][i])
	}
	return record, nil
}

func (r *SheetReader) readHeader() (string, error) {
	readRange := fmt.Sprintf("%s!1:1", r.sheetName)
	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetId, readRange).
		ValueRenderOption(r.ValueRenderOption).
		DateTimeRenderOption(r.DateTimeRenderOption).
		Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return "", io.EOF
	}
	var sb strings.Builder
	sb.WriteRune(rune('A' + len(resp.Values[0])))
	return sb.String(), nil
}

// ReadAll reads all the remaining records from r.
// Each record is a slice of fields.
// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
// defined to read until EOF, it does not treat end of file as an error to be
// reported.
func (r *SheetReader) ReadAll() (records [][]string, err error) {
	if r.columnEnd == "" {
		columnEnd, err := r.readHeader()
		if err != nil {
			return nil, err
		}
		r.columnEnd = columnEnd
	}

	readRange := fmt.Sprintf("%s!%s%d:%s", r.sheetName, r.columnStart, r.idx, r.columnEnd)
	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetId, readRange).
		ValueRenderOption(r.ValueRenderOption).
		DateTimeRenderOption(r.DateTimeRenderOption).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	if resp.Values == nil {
		return nil, io.EOF
	}

	offset := 0
	if !r.header && r.idx > 1 {
		records = make([][]string, len(resp.Values)+1)
		records[0], err = r.read(1)
		if err != nil {
			return nil, err
		}
		r.header = true
		offset = 1
	} else {
		records = make([][]string, len(resp.Values))
		if r.idx == 1 {
			r.header = true
		}
	}

	r.idx += len(resp.Values)
	for i, row := range resp.Values {
		records[i] = make([]string, len(row))
		for j := range row {
			records[i+offset][j] = fmt.Sprintf("%v", row[j])
		}
	}
	return records, nil
}
