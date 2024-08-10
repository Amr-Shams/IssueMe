package util

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/spf13/viper"
)

type Cache struct {
	CommitHash      string       `json:"commit_hash"`
	ReportedTodos   []*Todo.Todo `json:"reported_todos"`
	UnreportedTodos []*Todo.Todo `json:"unreported_todos"`
}

func (c *Cache) Save(projDir string) error {
	filePath := viper.GetString("cache")
	filePath = filepath.Join(projDir, filePath)
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func LoadCacheFromFile(filePath string) (*Cache, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var c Cache
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func LoadCache(projetDir string) (*Cache, error) {
	filePath := viper.GetString("cache")
	commitHash, err := GetCommitHash()
	if err != nil {
		return nil, err
	}
	absPath := filepath.Join(projetDir, filePath)
	cached, err := LoadCacheFromFile(absPath)
	if err != nil {
		return nil, err
	}
	if cached.CommitHash == commitHash {
		return cached, nil
	}
	return nil, nil
}
func (c *Cache) UpdateCache(projDir string, reportedTodos []*Todo.Todo, unreportedTodos []*Todo.Todo) error {
	for _, todo := range reportedTodos {
		if !containsTodo(c.ReportedTodos, todo) {
			c.ReportedTodos = append(c.ReportedTodos, todo)
		}
	}

	for _, todo := range unreportedTodos {
		if !containsTodo(c.UnreportedTodos, todo) {
			c.UnreportedTodos = append(c.UnreportedTodos, todo)
		}
	}
	return c.Save(projDir)
}

func containsTodo(todos []*Todo.Todo, todo *Todo.Todo) bool {
	for _, t := range todos {
		if t.String() == todo.String() {
			return true
		}
	}
	return false
}

func GetCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
func GetFiles(projectDir string, allFiles bool) ([]string, error) {
	if allFiles {
		return GetAllFiles(projectDir)
	}
	return GetModifiedFiles()
}

func GetModifiedFiles() ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var modifiedFiles []string
	for i, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if len(line) > 3 {
			if i == 0 {
				modifiedFiles = append(modifiedFiles, line[2:])
				continue
			}
			modifiedFiles = append(modifiedFiles, line[3:])
		}
	}
	return modifiedFiles, nil
}

func GetAllFiles(projectDir string) ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}
