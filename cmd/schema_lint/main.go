package schema_lint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-lint",
	Short: "Schema lint",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileList, err := filepath.Glob(filepath.Join(args[0], "*.yaml"))
		if err != nil {
			return err
		}
		compiler := jsonschema.NewCompiler()
		compiler.ExtractAnnotations = true
		for _, f := range fileList {
			source, _ := ioutil.ReadFile(f)
			d := map[string]any{}
			yaml.Unmarshal(source, &d)
			schemaText, _ := json.Marshal(d)
			if err := compiler.AddResource(f, strings.NewReader(string(schemaText))); err != nil {
				fmt.Printf("Error loading: %s: %s\n", f, err)
			}
		}

		for _, f := range fileList {
			sch, err := compiler.Compile(f)
			if err != nil {
				fmt.Printf("Error compiling %s : %s\n", f, err)
			} else {
				if len(sch.Types) == 1 && sch.Types[0] == "object" {
					fmt.Printf("OK: %s %s (%s)\n", f, sch.Id, sch.Title)
				}
			}
		}
		return nil
	},
}
