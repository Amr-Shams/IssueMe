package project 
import (
    "github.com/spf13/viper"
    "os"
    "path/filepath"
    "log"
    "regexp"
    "os/exec"
    "strings"
    "bufio"
    "gopkg.in/yaml.v3"
)
func locateDotGit()(string,error){
    path := viper.GetString("input")
    gitPath := filepath.Join(path,".git")
    if stat,err := os.Stat(gitPath); err==nil && stat.IsDir(){
        return gitPath,nil
    }
    log.Println("No .git directory found in the project")
    return "",os.ErrNotExist
}

type TransformRule struct {
    match string 
    replace string
}
 func (r *TransformRule) Apply(s string) string {
    re := regexp.MustCompile(r.match)
    return re.ReplaceAllString(s,r.replace)
}
type Project struct {
    Transforms []TransformRule 
    Keywords []string 
    Remote string 
}
func (p *Project) LocateProject() string{
    gitPath,err := locateDotGit();
    if err != nil {
        log.Fatal("Failed to locate project")
        return ""
    }
    return filepath.Dir(gitPath)
}
func listAllTodos(keywords []string, filePath string) ([]string,error){
    file,err := os.Open(filePath)
    if err != nil {
        log.Fatalf("Failed to open file %s",filePath)
        return nil,err
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    todos := make([]string,0)
    for scanner.Scan() {
        line := scanner.Text()
        for _,keyword := range keywords {      
            if strings.Contains(line,keyword) {
                todos = append(todos,line)
                break
            }
        }
    }
    return todos,nil
}
// func to list all the files in the project using git ls-files 
func (p *Project) ListFiles() ([]string,error){
    projectPath := p.LocateProject()
    if projectPath == "" {
        return nil,os.ErrNotExist
    }
    cmd := exec.Command("git","ls-files")
    cmd.Dir = projectPath
    out,err := cmd.Output()
    if err != nil {
        log.Fatalf("Failed to list files in project %s",err.Error())
        return nil,err
    }
    files := strings.Split(string(out),"\n")
    return files,nil
}

func (p* Project) WalkFiles() error {
    files,err := p.ListFiles()
    if err != nil {
        log.Fatalf("Failed to list files in project %s",err.Error())
        return err
    }
    for _,file:= range files {
        stat,err := os.Stat(file)
        if err != nil {
            log.Printf("Failed to stat file %s",file)
            return err
        }
        if stat.IsDir() {
          log.Printf("Skipping directory %s",file)
        }
        // print the file name
        log.Println(file)
    }
    return nil
}
func NewProject() *Project {
    project := &Project{
        Transforms: make([]TransformRule,0),
        Keywords: make([]string,0),
        Remote: "origin",
    }
    configPth := viper.GetString("config")
    if configPth == "" {
        log.Fatal("No config file specified")
    }
    config,err := os.Open(configPth)
    if err != nil {
        log.Fatalf("Failed to open config file %s",configPth)
    }
    defer config.Close()
    decoder := yaml.NewDecoder(config)
  
    if err:= decoder.Decode(&project); err != nil {
        log.Fatalf("Failed to decode config file %s with error %s",configPth,err.Error())
    }
    if len(project.Keywords) == 0 {
        project.Keywords = []string{"TODO","FIXME","BUG"}
    }
    return project
}
//TODO: Add functions to walk and retrieve the files in the project directory 
