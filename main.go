package main

import (
	"github.com/Amr-Shams/IssueMe/cmd"
)
// TODO(73): Make your own LSP for the project
func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		panic(err)
	}
}
