package main

type SubjectCode struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	LinkedGroup string `json:"linkedGroup"`
}

type Course struct {
	CourseID    string      `json:"__catalogCourseId"`
	PID         string      `json:"pid"`
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	SubjectCode SubjectCode `json:"subjectCode"`
}

type CRN struct {
	CourseReferenceNumber string `json:"courseReferenceNumber"`
}

type SessionCourse struct {
	CRNS []CRN `json:"data"`
}

type DetailedCourseInfo struct {
	Fmt []struct {
		Category              string `json:"category"`
		CourseReferenceNumber string `json:"courseReferenceNumber"`
		Faculty               []struct {
			BannerID         string `json:"bannerId"`
			Category         string `json:"category"`
			DisplayName      string `json:"displayName"`
			EmailAddress     string `json:"emailAddress"`
			PrimaryIndicator bool   `json:"primaryIndicator"`
		} `json:"faculty"`
		MeetingTime struct {
			BeginTime           string `json:"beginTime"`
			EndTime             string `json:"endTime"`
			Building            string `json:"building"`
			BuildingDescription string `json:"buildingDescription"`
			Room                string `json:"room"`
			Monday              bool   `json:"monday"`
			Tuesday             bool   `json:"tuesday"`
			Wednesday           bool   `json:"wednesday"`
			Thursday            bool   `json:"thursday"`
			Friday              bool   `json:"friday"`
			MeetingScheduleType string `json:"meetingScheduleType"`
		} `json:"meetingTime"`
		Term string `json:"term"`
	} `json:"fmt"`
}
