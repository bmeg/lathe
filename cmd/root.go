package cmd

import (
	"os"

	"github.com/bmeg/lathe/cmd/build"
	"github.com/bmeg/lathe/cmd/lint"
	"github.com/bmeg/lathe/cmd/plan"
	"github.com/bmeg/lathe/cmd/plangraph"

	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "lathe",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(plan.Cmd)
	RootCmd.AddCommand(lint.Cmd)
	RootCmd.AddCommand(build.Cmd)
	RootCmd.AddCommand(plangraph.Cmd)
}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
