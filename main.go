package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type CSVExportRow struct {
	Term                string
	Subject             string
	CourseName          string
	CourseNumber        string
	CRN                 string
	Section             string
	Time                string
	Days                string
	Location            string
	DateRange           string
	ScheduleType        string
	Instructor          string
	InstructionalMethod string
	Units               string
	Available           bool
}

func formatTime(time string) string {
	if time == "" {
		return ""
	}
	// Convert "1330" to "13:30"
	return fmt.Sprintf("%s:%s", time[:2], time[2:])
}

func getDays(mt MeetingTime) string {
	days := ""
	if mt.Monday {
		days += "M"
	}
	if mt.Tuesday {
		days += "T"
	}
	if mt.Wednesday {
		days += "W"
	}
	if mt.Thursday {
		days += "R"
	}
	if mt.Friday {
		days += "F"
	}
	return days
}

func getInstructors(faculty []struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}) string {
	var names []string
	for _, f := range faculty {
		if f.DisplayName != "" {
			names = append(names, f.DisplayName)
		}
	}
	return strings.Join(names, "; ")
}

func loadCoursesFromJSON(filename string) ([]Course, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var courses []Course
	if err := json.Unmarshal(bytes, &courses); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}
	return courses, nil
}

func exportToCSV(rows []CSVExportRow, filename string) error {
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	writer := csv.NewWriter(out)
	defer writer.Flush()

	header := []string{
		"Term", "Subject", "Course Name", "Course Number", "CRN", "Section",
		"Time", "Days", "Location", "Date Range", "Schedule Type",
		"Instructor", "Instructional Method", "Units", "Available",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	for _, row := range rows {
		record := []string{
			row.Term,
			row.Subject,
			row.CourseName,
			row.CourseNumber,
			row.CRN,
			row.Section,
			row.Time,
			row.Days,
			row.Location,
			row.DateRange,
			row.ScheduleType,
			row.Instructor,
			row.InstructionalMethod,
			row.Units,
			fmt.Sprintf("%v", row.Available),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %v", err)
		}
	}
	return nil
}

func main() {
	courseFlag := flag.Bool("course", false, "fetch course info")
	allCoursesFlag := flag.Bool("all", false, "fetch all courses and export to CSV")
	flag.Parse()

	if *courseFlag {
		session, err := NewSession()

		if err != nil {
			fmt.Println("Error creating session:", err)
			return
		}

		courseSubject, courseID, err := getCourseDetails(os.Stdin, flag.Args()...)
		fmt.Println(courseSubject, courseID)
		if err != nil {
			fmt.Printf("Error fetching courses sessions: %v\n", err)
			return
		}
		crn, err := session.fetchCourseInfo("202501", courseSubject, courseID)
		if err != nil {
			fmt.Printf("Error fetching course info: %v\n", err)
			return
		}

		for _, course := range crn.CRNS {
			if course.CourseReferenceNumber != "" {
				fmt.Println(course.CourseReferenceNumber)
				details, err := session.fetchSessions("202501", course.CourseReferenceNumber)
				if err != nil {
					fmt.Printf("Error fetching course details: %v\n", err)
					return
				}
				if len(details.Fmt) > 0 {
					for _, section := range details.Fmt {
						mt := section.MeetingTime
						if mt.BeginTime != "" {
							fmt.Printf("Schedule: %s-%s\n", mt.BeginTime, mt.EndTime)
							fmt.Printf("Location: %s (%s) Room %s\n", mt.Building, mt.BuildingDescription, mt.Room)
						}
					}
				}
			}
		}
		return
	}

	if *allCoursesFlag {
		fmt.Println("Loading all courses from courses.json")

		courses, err := loadCoursesFromJSON("courses.json")
		if err != nil {
			fmt.Printf("Error loading courses: %v\n", err)
			return
		}

		session, err := NewSession()
		if err != nil {
			fmt.Printf("Error creating session: %v\n", err)
			return
		}

		sem := make(chan struct{}, 10)

		results := make(chan CSVExportRow)

		var wg sync.WaitGroup

		errorCh := make(chan error)

		var csvRows []CSVExportRow
		done := make(chan bool)
		go func() {
			for row := range results {
				csvRows = append(csvRows, row)
			}
			done <- true
		}()

		// Process each course
		for _, c := range courses {
			subject := c.SubjectCode.Name
			number := strings.TrimPrefix(c.CourseID, subject)

			wg.Add(1)
			go func(c Course, subject, number string) {
				defer wg.Done()

				sem <- struct{}{}
				defer func() { <-sem }()

				fmt.Printf("Fetching details for %s %s (%s)...\n", subject, number, c.Title)

				crn, err := session.fetchCourseInfo("202501", subject, number)
				if err != nil {
					errorCh <- fmt.Errorf("error fetching course info for %s %s: %v", subject, number, err)
					// Even if there's an error, we'll record the course as unavailable
					results <- CSVExportRow{
						Term:         "202501",
						Subject:      subject,
						CourseName:   c.Title,
						CourseNumber: number,
						Available:    false,
					}
					return
				}

				// If no CRNs returned, course isn't available this term
				if len(crn.CRNS) == 0 {
					results <- CSVExportRow{
						Term:         "202501",
						Subject:      subject,
						CourseName:   c.Title,
						CourseNumber: number,
						Available:    false,
					}
					return
				}

				for _, courseCRN := range crn.CRNS {
					if courseCRN.CourseReferenceNumber == "" {
						continue
					}

					details, err := session.fetchSessions("202501", courseCRN.CourseReferenceNumber)
					if err != nil {
						errorCh <- fmt.Errorf("error fetching session for CRN %s: %v",
							courseCRN.CourseReferenceNumber, err)
						continue
					}

					for _, section := range details.Fmt {
						mt := section.MeetingTime
						location := ""
						if mt.Building != "" && mt.Room != "" {
							location = fmt.Sprintf("%s %s", mt.Building, mt.Room)
						}

						row := CSVExportRow{
							Term:         "202501",
							Subject:      subject,
							CourseName:   c.Title,
							CourseNumber: number,
							CRN:          courseCRN.CourseReferenceNumber,
							Section:      section.Section,
							Time:         fmt.Sprintf("%s-%s", formatTime(mt.BeginTime), formatTime(mt.EndTime)),
							Days:         getDays(mt),
							Location:     location,
							DateRange:    "", // Remove date range if not available in API
							ScheduleType: mt.MeetingScheduleType,
							Instructor:   getInstructors(section.Faculty),
							Available:    true,
						}
						results <- row
					}
				}
			}(c, subject, number)
		}

		errDone := make(chan bool)
		var errors []error
		go func() {
			for err := range errorCh {
				errors = append(errors, err)
			}
			errDone <- true
		}()

		wg.Wait()

		close(results)
		close(errorCh)

		<-done
		<-errDone

		if len(errors) > 0 {
			fmt.Println("\nErrors occurred during processing:")
			for _, err := range errors {
				fmt.Printf("- %v\n", err)
			}
		}

		if err := exportToCSV(csvRows, "courses.csv"); err != nil {
			fmt.Printf("Error exporting to CSV: %v\n", err)
			return
		}

		fmt.Printf("Exported %d course sections to courses.csv\n", len(csvRows))
		return
	}

	// If neither flag is set, just show usage
	fmt.Println("Usage:")
	fmt.Println("  --course [SUBJECT COURSE#] : fetch a single course (e.g., --course CSC 110).")
	fmt.Println("  --all                      : fetch all courses from courses.json and export to CSV.")
}
