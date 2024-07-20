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

	"sort"
    "slice"

	"github.com/AlecAivazis/survey/v2"
	Project "github.com/Amr-Shams/IssueMe/Project"
	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/google/go-github/v39/github"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type Issue struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
    State string `json:"state"`
}

func ExportCommand(root *cobra.Command) {
	reportCmd := reportCommand()
	root.AddCommand(reportCmd)
}
func reportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "List all the todos in the project and create issues for them",
		Run: func(cmd *cobra.Command, args []string) {
			project := Project.NewProject()
			todos, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project %s", err.Error())
			}
			sort.Slice(todos, func(i, j int) bool {
				return todos[i].Uergency > todos[j].Uergency
			})
			selected := CheckBoxes("Select the todos you want to create issues for", todos)
			fmt.Printf("Selected %d todos\n", len(selected))
			for _, todo := range selected {
				fmt.Printf("Creating issue for %s\n", todo.String())
				err := FireIssue(todo)
				if err != nil {
					log.Fatalf("Failed to create issue: %v", err)
				}
				fmt.Println("Issue created successfully")
				fmt.Println("Issue ID: ", *todo.ID)
			}
		},
	}
}
func convertIssuesToOptions(issues []Issue) []string {
    var options []string
    for _, issue := range issues {
        options = append(options, issue.Title)
    }
    return options
}
func DeleteIssues() []Issue {
    issues := GetIssues() 
    issues = slice.Delete(issues, func(i int) bool {
        return issues[i].State == "open"
    })
    options := convertIssuesToOptions(issues) 
    prompt := &survey.MultiSelect{ 
        Message: "Select the issues you want to delete",
        Options: options,
    }
    var selectedIndices []int 
    survey.AskOne(prompt, &selectedIndices) 
    var selectedIssues []Issue 
    for _, index := range selectedIndices {
        selectedIssues = append(selectedIssues, issues[index])
    }
    return selectedIssues
    
}
func PurgeIssues(todo *Todo.Todo, client *github.Client, owner, repo string) {
   issueURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s", owner, repo, *todo.ID)
    _, err := client.Issues.Delete(ctx, owner, repo, *todo.ID)
    if err != nil {
        log.Fatalf("Failed to delete issue: %v", err)
    }
    todo.Remove()
}
func convertTodosToOptions(todos []*Todo.Todo) []string {
	var options []string
	for _, todo := range todos {
		options = append(options, todo.LogString())
	}
	return options
}
func CheckBoxes(label string, todos []*Todo.Todo) []*Todo.Todo {
    var selectedIndices []int 
    options := convertTodosToOptions(todos)
    prompt := &survey.MultiSelect{
        Message: label,
        Options: options,
    }
    survey.AskOne(prompt, &selectedIndices)
	fmt.Println(selectedIndices)
    var selectedTodos []*Todo.Todo
    for _, index := range selectedIndices {
        selectedTodos = append(selectedTodos, todos[index])
    }
    return selectedTodos
}
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
func createClient() *github.Client {
    projectDir := viper.GetString("input")
    err := godotenv.Load(projectDir + "/.env") 
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
    token := os.Getenv("GITHUB_TOKEN")
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)
    return Client
}
func FireIssue(todo *Todo.Todo) error {
    client := createClient()
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
	id := strconv.Itoa(issu2.GetNumber())
	todo.ID = &id
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
