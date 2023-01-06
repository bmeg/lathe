package plan_list

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

var changeDir = ""
var exclude = []string{}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan-list",
	Short: "Scan directory and list output objects",
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
							pb := playbook.Playbook{}
							if sifterErr := playbook.ParseFile(path, &pb); sifterErr == nil {

								if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

									config, err := pb.PrepConfig(userInputs, baseDir)
									if err == nil {
										task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), config)
										//sourcePath, _ := filepath.Abs(path)
										//cmdPath, _ := filepath.Rel(baseDir, sourcePath)

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
														schema, _ := evaluate.ExpressionString(s.ObjectValidate.Schema, task.GetConfig(), map[string]any{})
														outdir := pb.GetDefaultOutDir()
														outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)

														outpath := filepath.Join(outdir, outname)
														outpath, _ = filepath.Rel(baseDir, outpath)

														schemaPath, _ := filepath.Rel(baseDir, schema)

														_ = schemaPath
														fmt.Printf("%s\t%s\n", s.ObjectValidate.Title, outpath)
														//objCreate := ObjectConvertStep{Name: pname, Input: outpath, Class: s.ObjectValidate.Title, Schema: schemaPath}
														//gb.Objects = append(gb.Objects, objCreate)

													}
												}
											}
										}

									}
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
