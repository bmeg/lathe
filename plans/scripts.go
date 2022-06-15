package plans

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmeg/sifter/task"
	"github.com/google/shlex"
)

type Script struct {
	CommandLine string   `json:"commandLine"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	Workdir     string   `json:"workdir"`
	Order       int      `json:"order"`
}

type Plan struct {
	Name    string            `json:"name"`
	Scripts map[string]Script `json:"scripts"`
	path    string
}

func (pl *Plan) Execute() error {
	scripts := []string{}
	for k := range pl.Scripts {
		scripts = append(scripts, k)
	}

	sort.Slice(scripts, func(x, y int) bool { return pl.Scripts[scripts[x]].Order < pl.Scripts[scripts[y]].Order })

	for _, s := range scripts {
		log.Printf("Running script %s", s)
		err := pl.RunScript(s)
		if err != nil {
			log.Printf("Scripting error: %s", err)
			return err
		}
	}
	return nil
}

func (pl *Plan) RunScript(name string) error {
	if sc, ok := pl.Scripts[name]; ok {
		path, _ := filepath.Abs(pl.path)
		workdir := filepath.Join(filepath.Dir(path), sc.Workdir)
		cmdLine, err := shlex.Split(sc.CommandLine)
		if err != nil {
			return err
		}
		cmd := exec.Command(cmdLine[0], cmdLine[1:len(cmdLine)]...)
		cmd.Dir = workdir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		log.Printf("(%s) %s %s", cmd.Dir, cmd.Path, strings.Join(cmd.Args, " "))
		return cmd.Run()
	}
	return fmt.Errorf("Script %s not found", name)
}

func (pb *Plan) GetScriptInputs(task task.RuntimeTask) map[string][]string {
	out := map[string][]string{}
	for k, v := range pb.Scripts {
		o := []string{}
		for _, p := range v.Inputs {
			path := filepath.Join(filepath.Dir(pb.path), p)
			npath, _ := filepath.Abs(path)
			o = append(o, npath)
		}
		out[k] = o
	}
	return out
}

func (pb *Plan) GetScriptOutputs(task task.RuntimeTask) map[string][]string {
	out := map[string][]string{}
	for k, v := range pb.Scripts {
		o := []string{}
		for _, p := range v.Outputs {
			path := filepath.Join(filepath.Dir(pb.path), p)
			npath, _ := filepath.Abs(path)
			o = append(o, npath)
		}
		out[k] = o
	}
	return out
}
