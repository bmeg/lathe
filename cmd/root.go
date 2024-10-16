package cmd

import (
	"os"

	"github.com/bmeg/lathe/cmd/inputs"
	"github.com/bmeg/lathe/cmd/outputs"
	"github.com/bmeg/lathe/cmd/prep_upload"
	"github.com/bmeg/lathe/cmd/run"
	"github.com/bmeg/lathe/cmd/viz"

	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "lathe",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(prep_upload.Cmd)
	RootCmd.AddCommand(inputs.Cmd)
	RootCmd.AddCommand(outputs.Cmd)
	RootCmd.AddCommand(run.Cmd)
	RootCmd.AddCommand(viz.Cmd)
}

var genBashCompletionCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completions file",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
