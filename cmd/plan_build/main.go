package plan_build

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/lathe/builder"
	"github.com/spf13/cobra"
)

var changeDir = ""
var exclude = []string{}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan-build",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		if changeDir != "" {
			baseDir, _ = filepath.Abs(changeDir)
		} else if len(args) > 1 {
			return fmt.Errorf("for multiple input directories, based dir must be defined")
		}

		if len(exclude) > 0 {
			log.Printf("Excluding %#v", exclude)
		}

		userInputs := map[string]string{}

		scanStats := builder.ScanStats{}
		steps := []builder.Step{}
		for _, dir := range args {
			t, err := builder.BuildScan(dir, baseDir, exclude, userInputs, &scanStats)
			if err == nil {
				steps = append(steps, t...)
			}
		}

		err := builder.RenderSnakefile(steps, baseDir)

		log.Printf("Sifter file count: %d", scanStats.SifterParseCount)
		log.Printf("Lathe file count: %d", scanStats.LatheParseCount)

		return err
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&changeDir, "dir", "C", changeDir, "Change Directory for script base")
	flags.StringArrayVarP(&exclude, "exclude", "e", exclude, "Paths to exclude")
}
