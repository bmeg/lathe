# Lathe Workflow Object Model - Schema Documentation

This document describes the workflow object model used by Lathe to declare jobs, tools, files, and workflows via JavaScript.

## Overview

The Lathe workflow engine is built around the concept of:
- **Files**: Data artifacts that can be inputs or outputs
- **Tools**: Reusable command templates with resource requirements
- **Jobs (Processes)**: Individual task executions with dependencies
- **Workflows**: Directed acyclic graphs (DAGs) of jobs connected by file dependencies
- **Futures**: Deferred results that resolve upon task completion
- **Callbacks**: Post-job analysis functions

All workflow declarations are made in JavaScript, which can call global functions provided by the `lathe` object to build the workflow model.

## Global API

### `lathe.Workflow(name: string): WorkflowDesc`

Creates a new named workflow. Returns a workflow object that can have steps added to it.

```javascript
const mainWorkflow = lathe.Workflow("main");
```

### `lathe.Process(spec: ProcessSpec): ProcessDesc`

Creates a job (process) that will be executed as part of the workflow. The spec object contains the job definition.

```javascript
const job1 = lathe.Process({
  name: "align_reads",
  commandLine: "bwa mem ref.fa input.fq > output.sam",
  image: "bwa:latest",
  inputs: {
    "ref": "reference.fa",
    "reads": "input.fq"
  },
  outputs: {
    "alignment": "output.sam"
  },
  cpus: 4,
  memoryMB: 8192
});
mainWorkflow.Add(job1);
```

### `lathe.File(spec: FileSpec): File`

Creates a reference to a file (input/output data).

```javascript
const inputFile = lathe.File({
  type: "local",
  path: "data/input.txt"
});

const s3File = lathe.File({
  type: "s3",
  path: "s3://bucket/data/file.txt"
});
```

### `lathe.FileCheck(spec: FileCheckSpec): FileCheck`

Creates a file existence check step, which ensures a file exists before dependent jobs run.

```javascript
const checkInput = lathe.FileCheck({
  file: {
    type: "local",
    path: "data/input.txt"
  }
});
mainWorkflow.Add(checkInput);
```

### `lathe.Tool(spec: ToolSpec): ToolCommand`

Creates a reusable tool/command template that can be instantiated multiple times.

```javascript
const bwaTool = lathe.Tool({
  name: "bwa_align",
  commandLine: "bwa mem {{ref}} {{reads}} > {{output}}",
  image: "bwa:latest",
  inputs: {
    "ref": "reference file",
    "reads": "reads file"
  },
  outputs: {
    "alignment": "SAM file"
  },
  resources: {
    cpus: 4,
    memoryMB: 8192,
    diskMB: 50000,
    timeout: 3600,
    retries: 1
  }
});
```

### `lathe.DockerImage(baseDir: string, tag: string, dockerfile?: string, buildArgs?: object): void`

Declares a Docker image specification. The image will be built before it's used.

```javascript
lathe.DockerImage("docker/bwa", "bwa:latest", "Dockerfile", {
  VERSION: "0.7.17"
});
```

### `lathe.LoadPlan(path: string): object`

Loads and executes a sub-workflow from an external JavaScript file. Returns a map of workflow names.

```javascript
const subWorkflows = lathe.LoadPlan("subworkflows/preprocessing.js");
mainWorkflow.Add(subWorkflows.alignReads);
```

### `lathe.Plugin(command: string): any`

Executes an external command and returns its JSON output. Useful for dynamic workflow generation.

```javascript
const config = lathe.Plugin("get_config --format json");
```

### `lathe.Params: object`

User parameters passed to the workflow. Can be used for dynamic configuration.

```javascript
const mode = lathe.Params.mode || "production";
const threads = lathe.Params.threads || 4;
```

## Type Specifications

### ProcessSpec

Specification for creating a job/process:

```typescript
interface ProcessSpec {
  // Required
  commandLine: string;        // Command to execute
  
  // Optional
  name?: string;              // Job name (auto-generated if not provided)
  description?: string;       // Human-readable description
  shell?: string;             // Shell interpreter (sh, bash, etc)
  image?: string;             // Docker image to run in
  inputs?: object;            // Map of input names to file paths
  outputs?: object;           // Map of output names to file paths
  cpus?: number;              // Number of CPU cores (default: 1)
  memoryMB?: number;          // Memory in MB (default: 1024)
}
```

### FileSpec

Specification for creating a file reference:

```typescript
interface FileSpec {
  // Required
  path: string;               // File path or URI
  
  // Optional
  type?: "local" | "s3" | "http";  // File location type (default: local)
  metadata?: object;          // Custom metadata
}
```

### FileCheckSpec

Specification for a file existence check:

```typescript
interface FileCheckSpec {
  file: FileSpec;             // File to check
}
```

### ToolSpec

Specification for creating a reusable tool:

```typescript
interface ToolSpec {
  // Required
  name: string;               // Unique tool name
  commandLine: string;        // Command template (supports {{var}} substitution)
  
  // Optional
  shell?: string;             // Shell interpreter
  image?: string;             // Docker image
  inputs?: object;            // Input specifications
  outputs?: object;           // Output specifications
  resources?: ResourceSpec;   // Resource requirements
  metadata?: object;          // Custom metadata
}
```

### ResourceSpec

Specification for compute resources:

```typescript
interface ResourceSpec {
  cpus?: number;              // Number of CPU cores
  memoryMB?: number;          // Memory in megabytes
  diskMB?: number;            // Disk space in megabytes
  gpus?: number;              // Number of GPU devices
  timeout?: number;           // Timeout in seconds
  retries?: number;           // Number of retries on failure
}
```

## Workflow Composition

### Adding Steps to a Workflow

```javascript
const wf = lathe.Workflow("myworkflow");

// Add a process
const job1 = lathe.Process({ /* ... */ });
wf.Add(job1);

// Add a file check
const check = lathe.FileCheck({ /* ... */ });
wf.Add(check);

// Add another job (automatically creates dependency on previous job through shared files)
const job2 = lathe.Process({
  name: "process2",
  inputs: { "data": "output.sam" }  // Depends on job1's output
});
wf.Add(job2);

// Add a sub-workflow
const sub = lathe.Workflow("sub");
const subJob = lathe.Process({ /* ... */ });
sub.Add(subJob);
wf.Add(sub);  // Inlines sub-workflow steps
```

### Dependency Resolution

Dependencies are automatically resolved based on:
1. **Explicit dependencies**: Can be specified in the job declaration
2. **File dependencies**: If job B's input matches job A's output, B depends on A
3. **File existence checks**: Jobs depend on any file checks that guarantee their inputs exist

## Futures and Callbacks

### Getting a Future for a Job

Every job returns a Future that will resolve when the job completes:

```javascript
const job = lathe.Process({
  name: "my_job",
  commandLine: "echo hello > output.txt",
  outputs: { "message": "output.txt" }
});

// When the workflow executes, this future will resolve
const jobFuture = job.GetFuture();
```

### Post-Job Analysis with Callbacks

Callbacks are functions executed after a job completes:

```javascript
const job = lathe.Process({
  name: "process_data",
  commandLine: "./process.sh",
  outputs: { "result": "result.json" }
});

// Register a callback (JavaScript function)
onComplete(job, function(result) {
  lathe.println("Job completed: " + result.jobName);
  lathe.println("Exit code: " + result.status.exitCode);
  
  if (result.status.state === "completed" && result.status.exitCode === 0) {
    lathe.println("Success!");
    // Can trigger additional workflows or analysis
  }
});

mainWorkflow.Add(job);
```

The callback receives a JobResult object:

```typescript
interface JobResult {
  jobName: string;            // Name of the completed job
  status: JobStatus;          // Final status
  outputFiles: object;        // Map of output names to Files
  logs: object;               // stdout and stderr logs
  metadata: object;           // Execution metadata
}

interface JobStatus {
  state: "completed" | "failed" | "cancelled";
  exitCode: number;
  error?: string;
  startTime?: Date;
  endTime?: Date;
  metadata?: object;
}
```

## Example Workflow

```javascript
// Declare workflow
const mainWf = lathe.Workflow("genomics_pipeline");

// Define input parameters
const refGenome = lathe.Params.reference || "hg38.fa";
const sampleFile = lathe.Params.sample || "sample.fq";

// Check that inputs exist
mainWf.Add(lathe.FileCheck({
  file: { path: refGenome }
}));

mainWf.Add(lathe.FileCheck({
  file: { path: sampleFile }
}));

// Job 1: Index reference
const indexJob = lathe.Process({
  name: "index_reference",
  commandLine: "bwa index " + refGenome,
  image: "bwa:latest",
  cpus: 2,
  memoryMB: 4096
});
mainWf.Add(indexJob);

// Job 2: Align reads
const alignJob = lathe.Process({
  name: "align_reads",
  commandLine: "bwa mem " + refGenome + " " + sampleFile + " > aligned.sam",
  image: "bwa:latest",
  inputs: {
    "genome_indexed": refGenome,
    "reads": sampleFile
  },
  outputs: {
    "alignment": "aligned.sam"
  },
  cpus: 4,
  memoryMB: 8192
});
mainWf.Add(alignJob);

// Register callback for alignment job
onComplete(alignJob, function(result) {
  if (result.status.exitCode === 0) {
    lathe.println("Alignment successful!");
  } else {
    lathe.println("Alignment failed!");
  }
});

// Job 3: Convert SAM to BAM
const convertJob = lathe.Process({
  name: "convert_to_bam",
  commandLine: "samtools view -b -o aligned.bam aligned.sam",
  image: "samtools:latest",
  inputs: {
    "sam_file": "aligned.sam"  // Automatic dependency on alignJob
  },
  outputs: {
    "bam_file": "aligned.bam"
  },
  cpus: 2,
  memoryMB: 4096
});
mainWf.Add(convertJob);
```

## Running Workflows

Workflows are executed using the runner:

```bash
# Execute workflow script with default parameters
lathe run workflow.js

# Execute with parameters
lathe run workflow.js --params mode=test --params threads=8
```

Parameters are accessible in the script as `lathe.Params`:

```javascript
const threads = lathe.Params.threads || 4;
const mode = lathe.Params.mode || "production";
```

## Job Runners

Lathe supports multiple job runners:

1. **Local Runner**: Executes jobs via `os.exec` on the local machine
   - Suitable for testing and single-machine deployments
   - Simple configuration with CPU/memory constraints

2. **TES Runner**: Executes jobs via the GA4GH Task Execution Service
   - Suitable for HPC and cloud deployments
   - Compatible with cloud providers (Google Cloud, AWS, Azure)
   - Supports containerized execution

Both runners are selected at execution time and receive the same workflow model.

## Notes and Best Practices

1. **Naming**: Always provide meaningful names to jobs for easier debugging
2. **Dependencies**: Let Lathe resolve dependencies through file inputs/outputs when possible
3. **Resources**: Specify accurate resource requirements for better scheduling
4. **Callbacks**: Use callbacks for post-job validation and dynamic workflows
5. **Modularity**: Use LoadPlan to organize large workflows into sub-workflows
6. **Error Handling**: Check job status in callbacks to implement error handling
