package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
	readRange := "Class Data!A34"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s\n", row[0], row[4])
		}
	}
}

type SheetReader struct {
	srv           *sheets.Service
	spreadsheetId string
	sheetName     string
	columnStart   string
	columnEnd     string
	rowStart      int

	idx           int
	header bool

	ValueRenderOption    string
	DateTimeRenderOption string
}

func New(srv *sheets.Service, spreadsheetId, sheetName, columnStart, columnEnd string, rowStart int) *SheetReader {
	return &SheetReader{
		srv:                  srv,
		spreadsheetId:        spreadsheetId,
		sheetName:            sheetName,
		columnStart:          columnStart,
		columnEnd:            columnEnd,
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


// ReadAll reads all the remaining records from r.
// Each record is a slice of fields.
// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
// defined to read until EOF, it does not treat end of file as an error to be
// reported.
func (r *SheetReader) ReadAll() (records [][]string, err error) {
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
