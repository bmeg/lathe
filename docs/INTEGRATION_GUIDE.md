# Lathe Workflow Engine - Integration Guide

This guide describes how to integrate with Lathe using the current TES-aligned `jflow` API.

## Architecture Overview

Core packages:

1. `jflow/`
   - JavaScript runtime and API bindings
   - Workflow model types (`ProcessDesc`, `WorkflowDesc`, `File`, `ToolCommand`)
2. `workflow/`
   - DAG construction from declared steps
   - Dependency resolution via inputs/outputs and file checks
3. `runner/`
   - Local execution (`SingleMachineRunner`)
   - TES execution (`TESRunner`)

## API Surface

Primary globals available in workflow scripts:

- `jflow.Workflow(name)`
- `jflow.Process(spec)`
- `jflow.File(spec)`
- `jflow.FileCheck(spec)`
- `jflow.Path(name)`
- `jflow.Object(name)`
- `jflow.Tool(spec)`
- `jflow.Import(path, paramsOverride?)`
- `jflow.Plugin(command)`
- `jflow.GetParams(schema)`
- `jflow.Params`
- `onComplete(job, callback)`

## Process Specification (TES-aligned)

```javascript
const job = jflow.Process({
  name: "analyze",
  description: "Analyze one sample",
  executors: [{
    image: "python:3.11",
    command: "python analyze.py /data/input.csv > /data/output.csv"
  }],
  inputs: [{ name: "input", path: "/data/input.csv" }],
  outputs: [{ name: "output", path: "/data/output.csv" }],
  resources: { cpu_cores: 4, ram_gb: 8, disk_gb: 50 },
  tags: { project: "demo" }
});
```

Notes:
- `executor.command` supports string or string array.
- Multiple executors can be declared; local execution currently runs the first executor.

## Dependency Resolution

Dependencies are inferred automatically when:

- A process input path equals another process output path.
- A process input path is guarded by a `jflow.FileCheck` step.

Example:

```javascript
const wf = jflow.Workflow("pipeline");

wf.Add(jflow.FileCheck({ file: { path: "data/input.txt" } }));

const step1 = jflow.Process({
  name: "step1",
  executors: [{ image: "ubuntu:20.04", command: "cat data/input.txt > data/out.txt" }],
  inputs: [{ name: "in", path: "data/input.txt" }],
  outputs: [{ name: "out", path: "data/out.txt" }]
});

const step2 = jflow.Process({
  name: "step2",
  executors: [{ image: "ubuntu:20.04", command: "wc -l data/out.txt > data/stats.txt" }],
  inputs: [{ name: "out", path: "data/out.txt" }],
  outputs: [{ name: "stats", path: "data/stats.txt" }]
});

wf.Add(step1);
wf.Add(step2);
```

## Parameters

Use typed parameter validation for robust workflows:

```javascript
const params = jflow.GetParams({
  sample: "String",
  threads: "Number",
  reads: "File"
});
```

`File` parameters are normalized to absolute paths for local files.

## Tool Templates

`jflow.Tool` returns a callable process factory:

```javascript
const bwaMem = jflow.Tool({
  name: "bwa_mem",
  commandLine: "bwa mem -t {{threads}} {{ref}} {{r1}} {{r2}} > {{bam}}",
  image: "quay.io/biocontainers/bwa:0.7.17--hed695b0_7",
  inputs: {
    ref: "File",
    r1: "File",
    r2: "File",
    threads: "Value",
    bam: "Value"
  },
  outputs: { bam: "{{bam}}" },
  resources: { cpu_cores: 4, ram_gb: 8 }
});
```

## Module Imports

Use `jflow.Import` with explicit exports:

```javascript
// main.js
const mod = jflow.Import("./module.js", { threads: 8 });
const wf = jflow.Workflow("main");
wf.Add(mod.align);
```

```javascript
// module.js
const align = jflow.Process({
  name: "align",
  executors: [{ image: "ubuntu:20.04", command: "echo ok > /tmp/ok.txt" }],
  outputs: [{ name: "ok", path: "/tmp/ok.txt" }]
});

export { align };
```

## Callbacks

```javascript
onComplete(job, function(result) {
  if (result.status.state === "COMPLETE" && result.status.exitCode === 0) {
    println("Job succeeded: " + result.jobName);
  } else {
    println("Job failed: " + result.status.error);
  }
});
```

## Running Workflows

```bash
lathe run workflow.js
lathe run workflow.js --params sample=s1 --params threads=8
lathe run workflow.js --params-file params.yaml
lathe run workflow.js --tes http://tes-server:8000
```

## Programmatic Use (Go)

```go
package main

import (
    "fmt"
    "github.com/bmeg/lathe/jflow"
)

func main() {
    params := map[string]any{"threads": 8}
    plan, err := jflow.RunFileWithParams("workflow.js", params)
    if err != nil {
        panic(err)
    }
    fmt.Println("workflows:", len(plan.Workflows))
}
```
