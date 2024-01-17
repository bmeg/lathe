package class_list

import (
	"fmt"
	"path/filepath"

	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "class-list <path_to_playbook(s)>",
	Short: "List output files with class type",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir := args[0]

		util.ScanSifter(baseDir, func(pb *playbook.Playbook) {

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
							outdir := pb.GetDefaultOutDir()
							outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
							outpath := filepath.Join(outdir, outname)
							//outpath, _ = filepath.Rel(baseDir, outpath)
							fmt.Printf("%s\t%s\n", s.ObjectValidate.Title, outpath)
						}
					}
				}
			}

		})

		return nil
	},
}
