package scriptfile

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/google/shlex"
)

type ProcessDesc struct {
	BasePath    string
	Name        string
	Desc        map[string]any
	CommandLine string
	Inputs      map[string]string
	Outputs     map[string]string
	MemMB       uint
	NCpus       uint
}

func (pl *Plan) Process(data map[string]any) *ProcessDesc {
	if pl.Verbose {
		fmt.Printf("Process: %#v\n", data)
	}
	out := &ProcessDesc{BasePath: filepath.Dir(pl.Path)}

	out.Desc = data
	//out.Dependencies = []*ProcessDesc{}
	out.Inputs = map[string]string{}
	out.Outputs = map[string]string{}

	if cmd, ok := data["commandLine"]; ok {
		if cmdStr, ok := cmd.(string); ok {
			out.CommandLine = cmdStr
		}
	}

	if inputs, ok := data["inputs"]; ok {
		if inputsMap, ok := inputs.(map[string]any); ok {
			for k, v := range inputsMap {
				if vStr, ok := v.(string); ok {
					out.Inputs[k] = vStr
				}
			}
		}
	}

	if outputs, ok := data["outputs"]; ok {
		if outputMap, ok := outputs.(map[string]any); ok {
			for k, v := range outputMap {
				if vStr, ok := v.(string); ok {
					out.Outputs[k] = vStr
				}
			}
		}
	}

	out.MemMB = 1024
	if memMb, ok := data["memMB"]; ok {
		if memMbInt, ok := memMb.(int); ok {
			out.MemMB = uint(memMbInt)
		} else if memMbInt, ok := memMb.(int64); ok {
			out.MemMB = uint(memMbInt)
		}
	}

	out.NCpus = 1
	if ncpus, ok := data["ncpus"]; ok {
		if ncpusInt, ok := ncpus.(int); ok {
			out.NCpus = uint(ncpusInt)
		} else if ncpusInt, ok := ncpus.(int64); ok {
			out.NCpus = uint(ncpusInt)
		}
	}

	if name, ok := data["name"]; ok {
		if nameStr, ok := name.(string); ok {
			out.Name = nameStr
		}
	}

	return out
}

type WorkflowDesc struct {
	Name      string
	Processes []*ProcessDesc
}

//func (pd *ProcessDesc) Depends(p *ProcessDesc) {
//fmt.Printf("Adding process dependency: %#v", pd)
//	pd.Dependencies = append(pd.Dependencies, p)
//}

func (wd *WorkflowDesc) Add(call goja.ConstructorCall) *goja.Object {
	if len(call.Arguments) != 1 {
		return nil
	}
	//fmt.Printf("Adding %#v\n", call.Arguments[0])
	e := call.Arguments[0].Export()
	if proc, ok := e.(*ProcessDesc); ok {
		if proc.Name == "" {
			proc.Name = fmt.Sprintf("%s:%d", wd.Name, len(wd.Processes))
		}
		wd.Processes = append(wd.Processes, proc)
	} else if wf, ok := e.(*WorkflowDesc); ok {
		wd.Processes = append(wd.Processes, wf.Processes...)
	} else {
		fmt.Printf("Unknown object: %#v\n", e)
	}
	return nil
}

func (pl *Plan) Workflow(name string) *WorkflowDesc {
	if pl.Verbose {
		fmt.Printf("Workflow\n")
	}
	w := &WorkflowDesc{Name: fmt.Sprintf("%s:%s", pl.Path, name)}
	pl.Workflows[name] = w
	return w
}

func (pl *Plan) Print(x any) {
	fmt.Printf("%s", x)
}

func (pl *Plan) Println(x any) {
	fmt.Printf("%s\n", x)
}

type Plan struct {
	Workflows map[string]*WorkflowDesc
	Verbose   bool
	Path      string
	VM        *goja.Runtime
}

func (pl *Plan) Glob(pattern string) []string {
	gp := filepath.Join(filepath.Dir(pl.Path), pattern)
	matches, _ := filepath.Glob(gp)
	return matches
}

func (pl *Plan) LoadPlan(path string) map[string]*WorkflowDesc {
	fmt.Printf("Loading sub-workflow %s\n", path)
	if x, err := RunFile(path); err == nil {
		return x
	} else {
		fmt.Printf("Error Loading sub-workflow %s : %s\n", path, err)
	}
	return map[string]*WorkflowDesc{}
}

func (pl *Plan) Plugin(cmdLine string) goja.Value {

	cmdArgs, err := shlex.Split(cmdLine)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return nil
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = filepath.Dir(pl.Path)
	stdout, _ := cmd.StdoutPipe()

	go cmd.Run()

	data, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	//fmt.Printf("Plugin output: %s\n", data)

	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err == nil {
		return pl.VM.ToValue(m)
	}
	a := []any{}
	if err := json.Unmarshal(data, &a); err == nil {
		return pl.VM.ToValue(a)
	}
	return nil
}

func RunFile(path string) (map[string]*WorkflowDesc, error) {

	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(path)
	if abserr != nil {
		return nil, abserr
	}

	// Read file
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config at path %s: \n%v", path, err)
	}

	vm := goja.New()

	pl := &Plan{Workflows: map[string]*WorkflowDesc{}, Path: path, VM: vm}

	latheObj := map[string]any{
		"Params": map[string]string{
			"mode": "prep",
		},
		"Workflow": pl.Workflow,
		"LoadPlan": pl.LoadPlan,
		"Process":  pl.Process,
		"Plugin":   pl.Plugin,
	}

	vm.Set("print", pl.Print)
	vm.Set("println", pl.Println)
	vm.Set("glob", pl.Glob)
	vm.Set("lathe", latheObj)

	_, err = vm.RunScript("main", string(source))
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%#v\n", pl.Workflows)
	return pl.Workflows, nil
}
