package builder

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/lathe/scriptfile"
	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"gopkg.in/yaml.v2"
)

type ScanStats struct {
	SifterParseCount int
	LatheParseCount  int
}

func BuildScan(dir string, baseDir string, exclude []string, userInputs map[string]string, scanStats *ScanStats) ([]Step, error) {
	steps := []Step{}

	names := []string{}

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
							if err != nil {
								log.Printf("sifter config error %s: %s ", path, err)
							} else {
								scanStats.SifterParseCount++
								scriptDir := filepath.Dir(path)
								task := task.NewTask(pb.Name, scriptDir, baseDir, pb.GetDefaultOutDir(), config)
								sourcePath, _ := filepath.Abs(path)
								cmdPath, _ := filepath.Rel(baseDir, sourcePath)

								inputs := []string{}
								outputs := []string{}
								for _, p := range pb.GetConfigFields() {
									if p.IsDir() || p.IsFile() {
										inputs = append(inputs, config[p.Name])
									}
								}
								inputs = append(inputs, sourcePath)

								sinks, _ := pb.GetOutputs(task)
								for _, v := range sinks {
									outputs = append(outputs, v...)
								}

								emitters, _ := pb.GetEmitters(task)
								for _, v := range emitters {
									outputs = append(outputs, v)
								}

								sName := uniqueName(pb.Name, names)
								names = append(names, sName)
								steps = append(steps, Step{
									Name:    sName,
									Command: fmt.Sprintf("sifter run %s", cmdPath),
									Inputs:  inputs,
									Outputs: outputs,
									MemMB:   pb.MemMB,
								})
							}
						}
					} else {
						pl := scriptfile.ScriptFile{}
						if latheErr := scriptfile.ParseFile(path, &pl); latheErr == nil {
							scanStats.LatheParseCount++
							for i, sc := range pl.GetScripts() {
								inputs := []string{}
								inputNames := []string{}
								outputs := []string{}
								outputNames := []string{}

								for k, v := range sc.GetInputs() {
									inputs = append(inputs, v)
									inputNames = append(inputNames, k)
								}

								for k, v := range sc.GetOutputs() {
									outputs = append(outputs, v)
									outputNames = append(outputNames, k)
								}

								sName := uniqueName(fmt.Sprintf("%s_%s", pl.Name, i), names)
								names = append(names, sName)
								newStep := Step{
									Name:         sName,
									Command:      sc.GetCommand(),
									Inputs:       inputs,
									InputNames:   inputNames,
									Outputs:      outputs,
									OutputNames:  outputNames,
									Workdir:      sc.GetWorkdir(),
									MemMB:        sc.MemMB,
									ScatterName:  sc.GetScatterName(),
									ScatterCount: sc.GetScatterCount(),
									ScriptType:   sc.GetScriptType(),
								}
								if sc.Docker != nil {
									newStep.Container = "docker://" + sc.Docker.Image
								}
								steps = append(steps, newStep)
							}
							for i, concat := range pl.GetCollections() {
								fmt.Printf("Collection: %s %d\n", concat.GetOutputPath(), i)

								sName := uniqueName(fmt.Sprintf("%s_collect_%d", pl.Name, i), names)
								names = append(names, sName)
								//./lathe class-concat ../bmeg-etl/transform/ allele -o ../bmeg-etl/output/allele/allele.json.gz
								inputs := []string{}

								util.ScanSifter(dir, func(pb *playbook.Playbook) {
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
													if s.ObjectValidate.Title == concat.Class {
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

								outFile, _ := filepath.Rel(baseDir, concat.GetOutputPath())
								steps = append(steps, Step{
									Name:    sName,
									Command: fmt.Sprintf("lathe class-concat %s %s -o %s", dir, concat.Class, outFile),
									Inputs:  inputs,
									Outputs: []string{outFile},
								})

							}
						} else {
							source, _ := ioutil.ReadFile(path)
							d := map[string]any{}
							yaml.Unmarshal(source, &d)
							if cl, ok := d["class"]; ok {
								if cls, ok := cl.(string); ok {
									if cls == "lathe" {
										log.Printf("Skipping lathe %s : %s\n", path, latheErr)
									}
									if cls == "sifter" {
										log.Printf("Skipping sifter %s : %s\n", path, sifterErr)
									}
								}
							}
						}
					}
				}
			}
			return nil
		})

	return steps, nil
}
