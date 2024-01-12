package graph_check

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/bmeg/golib"
	jsgraph "github.com/bmeg/jsonschemagraph/util"
	"github.com/bmeg/lathe/util"
	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/playbook"

	"github.com/bmeg/sifter/task"
	"github.com/spf13/cobra"

	"github.com/cockroachdb/pebble"
)

type output struct {
	Class  string
	Path   string
	Schema string
}

var nLoaderThreads int = 4
var nGenThreads int = 4

var vPrefix string = "v"
var tPrefix string = "t"
var fPrefix string = "f"

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
			schema *jsgraph.GraphSchema
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
					if sch, err := jsgraph.Load(o.Schema); err == nil {
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
		type dbEntry struct {
			vertexID string
			edgeFrom string
			edgeTo   string
		}
		logInput := make(chan logData, 10)
		dbInput := make(chan dbEntry, 100)
		genWG := &sync.WaitGroup{}
		for i := 0; i < nGenThreads; i++ {
			genWG.Add(1)
			go func() {
				for gen := range genInput {
					var schema jsgraph.GraphSchema
					elems, err := schema.Generate(gen.class, gen.data, true)
					if err == nil {
						for _, e := range elems {
							if e.Vertex != nil {
								dbInput <- dbEntry{vertexID: e.Vertex.Gid}
							} else if e.InEdge != nil {
								dbInput <- dbEntry{edgeFrom: e.InEdge.From, edgeTo: e.InEdge.To}
							} else if e.OutEdge != nil {
								dbInput <- dbEntry{edgeFrom: e.OutEdge.From, edgeTo: e.OutEdge.To}
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
			close(dbInput)
		}()

		go func() {
			for l := range logInput {
				log.Printf("%s:%d %s", l.path, l.line, l.err)
			}
		}()

		db, err := pebble.Open("keystore.db", &pebble.Options{})
		if err != nil {
			log.Printf("%s", err)
		}

		batch := db.NewBatch()
		pbw := &pebbleBulkWrite{db, batch, nil, nil, 0}

		for d := range dbInput {
			if d.vertexID != "" {
				k := vPrefix + d.vertexID
				pbw.Set([]byte(k), []byte{})
			} else {
				k1 := tPrefix + d.edgeTo
				pbw.Set([]byte(k1), []byte{})
				k2 := fPrefix + d.edgeFrom
				pbw.Set([]byte(k2), []byte{})
			}
			//fmt.Printf("%d\n", pbw.curSize)
		}
		pbw.Close()

		it := db.NewIter(&pebble.IterOptions{LowerBound: []byte(tPrefix)})
		for it.First(); it.Valid() && bytes.HasPrefix(it.Key(), []byte(tPrefix)); it.Next() {
			k := it.Key()
			k[0] = 'v'
			_, cl, err := db.Get(k)
			if err != nil {
				log.Printf("Vertex %s not found", k[1:])
			} else {
				cl.Close()
			}
		}

		it = db.NewIter(&pebble.IterOptions{LowerBound: []byte(fPrefix)})
		for it.First(); it.Valid() && bytes.HasPrefix(it.Key(), []byte(fPrefix)); it.Next() {
			k := it.Key()
			k[0] = 'v'
			_, cl, err := db.Get(k)
			if err != nil {
				log.Printf("Vertex %s not found", k[1:])
			} else {
				cl.Close()
			}
		}

		it.Close()

		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.IntVarP(&nGenThreads, "nworkers", "n", nGenThreads, "Number of workers")
}
