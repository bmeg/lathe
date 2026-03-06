package scriptfile

import (
	"errors"
	"sync"
	"time"
)

// Error types
var (
	ErrInvalidCommandType = errors.New("command must be string or []string")
)

// ============================================================================
// Workflow Object Model Schema
// ============================================================================
//
// This module defines the core data structures used to declare workflows,
// tools, jobs, and files in the Lathe workflow engine.
//
// Key Concepts:
//   - Files: Represent both local files and cloud storage objects (S3)
//   - Tools: Encapsulate containerized commands with resource requirements
//   - Jobs: Individual task executions with inputs, outputs, and dependencies
//   - Workflows: Directed acyclic graphs of Jobs connected by file dependencies
//   - Futures: Deferred results that resolve upon task completion
// ============================================================================

// ============================================================================
// File Types
// ============================================================================

// FileType indicates whether a file is local or in cloud storage
type FileType string

const (
	FileTypeLocal FileType = "local"
	FileTypeS3    FileType = "s3"
	FileTypeHTTP  FileType = "http"
)

// File represents a data file that can be used as input/output for jobs.
// Files can be local filesystem paths or cloud storage objects (S3, etc).
type File struct {
	// Type indicates the file location type (local, s3, http)
	Type FileType `json:"type"`

	// Path is the file path/URI. For local files, relative paths are
	// resolved against BasePath. For S3, format is s3://bucket/key
	Path string `json:"path"`

	// BasePath is the base directory for resolving relative paths (local files only)
	BasePath string `json:"-"`

	// Optional metadata
	Size         int64             `json:"size,omitempty"`
	LastModified time.Time         `json:"lastModified,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// MustExist indicates this file must exist before a job runs
type FileCheck struct {
	File *File `json:"file"`
}

// ============================================================================
// Resource Requirements
// ============================================================================

// ResourceRequirements specifies compute resources needed for a job.
// Aligned with GA4GH TES API specification.
type ResourceRequirements struct {
	// CPUCores: number of CPU cores requested
	CPUCores uint `json:"cpu_cores,omitempty"`

	// RamGB: memory in gigabytes
	RamGB float64 `json:"ram_gb,omitempty"`

	// DiskGB: disk space in gigabytes
	DiskGB float64 `json:"disk_gb,omitempty"`

	// Preemptible: allow running on preemptible/spot instances
	Preemptible bool `json:"preemptible,omitempty"`

	// Zones: compute zones where the task should run
	Zones []string `json:"zones,omitempty"`

	// BackendParameters: key/value pairs for backend configuration
	BackendParameters map[string]string `json:"backend_parameters,omitempty"`

	// BackendParametersStrict: fail if backend parameters are unsupported
	BackendParametersStrict bool `json:"backend_parameters_strict,omitempty"`

	// Timeout: maximum execution time in seconds (jflow-specific)
	Timeout uint `json:"timeout,omitempty"`

	// Retries: number of times to retry failed execution (jflow-specific)
	Retries uint `json:"retries,omitempty"`
}

// Default resource requirements if not specified
func DefaultResourceRequirements() ResourceRequirements {
	return ResourceRequirements{
		CPUCores:    1,
		RamGB:       1.0,
		DiskGB:      10.0,
		Preemptible: false,
		Timeout:     0, // no timeout
		Retries:     0,
	}
}

// ToResources converts to native resource format
func (r *ResourceRequirements) ToResources() *Resources {
	return &Resources{
		CPUCores:                r.CPUCores,
		RamGB:                   r.RamGB,
		DiskGB:                  r.DiskGB,
		Preemptible:             r.Preemptible,
		Zones:                   r.Zones,
		BackendParameters:       r.BackendParameters,
		BackendParametersStrict: r.BackendParametersStrict,
	}
}

// ============================================================================
// Docker/Container Configuration
// ============================================================================

// DockerImage represents a container image specification
type DockerImage struct {
	// BaseDir is the local directory containing the Dockerfile
	BaseDir string `json:"baseDir"`

	// Tag is the image tag (repository:tag format)
	Tag string `json:"tag"`

	// Dockerfile path relative to BaseDir
	Dockerfile string `json:"dockerfile,omitempty"`

	// BuildArgs for docker build
	BuildArgs map[string]string `json:"buildArgs,omitempty"`

	// PullAlways forces pulling the image even if it exists locally
	PullAlways bool `json:"pullAlways,omitempty"`
}

// ============================================================================
// Tool/Command Definition
// ============================================================================

// ToolCommand represents a command-line tool with inputs, outputs, and resource requirements
type ToolCommand struct {
	// Name is the unique identifier for this tool
	Name string `json:"name"`

	// CommandLine is the command to execute (may contain template variables)
	CommandLine string `json:"commandLine"`

	// Shell is the shell interpreter to use (sh, bash, etc). If empty, command is executed directly.
	Shell string `json:"shell,omitempty"`

	// Inputs maps input parameter names to their file paths/locations
	Inputs map[string]string `json:"inputs"`

	// Outputs maps output parameter names to their file paths/locations
	Outputs map[string]string `json:"outputs"`

	// Image specifies the Docker image to run this command in
	Image string `json:"image,omitempty"`

	// Resources specifies CPU, memory, and other requirements
	Resources ResourceRequirements `json:"resources,omitempty"`

	// BuildArgs for image building if applicable
	BuildArgs map[string]string `json:"buildArgs,omitempty"`

	// Metadata for this tool
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ============================================================================
// Job/Process Definition
// ============================================================================

// JobState represents the state of a job execution
// Uses TES-aligned state names
type JobState string

const (
	// TES-aligned states
	JobStateUnknown       JobState = "UNKNOWN"
	JobStateQueued        JobState = "QUEUED"
	JobStateInitializing  JobState = "INITIALIZING"
	JobStateRunning       JobState = "RUNNING"
	JobStatePaused        JobState = "PAUSED"
	JobStateComplete      JobState = "COMPLETE"
	JobStateExecutorError JobState = "EXECUTOR_ERROR"
	JobStateSystemError   JobState = "SYSTEM_ERROR"
	JobStateCanceled      JobState = "CANCELED"
	JobStateCanceling     JobState = "CANCELING"
	JobStatePreempted     JobState = "PREEMPTED"
)

// ToState converts JobState to State
func (js JobState) ToState() State {
	return State(js)
}

// JobStatus represents the execution status of a job
type JobStatus struct {
	State     JobState       `json:"state"`
	ExitCode  int            `json:"exitCode,omitempty"`
	Error     string         `json:"error,omitempty"`
	StartTime time.Time      `json:"startTime,omitempty"`
	EndTime   time.Time      `json:"endTime,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ProcessDesc represents a job declaration in the workflow.
// Uses GA4GH TES-aligned format.
type ProcessDesc struct {
	// BasePath is the working directory for this process
	BasePath string `json:"-"`

	// Name is the unique identifier for this process/job
	Name string `json:"name"`

	// Description of what this job does
	Description string `json:"description,omitempty"`

	// The raw declaration data (for extensibility)
	Desc map[string]any `json:"-"`

	// Inputs are input files (TES format)
	Inputs []Input `json:"inputs,omitempty"`

	// Outputs are output files (TES format)
	Outputs []Output `json:"outputs,omitempty"`

	// Executors are commands to run sequentially (TES format)
	Executors []Executor `json:"executors,omitempty"`

	// Volumes are shared directories between executors
	Volumes []string `json:"volumes,omitempty"`

	// Resources specifies compute requirements
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// Tags are arbitrary key-value metadata
	Tags map[string]string `json:"tags,omitempty"`

	// ===== Runtime fields =====

	// Dependencies on other processes (job names)
	Dependencies []string `json:"-"`

	// Status tracking
	Status *JobStatus `json:"-"`

	// Future for this job's result
	future *Future[*JobResult] `json:"-"`

	// Callback function to execute after job completion
	onComplete CallbackFunc `json:"-"`

	mu sync.RWMutex
}

// GetName implements the Step interface
func (pd *ProcessDesc) GetName() string {
	return pd.Name
}

// GetBasePath implements the Step interface
func (pd *ProcessDesc) GetBasePath() string {
	return pd.BasePath
}

// GetInputs implements the Step interface
func (pd *ProcessDesc) GetInputs() map[string]string {
	// Convert Input array to map for compatibility
	result := make(map[string]string)
	for _, input := range pd.Inputs {
		key := input.Name
		if key == "" {
			key = input.Path
		}
		result[key] = input.Path
	}
	return result
}

// GetProcess implements the Step interface
func (pd *ProcessDesc) GetProcess() *ProcessDesc {
	return pd
}

// ToTask converts ProcessDesc to Task format
func (pd *ProcessDesc) ToTask() *Task {
	task := &Task{
		Name:        pd.Name,
		Description: pd.Description,
		Inputs:      pd.Inputs,
		Outputs:     pd.Outputs,
		Executors:   pd.Executors,
		Tags:        pd.Tags,
		Volumes:     pd.Volumes,
	}

	// Convert resources
	if pd.Resources != nil {
		task.Resources = pd.Resources.ToResources()
	}

	return task
}

// ============================================================================
// Workflow Definition
// ============================================================================

// WorkflowDesc represents a complete workflow definition
type WorkflowDesc struct {
	// Name is the unique identifier for this workflow
	Name string `json:"name"`

	// Steps are the jobs/processes that make up the workflow
	Steps []Step `json:"steps"`

	// InputParams are parameters passed to the workflow
	InputParams map[string]any `json:"-"`

	// Metadata about the workflow
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Note: Add() method is defined in workflow.go to support Goja integration

// ============================================================================
// Job Execution Results and Futures
// ============================================================================

// JobResult represents the result of a completed job
type JobResult struct {
	// JobName is the name of the job that produced this result
	JobName string

	// Status is the final status of the job
	Status JobStatus

	// OutputFiles maps output names to the files that were produced
	OutputFiles map[string]*File

	// Logs contains stdout and stderr output
	Logs map[string]string

	// Metadata about the execution
	Metadata map[string]any
}

// Future represents a deferred job result that will be resolved when the job completes
type Future[T any] struct {
	result T
	err    error
	done   chan struct{}
	mu     sync.RWMutex
}

// NewFuture creates a new Future for a deferred result
func NewFuture[T any]() *Future[T] {
	return &Future[T]{
		done: make(chan struct{}),
	}
}

// Resolve sets the result value and marks the Future as complete
func (f *Future[T]) Resolve(result T) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.result = result
	close(f.done)
}

// Reject sets an error and marks the Future as complete
func (f *Future[T]) Reject(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
	close(f.done)
}

// Wait blocks until the Future is resolved and returns the result
func (f *Future[T]) Wait() (T, error) {
	<-f.done
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.result, f.err
}

// GetFuture returns the Future for this ProcessDesc (for job results)
func (pd *ProcessDesc) GetFuture() *Future[*JobResult] {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	if pd.future == nil {
		pd.future = NewFuture[*JobResult]()
	}
	return pd.future
}

// ============================================================================
// Callbacks for Post-Job Analysis
// ============================================================================

// CallbackFunc is a function to be executed after a job completes
type CallbackFunc func(*JobResult) error

// SetCallback registers a callback to be executed after this job completes
func (pd *ProcessDesc) SetCallback(cb CallbackFunc) {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.onComplete = cb
}

// ExecuteCallback calls the registered callback (if any) with the job result
func (pd *ProcessDesc) ExecuteCallback(result *JobResult) error {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	if pd.onComplete != nil {
		return pd.onComplete(result)
	}
	return nil
}

// ============================================================================
// Workflow Execution Plan
// ============================================================================

// ExecutionPlan represents the complete plan for executing a workflow
type ExecutionPlan struct {
	// Workflows is the map of workflow definitions
	Workflows map[string]*WorkflowDesc

	// Images are the Docker images referenced by jobs
	Images []*DockerImage

	// Parameters passed to the workflow
	Parameters map[string]any

	// Metadata about the execution plan
	Metadata map[string]any
}

// ============================================================================
// Step Interface
// ============================================================================

// Step is the interface for workflow steps (both jobs and file checks)
type Step interface {
	GetName() string
	GetBasePath() string
	GetInputs() map[string]string
	GetProcess() *ProcessDesc
}

// FileCheck implements the Step interface for file existence checks
func (fc *FileCheck) GetName() string {
	return fc.File.Path
}

func (fc *FileCheck) GetBasePath() string {
	return fc.File.BasePath
}

func (fc *FileCheck) GetInputs() map[string]string {
	return map[string]string{
		"file": fc.File.Path,
	}
}

func (fc *FileCheck) GetProcess() *ProcessDesc {
	return nil
}
