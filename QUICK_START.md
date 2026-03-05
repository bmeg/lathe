# Lathe Workflow Engine - Quick Start Examples

This file contains practical examples to get started with the refactored Lathe workflow engine.

## Example 1: Hello World

**File: hello.js**

```javascript
const wf = lathe.Workflow("hello");

const sayHello = lathe.Process({
  name: "greet",
  commandLine: "echo 'Hello, World!' > greeting.txt",
  outputs: { greeting: "greeting.txt" },
  cpus: 1,
  memoryMB: 256
});

wf.Add(sayHello);
```

**Run:**
```bash
lathe run hello.js
```

## Example 2: Data Pipeline with Dependencies

**File: pipeline.js**

```javascript
const wf = lathe.Workflow("pipeline");

// Ensure input exists
wf.Add(lathe.FileCheck({
  file: { path: "data/input.csv" }
}));

// Step 1: Clean data
const cleanJob = lathe.Process({
  name: "clean_data",
  commandLine: "python clean.py data/input.csv > data/cleaned.csv",
  image: "python:3.11",
  inputs: { raw: "data/input.csv" },
  outputs: { cleaned: "data/cleaned.csv" },
  cpus: 2,
  memoryMB: 1024
});
wf.Add(cleanJob);

// Step 2: Validate cleaned data
const validateJob = lathe.Process({
  name: "validate",
  commandLine: "python validate.py data/cleaned.csv > data/report.txt",
  image: "python:3.11",
  inputs: { data: "data/cleaned.csv" },  // Auto-depends on cleanJob
  outputs: { report: "data/report.txt" },
  cpus: 1,
  memoryMB: 512
});
wf.Add(validateJob);
```

**Run:**
```bash
lathe run pipeline.js
```

## Example 3: Using Callbacks for Post-Job Analysis

**File: with_callbacks.js**

```javascript
const wf = lathe.Workflow("analysis");

const job = lathe.Process({
  name: "compute_stats",
  commandLine: "python stats.py > output.json",
  image: "python:3.11",
  outputs: { result: "output.json" },
  cpus: 2,
  memoryMB: 2048
});

// Callback for post-job validation
onComplete(job, function(result) {
  lathe.println("=== Job Completed ===");
  lathe.println("Name: " + result.jobName);
  lathe.println("Status: " + result.status.state);
  
  if (result.status.state === "completed") {
    lathe.println("Exit Code: " + result.status.exitCode);
    
    if (result.status.exitCode === 0) {
      lathe.println("✓ SUCCESS");
    } else {
      lathe.println("✗ FAILED");
      if (result.status.error) {
        lathe.println("Error: " + result.status.error);
      }
    }
  } else {
    lathe.println("✗ " + result.status.error);
  }
  
  // Log execution time
  if (result.status.startTime && result.status.endTime) {
    const duration = result.status.endTime - result.status.startTime;
    lathe.println("Duration: " + duration + "ms");
  }
});

wf.Add(job);
```

**Run:**
```bash
lathe run with_callbacks.js
```

## Example 4: Parameterized Workflow

**File: parameterized.js**

```javascript
// Get parameters from command line or use defaults
const sampleId = lathe.Params.sample || "default";
const threads = lathe.Params.threads || 4;
const memory = lathe.Params.memory || 2048;

lathe.println("Running analysis for sample: " + sampleId);
lathe.println("Using " + threads + " threads");

const wf = lathe.Workflow("analysis");

const alignJob = lathe.Process({
  name: "align_" + sampleId,
  commandLine: `bwa mem -t ${threads} ref.fa ${sampleId}.fq > ${sampleId}.sam`,
  image: "bwa:latest",
  cpus: threads,
  memoryMB: memory
});

wf.Add(alignJob);
```

**Run:**
```bash
lathe run parameterized.js --params sample=sample1 --params threads=8 --params memory=4096
```

## Example 5: Multi-Stage Workflow with Error Handling

**File: multi_stage.js**

```javascript
const wf = lathe.Workflow("genomics");

// Stage 1: Index reference genome
const indexJob = lathe.Process({
  name: "index_genome",
  commandLine: "bwa index reference.fa",
  image: "bwa:latest",
  inputs: { ref: "reference.fa" },
  cpus: 2,
  memoryMB: 4096
});

onComplete(indexJob, function(result) {
  if (result.status.exitCode !== 0) {
    lathe.println("ERROR: Genome indexing failed!");
    lathe.println(result.status.error);
  } else {
    lathe.println("✓ Genome indexed successfully");
  }
});

wf.Add(indexJob);

// Stage 2: Align reads
const alignJob = lathe.Process({
  name: "align_reads",
  commandLine: "bwa mem reference.fa reads.fq > output.sam",
  image: "bwa:latest",
  inputs: { 
    ref: "reference.fa",
    reads: "reads.fq"
  },
  outputs: { alignment: "output.sam" },
  cpus: 4,
  memoryMB: 8192
});

onComplete(alignJob, function(result) {
  if (result.status.exitCode === 0) {
    lathe.println("✓ Read alignment completed");
  } else {
    lathe.println("✗ Read alignment failed");
  }
});

wf.Add(alignJob);

// Stage 3: Convert to BAM (depends on alignment output)
const convertJob = lathe.Process({
  name: "convert_bam",
  commandLine: "samtools view -b -o output.bam output.sam",
  image: "samtools:latest",
  inputs: { sam: "output.sam" },
  outputs: { bam: "output.bam" },
  cpus: 2,
  memoryMB: 2048
});

onComplete(convertJob, function(result) {
  if (result.status.exitCode === 0) {
    lathe.println("✓ BAM conversion completed");
    lathe.println("Output: output.bam");
  }
});

wf.Add(convertJob);
```

**Run:**
```bash
lathe run multi_stage.js
```

## Example 6: Sub-Workflows and Composition

**File: main_workflow.js**

```javascript
// Load sub-workflows from other files
const prepWfs = lathe.LoadPlan("preprocessing.js");
const analysisWfs = lathe.LoadPlan("analysis.js");

// Create main workflow
const mainWf = lathe.Workflow("main");

// Add sub-workflows
mainWf.Add(prepWfs.preprocess);
mainWf.Add(analysisWfs.analyze);
```

**File: preprocessing.js**

```javascript
const prepWf = lathe.Workflow("preprocess");

const cleanJob = lathe.Process({
  name: "clean_data",
  commandLine: "python clean.py input.raw > input.clean",
  image: "python:3.11",
  outputs: { cleaned: "input.clean" }
});

prepWf.Add(cleanJob);
```

**File: analysis.js**

```javascript
const analysisWf = lathe.Workflow("analyze");

const analyzeJob = lathe.Process({
  name: "analyze",
  commandLine: "python analyze.py input.clean > results.txt",
  image: "python:3.11",
  inputs: { data: "input.clean" },
  outputs: { results: "results.txt" }
});

analysisWf.Add(analyzeJob);
```

**Run:**
```bash
lathe run main_workflow.js
```

## Example 7: Docker Container with Image Building

**File: dockerized.js**

```javascript
// Build custom Docker image
lathe.DockerImage("docker/myapp", "myapp:latest", "Dockerfile", {
  VERSION: "1.0.0",
  PREFIX: "/app"
});

const wf = lathe.Workflow("containerized");

// Use the custom image
const appJob = lathe.Process({
  name: "run_app",
  commandLine: "/app/bin/myapp --input data.txt > results.json",
  image: "myapp:latest",
  inputs: { data: "data.txt" },
  outputs: { results: "results.json" },
  cpus: 4,
  memoryMB: 4096,
  memoryMB: 8192  // Container image might need more resources
});

wf.Add(appJob);
```

**Run:**
```bash
lathe run dockerized.js
```

## Example 8: Dynamic Workflow from Configuration

**File: dynamic.js**

```javascript
// Load configuration from external tool
const config = lathe.Plugin("get_samples --json");

lathe.println("Loaded " + config.samples.length + " samples");

const wf = lathe.Workflow("dynamic_analysis");

// Create a job for each sample
for (const sample of config.samples) {
  const job = lathe.Process({
    name: "analyze_" + sample.name,
    commandLine: `analyze.sh ${sample.id} > results_${sample.name}.txt`,
    image: sample.image || "analysis:latest",
    outputs: { 
      results: `results_${sample.name}.txt`
    },
    cpus: sample.cpus || 2,
    memoryMB: sample.memory || 2048
  });
  
  onComplete(job, function(result) {
    lathe.println("Analyzed sample: " + sample.name);
  });
  
  wf.Add(job);
}
```

**Run:**
```bash
lathe run dynamic.js
```

## Example 9: Advanced Resource Management

**File: resource_heavy.js**

```javascript
const wf = lathe.Workflow("big_data");

const job = lathe.Process({
  name: "big_compute",
  commandLine: "bigcompute --input huge_data.bin --output result.bin",
  image: "scientific:latest",
  inputs: {
    data: "huge_data.bin"
  },
  outputs: {
    result: "result.bin"
  },
  cpus: 32,              // 32 CPU cores
  memoryMB: 262144,      // 256 GB RAM
  description: "Heavy computational workload"
});

onComplete(job, function(result) {
  if (result.status.exitCode === 0) {
    lathe.println("✓ Big computation finished successfully");
    
    // Check metadata for resource usage
    for (const key in result.metadata) {
      lathe.println(key + ": " + result.metadata[key]);
    }
  }
});

wf.Add(job);
```

**Run with TES backend (for HPC/Cloud):**
```bash
lathe run --runner tes --tes-url http://tes-server:8000 resource_heavy.js
```

## Example 10: Error Recovery with Retries

**File: with_retries.js**

```javascript
const wf = lathe.Workflow("reliable");

const unreliableJob = lathe.Process({
  name: "download_data",
  commandLine: "wget https://example.com/data.zip -O data.zip",
  outputs: { data: "data.zip" },
  cpus: 1,
  memoryMB: 256,
  description: "Download data (might fail temporarily)"
});

onComplete(unreliableJob, function(result) {
  if (result.status.exitCode === 0) {
    lathe.println("✓ Download successful");
  } else {
    lathe.println("✗ Download failed");
    lathe.println("The runner will retry this job");
  }
});

wf.Add(unreliableJob);

// Next job depends on successful download
const processJob = lathe.Process({
  name: "process",
  commandLine: "unzip data.zip && process.sh > results.txt",
  inputs: { archive: "data.zip" },
  outputs: { results: "results.txt" },
  cpus: 2,
  memoryMB: 1024
});

wf.Add(processJob);
```

**Run:**
```bash
lathe run with_retries.js
```

## Quick Reference

### Global Functions

| Function | Purpose | Example |
|----------|---------|---------|
| `lathe.Workflow(name)` | Create workflow | `const w = lathe.Workflow("main")` |
| `lathe.Process(spec)` | Create job | `const j = lathe.Process({...})` |
| `lathe.File(spec)` | Create file ref | `lathe.File({path: "input.txt"})` |
| `lathe.FileCheck(spec)` | Check file exists | `lathe.FileCheck({file: {...}})` |
| `lathe.Tool(spec)` | Create tool template | `lathe.Tool({...})` |
| `lathe.DockerImage(...)` | Declare image | `lathe.DockerImage(".", "img:tag")` |
| `lathe.LoadPlan(path)` | Load sub-workflow | `lathe.LoadPlan("sub.js")` |
| `lathe.Plugin(cmd)` | Execute command | `lathe.Plugin("get_data --json")` |
| `onComplete(job, fn)` | Register callback | `onComplete(job, func(r) {...})` |
| `lathe.Params` | User parameters | `lathe.Params.sample` |

### Common Parameters

| Parameter | Default | Use |
|-----------|---------|-----|
| `name` | Auto-generated | Identify jobs |
| `commandLine` | (required) | What to execute |
| `image` | (local) | Docker image |
| `cpus` | 1 | CPU cores |
| `memoryMB` | 1024 | RAM in MB |
| `inputs` | {} | Input files |
| `outputs` | {} | Output files |

### Execution Modes

```bash
# Local execution
lathe run workflow.js

# With parameters
lathe run workflow.js --params key=value

# Dry-run (parse only)
lathe run workflow.js --dry-run

# Cloud/HPC via TES
lathe run --runner tes --tes-url http://server:8000 workflow.js

# Verbose logging
lathe run --verbose workflow.js
```

## Tips and Tricks

1. **Use meaningful names**: Helps with debugging and logging
2. **Specify resources accurately**: Better scheduling and failure handling
3. **Add callbacks**: Validate results and handle errors
4. **Test locally first**: Use local runner before cloud deployment
5. **Use parameters**: Make workflows reusable
6. **Organize sub-workflows**: Use LoadPlan for complex workflows
7. **Check exit codes**: Always validate job success in callbacks
8. **Log progress**: Use lathe.println() for workflow progress tracking

## See Also

- [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) - Complete API reference
- [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) - Detailed usage guide
- [ARCHITECTURE.md](ARCHITECTURE.md) - Architecture and design
