package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Core structs remain the same
type Course struct {
	CourseID    string      `json:"__catalogCourseId"`
	Title       string      `json:"title"`
	SubjectCode SubjectCode `json:"subjectCode"`
}

type SubjectCode struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	LinkedGroup string `json:"linkedGroup"`
}

// Combined course section struct
type CourseSection struct {
	// Basic identification
	ID                    int    `json:"id"`
	Term                  string `json:"term"`
	TermDesc              string `json:"termDesc"`
	CourseReferenceNumber string `json:"courseReferenceNumber"`
	PartOfTerm            string `json:"partOfTerm"`

	// Course information
	CourseNumber       string `json:"courseNumber"`
	Subject            string `json:"subject"`
	SubjectDescription string `json:"subjectDescription"`
	SubjectCourse      string `json:"subjectCourse"`
	Section            string `json:"sequenceNumber"`
	CourseTitle        string `json:"courseTitle"`

	// Location and schedule
	CampusDescription       string `json:"campusDescription"`
	ScheduleTypeDescription string `json:"scheduleTypeDescription"`

	// Enrollment information
	CreditHours       float64 `json:"creditHours"`
	MaximumEnrollment int     `json:"maximumEnrollment"`
	Enrollment        int     `json:"enrollment"`
	SeatsAvailable    int     `json:"seatsAvailable"`

	// Wait list information
	WaitCapacity  int `json:"waitCapacity"`
	WaitCount     int `json:"waitCount"`
	WaitAvailable int `json:"waitAvailable"`

	// Cross listing
	CrossList          interface{} `json:"crossList"`
	CrossListCapacity  interface{} `json:"crossListCapacity"`
	CrossListCount     interface{} `json:"crossListCount"`
	CrossListAvailable interface{} `json:"crossListAvailable"`

	// Credit hours
	CreditHourHigh      float64 `json:"creditHourHigh"`
	CreditHourLow       float64 `json:"creditHourLow"`
	CreditHourIndicator string  `json:"creditHourIndicator"`

	// Section status
	OpenSection     bool   `json:"openSection"`
	LinkIdentifier  string `json:"linkIdentifier"`
	IsSectionLinked bool   `json:"isSectionLinked"`

	// Instructional method
	InstructionalMethod            string `json:"instructionalMethod"`
	InstructionalMethodDescription string `json:"instructionalMethodDescription"`

	// Faculty and meetings
	Faculty         []Faculty        `json:"faculty"`
	MeetingsFaculty []MeetingFaculty `json:"meetingsFaculty"`

	// Reserved seats and attributes
	ReservedSeatSummary interface{} `json:"reservedSeatSummary"`
	SectionAttributes   interface{} `json:"sectionAttributes"`
}

type MeetingTime struct {
	BeginTime              string  `json:"beginTime"`
	EndTime                string  `json:"endTime"`
	Building               string  `json:"building"`
	BuildingDescription    string  `json:"buildingDescription"`
	Campus                 string  `json:"campus"`
	CampusDescription      string  `json:"campusDescription"`
	Category               string  `json:"category"`
	Class                  string  `json:"class"`
	Room                   string  `json:"room"`
	Monday                 bool    `json:"monday"`
	Tuesday                bool    `json:"tuesday"`
	Wednesday              bool    `json:"wednesday"`
	Thursday               bool    `json:"thursday"`
	Friday                 bool    `json:"friday"`
	Saturday               bool    `json:"saturday"`
	Sunday                 bool    `json:"sunday"`
	StartDate              string  `json:"startDate"`
	EndDate                string  `json:"endDate"`
	CreditHourSession      float64 `json:"creditHourSession"`
	HoursWeek              float64 `json:"hoursWeek"`
	MeetingScheduleType    string  `json:"meetingScheduleType"`
	MeetingType            string  `json:"meetingType"`
	MeetingTypeDescription string  `json:"meetingTypeDescription"`
	Term                   string  `json:"term"`
	CourseReferenceNumber  string  `json:"courseReferenceNumber"`
}

type Faculty struct {
	BannerId              string `json:"bannerId"`
	Category              string `json:"category"`
	Class                 string `json:"class"`
	CourseReferenceNumber string `json:"courseReferenceNumber"`
	DisplayName           string `json:"displayName"`
	EmailAddress          string `json:"emailAddress"`
	PrimaryIndicator      bool   `json:"primaryIndicator"`
	Term                  string `json:"term"`
}

type MeetingFaculty struct {
	Category                       string      `json:"category"`
	Class                          string      `json:"class"`
	CourseReferenceNumber          string      `json:"courseReferenceNumber"`
	Faculty                        []Faculty   `json:"faculty"`
	MeetingTime                    MeetingTime `json:"meetingTime"`
	Term                           string      `json:"term"`
	Section                        string      `json:"sequenceNumber"`
	Enrollment                     int         `json:"enrollment"`
	MaximumEnrollment              int         `json:"maximumEnrollment"`
	WaitCount                      int         `json:"waitCount"`
	WaitCapacity                   int         `json:"waitCapacity"`
	CreditHourHigh                 float64     `json:"creditHourHigh"`
	InstructionalMethodDescription string      `json:"instructionalMethodDescription"`
}

// Response wrappers
type CourseResponse struct {
	Success    bool            `json:"success"`
	TotalCount int             `json:"totalCount"`
	Data       []CourseSection `json:"data"`
}

type DetailedResponse struct {
	Fmt []MeetingFaculty `json:"fmt"`
}

type KualiCredits struct {
	Credits struct {
		Min string `json:"min"`
		Max string `json:"max"`
	} `json:"credits"`
	Value  string `json:"value"`
	Chosen string `json:"chosen"`
}

type KualiCrossListedCourse struct {
	CatalogCourseID string `json:"__catalogCourseId"`
	PID            string `json:"pid"`
	Title          string `json:"title"`
}

type KualiSubjectCode struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	LinkedGroup string `json:"linkedGroup"`
}

type KualiCourseInfo struct {
	PassedCatalogQuery bool                   `json:"__passedCatalogQuery"`
	Description        string                 `json:"description"`
	PID               string                 `json:"pid"`
	Title             string                 `json:"title"`
	SupplementalNotes string                 `json:"supplementalNotes"`
	CatalogCourseID   string                 `json:"__catalogCourseId"`
	Credits           KualiCredits           `json:"credits"`
	CrossListedCourses []KualiCrossListedCourse `json:"crossListedCourses"`
	DateStart         string                 `json:"dateStart"`
	SubjectCode       KualiSubjectCode       `json:"subjectCode"`
	HoursCatalogText  string                 `json:"hoursCatalogText"`
}

func fetchKualiCourseInfo(pid string) (*KualiCourseInfo, error) {
	url := fmt.Sprintf("https://uvic.kuali.co/api/v1/catalog/course/65eb47906641d7001c157bc4/%s", pid)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Kuali data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var info KualiCourseInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse Kuali data: %v", err)
	}

	return &info, nil
}

// Helper function to clean HTML tags from strings
func cleanHTML(input string) string {
	// Simple HTML tag removal - you might want to use a proper HTML parser for more complex cases
	noTags := strings.ReplaceAll(input, "</p>", "\n")
	noTags = strings.ReplaceAll(noTags, "</li>", "\n")
	noTags = strings.ReplaceAll(noTags, "<ul>", "")
	noTags = strings.ReplaceAll(noTags, "</ul>", "")
	noTags = strings.ReplaceAll(noTags, "<li>", "• ")
	
	// Remove any remaining HTML tags
	for strings.Contains(noTags, "<") {
		start := strings.Index(noTags, "<")
		end := strings.Index(noTags, ">")
		if end > start {
			noTags = noTags[:start] + noTags[end+1:]
		} else {
			break
		}
	}
	
	return strings.TrimSpace(noTags)
}
