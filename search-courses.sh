#!/bin/bash

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

jq -r '.[] | [.subjectCode.name, (.catalogCourseId // .__catalogCourseId | capture("(?<subject>[A-Z]+)(?<number>[0-9]+)").number), .title] | join(" ")' courses.json | \
    fzf --layout=reverse --height=50% \
        --preview='subject=$(echo {} | cut -d" " -f1); number=$(echo {} | cut -d" " -f2); go run . -courses "$subject" "$number"' \
        --preview-window=right:60%:wrap \
        --bind='ctrl-/:change-preview-window(down|hidden|)' \
        --header="Search courses (Ctrl-/ to toggle preview)"
