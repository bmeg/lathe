package jflow

import (
	"os"
	"path/filepath"
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
		input1: "matrix"
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
