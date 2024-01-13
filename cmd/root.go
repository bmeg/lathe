package cmd

import (
	"os"

	"github.com/bmeg/lathe/cmd/graph_build"
	"github.com/bmeg/lathe/cmd/graph_check"
	"github.com/bmeg/lathe/cmd/lint"
	"github.com/bmeg/lathe/cmd/plan_graph"
	"github.com/bmeg/lathe/cmd/plan_list"
	"github.com/bmeg/lathe/cmd/prep"
	"github.com/bmeg/lathe/cmd/prep_manifest"
	"github.com/bmeg/lathe/cmd/prep_upload"
	"github.com/bmeg/lathe/cmd/run"

	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "lathe",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(prep.Cmd)
	RootCmd.AddCommand(prep_manifest.Cmd)
	RootCmd.AddCommand(prep_upload.Cmd)
	RootCmd.AddCommand(lint.Cmd)
	RootCmd.AddCommand(plan_graph.Cmd)
	RootCmd.AddCommand(plan_list.Cmd)
	RootCmd.AddCommand(graph_check.Cmd)
	RootCmd.AddCommand(graph_build.Cmd)
	RootCmd.AddCommand(run.Cmd)

}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
