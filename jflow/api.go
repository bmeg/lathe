package jflow

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmeg/lathe/logger"
	"github.com/dop251/goja"
	"github.com/google/shlex"
)

var toolTemplatePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// Process creates a new ProcessDesc (job) from a JavaScript object declaration
// Uses GA4GH TES-aligned format
// The object should contain:
//
//	name, executors, inputs (array), outputs (array), resources, volumes, tags
func (pl *Plan) Process(data map[string]any) *ProcessDesc {
	logger.Debug("Creating process from declaration", "data", data)

	proc := &ProcessDesc{
		BasePath:     filepath.Dir(pl.Path),
		Desc:         data,
		Dependencies: []string{},
		Status: &JobStatus{
			State: JobStateQueued,
		},
	}

	// Parse process name
	if name, ok := data["name"].(string); ok {
		proc.Name = name
	}

	// Parse description
	if desc, ok := data["description"].(string); ok {
		proc.Description = desc
	}

	// Parse executors (array format)
	if executors, ok := data["executors"].([]any); ok {
		proc.Executors = make([]Executor, 0, len(executors))
		for _, execAny := range executors {
			if execMap, ok := execAny.(map[string]any); ok {
				executor := parseExecutor(execMap)
				proc.Executors = append(proc.Executors, executor)
			}
		}
	}

	// Parse inputs (array format)
	if inputs, ok := data["inputs"].([]any); ok {
		proc.Inputs = make([]Input, 0, len(inputs))
		for _, inputAny := range inputs {
			if inputMap, ok := inputAny.(map[string]any); ok {
				input := parseInput(inputMap)
				proc.Inputs = append(proc.Inputs, input)
			}
		}
	}

	// Parse outputs (array format)
	if outputs, ok := data["outputs"].([]any); ok {
		proc.Outputs = make([]Output, 0, len(outputs))
		for _, outputAny := range outputs {
			if outputMap, ok := outputAny.(map[string]any); ok {
				output := parseOutput(outputMap)
				proc.Outputs = append(proc.Outputs, output)
			}
		}
	}

	// Parse resources
	if resources, ok := data["resources"].(map[string]any); ok {
		proc.Resources = parseResources(resources)
	}

	// Parse volumes
	if volumes, ok := data["volumes"].([]any); ok {
		proc.Volumes = make([]string, 0, len(volumes))
		for _, vol := range volumes {
			if volStr, ok := vol.(string); ok {
				proc.Volumes = append(proc.Volumes, volStr)
			}
		}
	}

	// Parse tags
	if tags, ok := data["tags"].(map[string]any); ok {
		proc.Tags = make(map[string]string)
		for k, v := range tags {
			if vStr, ok := v.(string); ok {
				proc.Tags[k] = vStr
			}
		}
	}

	// Initialize future for deferred results
	// Initialize future for deferred results
	proc.future = NewFuture[*JobResult]()

	logger.Info("Process created", "name", proc.Name, "executors", len(proc.Executors))
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
func (pl *Plan) Tool(data map[string]any) goja.Value {
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
		if cpuCores, ok := resources["cpu_cores"].(float64); ok {
			tool.Resources.CPUCores = uint(cpuCores)
		}
		if ramGb, ok := resources["ram_gb"].(float64); ok {
			tool.Resources.RamGB = ramGb
		}
		if diskGb, ok := resources["disk_gb"].(float64); ok {
			tool.Resources.DiskGB = diskGb
		}
		if timeout, ok := resources["timeout"].(float64); ok {
			tool.Resources.Timeout = uint(timeout)
		}
		if retries, ok := resources["retries"].(float64); ok {
			tool.Resources.Retries = uint(retries)
		}
	}

	return pl.VM.ToValue(func(call goja.FunctionCall) goja.Value {
		values := map[string]any{}
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			if err := pl.VM.ExportTo(call.Arguments[0], &values); err != nil {
				if fallback, ok := call.Arguments[0].Export().(map[string]any); ok {
					values = fallback
				}
			}
		}

		proc := pl.instantiateTool(tool, values)
		return pl.VM.ToValue(proc)
	})
}

// PathFactory creates a reusable local file constructor that binds a logical file name.
// Usage: fileCtor = jflow.Path("input"); fileObj = fileCtor("/real/path")
func (pl *Plan) PathFactory(name string) goja.Value {
	return pl.VM.ToValue(func(path string) *File {
		file := &File{
			Type:     FileTypeLocal,
			Path:     path,
			BasePath: filepath.Dir(pl.Path),
			Metadata: map[string]string{"name": name},
		}
		return file
	})
}

// Object creates a reusable remote file constructor that binds a logical file name.
// The URI determines the file type (s3/http/local fallback).
func (pl *Plan) Object(name string) goja.Value {
	return pl.VM.ToValue(func(uri string) *File {
		file := &File{
			Type:     inferFileType(uri),
			Path:     uri,
			BasePath: filepath.Dir(pl.Path),
			Metadata: map[string]string{"name": name},
		}
		return file
	})
}

func (pl *Plan) instantiateTool(tool *ToolCommand, values map[string]any) *ProcessDesc {
	templateValues := make(map[string]string, len(values))
	for key, value := range values {
		templateValues[key] = stringifyTemplateValue(value)
	}

	commandLine := renderToolTemplate(tool.CommandLine, templateValues)
	commandSpec, err := NewCommandSpec(commandLine)
	if err != nil {
		commandSpec, _ = NewCommandSpec("")
	}

	processName := tool.Name
	if name, ok := values["name"].(string); ok && name != "" {
		processName = name
	}

	resources := tool.Resources
	if resources.CPUCores == 0 {
		resources.CPUCores = 1
	}
	if resources.RamGB == 0 {
		resources.RamGB = 1.0
	}
	if resources.DiskGB == 0 {
		resources.DiskGB = 10.0
	}

	proc := &ProcessDesc{
		BasePath: filepath.Dir(pl.Path),
		Name:     processName,
		Desc:     map[string]any{"tool": tool.Name, "values": values},
		Executors: []Executor{{
			Image:   tool.Image,
			Command: commandSpec.ToArray(tool.Shell),
		}},
		Resources:    &resources,
		Dependencies: []string{},
		Status: &JobStatus{
			State: JobStateQueued,
		},
		future: NewFuture[*JobResult](),
	}

	proc.Inputs = buildToolInputs(tool, values)
	proc.Outputs = buildToolOutputs(tool, values, templateValues)

	return proc
}

func buildToolInputs(tool *ToolCommand, values map[string]any) []Input {
	inputs := []Input{}

	if len(tool.Inputs) > 0 {
		for key := range tool.Inputs {
			if value, ok := values[key]; ok {
				if input, ok := valueToInput(key, value); ok {
					inputs = append(inputs, input)
				}
			}
		}
		return inputs
	}

	for key, value := range values {
		if input, ok := valueToInput(key, value); ok {
			inputs = append(inputs, input)
		}
	}

	return inputs
}

func buildToolOutputs(tool *ToolCommand, values map[string]any, templateValues map[string]string) []Output {
	outputs := []Output{}
	for key, spec := range tool.Outputs {
		path := ""
		if value, ok := values[key]; ok {
			path = stringifyTemplateValue(value)
		}
		if path == "" {
			path = renderToolTemplate(spec, templateValues)
		}
		if path == "" {
			continue
		}
		outputs = append(outputs, Output{Name: key, Path: path})
	}
	return outputs
}

func valueToInput(name string, value any) (Input, bool) {
	if file := fileFromValue(value); file != nil {
		return Input{Name: name, Path: file.Path}, true
	}
	return Input{}, false
}

func fileFromValue(value any) *File {
	if file, ok := value.(*File); ok && file != nil {
		return file
	}
	if m, ok := value.(map[string]any); ok {
		path, ok := m["path"].(string)
		if !ok || path == "" {
			return nil
		}
		fileType := inferFileType(path)
		if t, ok := m["type"].(string); ok && t != "" {
			fileType = FileType(strings.ToLower(t))
		}
		return &File{Type: fileType, Path: path}
	}
	return nil
}

func stringifyTemplateValue(value any) string {
	if file := fileFromValue(value); file != nil {
		return file.Path
	}
	return fmt.Sprint(value)
}

func renderToolTemplate(template string, values map[string]string) string {
	return toolTemplatePattern.ReplaceAllStringFunc(template, func(match string) string {
		tokens := toolTemplatePattern.FindStringSubmatch(match)
		if len(tokens) != 2 {
			return match
		}
		if replacement, ok := values[tokens[1]]; ok {
			return replacement
		}
		return ""
	})
}

func inferFileType(path string) FileType {
	lower := strings.ToLower(path)
	switch {
	case strings.HasPrefix(lower, "s3://"):
		return FileTypeS3
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"):
		return FileTypeHTTP
	default:
		return FileTypeLocal
	}
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

// ============================================================================
// Format Parsing Helpers
// ============================================================================

// parseExecutor parses an executor from a map
func parseExecutor(data map[string]any) Executor {
	executor := Executor{
		Env: make(map[string]string),
	}

	if image, ok := data["image"].(string); ok {
		executor.Image = image
	}

	// Parse command - support both string and array
	if cmd, ok := data["command"]; ok {
		switch v := cmd.(type) {
		case string:
			// String command - wrap in shell
			executor.Command = []string{"/bin/sh", "-c", v}
		case []any:
			executor.Command = make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					executor.Command = append(executor.Command, s)
				}
			}
		case []string:
			executor.Command = v
		}
	}

	if workdir, ok := data["workdir"].(string); ok {
		executor.Workdir = workdir
	}
	if stdin, ok := data["stdin"].(string); ok {
		executor.Stdin = stdin
	}
	if stdout, ok := data["stdout"].(string); ok {
		executor.Stdout = stdout
	}
	if stderr, ok := data["stderr"].(string); ok {
		executor.Stderr = stderr
	}
	if ignoreErr, ok := data["ignore_error"].(bool); ok {
		executor.IgnoreError = ignoreErr
	}

	// Parse environment variables
	if env, ok := data["env"].(map[string]any); ok {
		for k, v := range env {
			if vStr, ok := v.(string); ok {
				executor.Env[k] = vStr
			}
		}
	}

	return executor
}

// parseInput parses an input from a map
func parseInput(data map[string]any) Input {
	input := Input{}

	if name, ok := data["name"].(string); ok {
		input.Name = name
	}
	if desc, ok := data["description"].(string); ok {
		input.Description = desc
	}
	if url, ok := data["url"].(string); ok {
		input.URL = url
	}
	if path, ok := data["path"].(string); ok {
		input.Path = path
	}
	if content, ok := data["content"].(string); ok {
		input.Content = content
	}
	if fileType, ok := data["type"].(string); ok {
		input.Type = FileTypeEnum(fileType)
	}

	return input
}

// parseOutput parses an output from a map
func parseOutput(data map[string]any) Output {
	output := Output{}

	if name, ok := data["name"].(string); ok {
		output.Name = name
	}
	if desc, ok := data["description"].(string); ok {
		output.Description = desc
	}
	if url, ok := data["url"].(string); ok {
		output.URL = url
	}
	if path, ok := data["path"].(string); ok {
		output.Path = path
	}
	if pathPrefix, ok := data["path_prefix"].(string); ok {
		output.PathPrefix = pathPrefix
	}
	if fileType, ok := data["type"].(string); ok {
		output.Type = FileTypeEnum(fileType)
	}

	return output
}

// parseResources parses resources from a map
func parseResources(data map[string]any) *ResourceRequirements {
	res := &ResourceRequirements{}

	if cpuCores, ok := data["cpu_cores"].(float64); ok {
		res.CPUCores = uint(cpuCores)
	}
	if ramGb, ok := data["ram_gb"].(float64); ok {
		res.RamGB = ramGb
	}
	if diskGb, ok := data["disk_gb"].(float64); ok {
		res.DiskGB = diskGb
	}
	if preemptible, ok := data["preemptible"].(bool); ok {
		res.Preemptible = preemptible
	}

	// Parse zones
	if zones, ok := data["zones"].([]any); ok {
		res.Zones = make([]string, 0, len(zones))
		for _, z := range zones {
			if zStr, ok := z.(string); ok {
				res.Zones = append(res.Zones, zStr)
			}
		}
	} else if zone, ok := data["zones"].(string); ok {
		// Support single zone as string
		res.Zones = []string{zone}
	}

	// Parse backend parameters
	if backendParams, ok := data["backend_parameters"].(map[string]any); ok {
		res.BackendParameters = make(map[string]string)
		for k, v := range backendParams {
			if vStr, ok := v.(string); ok {
				res.BackendParameters[k] = vStr
			}
		}
	}

	if backendStrict, ok := data["backend_parameters_strict"].(bool); ok {
		res.BackendParametersStrict = backendStrict
	}

	if timeout, ok := data["timeout"].(float64); ok {
		res.Timeout = uint(timeout)
	}
	if retries, ok := data["retries"].(float64); ok {
		res.Retries = uint(retries)
	}

	return res
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
