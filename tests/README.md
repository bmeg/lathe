# jflow Test Suite

This directory contains a comprehensive unit testing framework for jflow workflows.

## Overview

The test suite includes:
- **11 documented example workflows** covering current jflow features
- **Go testing framework** for automated testing
- **Test utilities** for running and validating workflows

## Test Examples

All test examples are in the `test_examples/` directory:

| Test | File | Description | Features Tested |
|------|------|-------------|-----------------|
| 01 | `01_hello_world.js` | Basic hello world | Simple command execution, output files |
| 02 | `02_file_operations.js` | File operations | Input/output file handling, multi-step |
| 03 | `03_multiple_executors.js` | Multiple executors | Sequential executors, data passing |
| 04 | `04_array_commands.js` | Array commands | Command as array of strings |
| 05 | `05_resource_specification.js` | Resource specs | CPU, memory, disk requirements |
| 06 | `06_file_check.js` | File checking | FileCheck functionality |
| 07 | `07_metadata_tags.js` | Metadata tags | TES tags for metadata |
| 09 | `09_workflow_params.js` | Parameters | Runtime parameter passing |
| 10 | `10_complex_pipeline.js` | Complex pipeline | Multi-step data pipeline |
| 11 | `11_tool_template_callable.js` | Tool templates | Callable tool templates, typed tool inputs |
| 12 | `12_path_object_factories.js` | Path/Object | Path/Object constructors and value rendering |

## Running Tests

### Run All Tests

```bash
go test ./tests -v
```

### Run Specific Test

```bash
JFLOW_TEST_NAME=hello_world go test ./tests -v -run TestIndividualWorkflow
```

### Validation Only (No Execution)

```bash
go test ./tests -v -run TestWorkflowValidation
```

### Check Test Files Exist

```bash
go test ./tests -v -run TestWorkflowFilesExist
```

### Benchmark

```bash
go test ./tests -bench=. -benchmem
```

### Using the Helper Script

```bash
# Run all tests
./run_tests.sh

# Run specific test
./run_tests.sh hello_world

# Run with verbose output
./run_tests.sh -v

# List all available tests
./run_tests.sh --list
```

## Test Structure

Each test workflow is designed to:
1. Execute quickly (under 90 seconds)
2. Test specific jflow features
3. Be independent of other tests
4. Produce verifiable outputs

## Adding New Tests

To add a new test workflow:

1. Create a new `.js` file in `test_examples/`
2. Follow the naming convention: `NN_test_name.js`
3. Add the test to `GetTestWorkflows()` in `tests/workflow_test.go`
4. Run validation: `go test ./tests -run TestWorkflowValidation`

Example test structure:

```javascript
// Test XX: Description
// Tests: Feature list

const workflow = jflow.Workflow("test_name");

const task = jflow.Process({
    name: "task_name",
    description: "Task description",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'test' > /tmp/output.txt"
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

## CI/CD Integration

The test suite can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run jflow tests
  run: |
    go test ./tests -v -timeout 10m
```

```yaml
# Example GitLab CI
test:
  script:
    - go test ./tests -v -timeout 10m
```

## Test Output

Successful test output will show:

```
=== RUN   TestAllWorkflows/hello_world
    workflow_test.go:XX: Testing: hello_world - Basic hello world test
    workflow_test.go:XX: Running workflow: hello_world_test
    workflow_test.go:XX: Workflow completed successfully: hello_world_test
--- PASS: TestAllWorkflows/hello_world (2.34s)
```

## Troubleshooting

### Test Timeout

If tests timeout, increase the timeout in `workflow_test.go`:

```go
Timeout: 120 * time.Second,  // Increase as needed
```

### Missing Test Files

Run the file existence check:

```bash
go test ./tests -v -run TestWorkflowFilesExist
```

### Docker Issues

Tests require Docker to be running. Check Docker status:

```bash
docker ps
```

### Verbose Logging

Enable verbose logging in tests by modifying `logger.Init()`:

```go
logger.Init(true, false)  // Enable verbose mode
```

## Performance

Expected execution times (approximate):
- Individual simple test: 2-5 seconds
- Individual complex test: 5-15 seconds
- Full test suite: 30-90 seconds

## Dependencies

- Go 1.21+
- Docker (for containerized execution)
- Required Go modules (see go.mod)
