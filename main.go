package main

import (
	"github.com/Amr-Shams/IssueMe/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		panic(err)
	}
}
