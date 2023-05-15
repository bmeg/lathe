package scriptfile

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmeg/goatee"

	"github.com/google/shlex"
)

type Script struct {
	CommandLine  string            `json:"commandLine"`
	Inputs       map[string]string `json:"inputs"`
	Outputs      map[string]string `json:"outputs"`
	Workdir      string            `json:"workdir"`
	Order        int               `json:"order"`
	MemMB        int               `json:"memMB"`
	path         string
	name         string
	scatterName  string
	scatterCount int
}

type ScatterGather struct {
	CommandLine string `json:"commandLine"`
	Input       string `json:"input"`
	Output      string `json:"output"`
	MemMB       int    `json:"memMB"`
	Shards      int    `json:"shards"`
	ShardPath   string `json:"shardPath"`
}

type PrepStage struct {
	IfMissing   string `json:"ifMissing"`
	CommandLine string `json:"commandLine"`
}

type Template struct {
	Inputs  []map[string]any `json:"inputs"`
	Scripts map[string]any   `json:"scripts"`
}

type CollectData struct {
	Class  string `json:"class"`
	Output string `json:"output"`
	path   string
}

type FileRecord struct {
	DownloadDate string `json:"downloadDate"`
	Source       string `json:"source"`
	Path         string `json:"path"`
}

type ScriptFile struct {
	Class         string                    `json:"class"`
	Name          string                    `json:"name"`
	Scripts       map[string]*Script        `json:"scripts"`
	Templates     map[string]*Template      `json:"templates"`
	Prep          []PrepStage               `json:"prep"`
	Collections   []CollectData             `json:"collections"`
	Files         []FileRecord              `json:"files"`
	ScatterGather map[string]*ScatterGather `json:"scatterGather"`
	path          string
}

func (pl *ScriptFile) GetScripts() map[string]*Script {
	out := pl.GenerateTemplateScripts()
	for k, v := range pl.Scripts {
		out[k] = v
	}
	for k, v := range pl.GenerateScatterGatherScripts() {
		out[k] = v
	}
	return out
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (pl *ScriptFile) DoPrep() error {
	planPath, _ := filepath.Abs(pl.path)
	planDir := filepath.Dir(planPath)
	for _, s := range pl.Prep {
		log.Printf("Running script %s", s)

		if s.IfMissing == "" || !exists(filepath.Join(planDir, s.IfMissing)) {
			err := pl.runScript(s.CommandLine)
			if err != nil {
				log.Printf("Scripting error: %s", err)
				return err
			}
		} else {
			log.Printf("Skipping prep because file %s exists", s.IfMissing)
		}
	}
	return nil

}

func (pl *ScriptFile) runScript(command string) error {
	planPath, _ := filepath.Abs(pl.path)
	workdir := filepath.Dir(planPath)
	cmdLine, err := shlex.Split(command)
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
	cmd.Dir = workdir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	log.Printf("(%s) %s %s", cmd.Dir, cmd.Path, strings.Join(cmd.Args, " "))
	return cmd.Run()
}

func (sc *Script) GetCommand() string {
	inputs := map[string]string{}
	outputs := map[string]string{}
	for k := range sc.Inputs {
		inputs[k] = fmt.Sprintf("{input.%s}", k)
	}
	for k := range sc.Outputs {
		outputs[k] = fmt.Sprintf("{output.%s}", k)
	}
	cmd, _ := goatee.RenderString(sc.CommandLine, map[string]any{
		"inputs": inputs, "outputs": outputs,
	})
	return cmd
}

func (pl *ScriptFile) GetCollections() []CollectData {
	return pl.Collections
}

func (cl CollectData) GetOutputPath() string {
	path := filepath.Join(filepath.Dir(cl.path), cl.Output)
	npath, _ := filepath.Abs(path)
	return npath
}

func (sc *Script) GetInputs() map[string]string {
	o := map[string]string{}
	//if sc.scatterName != "" {
	//	path := filepath.Join(filepath.Dir(sc.path), "{scatteritem}")
	//	o = append(o, path)
	//} else {
	for k, p := range sc.Inputs {
		path := filepath.Join(filepath.Dir(sc.path), p)
		npath, _ := filepath.Abs(path)
		o[k] = npath
	}
	//}
	return o
}

func (sc *Script) GetOutputs() map[string]string {

	o := map[string]string{}
	for k, p := range sc.Outputs {
		path := filepath.Join(filepath.Dir(sc.path), p)
		npath, _ := filepath.Abs(path)
		fmt.Printf("output: %s\n", npath)
		o[k] = npath
	}
	return o
}

func (sc *Script) GetWorkdir() string {
	f, _ := filepath.Abs(sc.path)
	return filepath.Dir(f)
}

func (sc *Script) GetScatterName() string {
	return sc.scatterName
}

func (sc *Script) GetScatterCount() int {
	return sc.scatterCount
}

func (pl *ScriptFile) GenerateTemplateScripts() map[string]*Script {
	out := map[string]*Script{}

	for tName, t := range pl.Templates {
		outRender, err := goatee.Render(t.Scripts, map[string]any{"inputs": t.Inputs})
		if err == nil {
			js, _ := json.MarshalIndent(outRender, "", "  ")
			fmt.Printf("Template Render:\n%s\n", js)
			outScript := map[string]*Script{}
			err := json.Unmarshal(js, &outScript)
			if err == nil {
				for k, v := range outScript {
					name := fmt.Sprintf("%s_%s_%s", pl.Name, tName, k)
					v.path = pl.path
					fmt.Printf("step %s = %#v\n", name, v)
					out[name] = v
				}
			} else {
				fmt.Printf("Error: %s\n", err)
			}
		} else {
			fmt.Printf("Error: %s\n", err)
		}
	}

	/*
		for tName, t := range pl.Templates {
			for num, inputs := range t.Inputs {
				for k, v := range t.Scripts {
					o := Script{}
					name := fmt.Sprintf("%s_%s_%s_%d", pl.Name, tName, k, num)
					o.name = name
					o.path = v.path
					o.MemMB = v.MemMB
					o.CommandLine, _ = raymond.Render(v.CommandLine, inputs)
					o.Inputs = make([]string, len(v.Inputs))
					for i := range v.Inputs {
						p, _ := raymond.Render(v.Inputs[i], inputs)
						o.Inputs[i] = p
					}
					o.Outputs = make([]string, len(v.Outputs))
					for i := range v.Outputs {
						p, _ := raymond.Render(v.Outputs[i], inputs)
						o.Outputs[i] = p
					}
					out[name] = &o
				}
			}
		}
	*/
	return out
}

func (pl *ScriptFile) GenerateScatterGatherScripts() map[string]*Script {
	out := map[string]*Script{}

	for k, v := range pl.ScatterGather {
		scatterName := fmt.Sprintf("%s_%s_scatter", pl.Name, k)

		scatterCmdLine, err := goatee.Render(v.CommandLine,
			map[string]any{
				"threads": "{threads}",
				"input":   "{input}",
				"shard":   "{wildcards.shard}",
				"total":   "{wildcards.total}",
			},
		)
		if err != nil {
			log.Printf("Template Error: %s", err)
		}

		shardPrefix := "./shards/"
		if v.ShardPath != "" {
			s, err := goatee.Render(v.ShardPath, map[string]any{"shard": "{shard}"})
			if err == nil {
				v.ShardPath = s.(string)
			}
		}
		shardOutputName := fmt.Sprintf("%s{shard}-of-{total}", shardPrefix)
		scatterScript := &Script{
			CommandLine: scatterCmdLine.(string),
			Inputs:      map[string]string{"input": v.Input},
			Outputs:     map[string]string{"output": shardOutputName},
			MemMB:       v.MemMB,
			path:        pl.path,
			//Workdir     string   `json:"workdir"`
			//Order       int      `json:"order"`
		}

		gatherInput := fmt.Sprintf("%s{scatteritem}", shardPrefix)
		gatherName := fmt.Sprintf("%s_%s_gather", pl.Name, k)
		gatherScript := &Script{
			CommandLine:  "cat {input} > {output}",
			Inputs:       map[string]string{"input": gatherInput},
			Outputs:      map[string]string{"output": v.Output},
			path:         pl.path,
			scatterName:  fmt.Sprintf("%s_%s", pl.Name, k),
			scatterCount: v.Shards,
			//Workdir     string   `json:"workdir"`
			//Order       int      `json:"order"`
		}

		out[scatterName] = scatterScript
		out[gatherName] = gatherScript
	}
	return out
}

/*

rule intermediate:
    input:
        recordfile="../output/chembl/chemblTransform.records.compound.json.gz"
    output:
        "shards/{shard}-of-{total}.gz"
    threads: 8
    resources:
        mem_mb=30000,
        runtime=800
    shell:
        "./compound_distance.py -n {threads} -i {input.recordfile} -s {wildcards.shard} -t {wildcards.total} -o {output}"


rule gather:
    input:
        gather.split("shards/{scatteritem}.gz")
    output:
        "gathered/all.gz"
    shell:
        "cat {input} > {output}"
*/
