package schema_graph

import (
	"fmt"

	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-graph",
	Short: "Schema graph <dir>",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		sch, _ := jsgraph.Load(args[0])
		fmt.Printf("digraph {\n")
		for _, cls := range sch.Classes {
			if cls.Title != "" {
				fmt.Printf("\t%s\n", cls.Title)
			}
		}

		for _, cls := range sch.Classes {
			for propName, prop := range cls.Properties {
				if ext, ok := prop.Extensions[jsgraph.GraphExtensionTag]; ok {
					gExt := ext.(jsgraph.GraphExtension)
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
