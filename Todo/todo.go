package Todo 
import (
    "fmt"
    "strings"
)

type Todo struct {
    Prefix        string
    Suffix        string
    Keyword       string
    Description  []string
    Uergency      int
    ID            *string 
    FileName      string
    Line          int
    Title         string
}

func (t *Todo) String() string {
    urgencySuffix:=strings.Repeat(string(t.Keyword[len(t.Keyword)-1]), t.Uergency)
    idStr:=""
    if t.ID!=nil {
        idStr="("+*t.ID+")"
    }
   return fmt.Sprintf("%s%s%s%s: %s",
		t.Prefix, t.Keyword, urgencySuffix, idStr,
		t.Suffix)
}


