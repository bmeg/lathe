package cmd

import (
	"os"

	"github.com/bmeg/lathe/cmd/build"
	"github.com/bmeg/lathe/cmd/data_validate"
	"github.com/bmeg/lathe/cmd/lint"
	"github.com/bmeg/lathe/cmd/plan"
	"github.com/bmeg/lathe/cmd/plangraph"
	"github.com/bmeg/lathe/cmd/prep"
	"github.com/bmeg/lathe/cmd/schema_add"
	"github.com/bmeg/lathe/cmd/schema_create"
	"github.com/bmeg/lathe/cmd/schema_lint"

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
	RootCmd.AddCommand(prep.Cmd)
	RootCmd.AddCommand(lint.Cmd)
	RootCmd.AddCommand(build.Cmd)
	RootCmd.AddCommand(schema_create.Cmd)
	RootCmd.AddCommand(schema_lint.Cmd)
	RootCmd.AddCommand(schema_add.Cmd)
	RootCmd.AddCommand(data_validate.Cmd)
	RootCmd.AddCommand(plangraph.Cmd)
}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
