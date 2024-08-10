# IssueMe
a simple tool that list all TODOs and FIXMEs in your project
## How it works 
1. finds all the TODOs and FIXMEs in the project,
2. report the issue to the github
3. purge your closed issue in the project to sync with the remot jobs

## How to use
1. clone the project 
```bash
git clone https://github.com/Amr-Shams/IssueMe.git
```
2. run the following command in the terminal
```bash
go get . 
go build . 
``` 
3. run the tests 
```bash
go test ./...
```
### List all todos(tasks)
`go run main.go list`
- this will return all the open and non-reported tasks
### Report selected Todos
`go run main.go report`
- a checkbox will show up for each todo
- selected ones will be reported to github
- the issue id will be appended to the files
### Purge current closed todos 
` go run main.go purge`
- will remove all the todos on the project (closed on the repo)
#### Resterictions 
- it supports only a single line comment
- this line can be sufixed with live code or an empty space
