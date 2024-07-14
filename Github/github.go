package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/google/go-github/v39/github"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
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
func listAllIssues() {
	cmd := exec.Command("gh", "issue", "list")
	projectDir := viper.GetString("input")
	cmd.Dir = projectDir
	out, err := cmd.Output()
	if err != nil {
		fmt.Errorf("Failed to list issues: %v", err)
	}
	fmt.Println(string(out))
}

func getRepoInfo() (owner, repo string, err error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	url := strings.TrimSpace(string(output))
	parts := strings.Split(url, "/")
	repoPart := parts[len(parts)-1]
	repo = strings.TrimSuffix(repoPart, ".git")
	owner = parts[len(parts)-2]
	return owner, repo, nil
}
func FireIssue(todo *Todo.Todo) error {
	projectDir := viper.GetString("input")
	err := godotenv.Load(projectDir + "/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// load the environment variables GITHUB_TOKEN
	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	owner, repo, err := getRepoInfo()
	if err != nil {
		log.Fatalf("Failed to get github url: %v", err)
	}

	issue := &github.IssueRequest{
		Title: &todo.Title,
		Body:  &todo.Description,
	}
	var issu2 *github.Issue
	issu2, _, err = client.Issues.Create(ctx, owner, repo, issue)
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
    id:=strconv.Itoa(issu2.GetNumber())
    todo.ID=&id
	return nil
}
func GetIssues() []Issue {
	owner, repo, err := getRepoInfo()
	if err != nil {
		log.Fatalf("Failed to get github url: %v", err)
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
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
