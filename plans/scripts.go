package plans

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/google/shlex"
)

type Script struct {
	CommandLine string   `json:"commandLine"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	Workdir     string   `json:"workdir"`
	Order       int      `json:"order"`
	path        string
	name        string
}

type PrepStage struct {
	IfMissing   string `json:"ifMissing"`
	CommandLine string `json:"commandLine"`
}

type Template struct {
	Inputs  []map[string]any   `json:"inputs"`
	Scripts map[string]*Script `json:"scripts"`
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
		for num, inputs := range t.Inputs {
			for k, v := range t.Scripts {
				o := Script{}
				o.name = v.name
				o.path = v.path
				o.CommandLine, _ = raymond.Render(v.CommandLine, inputs)
				o.Inputs = make([]string, len(v.Inputs))
				for i := range v.Inputs {
					p, _ := raymond.Render(v.Inputs[i], inputs)
					o.Inputs[i] = p
					//path := filepath.Join(filepath.Dir(pl.path), p)
					//npath, _ := filepath.Abs(path)
					//o.Inputs[i] = npath
				}
				o.Outputs = make([]string, len(v.Outputs))
				for i := range v.Outputs {
					p, _ := raymond.Render(v.Outputs[i], inputs)
					o.Outputs[i] = p
					//path := filepath.Join(filepath.Dir(pl.path), p)
					//npath, _ := filepath.Abs(path)
					//o.Outputs[i] = npath
					//fmt.Printf("output format: %s %s\n", pl.path, npath)
				}
				name := fmt.Sprintf("%s_%s_%s_%d", pl.Name, tName, k, num)
				out[name] = &o
			}
		}
	}
	return out
}
