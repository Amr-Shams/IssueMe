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

	"github.com/AlecAivazis/survey/v2"
	Project "github.com/Amr-Shams/IssueMe/Project"
	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/google/go-github/v39/github"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	Log "github.com/rs/zerolog/log"
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
	purgeCmd := purgeCommand()
	root.AddCommand(reportCmd)
	root.AddCommand(purgeCmd)
}
func reportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "List all the todos in the project and create issues for them",
		Run: func(cmd *cobra.Command, args []string) {
			project := Project.NewProject()
			_, todos, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project %s", err.Error())
			}

			sort.Slice(todos, func(i, j int) bool {
				return todos[i].Uergency > todos[j].Uergency
			})
			selected := CheckBoxes("Select the todos you want to create issues for", todos)
			fmt.Printf("Selected %d todos\n", len(selected))
            commitMessage := "Create issues for the following todos: "
			for _, todo := range selected {
				fmt.Printf("Creating issue for %s\n", todo.String())
				err := FireIssue(todo)
				if err != nil {
					log.Fatalf("Failed to create issue: %v", err)
				}
				fmt.Println("Issue created successfully")
				fmt.Println("Issue ID: ", *todo.ID)
                commitMessage += *todo.ID + " "
			}
            if len(selected) != 0 {
               CreateCommit(commitMessage)
            }
		},
	}
}

func purgeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "purge",
		Short: "Delete all closed issues",
		Run: func(cmd *cobra.Command, args []string) {
			project := Project.NewProject()
			todo, _, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project %s", err.Error())
			}
			todos := FilterTodos(todo)
			sort.Slice(todos, func(i, j int) bool {
				return todos[i].Uergency > todos[j].Uergency
			})
			selected := CheckBoxes("Select the todos you want to delete issues for", todos)
            commitMessage:="removing the closed todos from the project: "

			projectDir := viper.GetString("input")
			for _, issue := range selected {
				issue.Remove(projectDir)
                commitMessage += *issue.ID + " "
			}
            if len(selected) != 0 {
                CreateCommit(commitMessage)
            }
        		},
	}
}
func CreateCommit(commitMessage string) {
    cmd := exec.Command("git", "commit", "-am", commitMessage)
    projectDir := viper.GetString("input")
    cmd.Dir = projectDir
    err := cmd.Run()
    if err != nil {
        log.Fatalf("Failed to create commit: %v", err)
    }
}
func FilterTodos(todos []*Todo.Todo) []*Todo.Todo {
	client, ctx := createClient()
	owner, repo, err := getRepoInfo()
	if err != nil {
		log.Fatalf("Failed to get github url: %v", err)
	}
	var filteredTodos []*Todo.Todo
	for _, todo := range todos {
		id, err := strconv.Atoi(*todo.ID)
		if err != nil {
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			Log.Info().Msgf("Failed to convert issue id to int: %v", err)
			continue
		}
		issue, _, err := client.Issues.Get(ctx, owner, repo, id)
		if err != nil {
			log.Fatalf("Failed to get issue: %v", err)
		}
		if issue.GetState() == "closed" {
			filteredTodos = append(filteredTodos, todo)
		}
	}
	return filteredTodos
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
	projectDir := viper.GetString("input")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	url := strings.TrimSpace(string(output))
	parts := strings.Split(url, "/")
	repoPart := parts[len(parts)-1]
	repo = strings.TrimSuffix(repoPart, ".git")
	owner = parts[len(parts)-2]
	if strings.HasPrefix(owner, "git@") {
		owner = strings.TrimPrefix(owner, "git@github.com:")
	}
	return owner, repo, nil
}
func createClient() (*github.Client, context.Context) {
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
	return client, ctx
}
func FireIssue(todo *Todo.Todo) error {
	client, ctx := createClient()
	projectDir := viper.GetString("input")
	owner, repo, err := getRepoInfo()
	if err != nil {
		log.Fatalf("Failed to get github url: %v", err)
	}
	body := Todo.StringifyDescription(todo.Description)
	issue := &github.IssueRequest{
		Title: &todo.Title,
		Body:  &body,
	}
	var issu2 *github.Issue
	issu2, _, err = client.Issues.Create(ctx, owner, repo, issue)
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	id := strconv.Itoa(issu2.GetNumber())
	todo.Update(id, projectDir)
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
