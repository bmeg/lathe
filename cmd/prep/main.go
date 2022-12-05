package prep

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/plans"
	"github.com/spf13/cobra"
)

var changeDir = ""
var exclude = []string{}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(exclude) > 0 {
			log.Printf("Excluding %#v", exclude)
		}

		for _, dir := range args {
			startDir, _ := filepath.Abs(dir)
			filepath.Walk(startDir,
				func(path string, info fs.FileInfo, err error) error {
					if strings.HasSuffix(path, ".yaml") {
						//log.Printf("Checking %s\n", path)
						doExclude := false

						for _, e := range exclude {
							ePath, _ := filepath.Abs(e)
							if match, err := filepath.Match(ePath, path); match && err == nil {
								doExclude = true
							}
						}
						if !doExclude {
							pl := plans.Plan{}
							if latheErr := plans.ParseFile(path, &pl); latheErr == nil {
								if err := pl.DoPrep(); err != nil {
									log.Printf("Prep Error: %s\n", err)
									return err
								}
							}
						}
					}
					return nil
				})
		}
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&changeDir, "dir", "C", changeDir, "Change Directory for script base")
	flags.StringArrayVarP(&exclude, "exclude", "e", exclude, "Paths to exclude")
}
