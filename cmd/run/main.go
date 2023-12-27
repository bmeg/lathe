package run

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/bmeg/flame"
	"github.com/bmeg/lathe/scriptfile"
	"github.com/spf13/cobra"
)

const (
	STATUS_OK   = 0
	STATUS_FAIL = 1
)

type WorkflowStatus struct {
	Name   string
	Status int
}

type DataFile struct {
	BaseDir string
	RelPath string
}

func (df *DataFile) Abs() string {
	s, _ := filepath.Abs(filepath.Join(df.BaseDir, df.RelPath))
	return s
}

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
	for _, i := range status {
		if i.Status != STATUS_OK {
			return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: i}
		}
	}
	output := &WorkflowStatus{}
	outputsFound := 0
	for _, o := range ws.GetOutputs() {
		if PathExists(o.Abs()) {
			outputsFound++
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
			fmt.Printf("Skipping command (%d of %d outputs found): %s\n", outputsFound, len(ws.GetOutputs()), cmdLine)
			output.Status = STATUS_OK
		} else {
			fmt.Printf("Running command: %s\n", cmdLine)
			toolCmd := CommandLineTool{
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
		}
	} else {
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

/*****/

type Workflow struct {
	Steps  map[string]WorkflowStep
	DepMap map[string][]string

	Runner CommandRunner
}

func (w *Workflow) AddStep(ws WorkflowStep) error {
	n := ws.GetName()
	if _, ok := w.Steps[n]; ok {
		return fmt.Errorf("non-unique workflow step name: %s", n)
	}
	w.Steps[n] = ws
	return nil
}

func (w *Workflow) AddDepends(step WorkflowStep, dep WorkflowStep) error {
	stepName := step.GetName()
	depName := dep.GetName()
	if x, ok := w.DepMap[stepName]; ok {
		w.DepMap[stepName] = append(x, depName)
	} else {
		w.DepMap[stepName] = []string{depName}
	}
	return nil
}

/*****/

type WorkflowFileCheck struct {
	File DataFile
}

func (ws *WorkflowFileCheck) Process(key string, status []*WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
	for _, i := range status {
		if i.Status != STATUS_OK {
			return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: i}
		}
	}
	output := &WorkflowStatus{}
	fmt.Printf("Checking for file: %s\n", ws.File.Abs())
	if !PathExists(ws.File.Abs()) {
		output.Status = STATUS_FAIL
	} else {
		output.Status = STATUS_OK
	}
	return flame.KeyValue[string, *WorkflowStatus]{Key: key, Value: output}
}

func (ws *WorkflowFileCheck) IsGenerator() bool {
	return true
}

func (ws *WorkflowFileCheck) GetInputs() map[string]DataFile {
	out := map[string]DataFile{}
	return out
}

func (ws *WorkflowFileCheck) GetOutputs() map[string]DataFile {
	out := map[string]DataFile{}
	out["file"] = ws.File
	return out
}

func (ws *WorkflowFileCheck) GetName() string {
	return ws.File.Abs()
}

func (ws *WorkflowFileCheck) GetDesc() string {
	return fmt.Sprintf("check-file: %s", ws.File.Abs())
}

/*****/

func prepWorkflow(basedir string, wd *scriptfile.WorkflowDesc) (*Workflow, error) {

	wf := &Workflow{Steps: map[string]WorkflowStep{}, DepMap: make(map[string][]string), Runner: NewSingleMachineRunner(16, 32000)}

	//map inputs and outputs
	inFileMap := map[string]WorkflowStep{}
	outFileMap := map[string]WorkflowStep{}
	for _, p := range wd.Processes {
		ws := NewWorkflowProcess(wf, basedir, p)
		if err := wf.AddStep(ws); err != nil {
			fmt.Printf("error: %s\n", err)
		}
		for _, path := range ws.GetInputs() {
			inFileMap[path.Abs()] = ws
		}
		for _, path := range ws.GetOutputs() {
			outFileMap[path.Abs()] = ws
		}
	}

	//connect inputs to existing outputs
	for _, p := range wf.Steps {
		for _, path := range p.GetInputs() {
			if inS, ok := outFileMap[path.Abs()]; ok {
				wf.AddDepends(p, inS)
			}
		}
	}

	//Identify input that map to existing files
	fileSteps := map[string]WorkflowStep{}
	for i, p := range wf.Steps {
		ready := true
		curInputs := wf.DepMap[p.GetName()]
		if len(curInputs) != len(p.GetInputs()) {
			for k, v := range p.GetInputs() {
				if _, ok := wf.DepMap[k]; !ok {
					inPath := v.Abs()
					if PathExists(inPath) {
						//fmt.Printf("Found %s\n", inPath)
						file := v
						fileSteps[v.Abs()] = &WorkflowFileCheck{file}
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
			fmt.Printf("Cannot find inputs for step: %s\n", i)
		}
	}

	for _, s := range fileSteps {
		if err := wf.AddStep(s); err != nil {
			fmt.Printf("error: %s\n", err)
		}
	}

	return wf, nil
}

type FlameWorkflow struct {
	Workflow  *flame.Workflow
	ProcessIn chan *WorkflowStatus
}

func (wf *Workflow) BuildFlame() (*FlameWorkflow, error) {
	out := flame.NewWorkflow()

	nodeMap := map[WorkflowStep]flame.Emitter[flame.KeyValue[string, *WorkflowStatus]]{}

	workChan := make(chan *WorkflowStatus, 10)
	startNode := flame.AddSourceChan(out, workChan)
	for _, v := range wf.Steps {
		if v.IsGenerator() {
			curV := v
			//fmt.Printf("Starting Node: %s %s\n", k, v.GetDesc())
			m := flame.AddMapper(out, func(x *WorkflowStatus) flame.KeyValue[string, *WorkflowStatus] {
				return curV.Process(x.Name, []*WorkflowStatus{x})
			})
			m.Connect(startNode)
			nodeMap[v] = m
		}
	}

	for found := true; found; {
		found = false
		fmt.Printf("Checking Steps\n")
		for _, v := range wf.Steps {
			if _, ok := nodeMap[v]; !ok {
				inNodes := []flame.Emitter[flame.KeyValue[string, *WorkflowStatus]]{}
				for _, dep := range wf.DepMap[v.GetName()] {
					depStep := wf.Steps[dep]
					if n, ok := nodeMap[depStep]; ok {
						inNodes = append(inNodes, n)
					}
				}
				if len(inNodes) == len(wf.DepMap[v.GetName()]) {
					curV := v
					//fmt.Printf("Found dependancy: %s\n", curV.GetDesc())
					j := flame.AddKeyJoinGroupAsync(out, func(key string, status []*WorkflowStatus) *WorkflowStatus {
						return curV.Process(key, status).Value
					})
					for _, i := range inNodes {
						j.Connect(i)
					}
					nodeMap[v] = j
					found = true
				}
			}
		}
	}

	return &FlameWorkflow{Workflow: out, ProcessIn: workChan}, nil
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
					fwf, err := wf.BuildFlame()
					if err != nil {
						fmt.Printf("workflow build error: %s\n", err)
					}
					fmt.Printf("%#v\n", fwf)

					go func() {
						fwf.ProcessIn <- &WorkflowStatus{Name: "run"}
						close(fwf.ProcessIn)
					}()

					fwf.Workflow.Start()

					fwf.Workflow.Wait()
					fmt.Printf("Workflow Done\n")
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
