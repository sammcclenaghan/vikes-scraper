package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

type Session struct {
	client *http.Client
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

func NewSession() (*Session, error) {
	client := &http.Client{}
	return &Session{client: client}, nil
}

func makeRequest(client *http.Client, method, url string, body io.Reader) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": {"en-US,en;q=0.9"},
		"Connection":      {"keep-alive"},
		"User-Agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"},
		"Cache-Control":   {"max-age=0"},
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	return nil
}

func (s *Session) fetchCourseInfo(term string, subject string, courseNumber string) (*SessionCourse, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	initURL := "https://banner.uvic.ca/StudentRegistrationSsb/ssb/term/termSelection?mode=search"
	if err := makeRequest(client, "GET", initURL, nil); err != nil {
		return nil, fmt.Errorf("init request failed: %v", err)
	}

	// Step 2: Set the term
	termURL := "https://banner.uvic.ca/StudentRegistrationSsb/ssb/term/search?mode=search"
	termData := strings.NewReader(fmt.Sprintf("term=%s&studyPath=&studyPathText=&startDatepicker=&endDatepicker=", term))
	if err := makeRequest(client, "POST", termURL, termData); err != nil {
		return nil, fmt.Errorf("term setup failed: %v", err)
	}

	// Step 3: Search for courses
	searchURL := fmt.Sprintf("https://banner.uvic.ca/StudentRegistrationSsb/ssb/searchResults/searchResults?txt_term=%s&txt_subject=%s&txt_courseNumber=%s&pageOffset=0&pageMaxSize=50&sortColumn=subjectDescription&sortDirection=asc",
		term, subject, courseNumber)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Accept":           {"application/json, text/javascript, */*; q=0.01"},
		"Accept-Language":  {"en-US,en;q=0.9"},
		"Connection":       {"keep-alive"},
		"X-Requested-With": {"XMLHttpRequest"},
		"User-Agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"},
		"Referer":          {"https://banner.uvic.ca/StudentRegistrationSsb/ssb/classSearch/classSearch"},
		"Sec-Fetch-Dest":   {"empty"},
		"Sec-Fetch-Mode":   {"cors"},
		"Sec-Fetch-Site":   {"same-origin"},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var cr SessionCourse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("failed to decode JSON %v", err)
	}

	return &cr, nil
}

func (s *Session) fetchSessions(term, crn string) (*DetailedCourseInfo, error) {
	termURL := "https://banner.uvic.ca/StudentRegistrationSsb/ssb/term/search?mode=search"
	termData := strings.NewReader(fmt.Sprintf("term=%s&studyPath=&studyPathText=&startDatepicker=&endDatepicker=", term))

	req, err := http.NewRequest("POST", termURL, termData)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	detailURL := fmt.Sprintf("https://banner.uvic.ca/StudentRegistrationSsb/ssb/searchResults/getFacultyMeetingTimes?term=%s&courseReferenceNumber=%s", term, crn)

	req, err = http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err = s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var info DetailedCourseInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to decode JSON %v", err)
	}

	return &info, nil
}
