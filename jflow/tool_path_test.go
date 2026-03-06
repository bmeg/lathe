package jflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolTemplateCallable(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "tool_template_test.js")
	script := `
const wf = jflow.Workflow("main");
const inputMatrix = jflow.Path("input_matrix");
const makeTask = jflow.Tool({
	name: "matrix_tool",
	commandLine: "cat {{input1}} > {{output1}} && echo {{param1}}",
	image: "ubuntu:20.04",
	inputs: {
		input1: "File",
		param1: "Value"
	},
	outputs: {
		output1: "{{output1}}"
	}
});

const proc = makeTask({
	input1: inputMatrix("/tmp/input_matrix.txt"),
	output1: "/tmp/output_matrix.txt",
	param1: 5
});

wf.Add(proc);
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	plan, err := RunFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	wf := plan.Workflows["main"]
	if wf == nil {
		t.Fatalf("workflow main not found")
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(wf.Steps))
	}

	proc := wf.Steps[0].GetProcess()
	if proc == nil {
		t.Fatalf("expected process step")
	}
	if proc.Name != "matrix_tool" {
		t.Fatalf("expected process name matrix_tool, got %s", proc.Name)
	}
	if len(proc.Executors) != 1 || len(proc.Executors[0].Command) < 3 {
		t.Fatalf("unexpected executor command: %#v", proc.Executors)
	}

	cmd := proc.Executors[0].Command[2]
	expectedCmd := "cat /tmp/input_matrix.txt > /tmp/output_matrix.txt && echo 5"
	if cmd != expectedCmd {
		t.Fatalf("unexpected rendered command. expected %q, got %q", expectedCmd, cmd)
	}

	if len(proc.Inputs) != 1 {
		t.Fatalf("expected 1 tool input, got %d", len(proc.Inputs))
	}
	if proc.Inputs[0].Name != "input1" || proc.Inputs[0].Path != "/tmp/input_matrix.txt" {
		t.Fatalf("unexpected input mapping: %#v", proc.Inputs[0])
	}

	if len(proc.Outputs) != 1 {
		t.Fatalf("expected 1 tool output, got %d", len(proc.Outputs))
	}
	if proc.Outputs[0].Name != "output1" || proc.Outputs[0].Path != "/tmp/output_matrix.txt" {
		t.Fatalf("unexpected output mapping: %#v", proc.Outputs[0])
	}
}

func TestPathAndObjectFactories(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "path_object_test.js")
	script := `
const wf = jflow.Workflow("main");
const localInput = jflow.Path("local_input");
const remoteInput = jflow.Object("remote_input");

wf.Add(localInput("/tmp/local.txt"));
wf.Add(remoteInput("s3://bucket/path/file.txt"));
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	plan, err := RunFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	wf := plan.Workflows["main"]
	if wf == nil {
		t.Fatalf("workflow main not found")
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("expected 2 file check steps, got %d", len(wf.Steps))
	}

	localStep, ok := wf.Steps[0].(*FileCheck)
	if !ok {
		t.Fatalf("expected first step to be FileCheck")
	}
	if localStep.File.Type != FileTypeLocal || localStep.File.Path != "/tmp/local.txt" {
		t.Fatalf("unexpected local file: %#v", localStep.File)
	}

	remoteStep, ok := wf.Steps[1].(*FileCheck)
	if !ok {
		t.Fatalf("expected second step to be FileCheck")
	}
	if remoteStep.File.Type != FileTypeS3 || remoteStep.File.Path != "s3://bucket/path/file.txt" {
		t.Fatalf("unexpected remote file: %#v", remoteStep.File)
	}
}

func TestImportExportsCallableTools(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "module_tools.js")
	parentPath := filepath.Join(tmpDir, "parent_workflow.js")

	moduleScript := `
export const echoTool = jflow.Tool({
	name: "echo_tool",
	commandLine: "echo {{msg}} > {{out}}",
	image: "ubuntu:20.04",
	outputs: {
		out: "{{out}}"
	}
});

const indexed = jflow.Tool({
	name: "index_tool",
	commandLine: "echo indexing {{bam}} > {{bai}}",
	image: "ubuntu:20.04",
	inputs: {
		bam: "File"
	},
	outputs: {
		bai: "{{bai}}"
	}
});

export { indexed as samtoolsIndex };
`

	parentScript := `
const wf = jflow.Workflow("import_test");
const tools = jflow.Import("./module_tools.js");
const p = jflow.Path("p");

const sayHello = tools.echoTool({
	name: "say_hello",
	msg: "hello",
	out: "/tmp/hello_import.txt"
});

const index = tools.samtoolsIndex({
	name: "index_from_import",
	bam: p("/tmp/sample.bam"),
	bai: "/tmp/sample.bam.bai"
});

wf.Add(sayHello);
wf.Add(index);
`

	if err := os.WriteFile(modulePath, []byte(moduleScript), 0o644); err != nil {
		t.Fatalf("failed to write module script: %v", err)
	}
	if err := os.WriteFile(parentPath, []byte(parentScript), 0o644); err != nil {
		t.Fatalf("failed to write parent script: %v", err)
	}

	plan, err := RunFile(parentPath)
	if err != nil {
		t.Fatalf("failed to run parent script: %v", err)
	}

	wf := plan.Workflows["import_test"]
	if wf == nil {
		t.Fatalf("workflow import_test not found")
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(wf.Steps))
	}

	first := wf.Steps[0].GetProcess()
	if first == nil {
		t.Fatalf("expected first step to be process")
	}
	if first.Name != "say_hello" {
		t.Fatalf("unexpected first process name: %s", first.Name)
	}
	if got := first.Executors[0].Command[2]; got != "echo hello > /tmp/hello_import.txt" {
		t.Fatalf("unexpected first command: %q", got)
	}

	second := wf.Steps[1].GetProcess()
	if second == nil {
		t.Fatalf("expected second step to be process")
	}
	if second.Name != "index_from_import" {
		t.Fatalf("unexpected second process name: %s", second.Name)
	}
	if len(second.Inputs) != 1 || second.Inputs[0].Path != "/tmp/sample.bam" {
		t.Fatalf("unexpected second process inputs: %#v", second.Inputs)
	}
}

func TestToolInputKindsFileAndValue(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "tool_input_kinds_test.js")
	script := `
const wf = jflow.Workflow("main");
const p = jflow.Path("input_file");

const typedTool = jflow.Tool({
	name: "typed_tool",
	commandLine: "cat {{inputFile}} > {{output}} && echo {{threshold}} >> {{output}} && echo {{flags}} >> {{output}}",
	image: "ubuntu:20.04",
	inputs: {
		inputFile: "File",
		threshold: "Value",
		flags: "Value"
	},
	outputs: {
		output: "{{output}}"
	}
});

const proc = typedTool({
	name: "typed_run",
	inputFile: p("/tmp/typed_input.txt"),
	threshold: 0.75,
	flags: ["--a", "--b"],
	output: "/tmp/typed_output.txt"
});

wf.Add(proc);
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	plan, err := RunFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	wf := plan.Workflows["main"]
	if wf == nil {
		t.Fatalf("workflow main not found")
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(wf.Steps))
	}

	proc := wf.Steps[0].GetProcess()
	if proc == nil {
		t.Fatalf("expected process step")
	}

	if len(proc.Inputs) != 1 {
		t.Fatalf("expected exactly 1 file input dependency, got %d", len(proc.Inputs))
	}
	if proc.Inputs[0].Name != "inputFile" || proc.Inputs[0].Path != "/tmp/typed_input.txt" {
		t.Fatalf("unexpected input mapping: %#v", proc.Inputs[0])
	}

	if len(proc.Executors) != 1 || len(proc.Executors[0].Command) < 3 {
		t.Fatalf("unexpected executor command: %#v", proc.Executors)
	}
	cmd := proc.Executors[0].Command[2]
	expectedCmd := "cat /tmp/typed_input.txt > /tmp/typed_output.txt && echo 0.75 >> /tmp/typed_output.txt && echo [--a --b] >> /tmp/typed_output.txt"
	if cmd != expectedCmd {
		t.Fatalf("unexpected rendered command. expected %q, got %q", expectedCmd, cmd)
	}
}

func TestImportWithParamOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "module_overrides.js")
	parentPath := filepath.Join(tmpDir, "parent_overrides.js")

	moduleScript := `
const p = jflow.GetParams({
	threads: "Number"
});

export const threadValue = p.threads;
export const echoTool = jflow.Tool({
	name: "echo_threads",
	commandLine: "echo {{value}} > {{out}}",
	image: "ubuntu:20.04",
	inputs: {
		value: "Value",
		out: "Value"
	},
	outputs: {
		out: "{{out}}"
	}
});
`

	parentScript := `
const wf = jflow.Workflow("main");
const imported = jflow.Import("./module_overrides.js", { threads: 9 });
const parentParams = jflow.GetParams({
	threads: "Number"
});

const proc = imported.echoTool({
	name: "show_override",
	value: imported.threadValue + "-" + parentParams.threads,
	out: "/tmp/threads.txt"
});

wf.Add(proc);
`

	if err := os.WriteFile(modulePath, []byte(moduleScript), 0o644); err != nil {
		t.Fatalf("failed to write module script: %v", err)
	}
	if err := os.WriteFile(parentPath, []byte(parentScript), 0o644); err != nil {
		t.Fatalf("failed to write parent script: %v", err)
	}

	plan, err := RunFileWithParams(parentPath, map[string]any{"threads": 2})
	if err != nil {
		t.Fatalf("failed to run parent script: %v", err)
	}

	wf := plan.Workflows["main"]
	if wf == nil {
		t.Fatalf("workflow main not found")
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(wf.Steps))
	}

	proc := wf.Steps[0].GetProcess()
	if proc == nil {
		t.Fatalf("expected process step")
	}
	if len(proc.Executors) != 1 || len(proc.Executors[0].Command) < 3 {
		t.Fatalf("unexpected executor command: %#v", proc.Executors)
	}

	cmd := proc.Executors[0].Command[2]
	expected := "echo 9-2 > /tmp/threads.txt"
	if cmd != expected {
		t.Fatalf("unexpected rendered command. expected %q, got %q", expected, cmd)
	}
}

func TestToolOutputGlobTemplate(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "tool_output_glob_test.js")
	script := `
const wf = jflow.Workflow("main");
const p = jflow.Path("reads");

const alignTool = jflow.Tool({
	name: "align_tool",
	commandLine: "echo {{sample}} > /tmp/{{sample}}.txt",
	image: "ubuntu:20.04",
	inputs: {
		reads: "File",
		sample: "Value"
	},
	outputs: {
		alignment: "/tmp/{{sample}}*.bam"
	}
});

const proc = alignTool({
	name: "align_sample",
	reads: p("/tmp/sample.fastq.gz"),
	sample: "S1"
});

wf.Add(proc);
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	plan, err := RunFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	wf := plan.Workflows["main"]
	if wf == nil {
		t.Fatalf("workflow main not found")
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(wf.Steps))
	}

	proc := wf.Steps[0].GetProcess()
	if proc == nil {
		t.Fatalf("expected process step")
	}
	if len(proc.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(proc.Outputs))
	}
	if proc.Outputs[0].Name != "alignment" {
		t.Fatalf("unexpected output name: %s", proc.Outputs[0].Name)
	}
	if proc.Outputs[0].Path != "/tmp/S1*.bam" {
		t.Fatalf("unexpected output glob pattern: %s", proc.Outputs[0].Path)
	}
}

func TestGetParamsNormalizesFileAndValidatesNestedSchema(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "get_params_test.js")

	script := `
const wf = jflow.Workflow("main");
const params = jflow.GetParams({
	input: "File",
	threads: "Number",
	nested: {
		mode: "String"
	}
});

const proc = jflow.Tool({
	name: "param_tool",
	commandLine: "cat {{input}} > {{output}} && echo {{threads}}",
	image: "ubuntu:20.04",
	inputs: {
		input: "File",
		threads: "Value"
	},
	outputs: {
		output: "{{output}}"
	}
})({
	input: params.input,
	threads: params.threads,
	output: "/tmp/out.txt"
});

wf.Add(proc);
`

	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	params := map[string]any{
		"input":   "inputs/sample.fastq.gz",
		"threads": 8,
		"nested": map[string]any{
			"mode": "test",
		},
	}

	plan, err := RunFileWithParams(scriptPath, params)
	if err != nil {
		t.Fatalf("failed to run script: %v", err)
	}

	wf := plan.Workflows["main"]
	if wf == nil {
		t.Fatalf("workflow main not found")
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(wf.Steps))
	}

	proc := wf.Steps[0].GetProcess()
	if proc == nil {
		t.Fatalf("expected process step")
	}
	if len(proc.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(proc.Inputs))
	}

	expectedInputPath := filepath.Join(tmpDir, "inputs", "sample.fastq.gz")
	if proc.Inputs[0].Path != expectedInputPath {
		t.Fatalf("expected normalized input path %q, got %q", expectedInputPath, proc.Inputs[0].Path)
	}
}

func TestGetParamsFailsOnTypeMismatch(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "get_params_fail_test.js")

	script := `
const wf = jflow.Workflow("main");
jflow.GetParams({
	threads: "Number"
});
`

	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	_, err := RunFileWithParams(scriptPath, map[string]any{"threads": "eight"})
	if err == nil {
		t.Fatalf("expected schema validation error")
	}
	if got := err.Error(); got == "" || !containsAll(got, []string{"threads", "number"}) {
		t.Fatalf("unexpected error message: %q", got)
	}
}

func containsAll(s string, terms []string) bool {
	lower := strings.ToLower(s)
	for _, term := range terms {
		if !strings.Contains(lower, strings.ToLower(term)) {
			return false
		}
	}
	return true
}
