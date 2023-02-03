package schema_lint

import (
	"fmt"

	"github.com/bmeg/grip/log"
	"github.com/bmeg/sifter/schema"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-lint",
	Short: "Schema lint",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		sch, err := schema.Load(args[0], schema.LoadOpt{
			LogError: func(uri string, err error) {
				log.Errorf("Error compiling %s : %s\n", uri, err)
			},
		})
		if err == nil {
			for _, cls := range sch.Classes {
				fmt.Printf("OK: %s (%s)\n", cls.Title, cls.Location)
			}
		}
		return nil
	},
}
