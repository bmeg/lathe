package workflow

import (
	"fmt"
	"log"

	"github.com/aymerick/raymond"
	"github.com/bmeg/flame"
	"github.com/bmeg/lathe/runner"
	"github.com/bmeg/lathe/scriptfile"
)

type WorkflowStep interface {
	GetName() string
	IsGenerator() bool
	Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus]
	GetInputs() map[string]DataFile
	GetOutputs() map[string]DataFile

	GetDesc() string
}

/*****/

type WorkflowProcess struct {
	BaseDir  string
	Desc     *scriptfile.ProcessDesc
	Workflow *Workflow
}

func NewWorkflowProcess(wf *Workflow, baseDir string, desc *scriptfile.ProcessDesc) *WorkflowProcess {
	return &WorkflowProcess{BaseDir: baseDir, Desc: desc, Workflow: wf}
}

func (ws *WorkflowProcess) Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
	log.Printf("Process %s\n", ws.Desc.Name)
	dryRun := false
	for _, i := range status {
		if i.Status != STATUS_OK {
			log.Printf("Received upstream FAIL, skipping: %s", ws.Desc.Name)
			return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: i}
		}
		if i.DryRun {
			dryRun = true
		}
	}
	output := &WorkflowStatus{DryRun: dryRun}
	outputsFound := 0
	notFound := []string{}
	for _, o := range ws.GetOutputs() {
		if PathExists(o.Abs()) {
			outputsFound++
		} else {
			notFound = append(notFound, o.RelPath)
		}
	}

	cmdInputs := map[string]any{}
	cmdOutputs := map[string]any{}

	for k, v := range ws.Desc.Inputs {
		cmdInputs[k] = v
	}

	for k, v := range ws.Desc.Outputs {
		cmdOutputs[k] = v
	}

	cmdParams := map[string]any{
		"inputs":  cmdInputs,
		"outputs": cmdOutputs,
	}

	cmdLine, err := raymond.Render(ws.Desc.CommandLine, cmdParams)
	if err == nil {
		if outputsFound == len(ws.GetOutputs()) {
			log.Printf("Skipping command (%d of %d outputs found): %s\n", outputsFound, len(ws.GetOutputs()), cmdLine)
			output.Status = STATUS_OK
		} else {
			if !dryRun {
				//fmt.Printf("Running command: %s missing outputs: (%s)\n", cmdLine, strings.Join(notFound, ","))
				toolCmd := runner.CommandLineTool{
					CommandLine: cmdLine,
					BaseDir:     ws.BaseDir,
					MemMB:       ws.Desc.MemMB,
					NCpus:       ws.Desc.NCpus,
				}
				_, err := ws.Workflow.Runner.RunCommand(&toolCmd)
				if err == nil {
					output.Status = STATUS_OK
				} else {
					output.Status = STATUS_FAIL
				}
			} else {
				log.Printf("Would run command: %s %#v\n", cmdLine, cmdParams)
				output.Status = STATUS_OK
			}
		}
	} else {
		log.Printf("Template error: %s\n", err)
		output.Status = STATUS_FAIL
	}
	return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: output}
}

func (ws *WorkflowProcess) GetName() string {
	return ws.Desc.Name
}

func (ws *WorkflowProcess) IsGenerator() bool {
	return len(ws.GetInputs()) == 0
}

func (ws *WorkflowProcess) GetInputs() map[string]DataFile {
	out := map[string]DataFile{}
	for k, v := range ws.Desc.Inputs {
		out[k] = DataFile{BaseDir: ws.BaseDir, RelPath: v}
	}
	return out
}

func (ws *WorkflowProcess) GetOutputs() map[string]DataFile {
	out := map[string]DataFile{}
	for k, v := range ws.Desc.Outputs {
		out[k] = DataFile{BaseDir: ws.BaseDir, RelPath: v}
	}
	return out
}

func (ws *WorkflowProcess) GetDesc() string {
	return fmt.Sprintf("run: %s", ws.Desc.CommandLine)
}
