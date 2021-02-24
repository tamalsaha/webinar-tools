package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	passgen "gomodules.xyz/password-generator"
	"github.com/himalayan-institute/zoom-lib-golang"
)

// ExampleWebinar contains examples for the /webinar endpoints
func main() {
	var (
		apiKey          = os.Getenv("ZOOM_API_KEY")
		apiSecret       = os.Getenv("ZOOM_API_SECRET")
		email           = os.Getenv("ZOOM_ACCOUNT_EMAIL")
		registrantEmail = os.Getenv("ZOOM_EXAMPLE_REGISTRANT_EMAIL")
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

	os.Exit(1)

	fifty := int(50)
	webinars, err := zoom.ListWebinars(zoom.ListWebinarsOptions{
		HostID:   user.ID,
		PageSize: &fifty,
	})
	if err != nil {
		log.Fatalf("got error listing webinars: %+v\n", err)
	}

	log.Printf("Got open webinars: %+v\n", webinars)

	webinars, err = zoom.ListWebinars(zoom.ListWebinarsOptions{
		HostID:   user.ID,
		PageSize: &fifty,
	})
	if err != nil {
		log.Fatalf("got error listing webinars: %+v\n", err)
	}

	log.Printf("Got registration webinars: %+v\n", webinars)

	webinar, err := zoom.GetWebinarInfo(webinars.Webinars[0].ID)

	if err != nil {
		log.Fatalf("got error getting single webinar: %+v\n", err)
	}

	log.Printf("Got single webinars: %+v\n", webinar)

	log.Printf("created at: %s\n", webinar.CreatedAt)
	log.Printf("first occurrence start: %s\n", webinar.Occurrences[0].StartTime)

	customQs := []zoom.CustomQuestion{
		{
			Title: "asdf foo bar",
			Value: "example custom question answer",
		},
	}

	registrantInfo := zoom.WebinarRegistrant{
		WebinarID:       webinar.ID,
		Email:           registrantEmail,
		FirstName:       "Foo",
		LastName:        "Bar",
		CustomQuestions: customQs,
	}

	registrant, err := zoom.RegisterForWebinar(registrantInfo)
	if err != nil {
		log.Fatalf("got error registering a user for webinar %d: %+v\n", webinar.ID, err)
	}

	log.Printf("Got registrant: %+v\n", registrant)

	getRegistrationOpts := zoom.ListWebinarRegistrantsOptions{
		WebinarID: webinar.ID,
	}

	registrationInfo, err := zoom.ListWebinarRegistrants(getRegistrationOpts)
	if err != nil {
		log.Fatalf("got error getting registration info for webinar %d: %+v\n", webinar.ID, err)
	}

	log.Printf("Got registration information: %+v\n", registrationInfo)

	panelists, err := zoom.GetWebinarPanelists(webinar.ID)
	if err != nil {
		log.Fatalf("got error listing webinar panelists for webinar %d: %+v\n", webinar.ID, err)
	}

	log.Printf("Got webinar panelists: %+v\n", panelists)
}