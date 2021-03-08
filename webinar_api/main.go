package main

import "time"

type WebinarSchedule struct {
	Title          string    `json:"title" csv:"Title" form:"title"`
	Schedule       time.Time `json:"schedule" csv:"Schedule" form:"schedule"`
	Summary        string    `json:"summary" csv:"Summary" form:"summary"`
	Speaker        string    `json:"speaker" csv:"Speaker" form:"speaker"`
	SpeakerBio     string    `json:"speaker_bio" csv:"Speaker Bio" form:"speaker_bio"`
	SpeakerPicture string    `json:"speaker_picture" csv:"Speaker Picture" form:"speaker_picture"`
}

type WebinarSignup struct {
	FirstName       string `json:"first_name" csv:"First Name" form:"first_name"`
	LastName        string `json:"last_name" csv:"Last Name" form:"last_name"`
	Phone           string `json:"phone" csv:"Phone" form:"phone"`
	JobTitle        string `json:"job_title" csv:"Job Title" form:"job_title"`
	WorkEmail       string `json:"work_email" csv:"Work Email" form:"work_email"`
	KnowsKubernetes bool   `json:"knows_kubernetes" csv:"Knows Kubernetes" form:"knows_kubernetes"`
}

func main() {

}
