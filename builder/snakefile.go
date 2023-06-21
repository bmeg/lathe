package builder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bmeg/lathe/scriptfile"
)

type Step struct {
	Name         string
	Command      string
	Inputs       []string
	Outputs      []string
	InputNames   []string
	OutputNames  []string
	Workdir      string
	MemMB        int
	ScatterName  string
	ScatterCount int
	ScriptType   int
}

var snakeFile string = `
{{- if .ScatterGather }}
scattergather:
{{- range $name, $count := .ScatterGather }}
	{{$name}}={{$count}}
{{end -}}
{{end -}}

{{range .Steps}}
rule {{.Name}}:
{{- $saveStep := . -}}
{{- if .Inputs }}
	input:
		{{range $index, $file := .Inputs -}}
		{{- if $index }},
		{{ end -}}
		{{- if $saveStep.InputNames }}{{ index $saveStep.InputNames $index}}= {{end}}{{- $file -}}
		{{- end}}
{{- end}}
{{- if .Outputs }}
	output:
		{{range $index, $file := .Outputs -}}
		{{- if $index }},
		{{ end -}}
		{{- if $saveStep.OutputNames }}{{ index $saveStep.OutputNames $index}}= {{end}}{{- $file -}}
		{{- end}}
{{- end}}
{{- if .MemMB }}
	resources:
		mem_mb={{ .MemMB }}
{{- end}}
{{- if .Command }}
	shell:
		"{{.Command}}"
{{- end}}
{{end}}
`

// currently ignoring workdir
// 		"{{if .Workdir}}cd {{.Workdir}} && {{end}}{{.Command}}"

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
			if s.ScriptType != scriptfile.ScatterScript {
				outputs[f] = 0
			}
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

	var scatterGather map[string]string

	for i := range steps {
		for j := range steps[i].Inputs {
			if steps[i].ScriptType == scriptfile.SifterScript || steps[i].ScriptType == scriptfile.ScatterScript {
				inPath, _ := filepath.Abs(steps[i].Inputs[j])
				if k, err := filepath.Rel(baseDir, inPath); err == nil {
					steps[i].Inputs[j] = fmt.Sprintf(`"%s"`, k)
				} else {
					log.Printf("rel error for input %s: %s", steps[i].Name, err)
				}
			} else if steps[i].ScriptType == scriptfile.GatherScript {
				if scatterGather == nil {
					scatterGather = map[string]string{}
				}
				inPath, _ := filepath.Abs(steps[i].Inputs[0])
				if k, err := filepath.Rel(baseDir, inPath); err == nil {
					steps[i].Inputs[0] = fmt.Sprintf(`gather.%s("%s")`, steps[i].ScatterName, k)
				}
				scatterGather[steps[i].ScatterName] = fmt.Sprintf("%d", steps[i].ScatterCount)
				//add scattergather directive
			}
		}
		for j := range steps[i].Outputs {
			if k, err := filepath.Rel(baseDir, steps[i].Outputs[j]); err == nil {
				steps[i].Outputs[j] = fmt.Sprintf(`"%s"`, k)
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
	err = tmpl.Execute(outfile, map[string]any{"ScatterGather": scatterGather, "Steps": steps})
	return err
}
