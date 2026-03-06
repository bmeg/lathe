# Lathe Workflow Engine - Quick Start

This guide shows the current jflow prototype API using TES-aligned `jflow.Process` specs.

## Example 1: Hello World

```javascript
const wf = jflow.Workflow("hello");

const sayHello = jflow.Process({
  name: "greet",
  executors: [{
    image: "ubuntu:20.04",
    command: "echo 'Hello, World!' > /tmp/greeting.txt"
  }],
  outputs: [{ name: "greeting", path: "/tmp/greeting.txt" }],
  resources: { cpu_cores: 1, ram_gb: 0.25 }
});

wf.Add(sayHello);
```

Run with:

```bash
lathe run hello.js
```

## Example 2: File Dependency Pipeline

```javascript
const wf = jflow.Workflow("pipeline");

wf.Add(jflow.FileCheck({ file: { path: "data/input.csv" } }));

const cleanJob = jflow.Process({
  name: "clean_data",
  executors: [{
    image: "python:3.11",
    command: "python clean.py data/input.csv > data/cleaned.csv"
  }],
  inputs: [{ name: "raw", path: "data/input.csv" }],
  outputs: [{ name: "cleaned", path: "data/cleaned.csv" }],
  resources: { cpu_cores: 2, ram_gb: 1 }
});

const validateJob = jflow.Process({
  name: "validate",
  executors: [{
    image: "python:3.11",
    command: "python validate.py data/cleaned.csv > data/report.txt"
  }],
  inputs: [{ name: "data", path: "data/cleaned.csv" }],
  outputs: [{ name: "report", path: "data/report.txt" }],
  resources: { cpu_cores: 1, ram_gb: 0.5 }
});

wf.Add(cleanJob);
wf.Add(validateJob);
```

## Example 3: Callbacks

```javascript
const wf = jflow.Workflow("analysis");

const job = jflow.Process({
  name: "compute_stats",
  executors: [{
    image: "python:3.11",
    command: "python stats.py > /tmp/output.json"
  }],
  outputs: [{ name: "result", path: "/tmp/output.json" }],
  resources: { cpu_cores: 2, ram_gb: 2 }
});

onComplete(job, function(result) {
  println("Name: " + result.jobName);
  println("State: " + result.status.state);
  println("Exit code: " + result.status.exitCode);
  if (result.status.state === "COMPLETE" && result.status.exitCode === 0) {
    println("SUCCESS");
  }
});

wf.Add(job);
```

## Example 4: Parameters and Type Validation

```javascript
const params = jflow.GetParams({
  sample: "String",
  threads: "Number",
  reads: "File"
});

const wf = jflow.Workflow("parameterized");

const align = jflow.Process({
  name: "align_" + params.sample,
  executors: [{
    image: "bwa:latest",
    command: `bwa mem -t ${params.threads} ref.fa ${params.reads} > /tmp/${params.sample}.sam`
  }],
  inputs: [{ name: "reads", path: params.reads }],
  outputs: [{ name: "sam", path: `/tmp/${params.sample}.sam` }],
  resources: { cpu_cores: params.threads, ram_gb: 4 }
});

wf.Add(align);
```

Run with:

```bash
lathe run parameterized.js --params sample=sample1 --params threads=8 --params reads=/abs/path/sample1.fq
```

## Example 5: Import Modules

```javascript
// main_workflow.js
const prep = jflow.Import("preprocessing.js");
const analysis = jflow.Import("analysis.js");

const wf = jflow.Workflow("main");
wf.Add(prep.preprocess);
wf.Add(analysis.analyze);
```

```javascript
// preprocessing.js
const prepWf = jflow.Workflow("preprocess");
prepWf.Add(jflow.Process({
  name: "clean_data",
  executors: [{ image: "python:3.11", command: "python clean.py input.raw > input.clean" }],
  outputs: [{ name: "cleaned", path: "input.clean" }]
}));
export { prepWf as preprocess };
```

## Notes

- Use TES-aligned `Process` fields: `executors`, `inputs`, `outputs`, `resources`, `tags`.
- For command polymorphism, `executor.command` can be either a string or string array.
- Local runner currently executes the first executor in a process.
- See `examples/` for larger, runnable workflows.
