package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Issue struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// // TODO: Add functions to get/post issues from github
// func postIssue(todo *Todo.Todo) {
// 	title := todo.Title
// 	description := todo.Description

// }

func GetIssues() []Issue {
	url := "https://api.github.com/repos/Amr-Shams/IssueMe/issues"

	// Create an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Unmarshal the response body into a slice of Issues
	var issues []Issue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	PrintIssues(issues)

	return issues
}

func PrintIssues(issues []Issue) {
	for _, issue := range issues {
		fmt.Println("--------------------- Issue ID:", issue.ID, "---------------------")
		fmt.Println("Title:", issue.Title)
		fmt.Println("Body:", issue.Body)
		fmt.Println("----------------------------------------------------------------")
		fmt.Println()
	}
}
