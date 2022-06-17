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
var doPrep = false

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

		userInputs := map[string]any{}

		steps := []Step{}

		names := []string{}

		for _, dir := range args {
			startDir, _ := filepath.Abs(dir)
			filepath.Walk(startDir,
				func(path string, info fs.FileInfo, err error) error {
					if strings.HasSuffix(path, ".yaml") {
						//log.Printf("Checking %s\n", path)

						pb := playbook.Playbook{}
						if sifterErr := playbook.ParseFile(path, &pb); sifterErr == nil {

							if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

								localInputs := pb.PrepConfig(userInputs, baseDir)
								task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), localInputs)

								//log.Printf("pb outdir %s", task.OutDir())

								taskInputs, _ := pb.GetConfig(task)

								inputs := []string{}
								outputs := []string{}
								for _, p := range taskInputs {
									inputs = append(inputs, p)
								}

								sinks, _ := pb.GetOutputs(task)
								for _, v := range sinks {
									for _, p := range v {
										outputs = append(outputs, p)
									}
								}

								emitters, _ := pb.GetEmitters(task)
								for _, v := range emitters {
									outputs = append(outputs, v)
								}

								cmdPath, _ := filepath.Rel(baseDir, path)

								sName := uniqueName(pb.Name, names)
								names = append(names, sName)
								steps = append(steps, Step{
									Name:    sName,
									Command: fmt.Sprintf("sifter run %s", cmdPath),
									Inputs:  inputs,
									Outputs: outputs,
								})
							}
						} else {
							pl := plans.Plan{}
							if latheErr := plans.ParseFile(path, &pl); latheErr == nil {

								if doPrep {
									if err := pl.DoPrep(); err != nil {
										log.Panicf("Prep Error: %s\n", err)
										return err
									}
								}

								for _, sc := range pl.GetScripts() {
									inputs := []string{}
									outputs := []string{}
									scriptInputs := sc.GetInputs()
									for _, v := range scriptInputs {
										inputs = append(inputs, v)
									}

									scriptOutputs := sc.GetOutputs()
									for _, v := range scriptOutputs {
										outputs = append(outputs, v)
									}
									sName := uniqueName(pb.Name, names)
									names = append(names, sName)
									steps = append(steps, Step{
										Name:    sName,
										Command: sc.GetCommand(),
										Inputs:  inputs,
										Outputs: outputs,
										Workdir: sc.GetWorkdir(),
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
				if k, err := filepath.Rel(baseDir, steps[i].Inputs[j]); err == nil {
					steps[i].Inputs[j] = k
				} else {
					log.Printf("rel error: %s", err)
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
		err = tmpl.Execute(outfile, steps)
		return err
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&changeDir, "dir", "C", changeDir, "Change Directory for script base")
	flags.BoolVarP(&doPrep, "prep", "p", doPrep, "Run prep scripts")
}
