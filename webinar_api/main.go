package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/himalayan-institute/zoom-lib-golang"
	"google.golang.org/api/calendar/v3"

	"github.com/fatih/structs"
	"github.com/go-macaron/binding"
	"github.com/gocarina/gocsv"
	"github.com/tamalsaha/webinar-tools/lib"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/sets"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/macaron.v1"
)

type WebinarSchedule struct {
	Title          string   `json:"title" csv:"Title" form:"title"`
	Schedule       DateTime `json:"schedule" csv:"Schedule" form:"schedule"`
	Summary        string   `json:"summary" csv:"Summary" form:"summary"`
	Speaker        string   `json:"speaker" csv:"Speaker" form:"speaker"`
	SpeakerBio     string   `json:"speaker_bio" csv:"Speaker Bio" form:"speaker_bio"`
	SpeakerPicture string   `json:"speaker_picture" csv:"Speaker Picture" form:"speaker_picture"`
}

type WebinarMeetingID struct {
	GoogleCalendarEventID string `json:"google_calendar_event_id" csv:"Google Calendar Event ID"`
	ZoomMeetingID         int    `json:"zoom_meeting_id" csv:"Zoom Meeting ID"`
	ZoomMeetingPassword   string `json:"zoom_meeting_password" csv:"Zoom Meeting Password"`
}

type WebinarInfo struct {
	WebinarSchedule
	WebinarMeetingID
}

type WebinarRegistrationForm struct {
	FirstName       string `json:"first_name" csv:"First Name" form:"first_name"`
	LastName        string `json:"last_name" csv:"Last Name" form:"last_name"`
	Phone           string `json:"phone" csv:"Phone" form:"phone"`
	JobTitle        string `json:"job_title" csv:"Job Title" form:"job_title"`
	WorkEmail       string `json:"work_email" csv:"Work Email" form:"work_email"`
	KnowsKubernetes bool   `json:"knows_kubernetes" csv:"Knows Kubernetes" form:"knows_kubernetes"`
}

type WebinarRegistrationEmail struct {
	WorkEmail string `json:"work_email" csv:"Work Email" form:"work_email"`
}

type DateTime struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (string, error) {
	return date.Time.Format("1/2/2006 15:04:05"), nil
}

// You could also use the standard Stringer interface
func (date *DateTime) String() string {
	return date.String() // Redundant, just for example
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("1/2/2006 15:04:05", csv)
	return err
}

func main() {
	var (
		apiKey           = os.Getenv("ZOOM_API_KEY")
		apiSecret        = os.Getenv("ZOOM_API_SECRET")
		zoomAccountEmail = os.Getenv("ZOOM_ACCOUNT_EMAIL")
	)
	zoom.Debug = true
	zc := zoom.NewClient(apiKey, apiSecret)

	hc, err := gdrive.DefaultClient("/home/tamal/go/src/github.com/tamalsaha/webinar-tools")
	if err != nil {
		panic(err)
	}

	srvCalendar, err := calendar.NewService(context.TODO(), option.WithHTTPClient(hc))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar gc: %v", err)
	}

	srvSheets, err := sheets.NewService(context.TODO(), option.WithHTTPClient(hc))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	spreadsheetId := "1VW9K1yRLw6IFnr4o9ZJqaEamBahfqnjfl79EHeAZBzg"

	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	// m.Use(macaron.Static("public"))

	m.Get("/", func() string {
		header := structs.New(WebinarSchedule{}).Field("Schedule").Tag("csv")
		reader, err := lib.NewRowReader(srvSheets, spreadsheetId, "Schedule", &lib.Filter{
			Header: header,
			By: func(column []interface{}) (int, error) {
				type TP struct {
					Schedule time.Time
					Pos      int
				}
				var upcoming []TP
				now := time.Now()
				for i, v := range column {
					// 3/11/2021 3:00:00
					t, err := time.Parse("1/2/2006 15:04:05", v.(string))
					if err != nil {
						panic(err)
					}
					if t.After(now) {
						upcoming = append(upcoming, TP{
							Schedule: t,
							Pos:      i,
						})
					}
				}
				if len(upcoming) == 0 {
					return -1, io.EOF
				}
				sort.Slice(upcoming, func(i, j int) bool {
					return upcoming[i].Schedule.Before(upcoming[j].Schedule)
				})
				return upcoming[0].Pos, nil
			}})
		if err == io.EOF {
			return "not found"
		} else if err != nil {
			panic(err)
		}

		clients := []*WebinarSchedule{}
		if err := gocsv.UnmarshalCSV(reader, &clients); err != nil { // Load clients from file
			panic(err)
		}

		var result *WebinarSchedule
		if len(clients) > 0 {
			result = clients[0]
		}
		data, err := json.MarshalIndent(result, "", " ")
		if err != nil {
			panic(err)
		}
		return string(data)
	})

	m.Post("/register", binding.Bind(WebinarRegistrationForm{}), func(ctx *macaron.Context, form WebinarRegistrationForm) string {
		sheetName := "webinar_2020_03_11"

		clients := []*WebinarRegistrationForm{
			&form,
		}
		writer := lib.NewWriter(srvSheets, spreadsheetId, sheetName)
		err = gocsv.MarshalCSV(clients, writer)
		if err != nil {
			panic(err)
		}
		// ctx.Redirect("/", http.StatusSeeOther)

		// create zoom, google calendar event if not exists,
		// add attendant if google calendar meeting exists

		date := "2021-3-15"
		tdate, err := time.Parse("2006-1-2", date)
		if err != nil {
			panic(err)
		}
		yw, mw, dw := tdate.Date()

		reader, err := lib.NewRowReader(srvSheets, spreadsheetId, "Schedule", &lib.Filter{
			Header: "Schedule",
			By: func(values []interface{}) (int, error) {
				for i, v := range values {
					t2, err := time.Parse("1/2/2006 15:04:05", v.(string))
					if err != nil {
						panic(err)
					}
					y2, m2, d2 := t2.Date()

					if yw == y2 && mw == m2 && dw == d2 {
						return i, nil
					}
				}
				return -1, io.EOF
			},
		})
		if err != nil {
			panic(err)
		}

		meetings := []*WebinarInfo{}
		if err := gocsv.UnmarshalCSV(reader, &meetings); err != nil { // Load clients from file
			panic(err)
		}

		var result *WebinarInfo
		if len(meetings) > 0 {
			result = meetings[0]
		}
		if result != nil && result.GoogleCalendarEventID != "" {
			wats, err := lib.NewColumnReader(srvSheets, spreadsheetId, sheetName, "Work Email")
			if err != nil {
				panic(err)
			}
			atts := []*WebinarRegistrationEmail{}
			if err := gocsv.UnmarshalCSV(wats, &atts); err != nil { // Load clients from file
				panic(err)
			}

			emails := make([]string, len(atts))
			for i, a := range atts {
				emails[i] = a.WorkEmail
			}
			err = AddEventAttendants(srvCalendar, result.GoogleCalendarEventID, emails)
			if err != nil {
				panic(err)
			}
			return "meeting id" + result.GoogleCalendarEventID
		}

		ww := lib.NewRowWriter(srvSheets, spreadsheetId, "Schedule", &lib.Filter{
			Header: "Schedule",
			By: func(values []interface{}) (int, error) {
				for i, v := range values {
					t2, err := time.Parse("1/2/2006 15:04:05", v.(string))
					if err != nil {
						panic(err)
					}
					y2, m2, d2 := t2.Date()

					if yw == y2 && mw == m2 && dw == d2 {
						return i, nil
					}
				}
				return -1, io.EOF
			},
		})

		meetinginfo, err := CreateZoomMeeting(srvCalendar, zc, zoomAccountEmail, &result.WebinarSchedule, 60*time.Minute, []string{
			form.WorkEmail,
		})
		if err != nil {
			panic(err)
		}

		meetings2 := []*WebinarMeetingID{
			meetinginfo,
		}
		err = gocsv.MarshalCSV(meetings2, ww)
		if err != nil {
			panic(err)
		}

		return "meeting " + meetings[0].GoogleCalendarEventID
	})

	m.Get("/emails", func() string {
		header := structs.New(WebinarRegistrationEmail{}).Field("WorkEmail").Tag("csv")
		reader, err := lib.NewColumnReader(srvSheets, spreadsheetId, "webinar_2020_03_11", header)
		if err == io.EOF {
			return "not found"
		} else if err != nil {
			panic(err)
		}

		clients := []*WebinarRegistrationEmail{}
		if err := gocsv.UnmarshalCSV(reader, &clients); err != nil { // Load clients from file
			panic(err)
		}

		result := sets.NewString()
		for _, v := range clients {
			result.Insert(v.WorkEmail)
		}
		data, err := json.MarshalIndent(result.List(), "", " ")
		if err != nil {
			panic(err)
		}
		return string(data)
	})

	m.Run()
}
