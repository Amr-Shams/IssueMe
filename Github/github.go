package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	Project "github.com/Amr-Shams/IssueMe/Project"
	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/google/go-github/v39/github"
	"github.com/joho/godotenv"
	Log "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

/*
- this projecet would be much easier if i used something like zod in ts
- transforming between the objects (issue in the remote repo and the local app version would be much simpiler)
- still this douable but much harder
*/
type Issue struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	State string `json:"state"`
}

type GitHubService struct {
	client *github.Client
	ctx    context.Context
	owner  string
	repo   string
}

func NewGitHubService() (*GitHubService, error) {
	projectDir := viper.GetString("input")

	err := godotenv.Load(projectDir + "/.env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN not found in environment")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner, repo, err := getRepoInfoFromGit()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %v", err)
	}

	return &GitHubService{
		client: client,
		ctx:    ctx,
		owner:  owner,
		repo:   repo,
	}, nil
}

func getRepoInfoFromGit() (owner, repo string, err error) {
	projectDir := viper.GetString("input")

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote URL: %v", err)
	}

	url := strings.TrimSpace(string(output))
	return parseGitURL(url)
}

// parseGitURL parses various Git URL formats
func parseGitURL(url string) (owner, repo string, err error) {
	// ssh format
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH Git URL format")
		}
		path := parts[1]
		pathParts := strings.Split(path, "/")
		if len(pathParts) != 2 {
			return "", "", fmt.Errorf("invalid repository path")
		}
		owner = pathParts[0]
		repo = strings.TrimSuffix(pathParts[1], ".git")
		return owner, repo, nil
	}

	// http/https format
	if strings.HasPrefix(url, "https://") {
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid HTTPS Git URL format")
		}
		repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
		owner = parts[len(parts)-2]
		return owner, repo, nil
	}

	return "", "", fmt.Errorf("unsupported Git URL format")
}

func (gs *GitHubService) CreateIssue(todo *Todo.Todo) error {
	body := Todo.StringifyDescription(todo.Description)
	issue := &github.IssueRequest{
		Title: &todo.Title,
		Body:  &body,
	}

	createdIssue, _, err := gs.client.Issues.Create(gs.ctx, gs.owner, gs.repo, issue)
	if err != nil {
		return fmt.Errorf("failed to create issue: %v", err)
	}

	id := strconv.Itoa(createdIssue.GetNumber())
	projectDir := viper.GetString("input")
	todo.Update(id, projectDir)
	return nil
}

func (gs *GitHubService) CloseIssue(todo *Todo.Todo) error {
	if todo.ID == nil {
		return fmt.Errorf("todo has no associated issue ID")
	}

	id, err := strconv.Atoi(*todo.ID)
	if err != nil {
		return fmt.Errorf("invalid issue ID: %v", err)
	}

	state := "closed"
	issueRequest := &github.IssueRequest{
		State: &state,
	}

	_, _, err = gs.client.Issues.Edit(gs.ctx, gs.owner, gs.repo, id, issueRequest)
	if err != nil {
		return fmt.Errorf("failed to close issue %d: %v", id, err)
	}

	return nil
}

func (gs *GitHubService) GetIssuesByState(state string) ([]*github.Issue, error) {
	opts := &github.IssueListByRepoOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allIssues []*github.Issue
	for {
		issues, resp, err := gs.client.Issues.ListByRepo(gs.ctx, gs.owner, gs.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %v", err)
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

func (gs *GitHubService) FilterTodosByIssueState(todos []*Todo.Todo, state string) ([]*Todo.Todo, error) {
	issues, err := gs.GetIssuesByState(state)
	if err != nil {
		return nil, err
	}
	issueMap := make(map[int]*github.Issue)
	for _, issue := range issues {
		issueMap[issue.GetNumber()] = issue
	}

	var filteredTodos []*Todo.Todo
	for _, todo := range todos {
		if todo.ID == nil {
			continue
		}

		id, err := strconv.Atoi(*todo.ID)
		if err != nil {
			Log.Info().Msgf("Failed to convert issue id to int: %v", err)
			continue
		}

		if _, exists := issueMap[id]; exists {
			filteredTodos = append(filteredTodos, todo)
		}
	}

	return filteredTodos, nil
}

func ExportCommand(root *cobra.Command) {
	reportCmd := reportCommand()
	purgeCmd := purgeCommand()
	closeCmd := closeCommand()
	root.AddCommand(reportCmd)
	root.AddCommand(closeCmd)
	root.AddCommand(purgeCmd)
}

func reportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "List all the todos in the project and create issues for them",
		Run: func(cmd *cobra.Command, args []string) {
			gs, err := NewGitHubService()
			if err != nil {
				log.Fatalf("Failed to initialize GitHub service: %v", err)
			}

			project := Project.NewProject()
			_, todos, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project: %v", err)
			}

			sort.Slice(todos, func(i, j int) bool {
				return todos[i].Uergency > todos[j].Uergency
			})

			selected := CheckBoxes("Select the todos you want to create issues for", todos)
			fmt.Printf("Selected %d todos\n", len(selected))

			var successfulIssues []string
			for _, todo := range selected {
				fmt.Printf("Creating issue for %s\n", todo.String())
				err := gs.CreateIssue(todo)
				if err != nil {
					log.Printf("Failed to create issue for todo %s: %v", todo.String(), err)
					continue
				}
				fmt.Println("Issue created successfully")
				fmt.Println("Issue ID: ", *todo.ID)
				successfulIssues = append(successfulIssues, *todo.ID)
			}

			if len(successfulIssues) > 0 {
				commitMessage := fmt.Sprintf("Create issues for todos: %s", strings.Join(successfulIssues, " "))
				if err := createGitCommit(commitMessage); err != nil {
					log.Printf("Failed to create commit: %v", err)
				}
			}
		},
	}
}

func purgeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "purge",
		Short: "Delete all closed issues from the project",
		Run: func(cmd *cobra.Command, args []string) {
			gs, err := NewGitHubService()
			if err != nil {
				log.Fatalf("Failed to initialize GitHub service: %v", err)
			}

			project := Project.NewProject()
			todos, _, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project: %v", err)
			}

			closedTodos, err := gs.FilterTodosByIssueState(todos, "closed")
			if err != nil {
				log.Fatalf("Failed to filter closed todos: %v", err)
			}

			if len(closedTodos) == 0 {
				fmt.Println("No closed todos found")
				return
			}

			sort.Slice(closedTodos, func(i, j int) bool {
				return closedTodos[i].Uergency > closedTodos[j].Uergency
			})

			selected := CheckBoxes("Select the closed todos you want to remove from the project", closedTodos)

			projectDir := viper.GetString("input")
			var removedIssues []string
			for _, todo := range selected {
				todo.Remove(projectDir)
				removedIssues = append(removedIssues, *todo.ID)
			}

			if len(removedIssues) > 0 {
				commitMessage := fmt.Sprintf("Remove closed todos from project: %s", strings.Join(removedIssues, " "))
				if err := createGitCommit(commitMessage); err != nil {
					log.Printf("Failed to create commit: %v", err)
				}
			}
		},
	}
}

func closeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "close",
		Short: "Close selected GitHub issues",
		Run: func(cmd *cobra.Command, args []string) {
			gs, err := NewGitHubService()
			if err != nil {
				log.Fatalf("Failed to initialize GitHub service: %v", err)
			}

			project := Project.NewProject()
			todos, _, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project: %v", err)
			}

			// Filter for open issues only
			openTodos, err := gs.FilterTodosByIssueState(todos, "open")
			if err != nil {
				log.Fatalf("Failed to filter open todos: %v", err)
			}

			if len(openTodos) == 0 {
				fmt.Println("No open issues found")
				return
			}

			sort.Slice(openTodos, func(i, j int) bool {
				return openTodos[i].Uergency > openTodos[j].Uergency
			})

			selected := CheckBoxes("Select the issues you want to close", openTodos)

			var closedIssues []string
			for _, todo := range selected {
				if err := gs.CloseIssue(todo); err != nil {
					log.Printf("Failed to close issue for todo %s: %v", todo.String(), err)
					continue
				}
				fmt.Printf("Successfully closed issue %s\n", *todo.ID)
				closedIssues = append(closedIssues, *todo.ID)
			}

			if len(closedIssues) > 0 {
				commitMessage := fmt.Sprintf("Close GitHub issues: %s", strings.Join(closedIssues, " "))
				if err := createGitCommit(commitMessage); err != nil {
					log.Printf("Failed to create commit: %v", err)
				}
			}
		},
	}
}

func createGitCommit(commitMessage string) error {
	projectDir := viper.GetString("input")

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = projectDir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %v", err)
	}

	// Commit changes
	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Dir = projectDir
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}

	fmt.Printf("Created commit: %s\n", commitMessage)
	return nil
}

func convertTodosToOptions(todos []*Todo.Todo) []string {
	var options []string
	for _, todo := range todos {
		options = append(options, todo.LogString())
	}
	return options
}

func CheckBoxes(label string, todos []*Todo.Todo) []*Todo.Todo {
	if len(todos) == 0 {
		fmt.Println("No todos available")
		return nil
	}

	var selectedIndices []int
	options := convertTodosToOptions(todos)
	prompt := &survey.MultiSelect{
		Message: label,
		Options: options,
	}

	err := survey.AskOne(prompt, &selectedIndices)
	if err != nil {
		log.Printf("Error in selection: %v", err)
		return nil
	}

	var selectedTodos []*Todo.Todo
	for _, index := range selectedIndices {
		if index >= 0 && index < len(todos) {
			selectedTodos = append(selectedTodos, todos[index])
		}
	}
	return selectedTodos
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
