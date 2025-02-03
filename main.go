package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Course struct {
	CourseID string `json:"__catalogCourseId"`
	PID      string `json:"pid"`
	ID       string `json:"id"`
	Title    string `json:"title"`
}

func getCourseDetails(r io.Reader, args ...string) (string, string, error) {
	if len(args) >= 2 {
		return args[0], args[1], nil
	}

	if r == nil {
		r = os.Stdin
	}

	s := bufio.NewScanner(r)
	if s.Scan() {
		line := s.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
		return "", "", fmt.Errorf("input must contain at least two fields")
	}
	if err := s.Err(); err != nil {
		return "", "", fmt.Errorf("error reading input: %v", err)
	}
	return "", "", fmt.Errorf("no input provided")
}

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
			}
		}
	}
}
