package github

import (
	"testing"

	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetIssues(t *testing.T) {
	owner, repo, error := getRepoInfo()
	require.NoError(t, error)
	require.NotEmpty(t, owner)
	require.NotEmpty(t, repo)
	url := "https://api.github.com/repos/" + owner + "/" + repo + "/issues"
	expected := "https://api.github.com/repos/Amr-Shams/IssueMe/issues"
	require.Equal(t, expected, url)
}

func TestFireIssue(t *testing.T) {
	// Mock the Todo struct
	todo := &Todo.Todo{
		Prefix:      "Test",
		Suffix:      "Issue",
		Keyword:     "TODO",
		Description: "This is a test issue.",
		Uergency:    1,
		ID:          nil,
		FileName:    "test_file.go",
		Line:        10,
		Title:       "Test Issue Title",
	}
	viper.Set("input", "../")
	err := FireIssue(todo)
	require.NoError(t, err)
	require.NotNil(t, todo.ID)
}
