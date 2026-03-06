package workflow

import (
	"fmt"
	"os"
	"time"

	"github.com/bmeg/flame"
	"github.com/bmeg/lathe/jflow"
	"github.com/bmeg/lathe/logger"
	"github.com/bmeg/lathe/runner"
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
	Desc     *jflow.ProcessDesc
	Workflow *Workflow
}

func NewWorkflowProcess(wf *Workflow, baseDir string, desc *jflow.ProcessDesc) *WorkflowProcess {
	return &WorkflowProcess{BaseDir: baseDir, Desc: desc, Workflow: wf}
}

func (ws *WorkflowProcess) Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
	logger.Info("Process", "name", ws.Desc.Name)

	// Normalize legacy fields from TES format if needed
	dryRun := false
	for _, i := range status {
		if i.Status != STATUS_OK {
			logger.Info("Received upstream FAIL, skipping", "name", ws.Desc.Name)
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

	// Convert TES Input/Output arrays to maps for template rendering
	for i, input := range ws.Desc.Inputs {
		name := input.Name
		if name == "" {
			name = fmt.Sprintf("input_%d", i)
		}
		cmdInputs[name] = input.Path
	}

	for i, output := range ws.Desc.Outputs {
		name := output.Name
		if name == "" {
			name = fmt.Sprintf("output_%d", i)
		}
		cmdOutputs[name] = output.Path
	}

	cmdParams := map[string]any{
		"inputs":  cmdInputs,
		"outputs": cmdOutputs,
	}

	cmdLine := []string{}
	output.Status = STATUS_OK

	// Build command from executors (TES format)
	if len(ws.Desc.Executors) > 0 {
		// For now, we use the first executor. In the future, we may support sequential execution
		executor := ws.Desc.Executors[0]

		// Handle polymorphic command: can be string or array
		if len(executor.Command) > 0 {
			cmdLine = executor.Command
		}
	}

	if output.Status != STATUS_FAIL && len(cmdLine) == 0 {
		logger.Error("No command specified in executors")
		output.Status = STATUS_FAIL
	}

	if output.Status != STATUS_FAIL {
		doRun := true
		if outputsFound == len(ws.GetOutputs()) {

			var outputDate time.Time
			for _, o := range ws.GetOutputs() {
				i, err := os.Stat(o.Abs())
				if err == nil {
					if i.ModTime().After(outputDate) {
						outputDate = i.ModTime()
					}
				}
			}

			var inputDate time.Time
			for _, o := range ws.GetInputs() {
				i, err := os.Stat(o.Abs())
				if err == nil {
					if i.ModTime().After(inputDate) {
						inputDate = i.ModTime()
					}
				}
			}
			if outputDate.Before(inputDate) {
				logger.Info("Output files outdated, running command", "inputDate", inputDate, "outputDate", outputDate, "outputsRequired", ws.GetOutputs(), "commandLine", cmdLine)
			} else {
				logger.Info("Skipping command", "outputsFound", outputsFound, "outputsRequired", ws.GetOutputs(), "commandLine", cmdLine)
				output.Status = STATUS_OK
				doRun = false
			}
		}
		if doRun {
			if !dryRun {
				//fmt.Printf("Running command: %s missing outputs: (%s)\n", cmdLine, strings.Join(notFound, ","))
				inputs := []string{}
				outputs := []string{}
				for _, input := range ws.Desc.Inputs {
					inputs = append(inputs, input.Path)
				}
				for _, output := range ws.Desc.Outputs {
					outputs = append(outputs, output.Path)
				}

				// Get resource requirements
				cpus := uint(1)
				memMB := uint(1024)
				image := ""
				if ws.Desc.Resources != nil {
					cpus = ws.Desc.Resources.CPUCores
					if ws.Desc.Resources.RamGB > 0 {
						memMB = uint(ws.Desc.Resources.RamGB * 1024)
					}
				}
				if len(ws.Desc.Executors) > 0 {
					image = ws.Desc.Executors[0].Image
				}

				toolCmd := runner.CommandLineTool{
					CommandLine: cmdLine,
					BaseDir:     ws.BaseDir,
					MemMB:       memMB,
					NCpus:       cpus,
					Image:       image,
					Inputs:      inputs,
					Outputs:     outputs,
				}
				_, err := ws.Workflow.Runner.RunCommand(&toolCmd)
				if err == nil {
					for k, v := range ws.GetOutputs() {
						if !PathExists(v.Abs()) {
							logger.Error("Missing output", "commandLine", fmt.Sprint(cmdLine), "name", k, "path", v.Abs())
							output.Status = STATUS_FAIL
							logger.AddSummaryError("Missing output", "commandLine", fmt.Sprint(cmdLine), "name", k, "path", v.Abs())
						}
					}
					if output.Status == STATUS_OK {
						logger.Info("Command suceeded", "commandLine", fmt.Sprint(cmdLine))
					}
				} else {
					output.Status = STATUS_FAIL
					logger.AddSummaryError("CommandFailed", "commandLine", cmdLine)
					//The command failed, so outputs might be partially completed. Delete them for safety
					//TODO: setup command line option to turn this off
					for _, i := range ws.GetOutputs() {
						if IsFile(i.Abs()) {
							os.Remove(i.Abs())
						}
					}
				}
			} else {
				logger.Info("Would run command: %s %#v\n", cmdLine, cmdParams)
				output.Status = STATUS_OK
			}
		}
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
	for i, input := range ws.Desc.Inputs {
		name := input.Name
		if name == "" {
			name = fmt.Sprintf("input_%d", i)
		}
		out[name] = DataFile{BaseDir: ws.BaseDir, RelPath: input.Path}
	}
	return out
}

func (ws *WorkflowProcess) GetOutputs() map[string]DataFile {
	out := map[string]DataFile{}
	for i, output := range ws.Desc.Outputs {
		name := output.Name
		if name == "" {
			name = fmt.Sprintf("output_%d", i)
		}
		out[name] = DataFile{BaseDir: ws.BaseDir, RelPath: output.Path}
	}
	return out
}

func (ws *WorkflowProcess) GetDesc() string {
	if len(ws.Desc.Executors) > 0 {
		executor := ws.Desc.Executors[0]
		if len(executor.Command) > 0 {
			return fmt.Sprintf("run: %s", executor.Command)
		}
	}
	return fmt.Sprintf("run: %s", ws.Desc.Name)
}
