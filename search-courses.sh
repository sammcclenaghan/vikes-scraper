#!/bin/bash

# Default semester is current
SEMESTER="current"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--semester)
            SEMESTER="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [-s|--semester semester]"
            exit 1
            ;;
    esac
done

if ! command -v fzf >/dev/null 2>&1; then
    echo "fzf is required but not installed. Please install it first."
    echo "Install with: brew install fzf"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "jq is required but not installed. Please install it first."
    echo "Install with: brew install jq"
    exit 1
fi

# Read from courses.json and process with jq/fzf
jq -r '.[] | [.subjectCode.name, (.catalogCourseId // .__catalogCourseId | capture("(?<subject>[A-Z]+)(?<number>[0-9]+)").number), .title] | join(" ")' courses.json | \
    fzf --layout=reverse --height=50% \
        --preview='subject=$(echo {} | cut -d" " -f1); number=$(echo {} | cut -d" " -f2); go run . -courses "$subject" "$number" -semester "'"$SEMESTER"'"' \
        --preview-window=right:60%:wrap \
        --bind='ctrl-/:change-preview-window(down|hidden|)' \
        --header="Search courses for $SEMESTER semester (Ctrl-/ to toggle preview)"
