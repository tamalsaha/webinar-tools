package main

import (
	"context"
	"encoding/json"
	"fmt"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	passgen "gomodules.xyz/password-generator"
	"github.com/himalayan-institute/zoom-lib-golang"
)

// ExampleWebinar contains examples for the /webinar endpoints
func main__() {
	var (
		apiKey          = os.Getenv("ZOOM_API_KEY")
		apiSecret       = os.Getenv("ZOOM_API_SECRET")
		email           = os.Getenv("ZOOM_ACCOUNT_EMAIL")
	)

	zoom.APIKey = apiKey
	zoom.APISecret = apiSecret
	zoom.Debug = true

	user, err := zoom.GetUser(zoom.GetUserOpts{EmailOrID: email})
	if err != nil {
		log.Fatalf("got error listing users: %+v\n", err)
	}
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	ms, err := zoom.ListMeetings(zoom.ListMeetingsOptions{
		HostID:     user.ID,
		Type:       zoom.ListMeetingTypeUpcoming,
		PageSize:   nil,
		PageNumber: nil,
	})
	if err != nil {
		panic(err)
	}
	for _, m := range ms.Meetings {
		data, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		fmt.Println("_______________________________*************")
	}

	start := time.Now().Add(60 *time.Minute)

	meeting, err := zoom.CreateMeeting(zoom.CreateMeetingOptions{
		HostID:         user.ID,
		Topic:          "Test Zoom API",
		Type:           zoom.MeetingTypeScheduled,
		StartTime:     &zoom.Time{
			Time:start,
		},
		Duration:       25,
		Timezone:       start.Location().String(),
		Password:       passgen.GenerateForCharset(10, passgen.AlphaNum),
		Agenda:         `Solve World Hunger
and also
Corona`,
		TrackingFields: nil,
		Settings:       zoom.MeetingSettings{
			HostVideo:                    false,
			ParticipantVideo:             false,
			ChinaMeeting:                 false,
			IndiaMeeting:                 false,
			JoinBeforeHost:               true,
			MuteUponEntry:                true,
			Watermark:                    false,
			UsePMI:                       false,
			ApprovalType:                 zoom.ApprovalTypeManuallyApprove,
			RegistrationType:             zoom.RegistrationTypeRegisterEachTime,
			Audio:                        zoom.AudioBoth,
			AutoRecording:                zoom.AutoRecordingLocal,
			CloseRegistration:            false,
			WaitingRoom:                  false,
		},
	})
	if err != nil {
		panic(err)
	}
	data2, err := json.MarshalIndent(meeting, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data2))
}

func main() {
	user := zoom.User{
		Email:                            "tamal@appscode.com",
	}

	content, err := ioutil.ReadFile("create_meeting_response.json")
	if err != nil {
		panic(err)
	}

	var meeting zoom.Meeting
	err = json.Unmarshal(content, &meeting)
	if err != nil {
		panic(err)
	}

	create_calendar_event(user, meeting)
}

func create_calendar_event(user zoom.User, meeting zoom.Meeting) {
	client, err := gdrive.DefaultClient(".")
	if err != nil {
		log.Fatalf("Unable to create client: %v", err)
	}
	srv, err := calendar.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	calendarId := "c_oravu1d4snmip0784jpfkit8go@group.calendar.google.com" // Test

	// Refer to the Go quickstart on how to setup the environment:
	// https://developers.google.com/calendar/quickstart/go
	// Change the scope to calendar.CalendarScope and delete any stored credentials.

	start := time.Now().Add(60*time.Minute)
	end := start.Add(30 * time.Minute)

	fmt.Println("*****", start.UTC().Format(time.RFC3339))
	fmt.Println("*****", start.UTC().Location().String())

	var phones []string
	for _, num := range meeting.Settings.GobalDialInNumbers {
		if num.Country == "US" && num.Type == "toll" {
			phones = append(phones, num.Number)
		}
	}

	event := &calendar.Event{
		Summary:     "Google I/O 2015",
		// Location:    "800 Howard St., San Francisco, CA 94103",
		Description: "A chance to hear more about Google's developer products.",
		Start: &calendar.EventDateTime{
			DateTime: start.UTC().Format(time.RFC3339),
			TimeZone: start.UTC().Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: end.UTC().Format(time.RFC3339),
			TimeZone: end.UTC().Location().String(),
		},
		// Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=2"},
		Attendees: []*calendar.EventAttendee{
			{Email: "lpage@example.com"},
		},
		ConferenceData: &calendar.ConferenceData{
			ConferenceId: fmt.Sprintf("%d", meeting.ID),
			ConferenceSolution: &calendar.ConferenceSolution{
				IconUri: "https://lh3.googleusercontent.com/ugWKRyPiOCwjn5jfaoVgC-O80F3nhKH1dKMGsibXvGV1tc6pGXLOJk9WO7dwhw8-Xl9IwkKZEFDbeMDgnx-kf8YGJZ9uhrJMK9KP8-ISybmbgg1LK121obq2o5ML0YugbWh-JevWMu4FxxTKzM2r68bfDG_NY-BNnHSG7NcOKxo-RE7dfObk3VkycbRZk_GUK_1UUI0KitNg7HBfyqFyxIPOmds0l-h2Q1atWtDWLi29n_2-s5uw_kV4l2KeeaSGws_x8h0zsYWLDP5wdKWwYMYiQD2AFM32SHJ4wLAcAKnwoZxUSyT_lWFTP0PHJ6PwETDGNZjmwh3hD6Drn7Z3mnX662S6tkoPD92LtMDA0eNLr6lg-ypI2fhaSGKOeWFwA5eFjds7YcH-axp6cleuiEp05iyPO8uqtRDRMEqQhPaiRTcw7pY5UAVbz2yXbMLVofojrGTOhdvlYvIdDgBOSUkqCshBDV4A2TJyDXxFjpSYaRvwwWIT0JgrIxLpBhnyd3_w6m5My5QtgBJEt_S2Dq4bXwCAA7VcRiD61WmDyHfU3dBiWQUNjcH39IKT9V1fbUcUkfDPM_AGjp7pwgG3w9yDemGi1OGlRXS4pU7UwF24c2dozrmaK17iPaExn0cmIgtBnFUXReY48NI8h2MNd_QysNMWYNYbufoPD7trSu6nS39wlUDQer2V",
				Key: &calendar.ConferenceSolutionKey{
					Type: "addOn",
				},
				Name: "Zoom Meeting",
			},
			EntryPoints: []*calendar.EntryPoint{
				{
					EntryPointType: "video",
					Label:          strings.TrimPrefix(meeting.JoinURL, "https://"),
					MeetingCode:    fmt.Sprintf("%d", meeting.ID),
					Passcode:       fmt.Sprintf("%s", meeting.Password),
					Uri:            meeting.JoinURL,
				},
				//{
				//	EntryPointType: "phone",
				//	Label:          phones[0],
				//	RegionCode:     "US",
				//	Passcode:       fmt.Sprintf("%d", meeting.Password),
				//	Uri:            fmt.Sprintf("tel:%s", strings.Join(phones, ",")),
				//},
				//{
				//	EntryPointType: "more",
				//	Uri:            "https://us02web.zoom.us/u/kp0VS4U41",
				//},
			},
			// Notes:              "",
			Parameters: &calendar.ConferenceParameters{
				AddOnParameters: &calendar.ConferenceParametersAddOnParameters{
					Parameters: map[string]string{
						"meetingCreatedBy": user.Email,
						"meetingType":      fmt.Sprintf("%d", meeting.Type),
						"meetingUuid":      meeting.UUID,
						"realMeetingId":    fmt.Sprintf("%d", meeting.ID),
					},
				},
			},
		},
	}

	event, err = srv.Events.Insert(calendarId, event).ConferenceDataVersion(1).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)


	event2 := &calendar.Event{
		Id: event.Id,
		Attendees: []*calendar.EventAttendee{
			{Email: "lpage@example.com"},
			{Email: "sbrin+89@example.com"},
		},
	}

	e2, err := srv.Events.Patch(calendarId, event2.Id, event2).ConferenceDataVersion(1).Do() // .SendUpdates()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", e2.HtmlLink)
}