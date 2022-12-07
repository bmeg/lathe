package data_validate

import (
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "data-validate",
	Short: "Data Validate",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		//fileList, err := filepath.Glob(filepath.Join(args[0], "*.yaml"))
		//if err != nil {
		//	return err
		//}
		return nil
	},
}
