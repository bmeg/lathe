# Lathe Workflow Engine - Documentation Index

Welcome to the Lathe workflow engine documentation. This guide helps you navigate all available resources.

## Quick Navigation

### 🚀 Getting Started

- **[QUICK_START.md](QUICK_START.md)** - 10 runnable examples to get started immediately
  - Hello World workflow
  - Data pipelines with dependencies
  - Using callbacks for post-job analysis
  - Parameterized workflows
  - Multi-stage pipelines
  - Sub-workflow composition
  - Docker containerization
  - Dynamic workflow generation
  - Resource management
  - Error recovery

### 📖 Complete Documentation

- **[WORKFLOW_MODEL.md](WORKFLOW_MODEL.md)** - Complete API reference
  - Global API (`lathe.*` functions)
  - Type specifications
  - Schema documentation
  - Workflow composition patterns
  - Futures and callbacks usage
  - Workflow execution

- **[INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)** - Usage patterns and integration
  - Architecture overview
  - Detailed usage examples
  - Execution patterns
  - Implementation details
  - Best practices
  - Migration from old API
  - Troubleshooting guide

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Design and architecture
  - Design philosophy
  - Component descriptions
  - Type system design
  - Execution flow diagrams
  - Performance considerations
  - Testing strategy
  - Future enhancements

### 📋 Reference

- **[REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md)** - High-level summary of changes
  - Key changes overview
  - Benefits of refactoring
  - Migration path
  - Statistics
  - Conclusion

- **[CHANGELOG.md](CHANGELOG.md)** - Detailed change log
  - All file changes
  - API changes
  - Type system updates
  - Backward compatibility matrix
  - Testing status
  - Deployment checklist

---

## For Different Use Cases

### I just want to run a workflow

👉 Start here: **[QUICK_START.md](QUICK_START.md)**

Examples cover:
- Basic workflows
- Parameterization
- Error handling
- Docker containers
- Multi-stage pipelines

### I want to understand how it works

👉 Read these in order:
1. **[QUICK_START.md](QUICK_START.md)** - See examples
2. **[INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)** - Understand patterns
3. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Deep dive into design

### I need the complete API reference

👉 Go to: **[WORKFLOW_MODEL.md](WORKFLOW_MODEL.md)**

Covers:
- Global API functions
- Type specifications
- All parameter options
- Complete examples

### I'm integrating Lathe into my project

👉 Read: **[INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)**

Sections for:
- Architecture overview
- Programmatic usage
- Different execution models
- Implementation details

### I want to understand what changed

👉 Check: **[REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md)** and **[CHANGELOG.md](CHANGELOG.md)**

Get details about:
- New features (futures, callbacks)
- API enhancements
- Type system changes
- Backward compatibility

### I'm contributing or extending the engine

👉 Review: **[ARCHITECTURE.md](ARCHITECTURE.md)** and source code

- Component architecture
- Type system design
- Extension points
- Code organization

---

## Key Concepts

### Workflows
A collection of jobs (processes) organized as a directed acyclic graph (DAG). Jobs can be executed in parallel if they have no dependencies.

**Documentation**: See WORKFLOW_MODEL.md section "Workflow Composition"

### Processes (Jobs)
Individual tasks that execute commands, possibly in Docker containers. Can have inputs, outputs, resource requirements.

**Documentation**: See WORKFLOW_MODEL.md section "ProcessSpec"

### Futures
Deferred result containers that return immediately from job declaration but resolve when the job completes. Enable asynchronous job submission.

**Documentation**: See WORKFLOW_MODEL.md section "Futures and Callbacks"

### Callbacks
Functions executed after a job completes for post-job analysis, validation, or triggering dependent workflows.

**Documentation**: See WORKFLOW_MODEL.md section "Post-Job Analysis with Callbacks"

### Dependencies
Automatically resolved through:
1. **File dependencies**: Job B depends on A if B's input matches A's output
2. **Explicit dependencies**: Can be specified in job declaration
3. **File checks**: Jobs depend on any file checks that guarantee their inputs

**Documentation**: See INTEGRATION_GUIDE.md section "How Futures Work"

### Runners
Different backends for executing workflows:
- **Local runner**: Execute on local machine using os.exec
- **TES runner**: Execute via GA4GH TES API (cloud/HPC)

**Documentation**: See WORKFLOW_MODEL.md section "Job Runners"

---

## Common Tasks

### Create a simple workflow

1. Read: [QUICK_START.md](QUICK_START.md) Example 1
2. Reference: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) ProcessSpec section
3. Command: `lathe run workflow.js`

### Add error handling

1. Read: [QUICK_START.md](QUICK_START.md) Example 3 and 5
2. Reference: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) Callbacks section
3. Use: `onComplete(job, function(result) {...})`

### Use parameters

1. Read: [QUICK_START.md](QUICK_START.md) Example 4
2. Reference: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) Parameter section
3. Use: `lathe.Params.paramName`

### Use Docker containers

1. Read: [QUICK_START.md](QUICK_START.md) Example 7
2. Reference: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) DockerImage section
3. Use: `image: "image:tag"` in Process spec

### Compose multiple workflows

1. Read: [QUICK_START.md](QUICK_START.md) Example 6
2. Reference: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) LoadPlan section
3. Use: `lathe.LoadPlan("sub.js")`

### Deploy to cloud

1. Read: [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) section "Executing Workflows"
2. Command: `lathe run --runner tes --tes-url http://server:8000 workflow.js`

---

## API Quick Reference

### Global Functions

```javascript
lathe.Workflow(name)           // Create workflow
lathe.Process(spec)            // Create job
lathe.File(spec)               // Reference file
lathe.FileCheck(spec)          // File existence check
lathe.Tool(spec)               // Tool template
lathe.DockerImage(...)         // Docker image
lathe.LoadPlan(path)           // Load sub-workflow
lathe.Plugin(cmd)              // Execute external command
onComplete(job, callback)      // Register callback
lathe.Params                   // User parameters
print(x), println(x)           // Logging
glob(pattern)                  // Path globbing
```

**Full reference**: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) Global API section

### Process Specification

```javascript
{
  name: "job_name",           // Optional - auto-generated if omitted
  commandLine: "command",     // Required
  image: "image:tag",         // Optional
  cpus: 4,                    // Optional - default 1
  memoryMB: 8192,            // Optional - default 1024
  inputs: { "param": "path" }, // Optional
  outputs: { "param": "path" }, // Optional
  description: "..."         // Optional
}
```

**Full reference**: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) ProcessSpec section

### Callback Function Signature

```javascript
onComplete(job, function(result) {
  result.jobName          // string
  result.status.state     // "completed"|"failed"|"cancelled"
  result.status.exitCode  // integer
  result.status.error     // string (if state != "completed")
  result.outputFiles      // map of output files
  result.logs            // stdout/stderr logs
  result.metadata        // execution metadata
});
```

**Full reference**: [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) JobResult section

---

## Troubleshooting

### Problem: Workflow not executing

**Check**:
1. Is the script file valid JavaScript?
2. Are all functions called with correct parameters?
3. Check [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md) for parameter requirements

### Problem: Jobs not running in dependency order

**Check**:
1. Do output file paths exactly match input file paths?
2. Are file checks added before dependent jobs?
3. Review [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) Dependency Resolution section

### Problem: Callbacks not executing

**Check**:
1. Is the callback function valid JavaScript?
2. Does it have correct parameter name?
3. See [QUICK_START.md](QUICK_START.md) Example 3

### Problem: Parameters not passed to workflow

**Check**:
1. Use `lathe.Params.paramName` (not capital P in object)
2. Pass parameters: `--params key=value`
3. See [QUICK_START.md](QUICK_START.md) Example 4

### Full troubleshooting guide

**Read**: [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) Troubleshooting section

---

## Performance Tips

1. **Specify accurate resources**: Helps scheduler make better decisions
2. **Use file dependencies**: Automatic dependency resolution is efficient
3. **Organize with sub-workflows**: Logical organization improves maintainability
4. **Add callbacks selectively**: Only for important jobs
5. **Test locally first**: Use local runner before cloud deployment

**More tips**: [ARCHITECTURE.md](ARCHITECTURE.md) Performance Considerations section

---

## Additional Resources

### Source Code

- `scriptfile/model.go` - Core type definitions (412 lines)
- `scriptfile/js_vm.go` - JavaScript environment setup
- `scriptfile/api.go` - Global API functions
- `scriptfile/workflow.go` - Workflow composition

### Examples Directory

Check `examples/` folder for sample workflows

### Community

- Questions? Check [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) Troubleshooting
- Features? Check [ARCHITECTURE.md](ARCHITECTURE.md) Future Enhancements
- Contribute? Review [ARCHITECTURE.md](ARCHITECTURE.md) Code Organization

---

## Document Organization

```
Documentation
├── Quick Start (get running)
│   └── QUICK_START.md (10 examples)
│
├── Learning (understand concepts)
│   ├── WORKFLOW_MODEL.md (API reference)
│   └── INTEGRATION_GUIDE.md (patterns)
│
├── Deep Dive (architecture)
│   └── ARCHITECTURE.md (design decisions)
│
└── Reference (detailed info)
    ├── REFACTORING_SUMMARY.md (what changed)
    └── CHANGELOG.md (detailed change log)
```

---

## Version Information

- **Engine Version**: Refactored (see REFACTORING_SUMMARY.md)
- **Build Status**: ✅ All packages compile successfully
- **Backward Compatibility**: ✅ 100% maintained
- **Breaking Changes**: ⚠️ None

---

## Next Steps

1. **New to Lathe?** → Start with [QUICK_START.md](QUICK_START.md)
2. **Want examples?** → See QUICK_START.md or Examples directory
3. **Need API reference?** → Check [WORKFLOW_MODEL.md](WORKFLOW_MODEL.md)
4. **Integrating Lathe?** → Read [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md)
5. **Understanding design?** → Review [ARCHITECTURE.md](ARCHITECTURE.md)
6. **Migrating workflows?** → Check [REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md)

---

**Last Updated**: 2026-03-04  
**Documentation Status**: Complete ✅  
**Build Status**: Passing ✅
