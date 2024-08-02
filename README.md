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
