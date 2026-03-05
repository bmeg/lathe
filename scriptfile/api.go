package scriptfile

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/bmeg/lathe/logger"
	"github.com/dop251/goja"
	"github.com/google/shlex"
)

// Process creates a new ProcessDesc (job) from a JavaScript object declaration
// The object should contain: name, commandLine, shell (optional), inputs, outputs, image, cpus, memoryMB
func (pl *Plan) Process(data map[string]any) *ProcessDesc {
	logger.Debug("Creating process from declaration", "data", data)

	proc := &ProcessDesc{
		BasePath:     filepath.Dir(pl.Path),
		Desc:         data,
		Inputs:       make(map[string]string),
		Outputs:      make(map[string]string),
		Dependencies: []string{},
		Status: &JobStatus{
			State: JobStatePending,
		},
	}

	// Parse command line
	if cmd, ok := data["commandLine"].(string); ok {
		proc.CommandLine = cmd
	}

	// Parse shell interpreter
	if shell, ok := data["shell"].(string); ok {
		proc.Shell = shell
	}

	// Parse inputs map
	if inputs, ok := data["inputs"].(map[string]any); ok {
		for key, val := range inputs {
			if valStr, ok := val.(string); ok {
				proc.Inputs[key] = valStr
			}
		}
	}

	// Parse outputs map
	if outputs, ok := data["outputs"].(map[string]any); ok {
		for key, val := range outputs {
			if valStr, ok := val.(string); ok {
				proc.Outputs[key] = valStr
			}
		}
	}

	// Parse Docker image
	if image, ok := data["image"].(string); ok {
		proc.Image = image
	}

	// Parse resource requirements - memory
	proc.MemMB = 1024 // default
	if memMb, ok := data["memoryMB"].(float64); ok {
		proc.MemMB = uint(memMb)
	} else if memMb, ok := data["memMB"].(float64); ok {
		proc.MemMB = uint(memMb)
	}

	// Parse resource requirements - CPUs
	proc.NCpus = 1 // default
	if cpus, ok := data["cpus"].(float64); ok {
		proc.NCpus = uint(cpus)
	} else if cpus, ok := data["ncpus"].(float64); ok {
		proc.NCpus = uint(cpus)
	}

	// Parse process name
	if name, ok := data["name"].(string); ok {
		proc.Name = name
	}

	// Parse description
	if desc, ok := data["description"].(string); ok {
		proc.Description = desc
	}

	// Initialize future for deferred results
	proc.future = NewFuture[*JobResult]()

	logger.Info("Process created", "name", proc.Name, "command", proc.CommandLine)
	return proc
}

// File creates a new File reference from a JavaScript object declaration
// The object should contain: path, and optionally type (local|s3|http), metadata
func (pl *Plan) File(data map[string]any) *File {
	if path, ok := data["path"].(string); ok {
		fileType := FileTypeLocal
		if fType, ok := data["type"].(string); ok {
			fileType = FileType(fType)
		}

		file := &File{
			Type:     fileType,
			Path:     path,
			BasePath: filepath.Dir(pl.Path),
		}

		// Parse optional metadata
		if metadata, ok := data["metadata"].(map[string]any); ok {
			file.Metadata = make(map[string]string)
			for k, v := range metadata {
				if vStr, ok := v.(string); ok {
					file.Metadata[k] = vStr
				}
			}
		}

		logger.Debug("File declared", "path", path, "type", fileType)
		return file
	}
	return nil
}

// FileCheck creates a file existence check step
func (pl *Plan) FileCheck(data map[string]any) *FileCheck {
	if fileLike, ok := data["file"]; ok {
		if fileMap, ok := fileLike.(map[string]any); ok {
			file := pl.File(fileMap)
			if file != nil {
				return &FileCheck{File: file}
			}
		}
	}
	return nil
}

// Workflow creates a new named workflow
func (pl *Plan) Workflow(name string) *WorkflowDesc {
	logger.Debug("Creating workflow", "name", name)

	workflow := &WorkflowDesc{
		Name:        fmt.Sprintf("%s:%s", pl.Path, name),
		Steps:       []Step{},
		InputParams: make(map[string]any),
		Metadata:    make(map[string]any),
	}

	pl.Workflows[name] = workflow
	return workflow
}

// Tool creates a reusable tool/command template
// This is different from Process - it's a template that can be instantiated multiple times
func (pl *Plan) Tool(data map[string]any) *ToolCommand {
	logger.Debug("Creating tool template", "data", data)

	tool := &ToolCommand{
		Inputs:    make(map[string]string),
		Outputs:   make(map[string]string),
		BuildArgs: make(map[string]string),
		Metadata:  make(map[string]any),
	}

	// Parse tool metadata
	if name, ok := data["name"].(string); ok {
		tool.Name = name
	}
	if desc, ok := data["commandLine"].(string); ok {
		tool.CommandLine = desc
	}
	if shell, ok := data["shell"].(string); ok {
		tool.Shell = shell
	}
	if image, ok := data["image"].(string); ok {
		tool.Image = image
	}

	// Parse inputs
	if inputs, ok := data["inputs"].(map[string]any); ok {
		for k, v := range inputs {
			if vStr, ok := v.(string); ok {
				tool.Inputs[k] = vStr
			}
		}
	}

	// Parse outputs
	if outputs, ok := data["outputs"].(map[string]any); ok {
		for k, v := range outputs {
			if vStr, ok := v.(string); ok {
				tool.Outputs[k] = vStr
			}
		}
	}

	// Parse resources
	if resources, ok := data["resources"].(map[string]any); ok {
		if cpus, ok := resources["cpus"].(float64); ok {
			tool.Resources.CPUs = uint(cpus)
		}
		if mem, ok := resources["memoryMB"].(float64); ok {
			tool.Resources.MemoryMB = uint(mem)
		}
		if disk, ok := resources["diskMB"].(float64); ok {
			tool.Resources.DiskMB = uint(disk)
		}
		if timeout, ok := resources["timeout"].(float64); ok {
			tool.Resources.Timeout = uint(timeout)
		}
		if retries, ok := resources["retries"].(float64); ok {
			tool.Resources.Retries = uint(retries)
		}
	}

	return tool
}

// DockerImage creates a new Docker image specification
func (pl *Plan) DockerImage(call goja.ConstructorCall) *goja.Object {
	if len(call.Arguments) < 2 {
		logger.Error("DockerImage requires at least 2 arguments: baseDir, tag")
		return nil
	}

	baseDir := call.Arguments[0].String()
	tag := call.Arguments[1].String()

	image := &DockerImage{
		BaseDir:   baseDir,
		Tag:       tag,
		BuildArgs: make(map[string]string),
	}

	// Optional Dockerfile path
	if len(call.Arguments) > 2 {
		image.Dockerfile = call.Arguments[2].String()
	}

	// Optional build arguments
	if len(call.Arguments) > 3 {
		if buildArgs, ok := call.Arguments[3].Export().(map[string]any); ok {
			for k, v := range buildArgs {
				if vStr, ok := v.(string); ok {
					image.BuildArgs[k] = vStr
				}
			}
		}
	}

	logger.Info("Docker image declared", "tag", tag, "baseDir", baseDir)
	pl.Images = append(pl.Images, image)
	return nil
}

// Print logs a message at info level
func (pl *Plan) Print(x any) {
	logger.Info(fmt.Sprintf("%v", x))
}

// Println logs a message with newline at info level
func (pl *Plan) Println(x any) {
	logger.Info(fmt.Sprintf("%v\n", x))
}

// Glob expands a glob pattern relative to the script directory
func (pl *Plan) Glob(pattern string) []string {
	gp := filepath.Join(filepath.Dir(pl.Path), pattern)
	matches, err := filepath.Glob(gp)
	if err != nil {
		logger.Error("Glob error", "pattern", pattern, "error", err)
		return []string{}
	}
	return matches
}

// OnComplete sets up a callback function to execute when a job completes.
// This function can be called from the JavaScript to attach post-job analysis.
func (pl *Plan) OnComplete(proc *ProcessDesc, callback goja.Value) error {
	if callback == nil || goja.IsNull(callback) || goja.IsUndefined(callback) {
		return nil
	}

	// Wrap the goja function to conform to CallbackFunc
	gojaFunc, ok := goja.AssertFunction(callback)
	if !ok {
		return fmt.Errorf("onComplete requires a function argument")
	}

	callbackFunc := func(result *JobResult) error {
		// Convert result to a goja value that the JS function can work with
		resultMap := map[string]any{
			"jobName": result.JobName,
			"status": map[string]any{
				"state":    string(result.Status.State),
				"exitCode": result.Status.ExitCode,
				"error":    result.Status.Error,
				"metadata": result.Status.Metadata,
			},
			"outputFiles": result.OutputFiles,
			"logs":        result.Logs,
			"metadata":    result.Metadata,
		}

		_, err := gojaFunc(goja.Null(), pl.VM.ToValue(resultMap))
		return err
	}

	proc.SetCallback(callbackFunc)
	logger.Debug("Callback registered for process", "name", proc.Name)
	return nil
}

// LoadPlan loads and executes a sub-workflow script from an external file
func (pl *Plan) LoadPlan(path string) map[string]*WorkflowDesc {
	logger.Debug("Loading sub-workflow", "path", path)

	// Resolve relative paths against the current script directory
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(pl.Path), path)
	}

	subplan, err := RunFileWithParams(path, pl.Parameters)
	if err != nil {
		logger.Error("Error loading sub-workflow", "path", path, "error", err)
		return map[string]*WorkflowDesc{}
	}

	return subplan.Workflows
}

// Plugin executes an external command and returns its JSON output.
// This allows integrating external tools and data generators into the workflow.
func (pl *Plan) Plugin(cmdLine string) goja.Value {
	cmdArgs, err := shlex.Split(cmdLine)
	if err != nil {
		logger.Error("Plugin parse error", "commandLine", cmdLine, "error", err)
		return nil
	}

	if len(cmdArgs) == 0 {
		logger.Error("Plugin error: empty command")
		return nil
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = filepath.Dir(pl.Path)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Plugin error creating stdout pipe", "error", err)
		return nil
	}

	go func() {
		if err := cmd.Run(); err != nil {
			logger.Error("Plugin execution error", "commandLine", cmdLine, "error", err)
		}
	}()

	data, err := io.ReadAll(stdout)
	if err != nil {
		logger.Error("Plugin read error", "error", err, "commandLine", cmdLine)
		return nil
	}

	// Try to parse as JSON object
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err == nil {
		return pl.VM.ToValue(m)
	}

	// Try to parse as JSON array
	a := []any{}
	if err := json.Unmarshal(data, &a); err == nil {
		return pl.VM.ToValue(a)
	}

	// Return as string if not valid JSON
	return pl.VM.ToValue(string(data))
}
