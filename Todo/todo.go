package Todo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// FIXME(69): The todo should have the keyword used for the comment to clean it up when purging
type Todo struct {
	Prefix      string
	Suffix      string
	Keyword     string
	Description []string
	Uergency    int
	ID          *string
	FileName    string
	Line        int
	Title       string
}

func (t *Todo) String() string {
	urgencySuffix := strings.Repeat(string(t.Keyword[len(t.Keyword)-1]), t.Uergency)
	idStr := ""
	if t.ID != nil {
		idStr = "(" + *t.ID + ")"
	}
	return fmt.Sprintf("%s%s%s:%s", t.Keyword, urgencySuffix, idStr,
		t.Title)
}
func StringifyDescription(description []string) string {
	var str strings.Builder
	for i := 2; i < len(description); i += 3 {
		str.WriteString(description[i])
	}
	return str.String()
}
func (t *Todo) Remove(projectPath string) {
	t.FileName = projectPath + "/" + t.FileName
	file, err := os.Open(t.FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	tempFile, err := os.Create(t.FileName + ".tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer tempFile.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 0
	todoFound := false
	descriptionIndex := 0
	for scanner.Scan() {
		lineNum++
		if lineNum == t.Line {
			tempFile.WriteString(t.Prefix + "\n")
			todoFound = true
			continue
		}
		if todoFound {
			if len(t.Description) > 0 && descriptionIndex < len(t.Description) {
				tempFile.WriteString(t.Description[descriptionIndex] + "\n")
				descriptionIndex += 3
				continue
			} else {
				todoFound = false
				descriptionIndex = 0
			}
		}
		tempFile.WriteString(scanner.Text() + "\n")
	}
	err = os.Rename(t.FileName+".tmp", t.FileName)
	if err != nil {
		log.Fatal(err)
	}
}
func (t *Todo) LogString() string {
	return fmt.Sprintf("%s:%d %s\n%s", t.FileName, t.Line, t.String(), StringifyDescription(t.Description))
}
func (t *Todo) Update(id string, projectPath string) {
	t.ID = &id
	t.FileName = projectPath + "/" + t.FileName
	file, err := os.Open(t.FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	tempFile, err := os.Create(t.FileName + ".tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer tempFile.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum == t.Line {
			tempFile.WriteString(t.Prefix + t.String() + "\n")
		} else {
			tempFile.WriteString(scanner.Text() + "\n")
		}
	}
	err = os.Rename(t.FileName+".tmp", t.FileName)
	if err != nil {
		log.Fatal(err)
	}
}
