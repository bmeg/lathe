package plans

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
	CommandLine string   `json:"commandLine"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	Workdir     string   `json:"workdir"`
	Order       int      `json:"order"`
	MemMB       int      `json:"memMB"`
	path        string
	name        string
}

type PrepStage struct {
	IfMissing   string `json:"ifMissing"`
	CommandLine string `json:"commandLine"`
}

type Template struct {
	Inputs  []map[string]any `json:"inputs"`
	Scripts map[string]any   `json:"scripts"`
}

type Plan struct {
	Class     string               `json:"class"`
	Name      string               `json:"name"`
	Scripts   map[string]*Script   `json:"scripts"`
	Templates map[string]*Template `json:"templates"`
	Prep      []PrepStage          `json:"prep"`
	path      string
}

func (pl *Plan) GetScripts() map[string]*Script {
	out := pl.GenerateScripts()
	for k, v := range pl.Scripts {
		out[k] = v
	}
	return out
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (pl *Plan) DoPrep() error {
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

func (pl *Plan) runScript(command string) error {
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
	return sc.CommandLine
}

func (sc *Script) GetInputs() []string {
	o := []string{}
	for _, p := range sc.Inputs {
		path := filepath.Join(filepath.Dir(sc.path), p)
		npath, _ := filepath.Abs(path)
		o = append(o, npath)
	}
	return o
}

func (sc *Script) GetOutputs() []string {

	o := []string{}
	for _, p := range sc.Outputs {
		path := filepath.Join(filepath.Dir(sc.path), p)
		npath, _ := filepath.Abs(path)
		fmt.Printf("output: %s\n", npath)
		o = append(o, npath)
	}
	return o
}

func (sc *Script) GetWorkdir() string {
	f, _ := filepath.Abs(sc.path)
	return filepath.Dir(f)
}

func (pl *Plan) GenerateScripts() map[string]*Script {
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
