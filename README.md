
# Lathe
Dynamic build system implementing the jflow workflow standard with **GA4GH TES API alignment**.

Lathe allows a user to programmatically define a workflow of tasks using the **jflow** JavaScript API - an open standard for workflow definitions that is now aligned with the GA4GH Task Execution Service (TES) API v1.1.0 specification.

Workflow files are written in JavaScript using the jflow API.

## Key Features

- ✨ **TES API Alignment**: Compatible with GA4GH TES specification
- 🔄 **Backward Compatible**: Legacy jflow syntax still supported
- 🎯 **Polymorphic Commands**: Use strings or arrays for command specifications
- 🔗 **Multi-Executor Tasks**: Run sequential commands in the same task
- 🏷️ **Rich Metadata**: TES-style tags and resource specifications
- 📦 **Container Support**: Docker/Podman container execution
- ☁️ **Cloud Storage**: S3, GCS, HTTP file support

## Quick Start

### Legacy Format (Still Supported)
```javascript
const pipeline = jflow.Workflow("data_pipeline");

const job = jflow.Process({
    name: "analyze",
    commandLine: "python analyze.py input.csv > output.csv",
    image: "python:3.11",
    inputs: { data: "input.csv" },
    outputs: { result: "output.csv" },
    cpus: 4,
    memoryMB: 8192
});

pipeline.Add(job);
```

### TES-Aligned Format (Recommended)
```javascript
const pipeline = jflow.Workflow("data_pipeline");

const job = jflow.Process({
    name: "analyze",
    description: "Analyze genomic data",
    
    // TES executors
    executors: [{
        image: "python:3.11",
        command: "python analyze.py /data/input.csv > /data/output.csv",  // String or array
        env: { "PYTHONPATH": "/app" }
    }],
    
    // TES inputs/outputs
    inputs: [{
        url: "s3://bucket/input.csv",
        path: "/data/input.csv"
    }],
    outputs: [{
        url: "s3://bucket/output.csv",
        path: "/data/output.csv"
    }],
    
    // TES resources
    resources: {
        cpu_cores: 4,
        ram_gb: 8,
        disk_gb: 50
    },
    
    tags: {
        "project": "genomics",
        "version": "1.0"
    }
});

pipeline.Add(job);
```

## Documentation

- **[TES Alignment Guide](TES_ALIGNMENT_GUIDE.md)**: Comprehensive guide to TES format and migration
- **[jflow Standard](JFLOW_STANDARD.md)**: Complete jflow API specification
- **[Quick Start](QUICK_START.md)**: Getting started guide
- **[Architecture](ARCHITECTURE.md)**: System architecture overview

Example:
```javascript

prep = jflow.Workflow("prep")

projects = [
    {name: "BeatAML_2018"},
    {name: "FIMM_2016"},
    {name: "NCI60_2021"},
    {name: "Tavor_2020"},
    {name: "CCLE_2015"},
    {name: "GBM_scr2"},
    {name: "PDTX_2019"},
    {name: "UHNBreast_2019"},
    {name: "CTRPv2_2015"},
    {name: "GBM_scr3"},
    {name: "GRAY_2017"},
    {name: "PRISM_2020"},
    {name: "gCSI_2019"}
]

downloadOutputs = {}

projects.forEach( (element, index) => {
    downloadOutputs[`file_${index}`] = `../../source/pharmacodb/rdata/${element.name}.rdata`
})

p = jflow.Process({
    name: "download",
    commandLine: `cwltool --outdir ../../source/pharmacodb/rdata ./download_pharmaco.cwl`,
    outputs: downloadOutputs
})
prep.Add(p)
```


## jflow API 
The jflow global object provides the following functions:
```
	Params: map[string]string{}
	Workflow:    function(name)
	LoadPlan:    function(path)
	Process:     function(Process)  // Supports both legacy and TES formats
	File:        function(path)
	Plugin:      function(commandLine)
	DockerImage: function(path)
```


## Process Object

### Legacy Format
```
	BasePath    string
	Name        string
	Desc        map[string]any
	CommandLine string
	Inputs      map[string]string
	Outputs     map[string]string
	MemMB       uint
	NCpus       uint
```

### TES Format (GA4GH Aligned)
```
	Name        string
	Description string
	Executors   []TESExecutor     // Sequential commands in containers
	TESInputs   []TESInput        // Input files with URLs
	TESOutputs  []TESOutput       // Output files with destinations
	Resources   *TESResources     // CPU, RAM, disk specifications
	Volumes     []string          // Shared volumes
	Tags        map[string]string // Metadata
```


# Running lathe

```
lathe run <lathe_file> <workflow_name>
```