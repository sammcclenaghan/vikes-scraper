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

func NewSession() (*Session, error) {
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
	return &Session{client: client}, nil
}

func (s *Session) fetchCourseInfo(term string, subject string, courseNumber string) (*CourseResponse, error) {
	initURL := "https://banner.uvic.ca/StudentRegistrationSsb/ssb/term/termSelection?mode=search"
	if err := makeRequest(s.client, "GET", initURL, nil); err != nil {
		return nil, fmt.Errorf("init request failed: %v", err)
	}

	termURL := "https://banner.uvic.ca/StudentRegistrationSsb/ssb/term/search?mode=search"
	termData := strings.NewReader(fmt.Sprintf("term=%s&studyPath=&studyPathText=&startDatepicker=&endDatepicker=", term))
	if err := makeRequest(s.client, "POST", termURL, termData); err != nil {
		return nil, fmt.Errorf("term setup failed: %v", err)
	}

	searchURL := fmt.Sprintf("https://banner.uvic.ca/StudentRegistrationSsb/ssb/searchResults/searchResults?txt_term=%s&txt_subject=%s&txt_courseNumber=%s&pageOffset=0&pageMaxSize=50&sortColumn=subjectDescription&sortDirection=asc",
		term, subject, courseNumber)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"User-Agent": {"Mozilla/5.0"},
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response CourseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON %v", err)
	}

	return &response, nil
}

func (s *Session) fetchSessions(term, crn string) (*DetailedResponse, error) {
	termURL := "https://banner.uvic.ca/StudentRegistrationSsb/ssb/term/search?mode=search"
	termData := strings.NewReader(fmt.Sprintf("term=%s&studyPath=&studyPathText=&startDatepicker=&endDatepicker=", term))

	if err := makeRequest(s.client, "POST", termURL, termData); err != nil {
		return nil, fmt.Errorf("term setup failed: %v", err)
	}

	detailURL := fmt.Sprintf("https://banner.uvic.ca/StudentRegistrationSsb/ssb/searchResults/getFacultyMeetingTimes?term=%s&courseReferenceNumber=%s", term, crn)

	req, err := http.NewRequest("GET", detailURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response DetailedResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return &response, nil
}

func makeRequest(client *http.Client, method, url string, body io.Reader) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"User-Agent":   {"Mozilla/5.0"},
		"Content-Type": {"application/x-www-form-urlencoded"},
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
