package graph_check

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/bmeg/golib"
	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/playbook"
	"github.com/bmeg/sifter/schema"
	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"
)

type output struct {
	Class  string
	Path   string
	Schema string
}

var nLoaderThreads int = 4
var nGenThreads int = 4

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "graph-check",
	Short: "process transform outputs, checking id linkages",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		baseDir := args[0]
		//localInputs := map[string]string{}

		outputs := []output{}
		util.ScanSifter(baseDir, func(pb *playbook.Playbook) {

			scriptDir := filepath.Dir(pb.GetPath())
			task := task.NewTask(pb.Name, scriptDir, baseDir, pb.GetDefaultOutDir(), pb.Config)

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
							schemaPath := filepath.Join(scriptDir, schema)

							outdir := pb.GetDefaultOutDir()
							outname := fmt.Sprintf("%s.%s.%s.json.gz", pb.Name, pname, emitName)
							outpath := filepath.Join(outdir, outname)
							//outpath, _ = filepath.Rel(baseDir, outpath)
							//fmt.Printf("%s\t%s\n", s.ObjectValidate.Title, outpath)

							outputs = append(outputs, output{
								Class:  s.ObjectValidate.Title,
								Path:   outpath,
								Schema: schemaPath,
							})
						}
					}
				}
			}
		})

		loaderInput := make(chan output, nLoaderThreads)
		go func() {
			for _, o := range outputs {
				loaderInput <- o
			}
			close(loaderInput)
		}()

		type genData struct {
			schema *schema.GraphSchema
			class  string
			data   map[string]any
			path   string
			line   int
		}

		genInput := make(chan genData, nLoaderThreads*100)
		loaderWG := &sync.WaitGroup{}
		for i := 0; i < nLoaderThreads; i++ {
			loaderWG.Add(1)
			go func() {
				for o := range loaderInput {
					fmt.Printf("Scaning %s %s %s\n", o.Schema, o.Class, o.Path)
					if sch, err := schema.Load(o.Schema); err == nil {
						reader, err := golib.ReadGzipLines(o.Path)
						if err == nil {
							lineNum := 0
							for line := range reader {
								if len(line) > 0 {
									data := map[string]any{}
									json.Unmarshal(line, &data)
									genInput <- genData{
										schema: &sch,
										class:  o.Class,
										data:   data,
										path:   o.Path,
										line:   lineNum,
									}
								}
								lineNum++
							}
						}
					}
				}
				loaderWG.Done()
			}()
		}

		go func() {
			loaderWG.Wait()
			close(genInput)
		}()

		type logData struct {
			err  error
			path string
			line int
		}
		logInput := make(chan logData, 10)
		genWG := &sync.WaitGroup{}
		for i := 0; i < nGenThreads; i++ {
			genWG.Add(1)
			go func() {
				for gen := range genInput {
					elems, err := gen.schema.Generate(gen.class, gen.data, true)
					if err == nil {
						for _, e := range elems {
							if e.Vertex != nil {
								//fmt.Printf("id: %s\n", e.Vertex.Gid)
							}
						}
					} else {
						logInput <- logData{
							err:  err,
							path: gen.path,
							line: gen.line,
						}
					}
				}
				genWG.Done()
			}()
		}
		go func() {
			genWG.Wait()
			close(logInput)
		}()

		for l := range logInput {
			log.Printf("%s:%d %s", l.path, l.line, l.err)
		}

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.IntVarP(&nGenThreads, "nworkers", "n", nGenThreads, "Number of workers")
}
