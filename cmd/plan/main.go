package plan

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bmeg/lathe/plans"
	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type Step struct {
	Name    string
	Command string
	Inputs  []string
	Outputs []string
	Workdir string
	MemMB   int
}

var snakeFile string = `

{{range .}}
rule {{.Name}}:
{{- if .Inputs }}
	input:
		{{range $index, $file := .Inputs -}}
		{{- if $index }},
		{{ end -}}
		"{{- $file -}}"
		{{- end}}
{{- end}}
{{- if .Outputs }}
	output:
		{{range $index, $file := .Outputs -}}
		{{- if $index }},
		{{ end -}}
		"{{- $file -}}"
		{{- end}}
{{- end}}
{{- if .MemMB }}
	resources:
		mem_mb={{ .MemMB }}
{{- end}}
{{- if .Command }}
	shell:
		"{{if .Workdir}}cd {{.Workdir}} && {{end}}{{.Command}}"
{{- end}}
{{end}}


`

func contains(n string, c []string) bool {
	for _, c := range c {
		if n == c {
			return true
		}
	}
	return false
}

func uniqueName(name string, used []string) string {
	name = strings.Replace(name, "-", "_", -1)
	if !contains(name, used) {
		return name
	}
	for i := 1; ; i++ {
		f := fmt.Sprintf("%s_%d", name, i)
		if !contains(f, used) {
			return f
		}
	}
	return name
}

var changeDir = ""
var exclude = []string{}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		if changeDir != "" {
			baseDir, _ = filepath.Abs(changeDir)
		} else if len(args) > 1 {
			return fmt.Errorf("For multiple input directories, based dir must be defined")
		}

		if len(exclude) > 0 {
			log.Printf("Excluding %#v", exclude)
		}

		userInputs := map[string]string{}

		steps := []Step{}

		names := []string{}

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
									if err != nil {
										log.Printf("sifter config error %s: %s ", path, err)
									} else {
										task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), config)
										sourcePath, _ := filepath.Abs(path)
										cmdPath, _ := filepath.Rel(baseDir, sourcePath)

										inputs := []string{}
										outputs := []string{}
										for _, p := range pb.GetConfigFields() {
											if p.IsDir() || p.IsFile() {
												inputs = append(inputs, config[p.Name])
											}
										}
										inputs = append(inputs, cmdPath)

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
										})
									}
								}
							} else {
								pl := plans.Plan{}
								if latheErr := plans.ParseFile(path, &pl); latheErr == nil {
									for i, sc := range pl.GetScripts() {
										inputs := []string{}
										outputs := []string{}
										scriptInputs := sc.GetInputs()
										inputs = append(inputs, scriptInputs...)

										scriptOutputs := sc.GetOutputs()
										outputs = append(outputs, scriptOutputs...)

										sName := uniqueName(fmt.Sprintf("%s_%s", pl.Name, i), names)
										names = append(names, sName)
										steps = append(steps, Step{
											Name:    sName,
											Command: sc.GetCommand(),
											Inputs:  inputs,
											Outputs: outputs,
											Workdir: sc.GetWorkdir(),
											MemMB:   sc.MemMB,
										})
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
		}

		//find all final outputs
		outputs := map[string]int{}
		for _, s := range steps {
			for _, f := range s.Outputs {
				outputs[f] = 0
			}
		}

		for _, s := range steps {
			for _, f := range s.Inputs {
				if x, ok := outputs[f]; ok {
					outputs[f] = x + 1
				}
			}
		}

		allStep := Step{
			Name:   "all",
			Inputs: []string{},
		}
		for k, v := range outputs {
			if v == 0 {
				allStep.Inputs = append(allStep.Inputs, k)
			}
		}
		steps = append([]Step{allStep}, steps...)

		for i := range steps {
			for j := range steps[i].Inputs {
				inPath, _ := filepath.Abs(steps[i].Inputs[j])
				if k, err := filepath.Rel(baseDir, inPath); err == nil {
					steps[i].Inputs[j] = k
				} else {
					log.Printf("rel error for input %s: %s", steps[i].Name, err)
				}
			}
			for j := range steps[i].Outputs {
				if k, err := filepath.Rel(baseDir, steps[i].Outputs[j]); err == nil {
					steps[i].Outputs[j] = k
				}
			}
			if steps[i].Workdir != "" {
				if k, err := filepath.Rel(baseDir, steps[i].Workdir); err == nil {
					steps[i].Workdir = k
				}
			}
		}

		tmpl, err := template.New("snakefile").Parse(snakeFile)
		if err != nil {
			panic(err)
		}

		outfile, err := os.Create(filepath.Join(baseDir, "Snakefile"))
		if err != nil {
			return err
		}
		err = tmpl.Execute(outfile, steps)
		return err
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&changeDir, "dir", "C", changeDir, "Change Directory for script base")
	flags.StringArrayVarP(&exclude, "exclude", "e", exclude, "Paths to exclude")
}
