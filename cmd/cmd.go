package cmd

import (
	"fmt"
	"time"

	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addCommand(root *cobra.Command) {

}

type prof interface {
	Stop()
}

func NewRootCommand() *cobra.Command {
	var (
		start    time.Time
		profiler prof
	)

	result := &cobra.Command{
		Use:     "IssueMe-2024",
		Short:   "Priorities of your project(life) in 2024 ",
		Long:    "Golang implementations for Todo priorities of your projects",
		Example: "go run main.go list  --input ./project1  --profile --config ./config.yaml",
		Args:    cobra.ExactArgs(1),
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if viper.GetBool("profile") {
				profiler = profile.Start()
			}
			start = time.Now()
		},
		PersistentPostRun: func(_ *cobra.Command, _ []string) {
			if profiler != nil {
				profiler.Stop()
			}

			fmt.Println("Took", time.Since(start))
		},
	}

	addCommand(result)

	flags := result.PersistentFlags()
	flags.StringP("input", "i", ".", "Input Project directory for the project if not provided it will use the current directory as the project directory")
	flags.Bool("profile", false, "Profile implementation performance")
	flags.StringP("config", "c", "config.yaml", "Config file for the project")
	// give a default value to the input flag
	_ = viper.BindPFlags(flags)

	return result
}
