# Lathe Workflow Engine - Architecture and Design

This document describes the architecture and design decisions of the refactored Lathe workflow engine.

## Design Philosophy

The Lathe workflow engine follows these principles:

1. **JavaScript-first declarations**: Workflows are declared in JavaScript, not YAML/JSON
   - More expressive and dynamic
   - Support for conditionals, loops, and functions
   - Easier to generate workflows programmatically

2. **Immediate return, deferred execution**: All job declarations return immediately
   - Enables building complex DAGs without waiting
   - Futures provide deferred result resolution
   - Callbacks enable post-job analysis

3. **Centralized object model**: Single source of truth for types
   - All workflow objects defined in `model.go`
   - Consistent across all packages
   - Easy to extend and maintain

4. **Automatic dependency resolution**: Dependencies through file inputs/outputs
   - Less boilerplate in workflow declarations
   - Intuitive mental model (data flow = dependencies)
   - Still allows explicit dependencies when needed

5. **Multiple execution backends**: Same workflow model for different runners
   - Local runner for development/testing
   - TES runner for cloud/HPC deployments
   - Easy to add new runners

## Project Structure

```
jflow/               # JavaScript parsing and workflow model building
├── model.go        # Central type definitions and interfaces
├── tes_model.go    # TES-aligned task model types
├── js_vm.go        # Goja JavaScript runtime setup
├── api.go          # API functions exposed to JavaScript
└── workflow.go     # Workflow composition logic

workflow/           # DAG execution and orchestration
├── workflow.go     # Workflow execution model
├── step_process.go # Process execution steps
├── step_filecheck.go # File check steps
└── util.go         # Utility functions

runner/             # Job execution backends
├── cmd_exec.go     # Local execution via os.exec
└── tes_exec.go     # GA4GH TES API execution

logger/             # Logging utilities

examples/           # Example workflows

docs/
├── JFLOW_STANDARD.md    # Complete API documentation (canonical)
└── INTEGRATION_GUIDE.md # Integration and usage guide
```

## Key Components

### 1. Centralized Object Model (`jflow/model.go`)

**Purpose**: Single source of truth for all workflow-related types

**Key Types**:
- `File`: Represents data files (local, S3, HTTP)
- `ProcessDesc`: Job/task definition with inputs, outputs, resources
- `WorkflowDesc`: Collection of steps forming a DAG
- `ToolCommand`: Reusable command template
- `Future[T]`: Deferred result container
- `JobResult`: Result of job execution
- `JobStatus`: Execution status tracking

**Design Decisions**:
- Use generic `Future[T]` for type-safe deferred results
- Separate concerns: declaration types vs. execution types
- Include metadata maps for extensibility
- Support multiple file types (local, S3, HTTP)

### 2. JavaScript Environment (`jflow/js_vm.go`)

**Purpose**: Initialize Goja runtime and set up jflow standard API

**Key Functions**:
- `RunFile(path)`: Parse and execute workflow script
- `RunFileWithParams(path, params)`: Pass parameters to script
- `setupVM()`: Configure global jflow API

**Design Decisions**:
- Separate `RunFile` and `RunFileWithParams` for clarity
- Global `jflow` object namespaces all API functions (jflow standard)
- Backward compatibility with `lathe` namespace
- Parameters passed as `jflow.Params` map
- Easy to add new global functions

### 3. API Functions (`jflow/api.go`)

**Purpose**: Implement the global functions exposed to JavaScript

**Key Functions**:
- `Process(spec)`: Create job declaration
- `File(spec)`: Create file reference
- `FileCheck(spec)`: Create file check step
- `Tool(spec)`: Create tool template
- `Workflow(name)`: Create workflow
- `DockerImage(...)`: Declare container image
- `OnComplete(job, callback)`: Register callback
- `Import(path)`: Import module exports
- `Plugin(command)`: Execute external command

**Design Decisions**:
- Consistent JavaScript object → Go struct conversion
- Flexible parsing with sensible defaults
- Support for both old and new parameter names (e.g., memMB → memoryMB)
- Type conversion handles JavaScript number quirks (float64, int64, etc.)

### 4. Workflow Composition (`jflow/workflow.go`)

**Purpose**: Manage workflow steps and support Goja integration

**Key Methods**:
- `Add(step)`: Add step with Goja call support
- `AddWithName(name, step)`: Add step with explicit name

**Design Decisions**:
- `Add()` uses Goja's ConstructorCall to support object-oriented JS
- Auto-generates step names if not provided
- Supports multiple step types: ProcessDesc, FileCheck, WorkflowDesc
- Can inline imported workflow modules into parent

### 5. Futures and Callbacks

**Purpose**: Enable deferred result handling and post-job analysis

**Key Types**:
- `Future[T]`: Generic container with `Resolve()`, `Reject()`, `Wait()`
- `CallbackFunc`: Function receiving `*JobResult`
- `JobResult`: Complete execution result
- `JobStatus`: Status information

**Design Decisions**:
- Generic `Future[T]` for type safety
- Separate resolve/reject for success/failure paths
- Callbacks attached to ProcessDesc, not separate
- Callback receives complete JobResult with metadata

**Execution Flow**:
```
1. JavaScript calls jflow.Process({...})
   → Creates ProcessDesc
   → Initializes Future[*JobResult]
   → Returns immediately

2. Script finishes building workflow
   → No jobs have executed yet
   → All futures are pending

3. Runner executes workflow
   → Executor processes each job
   → When complete, calls proc.GetFuture().Resolve(result)
   → Callback (if registered) is executed with result

4. Client can wait for future
   → future.Wait() blocks until resolved
   → Returns JobResult or error
```

## Execution Flow

### Script Parsing

```
JavaScript Script
    ↓
Goja Runtime (vm.RunScript)
    ↓
API Function Calls (jflow.Process, jflow.Workflow, etc.)
    ↓
Plan Object (Workflows + Images + Parameters)
    ↓
ExecutionPlan (exported for runners)
```

### Workflow Execution

```
ExecutionPlan
    ↓
Workflow Engine (PrepWorkflow)
    ↓
DAG with Dependency Resolution
    ↓
Runner Selection (Local / TES)
    ↓
Job Execution (for each job)
    ├─ Future.Resolve(result)
    ├─ Callback execution (if registered)
    └─ Log and track result
    ↓
Final Workflow Status
```

## Type Safety and Extensibility

### Type Safety

- **Workflow Object Model**: Centralized type definitions in `model.go`
- **Future[T]**: Generic containers for type-safe results
- **Step Interface**: Type-safe polymorphism for workflow steps
- **Consistent naming**: CamelCase for Go, snake_case for JSON tags

### Extensibility

- **Metadata maps**: All major types include metadata fields
  ```go
  Metadata map[string]any `json:"metadata,omitempty"`
  ```
- **Plugin system**: External commands can generate workflow configs
- **Custom runners**: Same workflow model works with different executors
- **New file types**: Can add S3, HTTP, FTP, etc. without API changes

## Resource Management

### Resource Requirements

Each job can specify:
- **CPUs**: Number of cores
- **Memory**: Megabytes of RAM
- **Disk**: Megabytes of disk space
- **GPUs**: Number of GPU devices
- **Timeout**: Maximum execution time
- **Retries**: Number of retry attempts

### Resource Constraints

- Local runner tracks available resources
- Respects CPU and memory pools
- Serializes resource-constrained job execution
- Cloud/HPC runners defer to backend scheduler

## File Handling

### File Types

1. **Local Files**: Relative to script directory or absolute paths
2. **S3 Files**: `s3://bucket/key` URIs
3. **HTTP Files**: `http://host/path` URIs

### File Resolution

- Relative paths resolved against script directory
- S3 paths require credentials (via environment or config)
- HTTP paths can be cached or streamed

### File Dependencies

- Job outputs become available to other jobs
- File checks ensure prerequisites exist
- Dependency graph built from file input/output matches

## Error Handling

### Job Failures

1. **Exit Code != 0**: Job failed during execution
2. **Timeout**: Job exceeded time limit
3. **Resource Error**: Insufficient resources available
4. **Container Error**: Image missing or failed to run

### Callback Error Handling

```javascript
onComplete(job, function(result) {
  if (result.status.state === "completed") {
    if (result.status.exitCode === 0) {
      // Success
    } else {
      // Failure - can log, retry, or skip dependent jobs
    }
  } else {
    // Failed or cancelled
    lathe.println("Error: " + result.status.error);
  }
});
```

### Workflow Failure Modes

1. **Job Failure**: Stop if critical, continue if non-critical
2. **Missing Input**: Dependent job blocked
3. **Resource Error**: Retry or fail workflow
4. **Container Build**: Fail before job execution

## Performance Considerations

### Optimization Strategies

1. **Parallel Execution**: Independent jobs run concurrently
2. **Resource Pooling**: Limit concurrent jobs by resource constraints
3. **Lazy Evaluation**: Parse script once, execute multiple times
4. **Container Caching**: Reuse built images across jobs

### Scalability

- **Local Runner**: Limited by machine resources
- **TES Runner**: Scales to cloud/HPC capabilities
- **DAG Size**: Thousands of jobs supported
- **File Size**: Depends on storage backend

## Testing Strategy

### Unit Tests

- Test type conversions and parsing
- Test Future resolution and callbacks
- Test dependency resolution logic

### Integration Tests

- End-to-end workflow execution
- Multi-job dependencies
- Callback execution and error handling

### Example Workflows

- Basic job with inputs/outputs
- Multiple dependent jobs
- Parameterized workflows
- Module import loading
- Error handling scenarios

## Future Enhancements

### Planned Features

1. **Caching**: Cache job outputs to avoid re-execution
2. **Streaming**: Stream large files between jobs
3. **Checkpointing**: Resume workflows from interruption points
4. **Monitoring**: Real-time progress tracking
5. **Logging**: Structured logging to external systems

### Potential Extensions

1. **Workflow Versioning**: Track workflow version with results
2. **Job Lineage**: Track data provenance through jobs
3. **Resource Optimization**: Auto-tune resource requirements
4. **Workflow Templates**: Pre-built common patterns
5. **IDE Support**: VS Code extension for workflow editing

## Backward Compatibility

### Old Type Locations

Backward compatibility is maintained through API aliasing (`lathe.*` and `jflow.*`).

### Migration Path

- Old scripts using `lathe.*` continue to run
- New code should use the `jflow` package and `jflow.*` API directly
- All functionality preserved and enhanced

### API Changes

| Old | New | Status |
|-----|-----|--------|
| `memMB` | `memoryMB` | Supported both |
| `ncpus` | `cpus` | Supported both |
| No callbacks | `onComplete()` | New feature |
| Basic File | Enhanced File | Extended |
| No futures | `Future[T]` | New feature |

## Conclusion

The refactored Lathe workflow engine provides:

1. **Clear separation of concerns**: Parsing, composition, execution
2. **Type-safe abstractions**: Futures, callbacks, structured types
3. **Extensibility**: Metadata, plugins, custom runners
4. **Consistency**: Centralized model, uniform API
5. **Scalability**: Supports local, cloud, and HPC execution

The design enables both simple workflows (local testing) and complex production deployments (cloud/HPC) using the same declarative model.
