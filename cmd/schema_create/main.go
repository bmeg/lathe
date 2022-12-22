package schema_create

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// content holds our base_schema.
//
//go:embed base_schema/*.yaml
var content embed.FS

var BaseDir = "base_schema"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-create",
	Short: "Schema create <dir>",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dest := args[0]
		fmt.Printf("Creating schema %s\n", dest)
		os.MkdirAll(dest, os.ModePerm)
		dir, _ := content.ReadDir(BaseDir)
		for _, ent := range dir {
			data, _ := content.ReadFile(filepath.Join(BaseDir, ent.Name()))
			fmt.Printf("Creating: %s\n", ent.Name())
			os.WriteFile(filepath.Join(dest, ent.Name()), data, os.ModePerm)
		}
		return nil
	},
}
