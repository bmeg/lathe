# jflow: JavaScript Workflow Standard

## Overview

**jflow** is an open standard for defining computational workflows using JavaScript. It provides a consistent API for workflow engines to implement, enabling workflow portability across different execution environments.

## Design Philosophy

jflow is designed to be:

- **Engine-agnostic**: Workflows written in jflow can run on any compliant implementation
- **JavaScript-native**: Leverages JavaScript's expressiveness for workflow logic
- **Declarative yet flexible**: Allows both simple declarative workflows and complex dynamic generation
- **Human-readable**: Workflows are easy to understand and maintain

## Standard API

The jflow standard defines a global `jflow` object with the following interface:

### Core Functions

```javascript
// Workflow declaration
jflow.Workflow(name: string): WorkflowDesc

// Job/Process declaration
jflow.Process(spec: ProcessSpec): ProcessDesc

// File operations
jflow.File(spec: FileSpec): File
jflow.FileCheck(spec: FileCheckSpec): FileCheck
jflow.Path(name: string): (path: string) => File
jflow.Object(name: string): (uri: string) => File

// Tool templates
jflow.Tool(spec: ToolSpec): ToolCommand

// Container management
jflow.DockerImage(baseDir: string, tag: string, dockerfile?: string, buildArgs?: object): void

// Workflow composition
jflow.Import(path: string, paramsOverride?: object): object

// Extensibility
jflow.Plugin(command: string): any

// Runtime parameters
jflow.Params: object
jflow.GetParams(schema: object): object
```

### Global Utilities

```javascript
// Output functions
print(message: string): void
println(message: string): void

// File system
glob(pattern: string): string[]

// Callbacks
onComplete(job: ProcessDesc, callback: (result: JobResult) => void): void
```

## Type Specifications

### ProcessSpec

jflow supports TES (Task Execution Service) aligned format for maximum compatibility.


#### TES-Aligned Format (GA4GH TES API v1.1.0)
```typescript
interface ProcessSpec {
  name?: string;            // Task name
  description?: string;     // Task description
  
  // TES-aligned arrays
  executors: Executor[];    // Array of executors
  inputs?: Input[];         // Array of input files
  outputs?: Output[];       // Array of output files
  volumes?: string[];       // Shared volumes between executors
  
  // TES-aligned resources
  resources?: Resources;    // Compute resources
  tags?: Record<string, string>; // Arbitrary key-value metadata
}

interface Executor {
  image: string;            // Required: Container image name
  command: string | string[]; // Required: Command (polymorphic: string or array)
  workdir?: string;         // Working directory
  stdin?: string;           // Stdin file path
  stdout?: string;          // Stdout file path
  stderr?: string;          // Stderr file path
  env?: Record<string, string>; // Environment variables
  ignore_error?: boolean;   // Continue on error
}

interface Input {
  name?: string;            // Input name
  description?: string;     // Input description
  url?: string;             // Source URL (s3://, gs://, file://, http://)
  path: string;             // Required: Path in container (absolute)
  content?: string;         // File content literal (alternative to url)
  type?: "FILE" | "DIRECTORY";
}

interface Output {
  name?: string;            // Output name
  description?: string;     // Output description
  url?: string;             // Optional in local execution, required by TES backends
  path: string;             // Required: Path in container (absolute)
  path_prefix?: string;     // Prefix for wildcard outputs
  type?: "FILE" | "DIRECTORY";
}

interface Resources {
  cpu_cores?: number;       // Number of CPU cores
  ram_gb?: number;          // Memory in gigabytes
  disk_gb?: number;         // Disk space in gigabytes
  preemptible?: boolean;    // Allow preemptible/spot instances
  zones?: string[];         // Compute zones
  backend_parameters?: Record<string, string>; // Backend-specific parameters
  backend_parameters_strict?: boolean; // Fail on unsupported params
  timeout?: number;         // Optional jflow extension (seconds)
  retries?: number;         // Optional jflow extension
}
```

Implementation note (Lathe): the data model accepts multiple executors, and TES backends may run them sequentially. The current local workflow runner executes the first executor.

#### Polymorphic Command Support

Commands can be specified as either strings or arrays for user convenience:

```javascript
// String format (automatically wrapped in shell)
jflow.Process({
  executors: [{
    image: "ubuntu:20.04",
    command: "echo hello world > output.txt"  // Becomes: ["/bin/sh", "-c", "echo hello world > output.txt"]
  }]
})

// Array format
jflow.Process({
  executors: [{
    image: "ubuntu:20.04",
    command: ["/bin/bash", "-c", "echo hello world > output.txt"]
  }]
})
```

### FileSpec
```typescript
interface FileSpec {
  path: string;             // Required: File path or URI
  type?: "local" | "s3" | "http";
  metadata?: object;
}
```

### FileCheckSpec
```typescript
interface FileCheckSpec {
  file: FileSpec;           // File to check existence
}
```

### ToolSpec
```typescript
interface ToolSpec {
  name: string;             // Required: Tool identifier
  commandLine: string;      // Required: Command template
  shell?: string;
  image?: string;
  inputs?: Record<string, "File" | "Value">; // Template variable name -> input kind
  outputs?: Record<string, string>; // Output name -> glob pattern, templated with input values
  resources?: ResourceSpec;
  metadata?: object;
}
```

## Implementation Requirements

A jflow-compliant workflow engine MUST:

1. Provide a JavaScript runtime (ES6 or higher)
2. Implement all core `jflow.*` functions
3. Support the standard type specifications
4. Resolve file-based dependencies automatically
5. Support the callback system for post-job hooks
6. Allow parameter passing via `jflow.Params`

A jflow-compliant workflow engine SHOULD:

1. Support Docker/container execution
2. Provide resource management (CPU, memory, disk)
3. Handle retries and error recovery
4. Support multiple execution backends
5. Provide logging and debugging capabilities

A jflow-compliant workflow engine MAY:

1. Extend the API with additional functions
2. Support additional file types beyond local/S3/HTTP
3. Provide workflow visualization tools
4. Implement custom plugin systems
5. Add engine-specific optimizations

## Example Workflow

### TES-Aligned Format
```javascript
// jflow workflow with TES-aligned format
const pipeline = jflow.Workflow("data_pipeline");

// Process data with TES format
const cleanJob = jflow.Process({
  name: "clean_data",
  description: "Clean and normalize input data",
  
  // TES executors
  executors: [{
    image: "python:3.11",
    command: "python clean.py /data/input.csv > /data/cleaned.csv",  // String or array
    workdir: "/workspace",
    env: {
      "PYTHONPATH": "/app",
      "LOG_LEVEL": "INFO"
    }
  }],
  
  // TES inputs (array of file specifications)
  inputs: [{
    name: "raw_data",
    url: "s3://my-bucket/input.csv",
    path: "/data/input.csv",
    type: "FILE"
  }],
  
  // TES outputs (array with destinations)
  outputs: [{
    name: "cleaned_data",
    url: "s3://my-bucket/output/cleaned.csv",
    path: "/data/cleaned.csv",
    type: "FILE"
  }],
  
  // TES resources
  resources: {
    cpu_cores: 2,
    ram_gb: 4,
    disk_gb: 10,
    preemptible: false
  },
  
  // TES tags
  tags: {
    "project": "data-pipeline",
    "stage": "cleaning"
  },
  
  // Shared volumes
  volumes: ["/data", "/workspace"]
});

pipeline.Add(cleanJob);

// Add callback
onComplete(cleanJob, function(result) {
  if (result.status.exitCode === 0) {
    println("Data cleaning successful!");
  } else {
    println("Error: " + result.status.error);
  }
});
```

## Workflow Composition

### Adding Steps to a Workflow

```javascript
const wf = jflow.Workflow("myworkflow");

// Add a process
const job1 = jflow.Process({ /* ... */ });
wf.Add(job1);

// Add a file check
const check = jflow.FileCheck({ /* ... */ });
wf.Add(check);

// Add another job (dependency inferred via shared file path)
const job2 = jflow.Process({
  name: "process2",
  executors: [{
    image: "ubuntu:20.04",
    command: "wc -l /data/output.sam > /data/count.txt"
  }],
  inputs: [{ name: "data", path: "/data/output.sam" }]
});
wf.Add(job2);

// Add a workflow exported from an imported module
const sub = jflow.Workflow("sub");
const subJob = jflow.Process({ /* ... */ });
sub.Add(subJob);
wf.Add(sub);
```

### Dependency Resolution

Dependencies are automatically resolved based on:

1. **File dependencies**: If job B's input path matches job A's output path, B depends on A.
2. **File checks**: Jobs depend on file-check steps that guarantee required input files exist.

## Futures and Callbacks

### Getting a Future for a Job

Every process has a future that resolves when execution completes:

```javascript
const job = jflow.Process({
  name: "my_job",
  executors: [{
    image: "ubuntu:20.04",
    command: "echo hello > /tmp/output.txt"
  }],
  outputs: [{ name: "message", path: "/tmp/output.txt" }]
});

const jobFuture = job.GetFuture();
```

### Callback Result Shape

```typescript
interface JobResult {
  jobName: string;
  status: JobStatus;
  outputFiles: object;
  logs: object;
  metadata: object;
}

interface JobStatus {
  state: "UNKNOWN" | "QUEUED" | "INITIALIZING" | "RUNNING" | "PAUSED" | "COMPLETE" | "EXECUTOR_ERROR" | "SYSTEM_ERROR" | "CANCELED" | "CANCELING" | "PREEMPTED";
  exitCode: number;
  error?: string;
  startTime?: Date;
  endTime?: Date;
  metadata?: object;
}
```

## Running Workflows

```bash
# Execute workflow script with default parameters
lathe run workflow.js

# Execute with inline parameters
lathe run workflow.js --params mode=test --params threads=8

# Execute with parameter file
lathe run workflow.js --params-file params.yaml
```

Parameters are accessible in scripts through `jflow.Params` and can be validated with `jflow.GetParams`.

## Job Runners

jflow implementations (including Lathe) can support multiple backends:

1. **Local Runner**
   - Executes commands on the local machine.
   - Suitable for development and single-machine execution.
2. **TES Runner**
   - Executes tasks through GA4GH TES backends.
   - Suitable for cloud/HPC environments.

Both backends consume the same jflow workflow model.

## Notes and Best Practices

1. Use meaningful process names for easier debugging.
2. Prefer file-based dependencies over manual orchestration.
3. Set realistic resources (`cpu_cores`, `ram_gb`, `disk_gb`) for scheduling.
4. Use callbacks for post-job validation and reporting.
5. Use `Import` + explicit exports for modular workflows.
6. Validate user parameters early with `GetParams`.

## Implementations

### Lathe

[Lathe](https://github.com/bmeg/lathe) is the reference implementation of the jflow standard. It provides:

- Full jflow API support
- Local and TES (Task Execution Service) runners
- Docker/container support
- Resource management
- Callback system
- Parameter passing
- Module import composition

## Versioning

This document describes jflow version 1.0.

Future versions will evolve the standard and clearly document any breaking changes.

## Contributing

The jflow standard is open for community input. Implementation feedback and improvement proposals are welcome.

## License

The jflow standard specification is released under the Apache 2.0 License, allowing free implementation by any workflow engine.
