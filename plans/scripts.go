package plans

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

type Plan struct {
	Class   string             `json:"class"`
	Name    string             `json:"name"`
	Scripts map[string]*Script `json:"scripts"`
	Prep    []PrepStage        `json:"prep"`
	path    string
}

func (pl *Plan) GetScripts() map[string]*Script {
	return pl.Scripts
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
