# TES Migration - Legacy Removal Complete

## Overview
Successfully removed all legacy workflow elements and TES prefixes, creating a clean, TES-native implementation.

## Type Renames
All type names have been cleaned up to remove the "TES" prefix:

- `TESInput` → `Input`
- `TESOutput` → `Output`
- `TESExecutor` → `Executor`
- `TESResources` → `Resources`
- `TESState` → `State`
- `TESTask` → `Task`
- `TESFileType` → `FileTypeEnum`

**Location**: `scriptfile/tes_model.go`

## State Constants Renamed
All state constants in `scriptfile/tes_model.go` now use the `State` prefix:

```go
StateUnknown      State = "UNKNOWN"
StateQueued       State = "QUEUED"
StateInitializing State = "INITIALIZING"
StateRunning      State = "RUNNING"
StatePaused       State = "PAUSED"
StateComplete     State = "COMPLETE"
StateExecutorError State = "EXECUTOR_ERROR"
StateSystemError  State = "SYSTEM_ERROR"
StateCanceled     State = "CANCELED"
StateCanceling    State = "CANCELING"
StatePreempted    State = "PREEMPTED"
```

## ProcessDesc Structure (scriptfile/model.go)
Removed all legacy fields and now uses only TES-aligned fields:

**Removed Fields:**
- `CommandLine` (string)
- `Shell` (string)
- `Image` (string)
- `MemMB` (uint)
- `NCpus` (uint)
- `Inputs` (map[string]string) - replaced
- `Outputs` (map[string]string) - replaced

**Current Fields (TES-aligned):**
- `Inputs []Input` - Array of input files with full metadata
- `Outputs []Output` - Array of output files with full metadata
- `Executors []Executor` - Array of commands to execute
- `Resources *ResourceRequirements` - Compute requirements
- `Volumes []string` - Shared directories between executors

## ResourceRequirements Structure (scriptfile/model.go)
Simplified to use only TES fields, removing legacy resource fields:

**Removed Fields:**
- `CPUs` (uint)
- `MemoryMB` (uint)
- `DiskMB` (uint)

**Current Fields (TES-aligned):**
- `CPUCores uint` - Number of CPU cores
- `RamGB float64` - RAM in gigabytes
- `DiskGB float64` - Disk size in gigabytes
- `Preemptible bool` - Can run on preemptible instances
- `Zones []string` - Compute zones
- `BackendParameters map[string]string` - Backend-specific config
- `BackendParametersStrict bool` - Fail if backend parameters unsupported
- `Timeout uint` - Execution timeout
- `Retries uint` - Number of retries

## API Changes (scriptfile/api.go)
Parser functions have been renamed to remove "TES" prefix:

- `parseTESExecutor()` → `parseExecutor()`
- `parseTESInput()` → `parseInput()`
- `parseTESOutput()` → `parseOutput()`
- `parseTESResources()` → `parseResources()`

All legacy parsing code has been removed.

## Workflow Execution Updates (workflow/step_process.go)
Updated to work with new TES structure:

1. Removed `NormalizeLegacyFields()` call
2. Changed Inputs/Outputs handling from map to array iteration:
   - Inputs now use `Input.Path` and `Input.Name`
   - Outputs now use `Output.Path` and `Output.Name`
3. Changed command execution to use `Executors[0].Command` array
4. Updated resource extraction to use `Resources.CPUCores` and `Resources.RamGB`
5. Removed unused imports: `raymond` and `shlex`

## Command Tool Updates (cmd/outputs/main.go)
Updated to handle new `Output` struct instead of string values:

- Output iteration now uses `output.Path` instead of string value
- Properly extracts Output.Name for JSON output

## API Field Names
All parser functions now expect TES-aligned field names:

- `cpu_cores` (not `cpus`)
- `ram_gb` (not `memoryMB`)
- `disk_gb` (not `diskMB`)

Example ProcessDesc JSON:
```json
{
  "name": "process_name",
  "description": "What this does",
  "inputs": [
    {
      "name": "input_file",
      "path": "/work/input.txt",
      "type": "FILE"
    }
  ],
  "outputs": [
    {
      "name": "output_file",
      "path": "/work/output.txt",
      "type": "FILE"
    }
  ],
  "executors": [
    {
      "image": "ubuntu:20.04",
      "command": ["bash", "-c", "cat /work/input.txt > /work/output.txt"]
    }
  ],
  "resources": {
    "cpu_cores": 2,
    "ram_gb": 4.0,
    "disk_gb": 10.0
  }
}
```

## Build Status
✅ **Successfully Builds** - No compilation errors

## Breaking Changes
⚠️ These are breaking changes if there's existing code using the old format:

1. `ProcessDesc` no longer has `CommandLine`, `Shell`, `Image`, `MemMB`, or `NCpus`
2. `Inputs` and `Outputs` are now arrays, not maps
3. Resource requirements use TES field names: `cpu_cores`, `ram_gb`, `disk_gb`
4. All type names have "TES" prefix removed

## Migration Path for Users
If migrating from legacy format:

1. **Old Format:**
   ```javascript
   Process({
     name: "mytask",
     commandLine: "cat input.txt > output.txt",
     memoryMB: 1024,
     cpus: 2,
     image: "ubuntu:20.04",
     inputs: {"input": "input.txt"},
     outputs: {"output": "output.txt"}
   })
   ```

2. **New Format:**
   ```javascript
   Process({
     name: "mytask",
     inputs: [{name: "input", path: "input.txt", type: "FILE"}],
     outputs: [{name: "output", path: "output.txt", type: "FILE"}],
     executors: [{
       image: "ubuntu:20.04",
       command: ["cat", "input.txt"]
     }],
     resources: {
       cpu_cores: 2,
       ram_gb: 1.0,
       disk_gb: 10.0
     }
   })
   ```

## Files Modified
- `scriptfile/tes_model.go` - Type renames, state constants
- `scriptfile/model.go` - ProcessDesc and ResourceRequirements structure
- `scriptfile/api.go` - Parser function renames, legacy code removal
- `workflow/step_process.go` - Executor handling, array iteration
- `cmd/outputs/main.go` - Output struct handling
