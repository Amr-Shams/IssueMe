package project 
import (
    "github.com/spf13/viper"
    "os"
    "path/filepath"
    "log"
    "regexp"
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
