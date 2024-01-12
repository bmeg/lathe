package schema_lint

import (
	"fmt"

	"github.com/bmeg/grip/log"
	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/spf13/cobra"
)

var httpDir string

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "schema-lint [schema dir]",
	Short: "Checks a directory of yaml schemas for syntax errors",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		sch, err := jsgraph.Load(args[0])
		if err == nil {
			for _, cls := range sch.Classes {
				fmt.Printf("OK: %s (%s)\n", cls.Title, cls.Location)
			}
		} else {
			log.Errorf("Loading error: %s", err)
		}
		return nil
	},
}

func init() {

}
