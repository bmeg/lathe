package scriptfile

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/google/shlex"
)

type Step interface {
	GetName() string
	GetBasePath() string
	GetInputs() map[string]string
	GetProcess() *ProcessDesc
}

type File struct {
	BasePath string
	Path     string
}

type Plan struct {
	Workflows map[string]*WorkflowDesc
	Verbose   bool
	Path      string
	VM        *goja.Runtime
	Images    []*DockerImage
}

func (pl *Plan) Process(data map[string]any) *ProcessDesc {
	if pl.Verbose {
		log.Printf("Process: %#v\n", data)
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

func (pl *Plan) File(data map[string]any) *File {
	if path, ok := data["path"]; ok {
		if pathStr, ok := path.(string); ok {
			return &File{
				Path:     pathStr,
				BasePath: pl.Path,
			}
		}
	}
	return nil
}

func (pl *Plan) Workflow(name string) *WorkflowDesc {
	if pl.Verbose {
		log.Printf("Workflow\n")
	}
	w := &WorkflowDesc{Name: fmt.Sprintf("%s:%s", pl.Path, name)}
	pl.Workflows[name] = w
	return w
}

func (pl *Plan) DockerImage(call goja.ConstructorCall) *goja.Object {
	if len(call.Arguments) != 2 {
		fmt.Printf("2 arguments required for DockerImage")
		return nil
	}

	baseDir := call.Arguments[0]
	tag := call.Arguments[1]
	out := &DockerImage{
		BaseDir: baseDir.String(),
		Tag:     tag.String(),
	}
	log.Printf("Found Image %#v", out)
	pl.Images = append(pl.Images, out)
	return nil
}

func (pl *Plan) Print(x any) {
	log.Printf("%s", x)
}

func (pl *Plan) Println(x any) {
	log.Printf("%s\n", x)
}

func (pl *Plan) Glob(pattern string) []string {
	gp := filepath.Join(filepath.Dir(pl.Path), pattern)
	matches, _ := filepath.Glob(gp)
	return matches
}

func (pl *Plan) LoadPlan(path string) *Plan {
	log.Printf("Loading sub-workflow %s\n", path)
	if x, err := RunFile(path); err == nil {
		return x
	} else {
		log.Printf("Error Loading sub-workflow %s : %s\n", path, err)
	}
	return &Plan{}
}

func (pl *Plan) Plugin(cmdLine string) goja.Value {

	cmdArgs, err := shlex.Split(cmdLine)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return nil
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = filepath.Dir(pl.Path)
	stdout, _ := cmd.StdoutPipe()

	go cmd.Run()

	data, err := io.ReadAll(stdout)
	if err != nil {
		log.Printf("Error: %s\n", err)
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
