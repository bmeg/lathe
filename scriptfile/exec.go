package scriptfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

type ProcessDesc struct {
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
	out := &ProcessDesc{}

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

func (wd *WorkflowDesc) Add(x *ProcessDesc) {
	if x.Name == "" {
		x.Name = fmt.Sprintf("%s:%d", wd.Name, len(wd.Processes))
	}
	wd.Processes = append(wd.Processes, x)
}

func (pl *Plan) Workflow(name string) *WorkflowDesc {
	if pl.Verbose {
		fmt.Printf("Workflow\n")
	}
	w := &WorkflowDesc{Name: name}
	pl.Workflows[name] = w
	return w
}

func (pl *Plan) Print(x any) {
	fmt.Printf("%s", x)
}

type Plan struct {
	Workflows map[string]*WorkflowDesc
	Verbose   bool
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

	pl := &Plan{Workflows: map[string]*WorkflowDesc{}}

	vm := goja.New()
	vm.Set("Process", pl.Process)
	vm.Set("Params", map[string]string{
		"mode": "prep",
	})
	vm.Set("Workflow", pl.Workflow)
	vm.Set("print", pl.Print)

	_, err = vm.RunScript("main", string(source))
	if err != nil {
		return nil, err
	}
	fmt.Printf("%#v\n", pl.Workflows)
	return pl.Workflows, nil
}
