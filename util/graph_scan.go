package util

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
)

type ObjectConvertStep struct {
	Name   string
	Input  string
	Class  string
	Schema string
}

type GraphBuildStep struct {
	Name    string
	Outdir  string
	Objects []ObjectConvertStep
}

var graphScript string = `

name: {{.Name}}
class: sifter

outdir: {{.Outdir}}

config:
{{range .Objects}}
    {{.Name}}: {{.Input}}
    {{.Name}}Schema: {{.Schema}}
{{- end}}

inputs:
{{range .Objects}}
    {{.Name}}:
        jsonLoad:
            input: "{{ "{{config." }}{{.Name}}{{"}}"}}"
{{- end}}

pipelines:
{{range .Objects}}
    {{.Name}}-graph:
        - from: {{.Name}}
        - graphBuild:
            schema: "{{ "{{config."}}{{.Name}}Schema{{ "}}" }}"
            title: {{.Class}}
{{- end}}
`

func ScanObjectToGraph(startDir string, baseDir string, outdir string) []GraphBuildStep {
	userInputs := map[string]string{}

	outSteps := []GraphBuildStep{}

	filepath.Walk(startDir,
		func(path string, info fs.FileInfo, err error) error {
			if strings.HasSuffix(path, ".yaml") {
				log.Printf("Scanning: %s", path)
				pb := playbook.Playbook{}
				if sifterErr := playbook.ParseFile(path, &pb); sifterErr == nil {
					if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

						localInputs, err := pb.PrepConfig(userInputs, baseDir)
						if err == nil {
							scriptDir := filepath.Dir(path)
							task := task.NewTask(pb.Name, scriptDir, baseDir, pb.GetDefaultOutDir(), localInputs)

							gb := GraphBuildStep{Name: pb.Name, Objects: []ObjectConvertStep{}, Outdir: outdir}

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

											schemaPath, err := filepath.Rel(baseDir, schema)
											if err != nil {
												log.Printf("Rel Error: %s", err)
											}

											objCreate := ObjectConvertStep{Name: pname, Input: outpath, Class: s.ObjectValidate.Title, Schema: schemaPath}
											gb.Objects = append(gb.Objects, objCreate)

										}
									}
								}
							}
							if len(gb.Objects) > 0 {
								outSteps = append(outSteps, gb)
							}

						}
					}
				} else {
					//log.Printf("Error: %s", sifterErr)
				}
			}
			return nil
		})
	return outSteps
}

func (gb *GraphBuildStep) GenPlan() ([]byte, error) {
	tmpl, err := template.New("graphscript").Parse(graphScript)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, gb)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	return buf.Bytes(), nil
}
