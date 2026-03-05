# Lathe Workflow Engine - Integration Guide

This guide describes how to use the refactored Lathe workflow engine to build and execute workflows using JavaScript declarations and futures-based callbacks for post-job analysis.

## Architecture Overview

### Core Components

The refactored Lathe engine consists of:

1. **Workflow Object Model** (`scriptfile/model.go`)
   - Centralized type definitions
   - `File`, `ProcessDesc`, `WorkflowDesc`, `ToolCommand`
   - `Future[T]` for deferred results
   - Job status tracking and callbacks

2. **JavaScript Environment** (`scriptfile/js_vm.go`)
   - Goja-based JavaScript runtime
   - Global `lathe` API for workflow declarations
   - Parameter passing and dynamic workflow generation

3. **API Functions** (`scriptfile/api.go`)
   - Job/process declaration via `lathe.Process()`
   - File declaration via `lathe.File()`
   - Tool templates via `lathe.Tool()`
   - Callback registration via `onComplete()`
   - Plugin system for external integration

4. **Workflow Composition** (`scriptfile/workflow.go`)
   - Workflow creation and step management
   - Automatic dependency resolution through file inputs/outputs
   - Support for inlining sub-workflows

5. **Job Execution** (`runner/`)
   - `SingleMachineRunner`: Local execution via `os.exec`
   - `TESRunner`: GA4GH TES API for cloud/HPC
   - Resource constraint management

## Using the Workflow Model

### 1. Basic Workflow Declaration

```javascript
// Declare a workflow
const mainWorkflow = lathe.Workflow("main");

// Create a job
const job1 = lathe.Process({
  name: "hello_world",
  commandLine: "echo 'Hello, World!'",
  cpus: 1,
  memoryMB: 256
});

// Add job to workflow
mainWorkflow.Add(job1);
```

### 2. File Input/Output Dependencies

```javascript
const wf = lathe.Workflow("data_pipeline");

// Define input file check
wf.Add(lathe.FileCheck({
  file: { path: "data/input.txt" }
}));

// Job 1: Process input
const processJob = lathe.Process({
  name: "process_data",
  commandLine: "cat data/input.txt | tr a-z A-Z > data/output.txt",
  inputs: { "raw": "data/input.txt" },
  outputs: { "processed": "data/output.txt" },
  cpus: 1,
  memoryMB: 512
});
wf.Add(processJob);

// Job 2: Validate output (auto-depends on processJob through file dependency)
const validateJob = lathe.Process({
  name: "validate",
  commandLine: "wc -l data/output.txt > data/stats.txt",
  inputs: { "data": "data/output.txt" },
  outputs: { "stats": "data/stats.txt" },
  cpus: 1,
  memoryMB: 256
});
wf.Add(validateJob);
```

### 3. Using Futures and Callbacks

```javascript
const wf = lathe.Workflow("with_callbacks");

// Create a job
const myJob = lathe.Process({
  name: "important_task",
  commandLine: "./script.sh > results.json",
  outputs: { "results": "results.json" },
  cpus: 2,
  memoryMB: 1024
});

// Register callback for post-job analysis
onComplete(myJob, function(result) {
  lathe.println("Job finished: " + result.jobName);
  
  if (result.status.state === "completed") {
    lathe.println("Exit code: " + result.status.exitCode);
    
    if (result.status.exitCode === 0) {
      lathe.println("SUCCESS: Job completed successfully");
      // Can trigger follow-up workflows or analysis
    } else {
      lathe.println("ERROR: Job failed with exit code " + result.status.exitCode);
    }
  } else {
    lathe.println("FAILED: " + result.status.error);
  }
  
  // Log execution details
  for (const key in result.metadata) {
    lathe.println(key + ": " + result.metadata[key]);
  }
});

wf.Add(myJob);

// Get future for programmatic access (if needed)
// const future = myJob.GetFuture();
// When executed, future will resolve with JobResult
```

### 4. Docker Container Execution

```javascript
const wf = lathe.Workflow("containerized");

// Declare Docker image
lathe.DockerImage("docker/bwa", "bwa:latest", "Dockerfile", {
  VERSION: "0.7.17",
  PREFIX: "/usr/local"
});

// Create job that runs in container
const bwaJob = lathe.Process({
  name: "bwa_align",
  commandLine: "bwa mem -t 8 reference.fa reads.fq > output.sam",
  image: "bwa:latest",
  inputs: {
    "ref": "reference.fa",
    "reads": "reads.fq"
  },
  outputs: {
    "alignment": "output.sam"
  },
  cpus: 8,
  memoryMB: 16384
});

wf.Add(bwaJob);
```

### 5. Parameterized Workflows

```javascript
// Access user parameters via lathe.Params
const referencePath = lathe.Params.reference || "default_ref.fa";
const samplePath = lathe.Params.sample || "default_sample.fq";
const threads = lathe.Params.threads || 4;
const memory = lathe.Params.memory || 4096;

const wf = lathe.Workflow("parameterized");

// Use parameters in job definitions
const alignJob = lathe.Process({
  name: "align",
  commandLine: `bwa mem -t ${threads} ${referencePath} ${samplePath} > output.sam`,
  cpus: threads,
  memoryMB: memory
});

wf.Add(alignJob);
```

### 6. Reusable Tool Templates

```javascript
// Define a tool template
const samtoolsTool = lathe.Tool({
  name: "samtools_view",
  commandLine: "samtools view {{args}} {{input}} > {{output}}",
  image: "samtools:latest",
  inputs: {
    "input": "SAM file"
  },
  outputs: {
    "output": "BAM file"
  },
  resources: {
    cpus: 2,
    memoryMB: 4096,
    timeout: 3600
  }
});

// Tools can be referenced in documentation or used to
// generate jobs dynamically
```

### 7. Loading Sub-Workflows

```javascript
// In main workflow script: workflow.js
const wf = lathe.Workflow("main");

// Load preprocessing sub-workflow
const preprocessWfs = lathe.LoadPlan("preprocessing.js");
wf.Add(preprocessWfs.preprocess);

// Load analysis sub-workflow
const analysisWfs = lathe.LoadPlan("analysis.js");
wf.Add(analysisWfs.analyze);

// preprocessing.js
const prepWf = lathe.Workflow("preprocess");
const cleanJob = lathe.Process({
  name: "clean_input",
  commandLine: "clean_data.sh"
});
prepWf.Add(cleanJob);

// analysis.js
const anaWf = lathe.Workflow("analyze");
const statsJob = lathe.Process({
  name: "compute_stats",
  commandLine: "compute_statistics.sh"
});
anaWf.Add(statsJob);
```

### 8. Dynamic Workflow Generation

```javascript
// Use plugins to generate workflow configuration
const config = lathe.Plugin("get_samples --json");

const wf = lathe.Workflow("dynamic_samples");

// Create a job for each sample
for (const sample of config.samples) {
  const job = lathe.Process({
    name: sample.name,
    commandLine: `process_sample.sh ${sample.id} > ${sample.name}.txt`,
    outputs: { result: `${sample.name}.txt` },
    cpus: sample.cpus || 2,
    memoryMB: sample.memory || 2048
  });
  
  wf.Add(job);
}
```

## Executing Workflows

### Command Line Usage

```bash
# Execute workflow with default parameters
lathe run workflow.js

# Execute with custom parameters
lathe run workflow.js \
  --params reference=hg38.fa \
  --params sample=sample1.fq \
  --params threads=8

# Dry-run (parse only, don't execute)
lathe run workflow.js --dry-run

# Use local runner
lathe run --runner local workflow.js

# Use TES runner with cloud backend
lathe run --runner tes --tes-url http://tes-server:8000 workflow.js
```

### Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/bmeg/lathe/scriptfile"
)

func main() {
    // Parse workflow script with parameters
    params := map[string]any{
        "reference": "hg38.fa",
        "sample": "sample1.fq",
        "threads": 8,
    }
    
    plan, err := scriptfile.RunFileWithParams("workflow.js", params)
    if err != nil {
        panic(err)
    }
    
    // Access parsed workflows
    for name, workflow := range plan.Workflows {
        fmt.Printf("Workflow: %s\n", name)
        fmt.Printf("  Steps: %d\n", len(workflow.Steps))
    }
    
    // Build execution DAG
    // (implementation depends on runner)
}
```

## Implementation Details

### How Futures Work

1. **Creation**: Every `ProcessDesc` has an associated `Future[*JobResult]`
   ```go
   proc.future = NewFuture[*JobResult]()
   ```

2. **Deferred Resolution**: The future is returned immediately from `lathe.Process()`
   ```javascript
   const job = lathe.Process({...});
   // Job doesn't execute yet - just creates the declaration
   ```

3. **Resolution**: When the job executes and completes, the runner resolves the future
   ```go
   result := &JobResult{...}
   proc.GetFuture().Resolve(result)
   ```

4. **Callback Execution**: After resolution, registered callbacks are invoked
   ```go
   proc.ExecuteCallback(result)
   ```

### Dependency Resolution

The workflow engine automatically resolves dependencies through:

1. **File-based dependencies**: If job B's inputs match job A's outputs, B depends on A
2. **Explicit dependencies**: Can be specified in job declaration
3. **File existence checks**: Jobs depend on any file checks that guarantee their inputs exist

### Callback Execution Context

Callbacks receive a `JobResult` object containing:
- Job name and final status
- Exit code and error messages
- Output files and their paths
- Execution logs (stdout/stderr)
- Execution metadata (timing, resources used, etc.)

## Best Practices

### 1. Naming and Documentation
- Always provide meaningful job names
- Use clear input/output descriptions
- Add comments explaining complex workflows

### 2. Resource Specification
- Provide accurate CPU and memory requirements
- This helps with scheduling and resource management
- Especially important for cloud/HPC runners

### 3. Error Handling
- Check job status in callbacks
- Implement retry logic for transient failures
- Log results for debugging

### 4. Modularity
- Break large workflows into sub-workflows
- Use LoadPlan to organize related workflows
- Reuse common patterns via Tool definitions

### 5. Testing
- Use `--dry-run` to validate workflow structure
- Test with small datasets first
- Use logging to trace execution

## Migration from Old API

### Old Style (Before Refactoring)

```javascript
const wf = lathe.Workflow("main");
const job = lathe.Process({commandLine: "...", memMB: 1024});
wf.Add(job);
```

### New Style (After Refactoring)

```javascript
const wf = lathe.Workflow("main");
const job = lathe.Process({
  commandLine: "...",
  memoryMB: 1024,  // Changed from memMB to memoryMB
  cpus: 1          // Now consistent naming
});

// Optional: Add callback for post-job analysis
onComplete(job, function(result) {
  lathe.println("Job completed: " + result.jobName);
});

wf.Add(job);
```

### Key Changes

| Old | New | Reason |
|-----|-----|--------|
| `memMB` | `memoryMB` | Consistency and clarity |
| `ncpus` | `cpus` | Shorter, clearer naming |
| No callbacks | `onComplete()` | Support for post-job analysis |
| Basic File | Enhanced File with type | Support for S3, HTTP, etc. |
| No futures | `Future[T]` | Deferred result handling |
| Generic metadata | Structured types | Type safety and clarity |

## Troubleshooting

### "Unknown object type in Workflow.Add"
- Ensure you're passing ProcessDesc, FileCheck, or WorkflowDesc
- Check that Process() was called to create the job definition

### "File not found" in callback
- Ensure output files are created during job execution
- Check job exit code first (exitCode != 0 means failure)

### Parameters not being passed
- Use `lathe.Params.paramName` (not capital P in Params object)
- Parameters are passed via command line: `--params key=value`

### Dependency not resolved
- Ensure output file path exactly matches input file path
- File checks must be added before dependent jobs
- Use explicit dependency specification if needed

## See Also

- [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) - Complete API documentation
- [scriptfile/model.go](scriptfile/model.go) - Type definitions and interfaces
- [scriptfile/js_vm.go](scriptfile/js_vm.go) - JavaScript environment setup
- [scriptfile/api.go](scriptfile/api.go) - API function implementations
