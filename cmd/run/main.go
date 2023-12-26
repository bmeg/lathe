package run

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bmeg/flame"
	"github.com/bmeg/lathe/scriptfile"
	"github.com/spf13/cobra"
)

type WorkflowStatus struct {
}

type StepInputMap map[string]WorkflowStep

type StepOutputMap map[string]string

type WorkflowStep interface {
	IsGenerator() bool
	Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus]
}

/*****/

type WorkflowProcess struct {
	Desc *scriptfile.ProcessDesc
}

func (ws *WorkflowProcess) Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
	output := &WorkflowStatus{}
	return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: output}
}

func (ws *WorkflowProcess) IsGenerator() bool {
	if len(ws.Desc.Inputs) == 0 {
		return true
	}
	return false
}

/*****/

type WorkflowFileCheck struct {
	Path string
}

func (ws *WorkflowFileCheck) Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
	output := &WorkflowStatus{}
	return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: output}
}

func (ws *WorkflowFileCheck) IsGenerator() bool {
	return true
}

/*****/

func prepWorkflow(basedir string, wd *scriptfile.WorkflowDesc) (*flame.Workflow, error) {
	wf := flame.NewWorkflow()

	workflowSteps := map[int]WorkflowStep{}

	inFileMap := map[string]WorkflowStep{}
	outFileMap := map[string]WorkflowStep{}
	for i, p := range wd.Processes {
		ws := &WorkflowProcess{p}
		workflowSteps[i] = ws
		for _, path := range p.Inputs {
			fPath, _ := filepath.Abs(filepath.Join(basedir, path))
			inFileMap[fPath] = ws
		}
		for _, path := range p.Outputs {
			fPath, _ := filepath.Abs(filepath.Join(basedir, path))
			outFileMap[fPath] = ws
		}
		//name := fmt.Sprintf("job_%d", i)
		//raymond.Render( p.CommandLine )
		//fmt.Printf("%s: %s\n", name, p.CommandLine)
		flame.AddKeyJoinGroupAsync(wf, ws.Process)
	}

	stepInputs := map[WorkflowStep]StepInputMap{}
	for i, p := range wd.Processes {
		stepInputs[workflowSteps[i]] = StepInputMap{}
		for name, path := range p.Inputs {
			fPath, _ := filepath.Abs(filepath.Join(basedir, path))
			if inS, ok := outFileMap[fPath]; ok {
				stepInputs[workflowSteps[i]][name] = inS
			}
		}
	}

	for i, p := range wd.Processes {
		ready := true
		curStep := workflowSteps[i]
		curInputs := stepInputs[curStep]
		if len(curInputs) != len(p.Inputs) {
			for k, v := range p.Inputs {
				if _, ok := curInputs[k]; !ok {
					inPath, _ := filepath.Abs(filepath.Join(basedir, v))
					if PathExists(inPath) {
						fmt.Printf("Found %s\n", inPath)
						curInputs[k] = &WorkflowFileCheck{inPath}
					} else {
						fmt.Printf("Missing %s: %s\n", k, v)
						ready = false
					}
				}
			}
		}
		if ready {
			//fmt.Printf("Ready: %#v\t%#v\n", stepInputs[workflowSteps[i]], stepOutputs[workflowSteps[i]])
		} else {
			//fmt.Printf("Not Ready: %#v\t%#v\n", stepInputs[workflowSteps[i]], stepOutputs[workflowSteps[i]])
			fmt.Printf("Cannot find inputs for step: %d\n", i)
		}
	}

	for i, p := range wd.Processes {
		curStep := workflowSteps[i]
		curInputs := stepInputs[curStep]

		isReady := true
		for _, k := range curInputs {
			if !k.IsGenerator() {
				isReady = false
			}
		}
		if isReady {
			fmt.Printf("ready: %#v\n", p)
		}

	}
	//for i, p := range wd.Processes {
	//}

	//for i := range wd.Processes {
	//	fmt.Printf("cmd: %s\n", wd.Processes[i].CommandLine)
	//	fmt.Printf("%d %#v\n", i, stepInputs[workflowSteps[i]])
	//}

	return wf, nil
}

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "run <plan file>",
	Short: "Scan directory to plan operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptPath, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		baseDir := filepath.Dir(scriptPath)
		names := []string{}
		if len(args) > 1 {
			names = args[1:]
		}
		workflows, err := scriptfile.RunFile(scriptPath)
		if err != nil {
			log.Printf("Script Error: %s\n", err)
			return err
		}

		for _, n := range names {
			if wfd, ok := workflows[n]; ok {
				wf, err := prepWorkflow(baseDir, wfd)
				if err == nil {
					//fmt.Printf("Running Workflow: %#v\n", wf)
					_ = wf
				}
			} else {
				fmt.Printf("Workflow %s not found\n", n)
			}
		}
		return nil
	},
}

func init() {
	//flags := Cmd.Flags()
}
