package Project

// TODO: Add functions to walk and retrieve the files in the project directory

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadDotgit(t *testing.T) {
	// set the viper string "input" to the current directory
	viper.Set("input", "../")
	// load the .git directory
	gitPth, err := locateDotGit()
	// check if there is an error
	require.NoError(t, err)
	// check if the path is not empty
	require.NotEmpty(t, gitPth)
}

func TestLocateProject(t *testing.T) {
	// set the viper string "input" to the current directory
	viper.Set("input", "../")
	project := &Project{
		Transforms: make([]TransformRule, 0),
		Keywords:   make([]string, 0),
		Remote:     "origin",
	}
	projectPath := project.LocateProject()
	require.NotEmpty(t, projectPath)
	require.Equal(t, "..", projectPath)
}

func TestNewProject(t *testing.T) {
	viper.Set("config", "config.yaml")
	viper.Set("input", "../")
	project := NewProject()
	require.NotNil(t, project)
	require.NotNil(t, project.Transforms)
	require.NotNil(t, project.Keywords)
	require.NotNil(t, project.Remote)
}

// func to test the listAllTodos

func TestApplyTransform(t *testing.T) {
	project := &Project{
		Transforms: []TransformRule{
			{match: "^test", replace: "passed"},
		},
		Keywords: []string{},
		Remote:   "origin",
	}
	input := "testApplyTransform"
	expected := "passedApplyTransform"
	result := project.applyTransform(input)
	require.Equal(t, expected, result)
}

func TestParseUnreportedTodoLine(t *testing.T) {
	project := NewProject()
	project.Keywords = []string{"TODO"}
	line := "TODO: This is a test todo"
	todo := project.parseUnreportedTodoLine(line)
	require.NotNil(t, todo)
	require.Equal(t, "This is a test todo", todo.Suffix)
	require.Nil(t, todo.ID)
	require.Equal(t, "TODO", todo.Keyword)
	require.Equal(t, "", todo.Prefix)
}

func TestParseReportedTodoLine(t *testing.T) {
	project := NewProject()
	project.Keywords = []string{"TODO"}
	line := "TODO(user): This is a reported todo"
	todo := project.parseReportedTodoLine(line)
	require.NotNil(t, todo)
	require.Equal(t, "This is a reported todo", todo.Suffix)
	require.Equal(t, "user", *todo.ID)
	require.Equal(t, "TODO", todo.Keyword)
	require.Equal(t, "", todo.Prefix)
}

func TestParseLine(t *testing.T) {
	project := NewProject()
	project.Keywords = []string{"TODO", "FIXME"}
	unreportedLine := "# TODO: This is an unreported todo"
	reportedLine := "#FIXME(user): This is a reported fixme"
	commentInTheMiddle := "This is a comment #TODO: This is a todo"

	unreportedTodo := project.parseLine(unreportedLine)
	reportedTodo := project.parseLine(reportedLine)

	require.NotNil(t, unreportedTodo)
	require.Equal(t, "This is an unreported todo", unreportedTodo.Suffix)
	require.Nil(t, unreportedTodo.ID)
	require.Equal(t, "TODO", unreportedTodo.Keyword)
	require.NotNil(t, reportedTodo)
	require.Equal(t, "This is a reported fixme", reportedTodo.Suffix)
	require.Equal(t, "user", *reportedTodo.ID)
	require.Equal(t, "FIXME", reportedTodo.Keyword)

	commentTodo := project.parseLine(commentInTheMiddle)
	require.NotNil(t, commentTodo)
	require.Equal(t, "This is a todo", commentTodo.Suffix)
	require.Nil(t, commentTodo.ID)
	require.Equal(t, "TODO", commentTodo.Keyword)
	require.Equal(t, "This is a comment ", commentTodo.Prefix)
}
func TestListAllTodos(t *testing.T) {
	project := NewProject()
	_, err := project.ListAllTodos()
	require.NoError(t, err)

}
