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

	"github.com/bmeg/sifter/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
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

		sch, _ := schema.Load(args[0])
		fmt.Printf("digraph {\n")
		for _, cls := range sch.Classes {
			if cls.Title != "" {
				fmt.Printf("\t%s\n", cls.Title)
			}
		}

		for _, cls := range sch.Classes {
			for propName, prop := range cls.Properties {
				if ext, ok := prop.Extensions[schema.GraphExtensionTag]; ok {
					gExt := ext.(schema.GraphExtension)
					for _, v := range gExt.Targets {
						fmt.Printf("\t%s -> %s [label=\"%s\"]\n", cls.Title, v.Schema.Title, propName)
						if v.Backref != "" {
							fmt.Printf("\t%s -> %s [label=\"%s\"]\n", v.Schema.Title, cls.Title, v.Backref)
						}
					}
				}
			}
		}
		fmt.Printf("}\n")
		return nil
	},
}
