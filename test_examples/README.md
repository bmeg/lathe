# jflow Test Examples

This directory contains a suite of test workflows for validating jflow functionality. Each test is designed to run quickly and test specific features of the jflow workflow engine.

## Quick Start

```bash
# From project root, run all tests
./run_tests.sh

# Run a specific test
./run_tests.sh hello_world

# List available tests
./run_tests.sh --list
```

## Test Catalog

### 01_hello_world.js
**Purpose:** Basic workflow functionality  
**Features Tested:**
- Simple command execution
- Output file creation
- Basic TES format

**Expected Output:** Creates `/tmp/hello.txt` with greeting message

### 02_file_operations.js
**Purpose:** File input/output handling  
**Features Tested:**
- Input file specification
- Output file specification
- Multi-step workflows with file dependencies
- Command chaining

**Expected Output:** Creates input file, processes it, outputs line count

### 03_multiple_executors.js
**Purpose:** Sequential executor execution  
**Features Tested:**
- Multiple executors in single process
- Data passing between executors
- Shared volume usage

**Expected Output:** Three sequential steps producing final output file

### 04_array_commands.js
**Purpose:** Command format variations  
**Features Tested:**
- Command as array of strings (non-shell mode)
- Array command execution

**Expected Output:** Executes command from array format

### 05_resource_specification.js
**Purpose:** Resource requirements  
**Features Tested:**
- CPU core specification
- RAM requirements
- Disk requirements
- Resource constraint enforcement

**Expected Output:** Task with specific resource allocation

### 06_file_check.js
**Purpose:** File existence validation  
**Features Tested:**
- FileCheck functionality
- File dependency verification
- Workflow step ordering

**Expected Output:** Creates file and validates its existence

### 07_metadata_tags.js
**Purpose:** Task metadata  
**Features Tested:**
- TES tags
- Arbitrary key-value metadata
- Tag preservation through execution

**Expected Output:** Task with metadata tags

### 08_legacy_format.js
**Purpose:** Backward compatibility  
**Features Tested:**
- Legacy jflow format (commandLine, cpus, memoryMB)
- Format conversion
- Mixed format support

**Expected Output:** Legacy format task executes successfully

### 09_workflow_params.js
**Purpose:** Parameter passing  
**Features Tested:**
- jflow.Params usage
- Runtime parameter injection
- Default parameter values
- Parameter interpolation in commands

**Expected Output:** Uses parameters with defaults

### 10_complex_pipeline.js
**Purpose:** Real-world workflow  
**Features Tested:**
- Multi-step data pipeline
- File dependencies between steps
- Complex workflow orchestration
- Data generation, processing, and summarization

**Expected Output:** Three-stage pipeline with final summary

## Test Design Principles

All tests follow these principles:

1. **Fast Execution:** Each test completes in under 90 seconds
2. **Isolated:** Tests don't depend on each other
3. **Deterministic:** Same input produces same output
4. **Minimal Dependencies:** Uses only `ubuntu:20.04` base image
5. **Clear Purpose:** Each test has a specific feature focus

## File Naming Convention

Tests follow the pattern: `NN_descriptive_name.js`

- `NN`: Two-digit number for ordering
- `descriptive_name`: Clear indication of what's tested
- `.js`: JavaScript workflow file

## Adding New Tests

When adding a new test:

1. **Choose a number:** Next available in sequence
2. **Write the test:** Follow existing patterns
3. **Document it:** Add entry to this README
4. **Update test suite:** Add to `tests/workflow_test.go`
5. **Validate:** Run `./run_tests.sh --validation`

Example template:

```javascript
// Test XX: Brief Description
// Tests: Feature 1, Feature 2

const workflow = jflow.Workflow("test_workflow_name");

const task = jflow.Process({
    name: "test_task",
    description: "What this task does",
    executors: [{
        image: "ubuntu:20.04",
        command: "your command here"
    }],
    outputs: [{
        name: "output",
        url: "file:///tmp/output.txt",
        path: "/tmp/output.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(task);
```

## Test Execution Flow

For each test, the framework:

1. **Parse:** Load and parse the JavaScript workflow file
2. **Validate:** Check syntax and structure
3. **Prepare:** Set up runner and workflow execution engine
4. **Execute:** Run the workflow with timeout protection
5. **Verify:** Check for successful completion
6. **Report:** Output results

## Expected Behavior

All tests should:
- ✅ Parse without errors
- ✅ Complete within timeout
- ✅ Produce expected output files
- ✅ Exit with success status

## Troubleshooting

### Test Fails to Parse
- Check JavaScript syntax
- Verify jflow API usage
- Run validation: `./run_tests.sh --validation`

### Test Times Out
- Check Docker is running: `docker ps`
- Verify image is available: `docker images ubuntu:20.04`
- Increase timeout in test configuration

### Output Files Missing
- Check file paths match between outputs and expected locations
- Verify container has write permissions
- Check workflow step ordering

## Integration Testing

These tests can be used for:

- **CI/CD Pipelines:** Automated testing on commits
- **Regression Testing:** Verify changes don't break functionality
- **Performance Monitoring:** Track execution time trends
- **Feature Validation:** Confirm new features work correctly

## Performance Baselines

Typical execution times (on standard hardware):

| Test | Expected Time |
|------|---------------|
| 01   | 2-5 seconds   |
| 02   | 3-8 seconds   |
| 03   | 3-8 seconds   |
| 04   | 2-5 seconds   |
| 05   | 2-5 seconds   |
| 06   | 3-6 seconds   |
| 07   | 2-5 seconds   |
| 08   | 2-5 seconds   |
| 09   | 2-5 seconds   |
| 10   | 5-15 seconds  |

**Total Suite:** 30-90 seconds

## Contributing

When contributing new tests:

1. Keep tests simple and focused
2. Use descriptive names
3. Add comprehensive comments
4. Update documentation
5. Test on clean environment
6. Ensure reproducibility
