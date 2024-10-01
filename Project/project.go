package Project

//
// 




// TODOOOO(72): this is the most important thing
// TODO(70): we should beutify the logs (slogan)
// TODO: we should beutify the logs (slogan)
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/Amr-Shams/IssueMe/Todo"
	"github.com/Amr-Shams/IssueMe/util"
	Log "github.com/charmbracelet/log"
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
				Log.Info(todo.LogString())
			}
			// print a separator between reported and unreported todos
			fmt.Println("-------------------------------------------------")
			for _, todo := range unreportedTodos {
				Log.Info(todo.LogString())
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
	cache      *util.Cache
	cacheMutex sync.Mutex
}

var projectInstance *Project

func (p *Project) LocateProject() string {
	gitPath, err := locateDotGit()
	if err != nil {
		log.Fatal("Failed to locate project")
		return ""
	}
	return filepath.Dir(gitPath)
}

func (p *Project) ListAllTodos() (reportedTodos []*Todo.Todo, unreportedTodos []*Todo.Todo, err error) {
	commitHash, err := util.GetCommitHash()
	if err != nil {
		log.Fatalf("Failed to get commit hash %s", err.Error())
	}
	projectDir := p.LocateProject()
	allFiles := viper.GetBool("clear-cache")

	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()
	p.cache, _ = util.LoadCache(projectDir)
	if p.cache == nil {
		allFiles = true
		p.cache = &util.Cache{
			CommitHash:      commitHash,
			ReportedTodos:   make([]*Todo.Todo, 0),
			UnreportedTodos: make([]*Todo.Todo, 0),
		}
	}

	err = p.processFiles(projectDir, allFiles, &reportedTodos, &unreportedTodos)
	if err != nil {
		return nil, nil, err
	}
	if err := p.cache.UpdateCache(projectDir, reportedTodos, unreportedTodos); err != nil {
		log.Fatalf("Failed to update cache %s", err.Error())
	}
	return p.cache.ReportedTodos, p.cache.UnreportedTodos, nil
}

func (p *Project) processFiles(projectDir string, allFiles bool, reportedTodos, unreportedTodos *[]*Todo.Todo) error {
	files, err := util.GetFiles(projectDir, allFiles)
	if err != nil {
		return err
	}
	numWorkers := 10
	fileChan := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				err := p.processFile(file, reportedTodos, unreportedTodos)
				if err != nil {
					log.Printf("Failed to process file %s: %s", file, err)
				}
			}
		}()
	}
	for _, file := range files {
		if strings.HasPrefix(file, ".") || file == "" {
			continue
		}
		absPath := filepath.Join(projectDir, file)
		fileChan <- absPath
	}
	close(fileChan)
	wg.Wait()
	return nil
}

func (p *Project) processFile(file string, reportedTodos, unreportedTodos *[]*Todo.Todo) error {
	line := 0
	input := util.FromFile(file)
	var todo *Todo.Todo
	for text := range input.Lines() {
		line++
		if todo == nil {
			todo = p.parseLine(text)
			if todo != nil {
				todo.Line = line
				todo.FileName = file
			}
		} else {
			if newTodo := p.parseLine(text); newTodo != nil {
				if todo.ID != nil {
					*reportedTodos = append(*reportedTodos, todo)
				} else {
					*unreportedTodos = append(*unreportedTodos, todo)
				}
				todo = newTodo
				todo.Line = line
				todo.FileName = file
			} else if body := checkComment(text); body != nil {
				todo.Description = append(todo.Description, body[1])
				todo.Description = append(todo.Description, body[3])
				todo.Description = append(todo.Description, body[2])
			} else {
				if todo.ID != nil {
					*reportedTodos = append(*reportedTodos, todo)
				} else {
					*unreportedTodos = append(*unreportedTodos, todo)
				}
				todo = nil
			}
		}
	}
	return nil
}

func NewProject() *Project {
	if projectInstance == nil {
		projectInstance = &Project{
			Transforms: make([]TransformRule, 0),
			Keywords:   make([]string, 0),
			Remote:     "origin",
		}
	}

	project := projectInstance
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
			groups = append(groups, prefix) //
			return groups
		}
	}
	return nil
}
