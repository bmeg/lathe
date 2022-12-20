package lint

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/plans"
	"github.com/bmeg/sifter/playbook"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "lint",
	Short: "Scan directory looking for errors",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		userInputs := map[string]string{}

		filepath.Walk(baseDir,
			func(path string, info fs.FileInfo, err error) error {
				if strings.HasSuffix(path, ".yaml") {
					pb := playbook.Playbook{}
					if parseErr := playbook.ParseFile(path, &pb); parseErr == nil {
						//log.Printf("Checking %s\n", path)
						if pb.Name == "" {
							log.Printf("Empty transform name: %s", path)
						}
						if pb.Outdir == "" {
							log.Printf("Empty output path: %s", path)
						}
						if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

							_, err := pb.PrepConfig(userInputs, baseDir)
							if err != nil {
								log.Printf("sifter config error %s: %s ", path, err)
							}
						}
					} else {
						pl := plans.Plan{}
						if latheErr := plans.ParseFile(path, &pl); latheErr == nil {
							if pl.Name == "" {
								log.Printf("lathe file %s missing name", path)
							}
						} else {
							// Double check if this was a sifter file in the first place
							// TODO: maybe do this check before trying to parse
							data, err := os.ReadFile(path)
							if err == nil {
								dst := map[string]any{}
								err = yaml.Unmarshal(data, &dst)
								if err == nil {
									if k, ok := dst["class"]; ok {
										if kStr, ok := k.(string); ok {
											if kStr == "sifter" {
												log.Printf("Error %s : %s\n", path, parseErr)
											}
										}
									} else {
										log.Printf("%s no class entry", path)
									}
								} else {
									log.Printf("%s is not valid yaml file", path)
								}
							}
						}
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
