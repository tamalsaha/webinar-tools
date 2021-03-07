package main

import (
	"context"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/tamalsaha/webinar-tools/lib"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
)

type Client2 struct { // Our example struct, you can use "-" to ignore a field
	Name                    string `csv:"Student Name"`
	Gender                  string `csv:"Gender"`
	ClassLevel              string `csv:"Class Level"`
	HomeState               string `csv:"Home State"`
	Major                   string `csv:"Major"`
	ExtracurricularActivity string `csv:"Extracurricular Activity"`
}

type Client struct { // Our example struct, you can use "-" to ignore a field
	Id      string `csv:"client_id"`
	Name    string `csv:"client_name"`
	Age     string `csv:"client_age"`
	City    string `csv:"city"`
	Country string `csv:"country"`
}

func main() {
	hc, err := gdrive.DefaultClient("/home/tamal/go/src/github.com/tamalsaha/webinar-tools")
	if err != nil {
		panic(err)
	}

	srv, err := sheets.NewService(context.TODO(), option.WithHTTPClient(hc))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	// spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"

	spreadsheetId := "18zl47TxgtdRxnzO-E47lyE_pV5Na1JbCAEVJKQ-PY20"

	//readRange := "Class Data!A1:F31"
	//resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	//if err != nil {
	//	log.Fatalf("Unable to retrieve data from sheet: %v", err)
	//}
	//
	//if len(resp.Values) == 0 {
	//	fmt.Println("No data found.")
	//} else {
	//	fmt.Println("Name, Major:")
	//	for _, row := range resp.Values {
	//		// Print columns A and E, which correspond to indices 0 and 4.
	//		fmt.Printf("%s, %s\n", row[0], row[4])
	//	}
	//}

	clients := []*Client{}

	reader := lib.NewReader(srv, spreadsheetId, "clients", "A", 1)
	if err := gocsv.UnmarshalCSV(reader, &clients); err != nil { // Load clients from file
		panic(err)
	}
	for _, client := range clients {
		fmt.Println("Hello", client.Name)
	}

	clients = []*Client{}
	clients = append(clients, &Client{Id: "12", Name: "John", Age: "21", City: "LV", Country: "US"}) // Add clients
	clients = append(clients, &Client{Id: "13", Name: "Fred", City: "Dhaka", Country: "BD"})
	clients = append(clients, &Client{Id: "14", Name: "James", Age: "32", City: "LA", Country: "US"})
	clients = append(clients, &Client{Id: "15", Name: "Danny"})

	writer := lib.NewWriter(srv, spreadsheetId, "clients2")
	err = gocsv.MarshalCSV(clients, writer)
	if err != nil {
		panic(err)
	}
}
