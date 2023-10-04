package run

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/lathe/builder"
	"github.com/spf13/cobra"
)

var changeDir = "./"
var exclude = []string{}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run scripts",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(exclude) > 0 {
			log.Printf("Excluding %#v", exclude)
		}

		baseDir, _ := filepath.Abs(changeDir)

		scanDir, _ := filepath.Abs(args[0])

		userInputs := map[string]string{}

		scanStats := builder.ScanStats{}

		t, err := builder.BuildScan(scanDir, baseDir, []string{}, userInputs, &scanStats)

		for _, i := range t {
			fmt.Printf("%#v\n", i)
		}

		return err
	},
}
