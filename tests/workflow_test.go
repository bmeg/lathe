package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmeg/lathe/jflow"
	"github.com/bmeg/lathe/logger"
	"github.com/bmeg/lathe/runner"
	"github.com/bmeg/lathe/workflow"
)

type TestWorkflow struct {
	Name        string
	FilePath    string
	Description string
	Timeout     time.Duration
}

func GetTestWorkflows() []TestWorkflow {
	testDir := filepath.Join("..", "test_examples")

	tests := []TestWorkflow{
		{
			Name:        "hello_world",
			FilePath:    filepath.Join(testDir, "01_hello_world.js"),
			Description: "Basic hello world test",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "file_operations",
			FilePath:    filepath.Join(testDir, "02_file_operations.js"),
			Description: "File input/output operations",
			Timeout:     60 * time.Second,
		},
		{
			Name:        "multiple_executors",
			FilePath:    filepath.Join(testDir, "03_multiple_executors.js"),
			Description: "Multiple sequential executors",
			Timeout:     60 * time.Second,
		},
		{
			Name:        "array_commands",
			FilePath:    filepath.Join(testDir, "04_array_commands.js"),
			Description: "Commands as array of strings",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "resource_specification",
			FilePath:    filepath.Join(testDir, "05_resource_specification.js"),
			Description: "Resource requirements specification",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "file_check",
			FilePath:    filepath.Join(testDir, "06_file_check.js"),
			Description: "File existence checking",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "metadata_tags",
			FilePath:    filepath.Join(testDir, "07_metadata_tags.js"),
			Description: "TES metadata tags",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "legacy_format",
			FilePath:    filepath.Join(testDir, "08_legacy_format.js"),
			Description: "Legacy jflow format support",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "workflow_params",
			FilePath:    filepath.Join(testDir, "09_workflow_params.js"),
			Description: "Workflow parameter passing",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "complex_pipeline",
			FilePath:    filepath.Join(testDir, "10_complex_pipeline.js"),
			Description: "Complex multi-step pipeline",
			Timeout:     90 * time.Second,
		},
		{
			Name:        "tool_template_callable",
			FilePath:    filepath.Join(testDir, "11_tool_template_callable.js"),
			Description: "Callable jflow.Tool template with Path factory",
			Timeout:     45 * time.Second,
		},
		{
			Name:        "path_object_factories",
			FilePath:    filepath.Join(testDir, "12_path_object_factories.js"),
			Description: "Path/Object constructors used in tool instantiation",
			Timeout:     45 * time.Second,
		},
	}

	return tests
}

func runWorkflow(t *testing.T, testWf TestWorkflow) error {
	t.Helper()

	logger.Init(false, false)
	defer logger.Close()

	if _, err := os.Stat(testWf.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("test file not found: %s", testWf.FilePath)
	}

	workflows, err := jflow.RunFile(testWf.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse workflow: %w", err)
	}

	if len(workflows.Workflows) == 0 {
		return fmt.Errorf("no workflows found in file")
	}

	var workflowName string
	var wfd *jflow.WorkflowDesc
	for name, desc := range workflows.Workflows {
		workflowName = name
		wfd = desc
		break
	}

	t.Logf("Running workflow: %s", workflowName)

	run := runner.NewSingleMachineRunner(16, 32000)

	wf, err := workflow.PrepWorkflow(wfd, run)
	if err != nil {
		return fmt.Errorf("failed to prepare workflow: %w", err)
	}

	fwf, err := wf.BuildFlame()
	if err != nil {
		return fmt.Errorf("failed to build workflow: %w", err)
	}

	done := make(chan error, 1)

	go func() {
		go func() {
			fwf.ProcessIn <- &workflow.WorkflowStatus{Name: "run", DryRun: false}
			close(fwf.ProcessIn)
		}()

		fwf.Workflow.Start()
		fwf.Workflow.Wait()

		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("workflow execution failed: %w", err)
		}
		t.Logf("Workflow completed successfully: %s", workflowName)
		return nil

	case <-time.After(testWf.Timeout):
		return fmt.Errorf("workflow timed out after %v", testWf.Timeout)
	}
}

func TestAllWorkflows(t *testing.T) {
	tests := GetTestWorkflows()

	for _, testWf := range tests {
		t.Run(testWf.Name, func(t *testing.T) {
			t.Logf("Testing: %s - %s", testWf.Name, testWf.Description)

			err := runWorkflow(t, testWf)
			if err != nil {
				t.Errorf("Test failed: %v", err)
			}
		})
	}
}

func TestIndividualWorkflow(t *testing.T) {
	testName := os.Getenv("JFLOW_TEST_NAME")
	if testName == "" {
		t.Skip("Set JFLOW_TEST_NAME environment variable to run individual test")
	}

	tests := GetTestWorkflows()
	for _, testWf := range tests {
		if testWf.Name == testName {
			t.Logf("Testing: %s - %s", testWf.Name, testWf.Description)
			err := runWorkflow(t, testWf)
			if err != nil {
				t.Errorf("Test failed: %v", err)
			}
			return
		}
	}

	t.Errorf("Test workflow not found: %s", testName)
}

func TestWorkflowValidation(t *testing.T) {
	tests := GetTestWorkflows()

	for _, testWf := range tests {
		t.Run(testWf.Name+"_validation", func(t *testing.T) {
			if _, err := os.Stat(testWf.FilePath); os.IsNotExist(err) {
				t.Errorf("Test file not found: %s", testWf.FilePath)
				return
			}

			workflows, err := jflow.RunFile(testWf.FilePath)
			if err != nil {
				t.Errorf("Failed to parse workflow: %v", err)
				return
			}

			if len(workflows.Workflows) == 0 {
				t.Errorf("No workflows found in file: %s", testWf.FilePath)
				return
			}

			var workflowName string
			for name := range workflows.Workflows {
				workflowName = name
				break
			}

			t.Logf("Workflow validated: %s (name: %s)", testWf.Name, workflowName)
		})
	}
}

func TestWorkflowFilesExist(t *testing.T) {
	tests := GetTestWorkflows()
	missing := []string{}

	for _, testWf := range tests {
		if _, err := os.Stat(testWf.FilePath); os.IsNotExist(err) {
			missing = append(missing, testWf.FilePath)
		}
	}

	if len(missing) > 0 {
		t.Errorf("Missing test files:\n%s", strings.Join(missing, "\n"))
	}
}

func BenchmarkWorkflowExecution(b *testing.B) {
	testWf := TestWorkflow{
		Name:        "hello_world",
		FilePath:    filepath.Join("..", "test_examples", "01_hello_world.js"),
		Description: "Basic hello world test",
		Timeout:     30 * time.Second,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		t := &testing.T{}
		err := runWorkflow(t, testWf)
		if err != nil {
			b.Errorf("Benchmark workflow failed: %v", err)
		}
	}
}
