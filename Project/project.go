package Project

// FIXMEEE(49): This is a bug
// TODO: This is a BUG
// FIXME(44): This is a hack
// BUG(43): This is a hack
// This is bug is made by me

// BUG(43): This is a hack
// This is bug is made by me

// This is bug is made by me
// TODOOOO: this is the most important thing
// TODO: we should beutify the logs (slogan)
import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func ExportCommand(root *cobra.Command) {
	listCmd := listingCommand()
	root.AddCommand(listCmd)
}
func listingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all the todos in the project",
		Run: func(cmd *cobra.Command, args []string) {
			project := NewProject()
			reportedTodos, unreportedTodos, err := project.ListAllTodos()
			if err != nil {
				log.Fatalf("Failed to list all todos in the project %s", err.Error())
			}
			sort.Slice(reportedTodos, func(i, j int) bool {
				return reportedTodos[i].Uergency > reportedTodos[j].Uergency
			})
			sort.Slice(unreportedTodos, func(i, j int) bool {
				return unreportedTodos[i].Uergency > unreportedTodos[j].Uergency
			})
			for _, todo := range reportedTodos {
				log.Printf(todo.LogString())
			}
			// print a separator between reported and unreported todos
			fmt.Println("-------------------------------------------------")
			for _, todo := range unreportedTodos {

				log.Printf(todo.LogString())
			}
		},
	}
}

func locateDotGit() (string, error) {
	path := viper.GetString("input")
	gitPath := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitPath); err == nil && stat.IsDir() {
		return gitPath, nil
	}
	log.Println("No .git directory found in the project")
	return "", os.ErrNotExist
}

type TransformRule struct {
	match   string `yaml:"match"`
	replace string `yaml:"replace"`
}

func (p *Project) applyTransform(s string) string {
	for _, rule := range p.Transforms {
		re := regexp.MustCompile(rule.match)
		s = re.ReplaceAllString(s, rule.replace)
	}
	return s
}

type Project struct {
	Transforms []TransformRule `yaml:"Transforms"`
	Keywords   []string        `yaml:"Keywords"`
	Remote     string          `yaml:"Remote"`
}

func (p *Project) LocateProject() string {
	gitPath, err := locateDotGit()
	if err != nil {
		log.Fatal("Failed to locate project")
		return ""
	}
	return filepath.Dir(gitPath)
}

func (p *Project) ListAllTodos() (reportedTodos []*Todo.Todo, unreportedTodos []*Todo.Todo, err error) {
	p.WalkFiles(func(file string) error {
		f, err := os.Open(file)
		if err != nil {
			log.Printf("Failed to open file %s", file)
			return err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		line := 0
		var todo *Todo.Todo
		for scanner.Scan() {
			line++
			if todo == nil {
				todo = p.parseLine(scanner.Text())
				if todo != nil {
					todo.Line = line
					todo.FileName = file
				}
			} else {
				if newTodo := p.parseLine(scanner.Text()); newTodo != nil {
					if todo.ID != nil {
						reportedTodos = append(reportedTodos, todo)
					} else {
						unreportedTodos = append(unreportedTodos, todo)
					}
					todo = newTodo
					todo.Line = line
					todo.FileName = file
				} else if body := checkComment(scanner.Text()); body != nil {
					todo.Description = append(todo.Description, body[1])
					todo.Description = append(todo.Description, body[3])
					todo.Description = append(todo.Description, body[2])
				} else {
					if todo.ID != nil {
						reportedTodos = append(reportedTodos, todo)
					} else {
						unreportedTodos = append(unreportedTodos, todo)
					}
					todo = nil
				}
			}
			if err := scanner.Err(); err != nil {
				log.Printf("Failed to scan file %s", file)
				return err
			}

		}
		return nil
	})
	return reportedTodos, unreportedTodos, nil

}

// func to list all the files in the project using git ls-files

func (p *Project) WalkFiles(visitor func(string) error) error {
	projectPath := p.LocateProject()
	if projectPath == "" {
		return os.ErrNotExist
	}
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to list files in project %s", err.Error())
		return err
	}
	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasPrefix(file, ".") || file == "" {
			continue
		}
		absPath := filepath.Join(projectPath, file)
		stat, err := os.Stat(absPath)
		if err != nil {
			log.Printf("Failed to stat file %s with error %s", absPath, err.Error())
			return err
		}
		if stat.IsDir() {
			log.Printf("Skipping directory %s", absPath)
		}
		err = visitor(absPath)
		if err != nil {
			return err
		}
	}
	return nil
}
func NewProject() *Project {
	project := &Project{
		Transforms: make([]TransformRule, 0),
		Keywords:   make([]string, 0),
		Remote:     "origin",
	}
	configPth := viper.GetString("config")
	projectPath := viper.GetString("input")
	configPth = filepath.Join(projectPath, configPth)
	if configPth == "" {
		configPth = "config.yaml"
	}
	config, err := os.Open(configPth)
	if err != nil {
		log.Fatalf("Failed to open config file %s", configPth)
	}
	defer config.Close()
	decoder := yaml.NewDecoder(config)
	if err := decoder.Decode(&project); err != nil {
		log.Fatalf("Failed to decode config file %s with error %s", configPth, err.Error())
	}
	if len(project.Keywords) == 0 {
		project.Keywords = []string{"TODO", "FIXME", "BUG"}
	}
	return project
}

func unreportedTodoRgex(keyword string) string {
	return "^(.*)" + regexp.QuoteMeta(keyword) + "(" + regexp.QuoteMeta(keyword[len(keyword)-1:]) + "*):(.*)$"
}

func reportedTodoRgex(keyword string) string {
	return "^(.*)" + regexp.QuoteMeta(keyword) + "(" + regexp.QuoteMeta(keyword[len(keyword)-1:]) + "*)" + "\\((.*)\\):(.*)$"
}
func (p *Project) parseUnreportedTodoLine(line string) *Todo.Todo {
	for _, k := range p.Keywords {
		re := regexp.MustCompile(unreportedTodoRgex(k))
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			perfix := matches[1]
			suffix := matches[3]
			urgency := len(matches[2])
			title := p.applyTransform(matches[3])
			return &Todo.Todo{
				Prefix:   perfix,
				Suffix:   suffix,
				Keyword:  k,
				Uergency: urgency,
				Line:     0,
				FileName: "",
				ID:       nil,
				Title:    title,
			}
		}
	}
	return nil
}

func (p *Project) parseReportedTodoLine(line string) *Todo.Todo {
	for _, k := range p.Keywords {
		re := regexp.MustCompile(reportedTodoRgex(k))
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			perfix := matches[1]
			suffix := matches[4]
			urgency := len(matches[2])
			id := matches[3]
			title := p.applyTransform(matches[4])
			return &Todo.Todo{
				Prefix:   perfix,
				Suffix:   suffix,
				Keyword:  k,
				Uergency: urgency,
				Line:     0,
				FileName: "",
				ID:       &id,
				Title:    title,
			}
		}
	}
	return nil
}

func (p *Project) parseLine(line string) *Todo.Todo {
	comment := checkComment(line)
	if comment == nil {
		return nil
	}
	if todo := p.parseUnreportedTodoLine(comment[2]); todo != nil {
		todo.Prefix = comment[1] + comment[3] + todo.Prefix
		return todo
	}
	if todo := p.parseReportedTodoLine(comment[2]); todo != nil {
		todo.Prefix = comment[1] + comment[3] + todo.Prefix
		return todo
	}
	return nil
}
func checkComment(line string) []string {
	commentPrefixes := []string{"//", "#", "--"}
	for _, prefix := range commentPrefixes {
		regex := "^(.*?)" + regexp.QuoteMeta(prefix) + "(.*)$"
		re := regexp.MustCompile(regex)
		groups := re.FindStringSubmatch(line)
		if groups != nil {
			groups = append(groups, prefix)
			return groups
		}
	}
	return nil
}
