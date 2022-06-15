package plan

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

type Step struct {
	Name    string
	Command string
	Inputs  []string
	Outputs []string
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
		"{{.Command}}"
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

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		userInputs := map[string]any{}

		steps := []Step{}

		names := []string{}

		filepath.Walk(baseDir,
			func(path string, info fs.FileInfo, err error) error {
				if strings.HasSuffix(path, ".yaml") {
					log.Printf("Checking %s\n", path)

					pb := playbook.Playbook{}
					if err := playbook.ParseFile(path, &pb); err == nil {

						if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

							localInputs := pb.PrepConfig(userInputs, baseDir)
							task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), localInputs)

							log.Printf("pb outdir %s", task.OutDir())

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
						log.Printf("Skipping %s : %s\n", path, err)
					}
				}
				return nil
			})

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
	//flags := Cmd.Flags()
}
