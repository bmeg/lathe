package build

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "build",
	Short: "Build snakefile for set of commands",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		for _, file := range args {
			filePath, _ := filepath.Abs(file)
			fmt.Printf("Processing: %s\n", filePath)
		}
		return nil
	},
}
