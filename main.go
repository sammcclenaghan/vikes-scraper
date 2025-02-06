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
	"time"
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
	Campus              string
	CampusDescription   string
	BuildingCode        string
	BuildingName        string
	RoomNumber          string
	MeetingType         string
	MeetingDescription  string
	InstructorEmail     string
	CreditHours         string
	StartDate           string
	EndDate             string
}

type CourseOutput struct {
	CRN             string `json:"crn"`
	Subject         string `json:"subject"`
	CourseNumber    string `json:"course_number"`
	Section         string `json:"section"`
	Title           string `json:"title"`
	Professor       string `json:"professor"`
	Email           string `json:"email"`
	Schedule        string `json:"schedule"`
	Location        string `json:"location"`
	Days            string `json:"days"`
	Enrollment      string `json:"enrollment"`
	CreditHours     string `json:"credit_hours"`
	InstructionType string `json:"instruction_type"`
	DateRange       string `json:"date_range"`
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

func getInstructors(faculty []Faculty) string {
	var names []string
	for _, f := range faculty {
		if f.EmailAddress != "" {
			names = append(names, fmt.Sprintf("%s (%s)", f.DisplayName, f.EmailAddress))
		} else {
			names = append(names, f.DisplayName)
		}
	}
	return strings.Join(names, ", ")
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
		"Campus", "Campus Description", "Building Code", "Building Name", "Room Number",
		"Meeting Type", "Meeting Description", "Instructor Email", "Credit Hours",
		"Start Date", "End Date",
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
			row.Campus,
			row.CampusDescription,
			row.BuildingCode,
			row.BuildingName,
			row.RoomNumber,
			row.MeetingType,
			row.MeetingDescription,
			row.InstructorEmail,
			row.CreditHours,
			row.StartDate,
			row.EndDate,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %v", err)
		}
	}

	return nil
}

func main() {
	courseFlag := flag.Bool("course", false, "fetch course info")
	coursesFlag := flag.Bool("courses", false, "fetch multiple courses info")
	allCoursesFlag := flag.Bool("all", false, "fetch all courses and export to CSV")
	dryRunFlag := flag.Bool("dry-run", false, "dry run")
	flag.Parse()

	if *courseFlag || *coursesFlag {
		session, err := NewSession()

		if err != nil {
			fmt.Println("Error creating session:", err)
			return
		}

		var courseQueries [][]string
		if *coursesFlag {
			// Process multiple courses from args
			args := flag.Args()
			if len(args) < 2 || len(args)%2 != 0 {
				fmt.Println("Usage: -courses SUBJECT1 NUMBER1 [SUBJECT2 NUMBER2 ...]")
				return
			}
			for i := 0; i < len(args); i += 2 {
				courseQueries = append(courseQueries, []string{args[i], args[i+1]})
			}
		} else {
			// Single course mode
			courseSubject, courseID, err := getCourseDetails(os.Stdin, flag.Args()...)
			if err != nil {
				fmt.Printf("Error getting course details: %v\n", err)
				return
			}
			courseQueries = append(courseQueries, []string{courseSubject, courseID})
		}

		for _, query := range courseQueries {
			courseSubject, courseID := query[0], query[1]
			response, err := session.fetchCourseInfo("202501", courseSubject, courseID)
			if err != nil {
				fmt.Printf("Error fetching course info for %s %s: %v\n", courseSubject, courseID, err)
				continue
			}

			// Track if we've printed the professor's email
			emailPrinted := make(map[string]bool)

			for _, section := range response.Data {
				if section.CourseReferenceNumber == "" {
					continue
				}

				details, err := session.fetchSessions("202501", section.CourseReferenceNumber)
				if err != nil {
					fmt.Printf("Error fetching course details: %v\n", err)
					continue
				}

				for _, meetingFaculty := range details.Fmt {
					// Print professor's email only once at the top
					if len(meetingFaculty.Faculty) > 0 && !emailPrinted[meetingFaculty.Faculty[0].EmailAddress] && meetingFaculty.Faculty[0].EmailAddress != "" {
						fmt.Printf("Email: %s\n", meetingFaculty.Faculty[0].EmailAddress)
						fmt.Println("------------------------")
						emailPrinted[meetingFaculty.Faculty[0].EmailAddress] = true
					}

					mt := meetingFaculty.MeetingTime

					fmt.Printf("Detailed Course Information for CRN %s:\n", section.CourseReferenceNumber)
					fmt.Printf("Course: %s %s-%s\n", courseSubject, courseID, section.Section)
					fmt.Printf("Title: %s\n", section.CourseTitle)

					if mt.BeginTime != "" {
						fmt.Printf("Schedule: %s-%s\n", formatTime(mt.BeginTime), formatTime(mt.EndTime))
						fmt.Printf("Location: %s (%s) Room %s\n", mt.Building, mt.BuildingDescription, mt.Room)
						fmt.Printf("Type: %s\n", mt.MeetingType)
						fmt.Printf("Days: %s\n", getDays(mt))
					}

					if len(meetingFaculty.Faculty) > 0 && meetingFaculty.Faculty[0].DisplayName != "" {
						fmt.Printf("Professor: %s\n", meetingFaculty.Faculty[0].DisplayName)
						if meetingFaculty.Faculty[0].EmailAddress != "" {
							fmt.Printf("Email: %s\n", meetingFaculty.Faculty[0].EmailAddress)
						}
					}

					// Add enrollment information from the section
					fmt.Printf("Enrollment: %d/%d", section.Enrollment, section.MaximumEnrollment)
					if section.WaitCount > 0 {
						fmt.Printf(" (Waitlist: %d/%d)", section.WaitCount, section.WaitCapacity)
					}
					fmt.Println()

					// Add credit hours
					if section.CreditHourHigh > 0 {
						fmt.Printf("Credit Hours: %.1f\n", section.CreditHourHigh)
					}

					// Add instructional method
					if section.InstructionalMethodDescription != "" {
						fmt.Printf("Instruction Type: %s\n", section.InstructionalMethodDescription)
					}

					// Add date range
					if mt.StartDate != "" && mt.EndDate != "" {
						fmt.Printf("Date Range: %s to %s\n", mt.StartDate, mt.EndDate)
					}

					fmt.Println("------------------------")
				}
			}
		}
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
		for i, c := range courses {
			if *dryRunFlag {
				if i >= 10 {
					break
				}
			}
			subject := c.SubjectCode.Name
			number := strings.TrimPrefix(c.CourseID, subject)

			wg.Add(1)
			go func(c Course, subject, number string) {
				defer wg.Done()

				sem <- struct{}{}
				defer func() { <-sem }()

				fmt.Printf("Fetching details for %s %s (%s)...\n", subject, number, c.Title)

				maxRetries := 3
				var response *CourseResponse
				var err error

				for i := 0; i < maxRetries; i++ {
					response, err = session.fetchCourseInfo("202501", subject, number)
					if err == nil {
						break
					}
					fmt.Printf("Retrying %s %s (%s)...\n", subject, number, c.Title)
					time.Sleep(1 * time.Second)
				}
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

				if len(response.Data) == 0 {
					results <- CSVExportRow{
						Term:         "202501",
						Subject:      subject,
						CourseName:   c.Title,
						CourseNumber: number,
						Available:    false,
					}
					return
				}

				for _, section := range response.Data {
					if section.CourseReferenceNumber == "" {
						continue
					}

					details, err := session.fetchSessions("202501", section.CourseReferenceNumber)
					if err != nil {
						errorCh <- fmt.Errorf("error fetching session for CRN %s: %v",
							section.CourseReferenceNumber, err)
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
							CRN:          section.CourseReferenceNumber,
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
	fmt.Println("  --courses [SUBJECT1 NUMBER1 SUBJECT2 NUMBER2 ...] : fetch multiple courses info.")
	fmt.Println("  --all                      : fetch all courses from courses.json and export to CSV.")
}
