package plangraph

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
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
    {{.Name}}:
        type: File
        default: {{.Input}}
    {{.Name}}Schema:
        type: Dir
        default: {{.Schema}}
{{end}}

inputs:
{{range .Objects}}
    {{.Name}}:
        jsonLoad:
            input: "{{ "{{config." }}{{.Name}}{{"}}"}}"
{{end}}

pipelines:
{{range .Objects}}
    {{.Name}}-graph:
        - from: {{.Name}}
        - graphBuild:
            schema: "{{ "{{config."}}{{.Name}}Schema{{ "}}" }}"
            class: {{.Class}}
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
}

var changeDir = ""
var outDir = "./"
var doPrep = false

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "plan-graph",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir, _ := filepath.Abs(args[0])

		if changeDir != "" {
			baseDir, _ = filepath.Abs(changeDir)
		} else if len(args) > 1 {
			return fmt.Errorf("For multiple input directories, based dir must be defined")
		}

		_ = baseDir

		outDir, _ := filepath.Abs(outDir)
		outDir, _ = filepath.Rel(baseDir, outDir)

		userInputs := map[string]any{}

		for _, dir := range args {
			startDir, _ := filepath.Abs(dir)
			filepath.Walk(startDir,
				func(path string, info fs.FileInfo, err error) error {
					if strings.HasSuffix(path, ".yaml") {
						pb := playbook.Playbook{}
						if sifterErr := playbook.ParseFile(path, &pb); sifterErr == nil {
							if len(pb.Pipelines) > 0 || len(pb.Inputs) > 0 {

								localInputs := pb.PrepConfig(userInputs, baseDir)
								task := task.NewTask(pb.Name, baseDir, pb.GetDefaultOutDir(), localInputs)

								gb := GraphBuildStep{Name: pb.Name, Objects: []ObjectConvertStep{}, Outdir: outDir}

								for pname, p := range pb.Pipelines {
									emitName := ""
									for _, s := range p {
										if s.Emit != nil {
											emitName = s.Emit.Name
										}
									}
									if emitName != "" {
										for _, s := range p {
											if s.ObjectCreate != nil {
												schema, _ := evaluate.ExpressionString(s.ObjectCreate.Schema, task.GetConfig(), map[string]any{})
												outdir := pb.GetDefaultOutDir()
												outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)

												outpath := filepath.Join(outdir, outname)
												outpath, _ = filepath.Rel(baseDir, outpath)

												schemaPath, _ := filepath.Rel(baseDir, schema)

												_ = schemaPath

												objCreate := ObjectConvertStep{Name: pname, Input: outpath, Class: s.ObjectCreate.Class, Schema: schemaPath}
												gb.Objects = append(gb.Objects, objCreate)

											}
										}
									}
								}

								if len(gb.Objects) > 0 {
									tmpl, err := template.New("graphscript").Parse(graphScript)
									if err != nil {
										panic(err)
									}

									outfile, err := os.Create(filepath.Join(baseDir, fmt.Sprintf("%s.yaml", pb.Name)))
									err = tmpl.Execute(outfile, gb)
									outfile.Close()
									if err != nil {
										fmt.Printf("Error: %s\n", err)
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
	flags.StringVarP(&outDir, "out", "o", outDir, "Change output Directory")
}
