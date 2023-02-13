package plan

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/lathe/builder"
	"github.com/bmeg/lathe/plan"
	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan <plan file>",
	Short: "Scan directory to plan operations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pln := plan.Plan{}
		plan.ParseFile(args[0], &pln)

		abDir, _ := filepath.Abs(args[0])
		baseDir := filepath.Dir(abDir)

		userInputs := map[string]string{}

		scanStats := builder.ScanStats{}
		steps := []builder.Step{}

		for _, step := range pln.Steps {
			if step.BuildCommands != nil {
				sDir := filepath.Join(baseDir, step.BuildCommands.Dir)
				t, err := builder.BuildScan(sDir, baseDir, []string{}, userInputs, &scanStats)
				if err == nil {
					steps = append(steps, t...)
				}
			} else if step.CollectClass != nil {
				inDir := filepath.Join(baseDir, step.CollectClass.Dir)

				inputs := []string{}

				util.ScanSifter(inDir, func(pb *playbook.Playbook) {
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
									if s.ObjectValidate.Title == step.CollectClass.Title {
										outdir := pb.GetDefaultOutDir()
										outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
										outpath := filepath.Join(outdir, outname)
										inputs = append(inputs, outpath)
									}
								}
							}
						}
					}
				})

				outDir := filepath.Join(baseDir, step.CollectClass.Output)
				rel, _ := filepath.Rel(baseDir, outDir)
				s := builder.Step{
					Name: fmt.Sprintf("collect_%s", step.CollectClass.Title),
					Command: fmt.Sprintf("lathe class-concat %s %s -o %s",
						step.CollectClass.Title, step.CollectClass.Dir,
						rel,
					),
					Outputs: []string{outDir},
					Inputs:  inputs,
				}
				steps = append(steps, s)
			}

		}

		err := builder.RenderSnakefile(steps, baseDir)
		if err != nil {
			log.Printf("%s\n", err)
		}

		log.Printf("Sifter file count: %d", scanStats.SifterParseCount)
		log.Printf("Lathe file count: %d", scanStats.LatheParseCount)

		return nil
	},
}

func init() {
	//flags := Cmd.Flags()
}
