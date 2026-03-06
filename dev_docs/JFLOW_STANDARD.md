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

// Tool templates
jflow.Tool(spec: ToolSpec): ToolCommand

// Container management
jflow.DockerImage(baseDir: string, tag: string, dockerfile?: string, buildArgs?: object): void

// Workflow composition
jflow.LoadPlan(path: string): object

// Extensibility
jflow.Plugin(command: string): any

// Runtime parameters
jflow.Params: object
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

jflow supports both legacy format and TES (Task Execution Service) aligned format for maximum compatibility.

#### Legacy Format
```typescript
interface ProcessSpec {
  commandLine: string;      // Required: Command to execute (string or array)
  name?: string;            // Job name
  description?: string;     // Human-readable description
  shell?: string;           // Shell interpreter
  image?: string;           // Container image
  inputs?: object;          // Input file mappings (key-value)
  outputs?: object;         // Output file mappings (key-value)
  cpus?: number;            // CPU cores
  memoryMB?: number;        // Memory in MB
}
```

#### TES-Aligned Format (GA4GH TES API v1.1.0)
```typescript
interface ProcessSpec {
  name?: string;            // Task name
  description?: string;     // Task description
  
  // TES-aligned arrays
  executors: Executor[];    // Required: Array of executors to run sequentially
  inputs?: TESInput[];      // Array of input files
  outputs?: TESOutput[];    // Array of output files
  volumes?: string[];       // Shared volumes between executors
  
  // TES-aligned resources
  resources?: TESResources; // Compute resources
  tags?: object;            // Arbitrary key-value metadata
}

interface Executor {
  image: string;            // Required: Container image name
  command: string | string[]; // Required: Command (polymorphic: string or array)
  workdir?: string;         // Working directory
  stdin?: string;           // Stdin file path
  stdout?: string;          // Stdout file path
  stderr?: string;          // Stderr file path
  env?: object;             // Environment variables
  ignore_error?: boolean;   // Continue on error
}

interface TESInput {
  name?: string;            // Input name
  description?: string;     // Input description
  url?: string;             // Source URL (s3://, gs://, file://, http://)
  path: string;             // Required: Path in container (absolute)
  content?: string;         // File content literal (alternative to url)
  type?: "FILE" | "DIRECTORY";
}

interface TESOutput {
  name?: string;            // Output name
  description?: string;     // Output description
  url: string;              // Required: Destination URL
  path: string;             // Required: Path in container (absolute)
  path_prefix?: string;     // Prefix for wildcard outputs
  type?: "FILE" | "DIRECTORY";
}

interface TESResources {
  cpu_cores?: number;       // Number of CPU cores
  ram_gb?: number;          // Memory in gigabytes
  disk_gb?: number;         // Disk space in gigabytes
  preemptible?: boolean;    // Allow preemptible/spot instances
  zones?: string[];         // Compute zones
  backend_parameters?: object; // Backend-specific parameters
  backend_parameters_strict?: boolean; // Fail on unsupported params
}
```

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

// Array format (TES native)
jflow.Process({
  executors: [{
    image: "ubuntu:20.04",
    command: ["/bin/bash", "-c", "echo hello world > output.txt"]
  }]
})

// Legacy format also supports both
jflow.Process({
  commandLine: "echo hello",  // String format
  // OR
  commandLine: ["echo", "hello"]  // Array format
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
  inputs?: object;
  outputs?: object;
  resources?: ResourceSpec;
  metadata?: object;
}
```

## Implementation Requirements

A jflow-compliant workflow engine MUST:

1. Provide a JavaScript runtime (ES5 or higher)
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

## Compatibility

Implementations may choose to provide backward compatibility with legacy namespace conventions. For example, Lathe supports both `jflow.*` and `lathe.*` namespaces, making existing workflows compatible with the new standard.

## Example Workflow

### Legacy Format
```javascript
// jflow-compliant workflow (legacy format)
const pipeline = jflow.Workflow("data_pipeline");

// Check inputs
pipeline.Add(jflow.FileCheck({
  file: { path: "input.csv" }
}));

// Process data
const cleanJob = jflow.Process({
  name: "clean_data",
  commandLine: "python clean.py input.csv > cleaned.csv",
  image: "python:3.11",
  inputs: { raw: "input.csv" },
  outputs: { cleaned: "cleaned.csv" },
  cpus: 2,
  memoryMB: 4096
});

pipeline.Add(cleanJob);
```

### TES-Aligned Format
```javascript
// jflow workflow with TES-aligned format
const pipeline = jflow.Workflow("data_pipeline");

// Process data with TES format
const cleanJob = jflow.Process({
  name: "clean_data",
  description: "Clean and normalize input data",
  
  // TES executors (run sequentially)
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

// Multi-executor example (sequential execution)
const multiStepJob = jflow.Process({
  name: "multi_step_analysis",
  
  executors: [
    {
      image: "python:3.11",
      command: ["python", "preprocess.py", "/data/input.csv", "/data/preprocessed.csv"]
    },
    {
      image: "r-base:4.2",
      command: "Rscript analyze.R /data/preprocessed.csv /data/results.csv"
    },
    {
      image: "python:3.11",
      command: ["python", "visualize.py", "/data/results.csv", "/data/report.html"]
    }
  ],
  
  inputs: [{
    url: "s3://my-bucket/data.csv",
    path: "/data/input.csv"
  }],
  
  outputs: [{
    url: "s3://my-bucket/report.html",
    path: "/data/report.html"
  }],
  
  volumes: ["/data"]
});

pipeline.Add(multiStepJob);

// Add callback
onComplete(cleanJob, function(result) {
  if (result.status.exitCode === 0) {
    println("Data cleaning successful!");
  } else {
    println("Error: " + result.status.error);
  }
});
```

## Implementations

### Lathe

[Lathe](https://github.com/bmeg/lathe) is the reference implementation of the jflow standard. It provides:

- Full jflow API support
- Local and TES (Task Execution Service) runners
- Docker/container support
- Resource management
- Callback system
- Parameter passing
- Sub-workflow composition

## Versioning

This document describes jflow version 1.0.

Future versions will maintain backward compatibility where possible and clearly document any breaking changes.

## Contributing

The jflow standard is open for community input. Implementation feedback and improvement proposals are welcome.

## License

The jflow standard specification is released under the Apache 2.0 License, allowing free implementation by any workflow engine.
