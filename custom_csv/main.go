package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	"io/ioutil"
	"sort"
	"strings"
	"time"
)

const (
	WebinarScheduleFormat = "1/2/2006 15:04:05"
)

type Dates []time.Time

// Convert the internal date as CSV string
func (date *Dates) MarshalCSV() (string, error) {
	if date == nil {
		return "", nil
	}

	dates := make([]time.Time, 0, len(*date))
	for _, d := range *date {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})
	parts := make([]string, 0, len(*date))
	for _, d := range dates {
		parts = append(parts, d.Format(WebinarScheduleFormat))
	}
	return strings.Join(parts, ","), nil
}

// Convert the CSV string as internal date
func (date *Dates) UnmarshalCSV(csv string) (err error) {
	parts := strings.Split(csv, ",")

	dates := make([]time.Time, 0, len(parts))
	for _, part := range parts {
		d, err := time.Parse(WebinarScheduleFormat, part)
		if err != nil {
			return err
		}
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	*date = dates
	return nil
}

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
	Title     string   `json:"title" csv:"Title" form:"title"`
	// Schedule  DateTime `json:"schedule" csv:"Schedule" form:"schedule"`
	Schedules Dates    `json:"schedules" csv:"Schedules" form:"schedules"`
}

//func main() {
//	clients := []*Client{}
//	now := time.Now()
//
//	clients = append(clients, &Client{Title: "w12", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
//	clients = append(clients, &Client{Title: "w13", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
//	clients = append(clients, &Client{Title: "w14", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
//	clients = append(clients, &Client{Title: "w15", Schedule: DateTime{now.Add(1 * 24 * time.Hour)}})
//	csvContent, err := gocsv.MarshalString(&clients) // Get all clients as CSV string
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(csvContent) // Display all clients as CSV string
//}

func main() {
	var schedules Dates
	err := schedules.UnmarshalCSV("4/18/2021 03:03:50,4/19/2021 03:03:50")
	if err != nil {
		panic(err)
	}

	clients := []*Client{}
	now := time.Now()

	clients = append(clients, &Client{Title: "w12", Schedules: Dates{now, now.Add(1 * 24 * time.Hour)}})
	clients = append(clients, &Client{Title: "w13", Schedules: Dates{now, now.Add(2 * 24 * time.Hour)}})
	clients = append(clients, &Client{Title: "w14", Schedules: Dates{now, now.Add(3 * 24 * time.Hour)}})
	clients = append(clients, &Client{Title: "w15", Schedules: Dates{now, now.Add(4 * 24 * time.Hour)}})
	csvContent, err := gocsv.MarshalString(&clients) // Get all clients as CSV string
	if err != nil {
		panic(err)
	}

	fmt.Println(csvContent) // Display all clients as CSV string

	data, err := json.MarshalIndent(clients, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	err = ioutil.WriteFile("schedules.csv", []byte(csvContent), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("-----------------------------------------------------------")

	c2 := []*Client{}
	if err := gocsv.UnmarshalCSV(gocsv.DefaultCSVReader(strings.NewReader(csvContent)), &c2); err != nil { // Load clients from file
		panic(err)
	}
	for _, client := range c2 {
		fmt.Printf("%+v\n", client)
	}
}
