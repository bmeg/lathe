
# Lathe
Dynamic build system implementing the jflow workflow standard with **GA4GH TES API alignment**.

Lathe allows a user to programmatically define a workflow of tasks using the **jflow** JavaScript API - an open standard for workflow definitions that is now aligned with the GA4GH Task Execution Service (TES) API v1.1.0 specification.

Workflow files are written in JavaScript using the jflow API.

## Key Features

- ✨ **TES API Alignment**: Compatible with GA4GH TES specification
- 🎯 **Polymorphic Commands**: Use strings or arrays for command specifications
- 🔗 **Multi-Executor Tasks**: Run sequential commands in the same task
- 🏷️ **Rich Metadata**: TES-style tags and resource specifications
- 📦 **Container Support**: Docker/Podman container execution
- ☁️ **Cloud Storage**: S3, GCS, HTTP file support

## Quick Start

### TES-Aligned Format
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

- **[TES Alignment Guide](docs/TES_ALIGNMENT_GUIDE.md)**: Comprehensive TES format guide
- **[jflow Standard](docs/JFLOW_STANDARD.md)**: Complete jflow API specification
- **[Quick Start](docs/QUICK_START.md)**: Getting started guide
- **[Architecture](docs/ARCHITECTURE.md)**: System architecture overview

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
    executors: [{
        image: "cwltool:latest",
        command: `cwltool --outdir ../../source/pharmacodb/rdata ./download_pharmaco.cwl`
    }],
    outputs: Object.entries(downloadOutputs).map(([name, path]) => ({ name, path })),
    resources: {
        cpu_cores: 1,
        ram_gb: 2
    }
})
prep.Add(p)
```


## jflow API 
The jflow global object provides the following functions:
```
	Params: map[string]string{}
	Workflow:    function(name)
    Import:      function(path)
    Process:     function(Process)
	File:        function(path)
	Plugin:      function(commandLine)
	DockerImage: function(path)
    GetParams:   function(schema)
    Path:        function(name)
    Object:      function(name)
```


## Process Object

### TES Format (GA4GH Aligned)
```
	Name        string
	Description string
    Executors   []Executor        // Commands in containers
    Inputs      []Input           // Input files with URLs
    Outputs     []Output          // Output files with destinations
    Resources   *Resources        // CPU, RAM, disk specifications
	Volumes     []string          // Shared volumes
	Tags        map[string]string // Metadata
```


# Running lathe

```
lathe run <lathe_file> <workflow_name>
```