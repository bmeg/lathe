package lint

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "lint",
	Short: "Scan directory looking for errors",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		filepath.Walk(baseDir,
			func(path string, info fs.FileInfo, err error) error {
				if strings.HasSuffix(path, ".yaml") {
					pb := playbook.Playbook{}
					if err := playbook.ParseFile(path, &pb); err == nil {
						//log.Printf("Checking %s\n", path)
						if pb.Name == "" {
							log.Printf("Empty transform name: %s", path)
						}
						if pb.Outdir == "" {
							log.Printf("Empty output path: %s", path)
						}
					} else {
						//log.Printf("Unknown file %s\n", path)
					}
				}
				return nil
			})

		return nil
	},
}

func init() {
	//flags := Cmd.Flags()
}
