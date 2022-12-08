package schema_add

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var classTemplate string = `
$schema: "http://json-schema.org/draft-04/schema#"

id: {{.class}}
title: {{.class}}
type: object

description: >
  Data Element
additionalProperties: false

properties:
  id:
    type: string

`

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-add",
	Short: "Schema add <dir> <class>",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dest := args[0]
		name := args[1]

		fmt.Printf("Creating %s in %s\n", name, dest)

		txt, err := template.New("yamlTemplate").Parse(classTemplate)
		if err != nil {
			return err
		}
		outFile, err := os.OpenFile(filepath.Join(dest, fmt.Sprintf("%s.yaml", name)), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer outFile.Close()
		return txt.Execute(outFile, map[string]any{"class": name})
	},
}
