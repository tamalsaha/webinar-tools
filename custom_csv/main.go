package main

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"time"
)

const (
	WebinarScheduleFormat = "1/2/2006 15:04:05"
)

type DateTime struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (string, error) {
	return date.Time.Format(WebinarScheduleFormat), nil
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(WebinarScheduleFormat, csv)
	return err
}

type Client struct {
	Title    string   `json:"title" csv:"Title" form:"title"`
	Schedule DateTime `json:"schedule" csv:"Schedule" form:"schedule"`
}

func main() {
	clients := []*Client{}
	now := time.Now()

	clients = append(clients, &Client{Title: "w12", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
	clients = append(clients, &Client{Title: "w13", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
	clients = append(clients, &Client{Title: "w14", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
	clients = append(clients, &Client{Title: "w15", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
	csvContent, err := gocsv.MarshalString(&clients) // Get all clients as CSV string
	if err != nil {
		panic(err)
	}

	fmt.Println(csvContent) // Display all clients as CSV string
}
