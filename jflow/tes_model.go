package jflow

// ============================================================================
// TES (Task Execution Service) API Aligned Model
// ============================================================================
//
// This module defines data structures aligned with the GA4GH TES API v1.1.0
// specification for describing tasks, executors, inputs, outputs, and resources.
//
// Reference: https://ga4gh.github.io/task-execution-schemas/docs/
//
// Key TES concepts mapped to jflow:
//   - tesTask -> ProcessDesc/Task
//   - tesExecutor -> Executor (command to run in container)
//   - tesInput -> Input (file to download and mount)
//   - tesOutput -> Output (file to upload to storage)
//   - tesResources -> Resources (compute requirements)
// ============================================================================

// ============================================================================
// Input/Output Types
// ============================================================================

// FileTypeEnum indicates whether an input/output is a file or directory
type FileTypeEnum string

const (
	FileTypeEnumFile      FileTypeEnum = "FILE"
	FileTypeEnumDirectory FileTypeEnum = "DIRECTORY"
)

// Input describes an input file to be used by the task.
// Inputs will be downloaded and mounted into the executor container.
type Input struct {
	// Name is an optional user-provided name for documentation
	Name string `json:"name,omitempty"`

	// Description is optional documentation about this input
	Description string `json:"description,omitempty"`

	// URL in long term storage (s3://bucket/key, gs://bucket/key, file:///path, http://...)
	// Required unless Content is set
	URL string `json:"url,omitempty"`

	// Path of the file inside the container (must be absolute)
	Path string `json:"path"`

	// Type indicates if this is a file or directory
	Type FileTypeEnum `json:"type,omitempty"`

	// Content is file content literal (alternative to URL)
	Content string `json:"content,omitempty"`
}

// Output describes an output file to be uploaded after task completion.
type Output struct {
	// Name is an optional user-provided name for documentation
	Name string `json:"name,omitempty"`

	// Description is optional documentation about this output
	Description string `json:"description,omitempty"`

	// URL where the TES server will upload the output
	// (s3://bucket/key, gs://bucket/key, file:///path)
	URL string `json:"url"`

	// Path of the file inside the container (must be absolute)
	// May contain wildcards (*, ?, [])
	Path string `json:"path"`

	// PathPrefix to be removed from matching outputs when Path contains wildcards
	PathPrefix string `json:"path_prefix,omitempty"`

	// Type indicates if this is a file or directory
	Type FileTypeEnum `json:"type,omitempty"`
}

// ============================================================================
// Executor
// ============================================================================

// Executor describes a command to be executed in a container.
// Multiple executors run sequentially, sharing the same inputs and volumes.
type Executor struct {
	// Image is the container image name (required)
	// Examples: "ubuntu:20.04", "quay.io/biocontainers/samtools:1.10"
	Image string `json:"image"`

	// Command is a sequence of program arguments to execute (required)
	// The first argument is the program to execute (argv)
	// Example: ["/bin/bash", "-c", "echo hello"]
	Command []string `json:"command"`

	// Workdir is the working directory inside the container
	Workdir string `json:"workdir,omitempty"`

	// Stdin is the path to a file inside the container to be used as stdin
	Stdin string `json:"stdin,omitempty"`

	// Stdout is the path inside the container where stdout will be written
	Stdout string `json:"stdout,omitempty"`

	// Stderr is the path inside the container where stderr will be written
	Stderr string `json:"stderr,omitempty"`

	// Env contains environment variables to set in the container
	Env map[string]string `json:"env,omitempty"`

	// IgnoreError allows the executor to continue even if this executor fails
	IgnoreError bool `json:"ignore_error,omitempty"`
}

// ============================================================================
// Resources
// ============================================================================

// Resources describes the compute resources requested by a task.
type Resources struct {
	// CPUCores is the requested number of CPUs
	CPUCores uint `json:"cpu_cores,omitempty"`

	// Preemptible indicates if the task can run on preemptible/spot instances
	Preemptible bool `json:"preemptible,omitempty"`

	// RamGB is the requested RAM in gigabytes
	RamGB float64 `json:"ram_gb,omitempty"`

	// DiskGB is the requested disk size in gigabytes
	DiskGB float64 `json:"disk_gb,omitempty"`

	// Zones are compute zones where the task should run
	Zones []string `json:"zones,omitempty"`

	// BackendParameters are key/value pairs for backend-specific configuration
	// (e.g., VM size, instance type, etc.)
	BackendParameters map[string]string `json:"backend_parameters,omitempty"`

	// BackendParametersStrict indicates whether to fail if backend parameters are unsupported
	BackendParametersStrict bool `json:"backend_parameters_strict,omitempty"`
}

// ============================================================================
// Task State
// ============================================================================

// State represents the execution state of a task
type State string

const (
	StateUnknown       State = "UNKNOWN"
	StateQueued        State = "QUEUED"         // Waiting for resources
	StateInitializing  State = "INITIALIZING"   // Preparing to run
	StateRunning       State = "RUNNING"        // Currently executing
	StatePaused        State = "PAUSED"         // Paused
	StateComplete      State = "COMPLETE"       // Successfully completed
	StateExecutorError State = "EXECUTOR_ERROR" // Executor failed
	StateSystemError   State = "SYSTEM_ERROR"   // System error
	StateCanceled      State = "CANCELED"       // Canceled by user
	StateCanceling     State = "CANCELING"      // Being canceled
	StatePreempted     State = "PREEMPTED"      // Preempted by system
)

// ============================================================================
// Task
// ============================================================================

// Task describes a complete task in TES format.
// This is the primary unit of work in jflow.
type Task struct {
	// ID is assigned by the server (read-only)
	ID string `json:"id,omitempty"`

	// State is the current execution state (read-only)
	State State `json:"state,omitempty"`

	// Name is a user-provided task name
	Name string `json:"name,omitempty"`

	// Description is optional documentation about the task
	Description string `json:"description,omitempty"`

	// Inputs are files to be downloaded and mounted into the executor
	Inputs []Input `json:"inputs,omitempty"`

	// Outputs are files to be uploaded after execution
	Outputs []Output `json:"outputs,omitempty"`

	// Resources describes compute requirements
	Resources *Resources `json:"resources,omitempty"`

	// Executors are commands to run sequentially
	Executors []Executor `json:"executors"`

	// Volumes are directories shared between executors
	Volumes []string `json:"volumes,omitempty"`

	// Tags are arbitrary key-value metadata
	Tags map[string]string `json:"tags,omitempty"`

	// CreationTime is when the task was created (read-only)
	CreationTime string `json:"creation_time,omitempty"`
}

// ============================================================================
// Polymorphic Command Support
// ============================================================================

// CommandSpec represents a command that can be specified as either a string
// or an array. This provides user-friendly alternatives to strict TES format.
type CommandSpec struct {
	// The command, which can be a string or []string
	value interface{}
}

// NewCommandSpec creates a CommandSpec from either a string or []string
func NewCommandSpec(cmd interface{}) (*CommandSpec, error) {
	switch v := cmd.(type) {
	case string:
		return &CommandSpec{value: v}, nil
	case []string:
		return &CommandSpec{value: v}, nil
	case []interface{}:
		// Convert []interface{} to []string (from JSON unmarshaling)
		strs := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				strs[i] = s
			} else {
				return nil, ErrInvalidCommandType
			}
		}
		return &CommandSpec{value: strs}, nil
	default:
		return nil, ErrInvalidCommandType
	}
}

// ToArray converts the command to []string array format.
// If the command is a string, it wraps it in a shell invocation.
func (cs *CommandSpec) ToArray(shell string) []string {
	if shell == "" {
		shell = "/bin/sh"
	}

	switch v := cs.value.(type) {
	case string:
		// Wrap string command in shell
		return []string{shell, "-c", v}
	case []string:
		return v
	default:
		return []string{shell, "-c", ""}
	}
}

// ToString converts the command to string format
func (cs *CommandSpec) ToString() string {
	switch v := cs.value.(type) {
	case string:
		return v
	case []string:
		// Join array into a shell command
		if len(v) == 0 {
			return ""
		}
		if len(v) == 1 {
			return v[0]
		}
		// If it's a shell invocation ["/bin/sh", "-c", "command"], extract the command
		if len(v) == 3 && (v[0] == "/bin/sh" || v[0] == "/bin/bash" || v[0] == "sh" || v[0] == "bash") && v[1] == "-c" {
			return v[2]
		}
		// Otherwise, join with spaces (simple approach)
		result := ""
		for i, arg := range v {
			if i > 0 {
				result += " "
			}
			// Quote if contains spaces
			if containsSpace(arg) {
				result += "\"" + arg + "\""
			} else {
				result += arg
			}
		}
		return result
	default:
		return ""
	}
}

func containsSpace(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' {
			return true
		}
	}
	return false
}

// IsArray returns true if the command is in array format
func (cs *CommandSpec) IsArray() bool {
	_, ok := cs.value.([]string)
	return ok
}

// IsString returns true if the command is in string format
func (cs *CommandSpec) IsString() bool {
	_, ok := cs.value.(string)
	return ok
}
