# TES API Alignment Guide

## Overview

The jflow workflow object model has been refactored to align with the **GA4GH Task Execution Service (TES) API v1.1.0** specification. This provides:

- **Standardization**: Compatibility with the widely-adopted TES API standard
- **Interoperability**: Easier integration with TES-compliant systems
- **Flexibility**: Support for both legacy jflow format and modern TES format
- **User-Friendly**: Polymorphic types allow strings or arrays for commands

## Key Concepts

### TES Task Model

The TES API defines tasks as the unit of work with:

- **Executors**: Sequential commands run in containers
- **Inputs**: Files downloaded and mounted into containers
- **Outputs**: Files uploaded from containers to storage
- **Resources**: Compute requirements (CPU, RAM, disk)
- **Volumes**: Shared directories between executors

### Mapping: jflow → TES

| jflow Legacy | TES Equivalent | Notes |
|--------------|----------------|-------|
| `ProcessDesc` | `tesTask` | Task/job definition |
| `commandLine` (string) | `executors[].command` (array) | Single executor |
| `inputs` (map) | `inputs[]` (array) | Key-value → array of objects |
| `outputs` (map) | `outputs[]` (array) | Key-value → array of objects |
| `image` (string) | `executors[].image` | Per-executor images |
| `cpus`, `memoryMB` | `resources.cpu_cores`, `resources.ram_gb` | Normalized units |

## TES Format

### Basic Task Structure

```javascript
const task = jflow.Process({
  name: "example_task",
  description: "Process genomic data",
  
  // Executors: commands to run sequentially
  executors: [{
    image: "ubuntu:20.04",
    command: ["echo", "hello", "world"],  // Array format
    workdir: "/workspace",
    env: { "VAR": "value" }
  }],
  
  // Inputs: files to download
  inputs: [{
    url: "s3://bucket/input.txt",
    path: "/data/input.txt",
    type: "FILE"
  }],
  
  // Outputs: files to upload
  outputs: [{
    url: "s3://bucket/output.txt",
    path: "/data/output.txt",
    type: "FILE"
  }],
  
  // Resources: compute requirements
  resources: {
    cpu_cores: 4,
    ram_gb: 8,
    disk_gb: 50
  },
  
  // Volumes: shared between executors
  volumes: ["/data"],
  
  // Tags: metadata
  tags: {
    "project": "genomics",
    "sample": "sample-001"
  }
});
```

## Polymorphic Command Support

Commands can be specified as strings or arrays:

### String Commands (User-Friendly)

```javascript
executors: [{
  image: "python:3.11",
  command: "python script.py input.txt > output.txt"
}]
// Automatically converted to: ["/bin/sh", "-c", "python script.py input.txt > output.txt"]
```

### Array Commands (TES Native)

```javascript
executors: [{
  image: "python:3.11",
  command: ["python", "script.py", "input.txt"]
}]
// Used directly without shell wrapper
```

### When to Use Each

- **String**: Shell features needed (pipes, redirects, wildcards)
- **Array**: Precise argument control, avoid shell escaping issues

## Multi-Executor Tasks

TES supports sequential execution of multiple containers:

```javascript
jflow.Process({
  name: "pipeline",
  
  executors: [
    {
      image: "python:3.11",
      command: "python preprocess.py /data/input.csv /data/prep.csv"
    },
    {
      image: "r-base:4.2",
      command: "Rscript analyze.R /data/prep.csv /data/results.csv"
    },
    {
      image: "python:3.11",
      command: "python visualize.py /data/results.csv /data/report.html"
    }
  ],
  
  volumes: ["/data"],  // Shared across all executors
  
  inputs: [{
    url: "s3://bucket/input.csv",
    path: "/data/input.csv"
  }],
  
  outputs: [{
    url: "s3://bucket/report.html",
    path: "/data/report.html"
  }]
});
```

## Resource Specifications

### TES-Aligned (Recommended)

```javascript
resources: {
  cpu_cores: 4,           // Number of CPU cores
  ram_gb: 8.0,            // Memory in gigabytes
  disk_gb: 100.0,         // Disk in gigabytes
  preemptible: true,      // Allow spot/preemptible instances
  zones: ["us-west-1"],   // Preferred compute zones
  backend_parameters: {   // Backend-specific settings
    "VmSize": "Standard_D4_v3"
  }
}
```

### Legacy Format (Still Supported)

```javascript
cpus: 4,
memoryMB: 8192,
diskMB: 102400
```

Legacy fields are automatically converted to TES format.

## Input/Output Specifications

### TES Format

```javascript
inputs: [
  {
    name: "reference_genome",
    description: "Human reference genome FASTA",
    url: "s3://genomics-data/hg38.fa",
    path: "/reference/genome.fa",
    type: "FILE"
  },
  {
    name: "sample_reads",
    url: "gs://samples/sample-001.fastq.gz",
    path: "/data/reads.fastq.gz"
  },
  {
    name: "config",
    content: "key=value\nflag=true",  // Inline content
    path: "/config/settings.conf"
  }
]

outputs: [
  {
    name: "alignment",
    url: "s3://results/sample-001.bam",
    path: "/output/alignment.bam",
    type: "FILE"
  },
  {
    name: "logs",
    url: "s3://logs/sample-001/",
    path: "/output/*.log",  // Wildcard pattern
    path_prefix: "/output/",
    type: "FILE"
  }
]
```

### Legacy Format

```javascript
inputs: {
  "reference": "/reference/genome.fa",
  "reads": "/data/reads.fastq.gz"
}

outputs: {
  "alignment": "/output/alignment.bam"
}
```

## Migration Examples

### Example 1: Simple Command

**Before (Legacy)**
```javascript
jflow.Process({
  name: "sort_file",
  commandLine: "sort input.txt > output.txt",
  image: "ubuntu:20.04",
  inputs: { input: "input.txt" },
  outputs: { output: "output.txt" },
  cpus: 2,
  memoryMB: 2048
})
```

**After (TES)**
```javascript
jflow.Process({
  name: "sort_file",
  executors: [{
    image: "ubuntu:20.04",
    command: "sort /data/input.txt > /data/output.txt"
  }],
  inputs: [{
    url: "file:///workspace/input.txt",
    path: "/data/input.txt"
  }],
  outputs: [{
    url: "file:///workspace/output.txt",
    path: "/data/output.txt"
  }],
  resources: {
    cpu_cores: 2,
    ram_gb: 2
  }
})
```

### Example 2: Multi-Step Pipeline

**TES Format (New Capability)**
```javascript
jflow.Process({
  name: "quality_control",
  
  executors: [
    {
      image: "biocontainers/fastqc:v0.11.9",
      command: ["fastqc", "/data/reads.fastq", "-o", "/data/qc"]
    },
    {
      image: "biocontainers/trimmomatic:0.39",
      command: "trimmomatic SE /data/reads.fastq /data/trimmed.fastq LEADING:20"
    },
    {
      image: "biocontainers/fastqc:v0.11.9",
      command: ["fastqc", "/data/trimmed.fastq", "-o", "/data/qc"]
    }
  ],
  
  inputs: [{
    url: "s3://reads/sample.fastq",
    path: "/data/reads.fastq"
  }],
  
  outputs: [
    {
      url: "s3://trimmed/sample.fastq",
      path: "/data/trimmed.fastq"
    },
    {
      url: "s3://qc/",
      path: "/data/qc/*",
      path_prefix: "/data/qc/"
    }
  ],
  
  volumes: ["/data"],
  
  resources: {
    cpu_cores: 4,
    ram_gb: 8
  }
})
```

## Backward Compatibility

The implementation maintains full backward compatibility:

1. **Legacy format still works**: Old workflows continue to run
2. **Automatic normalization**: Legacy fields are converted to TES internally
3. **Mixed format**: Can use TES features with legacy syntax where convenient

```javascript
// This works: mix of legacy and TES
jflow.Process({
  commandLine: "python script.py",  // Legacy
  image: "python:3.11",             // Legacy
  resources: {                       // TES
    cpu_cores: 4,
    ram_gb: 8
  },
  tags: {                           // TES
    "version": "1.0"
  }
})
```

## Best Practices

1. **Use TES format for new workflows**: Better standardization and features
2. **Leverage polymorphic commands**: Use strings for simple shell commands, arrays for complex ones
3. **Specify absolute paths**: TES requires absolute container paths (e.g., `/data/file.txt`)
4. **Use volumes for multi-executor tasks**: Share data between sequential executors
5. **Tag your tasks**: Use tags for tracking, filtering, and organization
6. **Prefer TES resource units**: Use `cpu_cores` and `ram_gb` over legacy fields

## TES State Model

TES defines standardized task states:

- `UNKNOWN`: State unknown
- `QUEUED`: Waiting for resources
- `INITIALIZING`: Preparing to run
- `RUNNING`: Currently executing
- `PAUSED`: Execution paused
- `COMPLETE`: Successfully finished
- `EXECUTOR_ERROR`: Command failed
- `SYSTEM_ERROR`: Infrastructure error
- `CANCELED`: User canceled
- `CANCELING`: Being canceled
- `PREEMPTED`: System preempted

These states are mapped from jflow's legacy states automatically.

## Reference

- **TES Specification**: https://ga4gh.github.io/task-execution-schemas/docs/
- **TES OpenAPI**: https://github.com/ga4gh/task-execution-schemas
- **GA4GH**: https://www.ga4gh.org/

## Support

For questions or issues with TES alignment, please file an issue on the project repository.
