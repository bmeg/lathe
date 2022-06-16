package plans

import (
	"fmt"
	"path/filepath"
)

type Script struct {
	Class       string   `json:"class"`
	CommandLine string   `json:"commandLine"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
	Workdir     string   `json:"workdir"`
	Order       int      `json:"order"`
	path        string
	name        string
}

type Plan struct {
	Name    string             `json:"name"`
	Scripts map[string]*Script `json:"scripts"`
	path    string
}

func (pl *Plan) GetScripts() map[string]*Script {
	return pl.Scripts
}

/*
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
*/

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
