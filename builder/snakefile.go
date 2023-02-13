package builder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

func RenderSnakefile(steps []Step, baseDir string) error {

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
}
