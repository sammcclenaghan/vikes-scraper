package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type CSVExportRow struct {
	Subject      string
	CourseNumber string
	Title        string
	CRN          string
	BeginTime    string
	EndTime      string
	Building     string
	Room         string
	Days         string
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

	header := []string{"Subject", "Course Number", "CRN", "BeginTime", "EndTime", "Building", "Room"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %v", err)
	}

	for _, row := range rows {
		record := []string{
			row.Subject,
			row.CourseNumber,
			row.Title,
			row.CRN,
			row.Days,
			row.BeginTime,
			row.EndTime,
			row.Building,
			row.Room,
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

		var csvRows []CSVExportRow
		for _, c := range courses {
			subject := c.SubjectCode.Name
			number := strings.TrimPrefix(c.CourseID, subject)
			fmt.Printf("Fetching details for %s %s (%s)...\n", subject, number, c.Title)

			crn, err := session.fetchCourseInfo("202501", subject, number)
			if err != nil {
				fmt.Printf("Error fetching course info for %s %s: %v\n", subject, number, err)
				continue
			}

			// For each CRN, fetch sessions
			for _, courseCRN := range crn.CRNS {
				if courseCRN.CourseReferenceNumber == "" {
					continue
				}
				details, err := session.fetchSessions("202501", courseCRN.CourseReferenceNumber)
				if err != nil {
					fmt.Printf("Error fetching session for CRN %s: %v\n", courseCRN.CourseReferenceNumber, err)
					continue
				}

				// Add a row for each meeting time
				for _, section := range details.Fmt {
					mt := section.MeetingTime
					row := CSVExportRow{
						Subject:      subject,
						CourseNumber: number,
						CRN:          courseCRN.CourseReferenceNumber,
						BeginTime:    mt.BeginTime,
						EndTime:      mt.EndTime,
						Building:     mt.Building,
						Room:         mt.Room,
					}
					csvRows = append(csvRows, row)
				}
			}
		}

		// Now write all rows to courses.csv
		if err := exportToCSV(csvRows, "courses.csv"); err != nil {
			fmt.Printf("Error exporting to CSV: %v\n", err)
			return
		}

		fmt.Println("Exported courses to courses.csv.")
		return
	}

	// If neither flag is set, just show usage
	fmt.Println("Usage:")
	fmt.Println("  --course [SUBJECT COURSE#] : fetch a single course (e.g., --course CSC 110).")
	fmt.Println("  --all                      : fetch all courses from courses.json and export to CSV.")
}
