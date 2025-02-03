package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	courseFlag := flag.Bool("course", false, "fetch course info")
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
	}
}
