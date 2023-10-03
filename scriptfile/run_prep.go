package scriptfile

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmeg/goatee"
	"github.com/bmeg/lathe/util"
	"github.com/google/shlex"
)

func (pl *ScriptFile) DoPrep() error {
	planPath, _ := filepath.Abs(pl.path)
	planDir := filepath.Dir(planPath)

	for _, d := range pl.DockerImages {
		cmdLine := []string{"docker", "build", "-t", d.Name, d.Dir}
		log.Printf("Docker Build: %s", cmdLine)

		cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
		cmd.Dir = planDir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		log.Printf("(%s) %s %s", cmd.Dir, cmd.Path, strings.Join(cmd.Args, " "))
		cmd.Run()
	}

	for _, s := range pl.Prep {
		if s.CommandLine != "" {
			doRun := true
			if len(s.Outputs) > 0 {
				doRun = false
				for _, v := range s.Outputs {
					iPath := filepath.Join(planDir, v)
					if !util.Exists(iPath) {
						log.Printf("File %s missing", iPath)
						doRun = true
					}
				}
			}

			if doRun {
				err := s.RunCommand()
				if err != nil {
					log.Printf("Scripting error: %s", err)
					return err
				}
			} else {
				cmd, _ := s.GetCommand()
				log.Printf("Skipping prep: %s", cmd)
			}
		} else {
			if s.Source != "" && s.Path != "" {
				log.Printf("Checking for %s", s.Path)
			}
		}
	}
	return nil

}

func (sc *PrepStage) GetCommand() (string, error) {
	inputs := map[string]string{}
	outputs := map[string]string{}

	for k, v := range sc.GetInputs() {
		inputs[k] = v
	}
	for k, v := range sc.GetOutputs() {
		outputs[k] = v
	}

	command, err := goatee.RenderString(sc.CommandLine, map[string]any{
		"inputs": inputs, "outputs": outputs,
	})

	return command, err
}

func (sc *PrepStage) RunCommand() error {
	planPath, _ := filepath.Abs(sc.path)
	workdir := filepath.Dir(planPath)

	command, err := sc.GetCommand()
	if err != nil {
		return err
	}

	log.Printf("Running prep command from %s : %s", sc.path, command)

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

func (sc *PrepStage) GetInputs() map[string]string {
	o := map[string]string{}
	for k, p := range sc.Inputs {
		path := filepath.Join(filepath.Dir(sc.path), p)
		npath, _ := filepath.Abs(path)
		o[k] = npath
	}
	return o
}

func (sc *PrepStage) GetOutputs() map[string]string {
	o := map[string]string{}
	for k, p := range sc.Outputs {
		path := filepath.Join(filepath.Dir(sc.path), p)
		npath, _ := filepath.Abs(path)
		o[k] = npath
	}
	return o
}
