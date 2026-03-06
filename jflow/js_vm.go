package jflow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmeg/lathe/logger"
	"github.com/dop251/goja"
)

// Plan represents the execution plan built from a script file.
// It contains all workflows, images, and metadata parsed from the JavaScript.
type Plan struct {
	// Workflows maps workflow names to their definitions
	Workflows map[string]*WorkflowDesc

	// Images are the Docker images referenced by jobs
	Images []*DockerImage

	// Verbose enables debug logging
	Verbose bool

	// Path is the absolute path to the source script file
	Path string

	// VM is the Goja runtime used for execution
	VM *goja.Runtime

	// Parameters passed into the workflow script
	Parameters map[string]any
}

// RunFile parses and executes a JavaScript workflow script file
// and returns the execution plan.
func RunFile(path string) (*Plan, error) {
	return RunFileWithParams(path, map[string]any{})
}

// RunFileWithParams parses and executes a JavaScript workflow script file
// with user-provided parameters and returns the execution plan.
func RunFileWithParams(path string, params map[string]any) (*Plan, error) {
	// Try to get absolute path. If it fails, fall back to relative path.
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	// Read file
	source, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script at path %s: %w", absPath, err)
	}

	// Create new VM and plan
	vm := goja.New()
	plan := &Plan{
		Workflows:  map[string]*WorkflowDesc{},
		Images:     []*DockerImage{},
		Path:       absPath,
		VM:         vm,
		Parameters: params,
		Verbose:    false,
	}

	// Set up the global Lathe API
	if err := plan.setupVM(); err != nil {
		return nil, err
	}

	// Execute the script
	logger.Debug("Executing workflow script", "path", absPath)
	_, err = vm.RunScript("main", string(source))
	if err != nil {
		return nil, fmt.Errorf("error executing script %s: %w", absPath, err)
	}

	logger.Info("Workflow plan created", "workflows", len(plan.Workflows), "images", len(plan.Images))
	return plan, nil
}

// setupVM initializes the JavaScript VM with the jflow API (open standard)
func (pl *Plan) setupVM() error {
	vm := pl.VM

	// Set up the jflow API object (open standard for workflow definitions)
	// Note: "lathe" is kept for backward compatibility
	jflowObj := map[string]any{
		// Workflow declaration
		"Workflow": pl.Workflow,

		// Job/Process declaration
		"Process": pl.Process,

		// File declaration and checking
		"File":      pl.File,
		"FileCheck": pl.FileCheck,
		"Path":      pl.PathFactory,
		"Object":    pl.Object,

		// Docker image declaration
		"DockerImage": pl.DockerImage,

		// Tool/command templates
		"Tool": pl.Tool,

		// Module import
		"Import": pl.Import,

		// Plugin system for extensibility
		"Plugin": pl.Plugin,

		// Configuration and parameters
		"Params":    pl.Parameters,
		"GetParams": pl.GetParams,
	}

	vm.Set("jflow", jflowObj)
	// Maintain backward compatibility with "lathe" namespace
	vm.Set("lathe", jflowObj)

	// Global utility functions
	vm.Set("print", pl.Print)
	vm.Set("println", pl.Println)
	vm.Set("glob", pl.Glob)

	// Callback utilities
	vm.Set("onComplete", pl.OnComplete)

	return nil
}

// ExecutionPlanFromScript parses a script file and returns an ExecutionPlan
func ExecutionPlanFromScript(path string, params map[string]any) (*ExecutionPlan, error) {
	plan, err := RunFileWithParams(path, params)
	if err != nil {
		return nil, err
	}

	return &ExecutionPlan{
		Workflows:  plan.Workflows,
		Images:     plan.Images,
		Parameters: plan.Parameters,
		Metadata: map[string]any{
			"scriptPath": plan.Path,
		},
	}, nil
}
