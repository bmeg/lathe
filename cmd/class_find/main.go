package class_find

import (
	"fmt"
	"path/filepath"

	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "class-find",
	Short: "Find output files of class type",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir := args[0]
		className := args[1]
		//userInputs := map[string]string{}

		util.ScanSifter(baseDir, func(pb *playbook.Playbook) {
			//localInputs, err := pb.PrepConfig(userInputs, baseDir)
			//task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), localInputs)

			for pname, p := range pb.Pipelines {
				emitName := ""
				for _, s := range p {
					if s.Emit != nil {
						emitName = s.Emit.Name
					}
				}
				if emitName != "" {
					for _, s := range p {
						if s.ObjectValidate != nil {
							if s.ObjectValidate.Class == className {
								outdir := pb.GetDefaultOutDir()
								outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
								outpath := filepath.Join(outdir, outname)
								fmt.Printf("%s\n", outpath)
							}
						}
					}
				}
			}

		})

		return nil
	},
}
