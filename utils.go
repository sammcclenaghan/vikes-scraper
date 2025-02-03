package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

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
