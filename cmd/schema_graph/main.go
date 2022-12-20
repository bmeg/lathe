package schema_graph

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/structtag"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func yamlLoader(s string) (io.ReadCloser, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	f := u.Path
	if runtime.GOOS == "windows" {
		f = strings.TrimPrefix(f, "/")
		f = filepath.FromSlash(f)
	}
	if strings.HasSuffix(f, ".yaml") {
		source, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		d := map[string]any{}
		yaml.Unmarshal(source, &d)
		schemaText, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(strings.NewReader(string(schemaText))), nil
	}
	return os.Open(f)
}

func isEdge(s string) bool {
	if strings.Contains(s, "_definitions.yaml#/to_many") {
		return true
	} else if strings.Contains(s, "_definitions.yaml#/to_one") {
		return true
	}
	return false
}

var referenceMeta = jsonschema.MustCompileString("referenceMeta.json", `{
	"properties" : {
		"reference_type_enum": {
			"type": "string"
		},
		"reference_backref": {
			"type" : "string"
		}
	}
}`)

type referenceCompiler struct{}

type referenceSchema struct {
	typeEnum []string
	backRef  string
}

func (s referenceSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	return nil
}

func (referenceCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	eString := ""
	if e, ok := m["reference_type_enum"]; ok {
		n, _ := e.(string)
		eString = n
	}
	brString := ""
	if e, ok := m["reference_backref"]; ok {
		n, _ := e.(string)
		brString = n
	}
	if eString == "" && brString == "" {
		return nil, nil
	}
	// nothing to compile, return nil
	return referenceSchema{[]string{eString}, brString}, nil
}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-graph",
	Short: "Schema graph <dir>",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		jsonschema.Loaders["file"] = yamlLoader

		compiler := jsonschema.NewCompiler()
		compiler.ExtractAnnotations = true

		compiler.RegisterExtension("reference_type_enum", referenceMeta, referenceCompiler{})

		fileList, err := filepath.Glob(filepath.Join(args[0], "*.yaml"))
		if err != nil {
			return err
		}

		for _, f := range fileList {
			sch, err := compiler.Compile(f)

			if err != nil {
				fmt.Printf("Error compiling %s : %s\n", f, err)
			} else {
				if len(sch.Types) == 1 && sch.Types[0] == "object" {
					fmt.Printf("%s\n", sch.Id)
					for k, v := range sch.Properties {
						fmt.Printf("\t%s - %#v\n", k, v.Title)
						if v.Ref != nil {
							if isEdge(v.Ref.Location) {
								//fmt.Printf("Ref: %#v\n", v.Ref.Location)
								fmt.Printf("%s --> %s\n", sch.Id, v.Title)
								fmt.Printf("Extension: %#v\n", v.Extensions)
								tags, err := structtag.Parse(string(v.Description))
								if err == nil {
									for _, t := range tags.Tags() {
										fmt.Printf("%s: %s\n", t.Key, t.Name)
									}
								}
							}
						}
					}

				}
			}
		}

		return nil
	},
}
