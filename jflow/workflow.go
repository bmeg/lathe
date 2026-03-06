package jflow

import (
	"fmt"

	"github.com/bmeg/lathe/logger"
	"github.com/dop251/goja"
)

// Add adds a step to the workflow. The step can be a ProcessDesc, FileCheck,
// or another WorkflowDesc (which will be inlined).
// This is called from JavaScript to compose the workflow.
func (wd *WorkflowDesc) Add(call goja.ConstructorCall) *goja.Object {
	if len(call.Arguments) != 1 {
		logger.Error("Workflow.Add requires exactly one argument")
		return nil
	}

	e := call.Arguments[0].Export()

	if proc, ok := e.(*ProcessDesc); ok {
		// Auto-generate a name if not provided
		if proc.Name == "" {
			proc.Name = fmt.Sprintf("%s_step_%d", wd.Name, len(wd.Steps))
		}
		logger.Debug("Adding process to workflow", "workflow", wd.Name, "process", proc.Name)
		wd.Steps = append(wd.Steps, proc)

	} else if wf, ok := e.(*WorkflowDesc); ok {
		// Inline sub-workflow steps
		logger.Debug("Inlining sub-workflow into parent", "parent", wd.Name, "subworkflow", wf.Name, "steps", len(wf.Steps))
		wd.Steps = append(wd.Steps, wf.Steps...)

	} else if fc, ok := e.(*FileCheck); ok {
		// Add file existence check step
		logger.Debug("Adding file check to workflow", "workflow", wd.Name, "file", fc.File.Path)
		wd.Steps = append(wd.Steps, fc)

	} else if file, ok := e.(*File); ok {
		// Wrap File in FileCheck if passed directly
		logger.Debug("Adding file (wrapped as check) to workflow", "workflow", wd.Name, "file", file.Path)
		wd.Steps = append(wd.Steps, &FileCheck{File: file})

	} else {
		logger.Error("Workflow.Add received unsupported step type", "workflow", wd.Name, "type", fmt.Sprintf("%T", e))
	}

	return nil
}

// AddWithName is a convenience method to add a step with a specific name
func (wd *WorkflowDesc) AddWithName(name string, step Step) error {
	if proc, ok := step.(*ProcessDesc); ok {
		proc.Name = name
	}
	wd.Steps = append(wd.Steps, step)
	return nil
}
