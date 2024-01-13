package workflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmeg/flame"
	"github.com/bmeg/lathe/runner"
	"github.com/bmeg/lathe/scriptfile"
)

const (
	STATUS_OK   = 0
	STATUS_FAIL = 1
)

type WorkflowStatus struct {
	Name   string
	Status int
	DryRun bool
}

type DataFile struct {
	BaseDir string
	RelPath string
}

func (df *DataFile) Abs() string {
	if filepath.IsAbs(df.RelPath) {
		fmt.Printf("Is abs: %s\n", df.RelPath)
		return df.RelPath
	}
	s, _ := filepath.Abs(filepath.Join(df.BaseDir, df.RelPath))
	return s
}

/*****/

type Workflow struct {
	Steps  map[string]WorkflowStep
	DepMap map[string][]string

	Runner runner.CommandRunner
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

func PrepWorkflow(wd *scriptfile.WorkflowDesc) (*Workflow, error) {
	fmt.Printf("Building Workflow DAG\n")
	wf := &Workflow{
		Steps:  map[string]WorkflowStep{},
		DepMap: make(map[string][]string),
		Runner: runner.NewSingleMachineRunner(16, 32000),
	}

	//map inputs and outputs
	inFileMap := map[string]WorkflowStep{}
	outFileMap := map[string]WorkflowStep{}
	for _, p := range wd.Processes {
		ws := NewWorkflowProcess(wf, p.BasePath, p)
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

	//fmt.Printf("InfileMap: %#v\n", inFileMap)
	//fmt.Printf("OutfileMap: %#v\n", outFileMap)

	//connect inputs to existing outputs
	fileSteps := map[string]WorkflowStep{}
	for _, p := range wf.Steps {
		for _, path := range p.GetInputs() {
			if inS, ok := outFileMap[path.Abs()]; ok {
				wf.AddDepends(p, inS)
			} else {
				fmt.Printf("File Check: %s\n", path.Abs())
				inPath := path.Abs()
				if x, ok := fileSteps[inPath]; ok {
					wf.AddDepends(p, x)
				} else {
					lPath := path
					s := &WorkflowFileCheck{lPath}
					fileSteps[inPath] = s
					if err := wf.AddStep(s); err != nil {
						fmt.Printf("error: %s\n", err)
					}
					wf.AddDepends(p, s)
				}
			}
		}
	}

	return wf, nil
}

type FlameWorkflow struct {
	Workflow  *flame.Workflow
	ProcessIn chan *WorkflowStatus
}

func (wf *Workflow) BuildFlame() (*FlameWorkflow, error) {
	fmt.Printf("Converting DAG to op-flow\n")
	out := flame.NewWorkflow()

	nodeMap := map[WorkflowStep]flame.Emitter[flame.KeyValue[string, *WorkflowStatus]]{}

	//create root node of DAG
	workChan := make(chan *WorkflowStatus, 10)
	startNode := flame.AddSourceChan(out, workChan)

	//Connect elements that can run immediately with no dependencies to the root node
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

	//Connect steps where all dependencies have been added to the DAG
	//repeat until no new steps can be added
	for found := true; found; {
		found = false
		for _, v := range wf.Steps {
			if _, ok := nodeMap[v]; !ok {
				//fmt.Printf("Checking Step: %s\n", v.GetName())
				inNodes := []flame.Emitter[flame.KeyValue[string, *WorkflowStatus]]{}
				for _, dep := range wf.DepMap[v.GetName()] {
					depStep := wf.Steps[dep]
					if n, ok := nodeMap[depStep]; ok {
						inNodes = append(inNodes, n)
					}
				}
				if len(inNodes) == len(wf.DepMap[v.GetName()]) {
					fmt.Printf("Adding Step: %s (%s) depends: (%s) %#v\n", v.GetName(), v.GetDesc(), strings.Join(wf.DepMap[v.GetName()], ","), v.GetInputs())

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

	for _, v := range wf.Steps {
		if _, ok := nodeMap[v]; !ok {
			fmt.Printf("Step %s not added to graph\n", v.GetName())
		}
	}

	return &FlameWorkflow{Workflow: out, ProcessIn: workChan}, nil
}
