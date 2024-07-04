package project 
import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/spf13/viper"
)

func TestLoadDotgit(t *testing.T) {
    // set the viper string "input" to the current directory 
    viper.Set("input", "../")
    // load the .git directory
    gitPth,err := locateDotGit()
    // check if there is an error
    require.NoError(t, err)
    // check if the path is not empty
    require.NotEmpty(t, gitPth)
}
func TestApply(t *testing.T) {
     r:=TransformRule{
         match: ".*",
         replace: "test",
     }
     example := "example"
     result := r.Apply(example)
     require.Equal(t, "test", result)
}
func TestLocateProject(t *testing.T) {
    // set the viper string "input" to the current directory 
    viper.Set("input", "../")
    project := &Project{
        Transforms: make([]TransformRule, 0),
        Keywords: make([]string, 0),
        Remote: "origin",
    }
    projectPath:=project.LocateProject()
    require.NotEmpty(t, projectPath)
    require.Equal(t, "..", projectPath)
}

func TestNewProject(t *testing.T) {
    viper.Set("config", "config.yaml")
    project := NewProject()
    require.NotNil(t, project)
    require.NotNil(t, project.Transforms)
    require.NotNil(t, project.Keywords)
    require.NotNil(t, project.Remote)
}
func TestListFiles(t*testing.T) {
    // set the viper string "input" to the current directory 
    viper.Set("input", "../")
    project := NewProject()
    files, err := project.ListFiles()
    require.NoError(t, err)
    require.NotEmpty(t, files)
}
// func to test the listAllTodos  
func TestListAllTodos(t *testing.T) {
    // set the viper string "input" to the current directory 
    viper.Set("input", "../")
    project := NewProject()
    _, err := listAllTodos(project.Keywords,"project.go")
    require.NoError(t, err)
}
