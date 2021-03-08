package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sort"
	"time"

	"github.com/go-macaron/binding"
	"github.com/gocarina/gocsv"
	"github.com/tamalsaha/webinar-tools/lib"
	gdrive "gomodules.xyz/gdrive-utils"
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

type WebinarRegistrationForm struct {
	FirstName       string `json:"first_name" csv:"First Name" form:"first_name"`
	LastName        string `json:"last_name" csv:"Last Name" form:"last_name"`
	Phone           string `json:"phone" csv:"Phone" form:"phone"`
	JobTitle        string `json:"job_title" csv:"Job Title" form:"job_title"`
	WorkEmail       string `json:"work_email" csv:"Work Email" form:"work_email"`
	KnowsKubernetes bool   `json:"knows_kubernetes" csv:"Knows Kubernetes" form:"knows_kubernetes"`
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
	hc, err := gdrive.DefaultClient("/home/tamal/go/src/github.com/tamalsaha/webinar-tools")
	if err != nil {
		panic(err)
	}

	srv, err := sheets.NewService(context.TODO(), option.WithHTTPClient(hc))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	spreadsheetId := "1VW9K1yRLw6IFnr4o9ZJqaEamBahfqnjfl79EHeAZBzg"

	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	// m.Use(macaron.Static("public"))

	m.Get("/", func() string {
		reader, err := lib.NewRowReader(srv, spreadsheetId, "webinar_schedule", "Schedule", func(column []interface{}) (int, error) {
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
		})
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

	m.Post("/register", binding.Bind(WebinarRegistrationForm{}), func(form WebinarRegistrationForm) string {
		clients := []*WebinarRegistrationForm{
			&form,
		}
		writer := lib.NewWriter(srv, spreadsheetId, "webinar_2020_03_11")
		err = gocsv.MarshalCSV(clients, writer)
		if err != nil {
			panic(err)
		}
		return "registration successful"
	})

	m.Run()
}
