package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Course struct {
	CourseID string `json:"__catalogCourseId"`
	PID      string `json:"pid'`
	ID       string `json:"id"`
	Title    string `json:"title"`
}

func main() {
	coursesFile, err := os.Open("courses.json")
	if err != nil {
		fmt.Println("Error opening courses.json:", err)
		return
	}
	defer coursesFile.Close()

	fmt.Println("Successfully opened courses.json")

	var courses []Course
	decoder := json.NewDecoder(coursesFile)
	err = decoder.Decode(&courses)
	if err != nil {
		fmt.Println("Error decoding courses.json:", err)
		return
	}

	for i, course := range courses {
		if i >= 10 {
			break
		}
		fmt.Println(course.CourseID, course.PID, course.ID, course.Title)
	}
}
