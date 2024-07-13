package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
    "os/exec"
    "github.com/spf13/viper"
    "strings"
    
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
func listAllIssues(){
    cmd:= exec.Command("gh", "issue", "list")
    projectDir := viper.GetString("input")
    cmd.Dir = projectDir
    out, err := cmd.Output()
    if err != nil {
        fmt.Errorf("Failed to list issues: %v", err)  
    }
    fmt.Println(string(out))
}


func getGithubURL() string{
    cmd := exec.Command("git", "remote", "-v")
    projectDir := viper.GetString("input")
    cmd.Dir = projectDir
    out, err := cmd.Output()
    if err != nil {
        fmt.Errorf("Failed to get remote url: %v", err)  
    }
    _,url := strings.Fields(string(out))[0], strings.Fields(string(out))[1]
    url = strings.TrimSuffix(url, ".git")
    url = strings.Replace(url, "github.com", "api.github.com/repos", 1)
    return url 
}

func GetIssues() []Issue {
	url := getGithubURL() + "/issues"

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
