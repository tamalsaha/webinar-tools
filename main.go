package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"log"
	"time"
)

func main() {
	client, err := gdrive.DefaultClient(".")
	if err != nil {
		log.Fatalf("Unable to create client: %v", err)
	}
	srv, err := calendar.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}

	fmt.Println("___________")

	calendarId := "c_oravu1d4snmip0784jpfkit8go@group.calendar.google.com" // Test

	es, err := srv.Events.List(calendarId).Do()
	if err != nil {
		panic(err)
	}
	for _, e2 := range es.Items {
		data, _ := json.MarshalIndent(e2.ConferenceData, "", "  ")
		fmt.Println(string(data))
	}

	cals, err := srv.CalendarList.List().ShowDeleted(false).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	if len(cals.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range cals.Items {
			fmt.Printf("%v|%v|%v|%v \n", item.Id, item.Summary, item.Primary, item.Description)
		}
	}

	fmt.Println("___________")

	// Refer to the Go quickstart on how to setup the environment:
	// https://developers.google.com/calendar/quickstart/go
	// Change the scope to calendar.CalendarScope and delete any stored credentials.

	start := time.Now()
	end := start.Add(30 * time.Minute)

	fmt.Println("*****", start.UTC().Format(time.RFC3339))
	fmt.Println("*****", start.UTC().Location().String())

	event := &calendar.Event{
		Summary: "Google I/O 2015",
		Location: "800 Howard St., San Francisco, CA 94103",
		Description: "A chance to hear more about Google's developer products.",
		Start: &calendar.EventDateTime{
			DateTime: start.UTC().Format(time.RFC3339),
			TimeZone: start.UTC().Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: end.UTC().Format(time.RFC3339),
			TimeZone: end.UTC().Location().String(),
		},
		Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=2"},
		Attendees: []*calendar.EventAttendee{
			&calendar.EventAttendee{Email:"lpage@example.com"},
			&calendar.EventAttendee{Email:"sbrin@example.com"},
		},
		ConferenceData: &calendar.ConferenceData{
			ConferenceId:       "",
			ConferenceSolution: nil,
			CreateRequest:      nil,
			EntryPoints:        nil,
			Notes:              "",
			Parameters:         nil,
			Signature:          "",
			ForceSendFields:    nil,
			NullFields:         nil,
		},
	}

	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)
}