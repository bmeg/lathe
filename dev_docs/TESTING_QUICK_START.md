# jflow Unit Testing - Quick Start Guide

## TL;DR

```bash
# Run all tests
./run_tests.sh

# List available tests
./run_tests.sh --list

# Run specific test
./run_tests.sh hello_world

# Quick validation (no execution)
./run_tests.sh --validation
```

## Installation

No installation needed! The testing framework is ready to use.

**Requirements:**
- Go 1.21+
- Docker

## Basic Commands

### Using the Shell Script

```bash
# Run all tests
./run_tests.sh

# Run with verbose output
./run_tests.sh -v

# Run a specific test
./run_tests.sh hello_world

# Validation only (fast, no execution)
./run_tests.sh --validation

# Run benchmarks
./run_tests.sh --bench

# List all tests
./run_tests.sh --list

# Help
./run_tests.sh --help
```

### Using Make

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run validation only
make test-validation

# Run a specific test
make test-one NAME=hello_world

# Run benchmarks
make bench

# List all tests
make list

# Clean up test outputs
make clean

# Show all make targets
make help
```

### Using Go Test Directly

```bash
# Run all tests
go test ./tests -run TestAllWorkflows -v -timeout 10m

# Run validation only
go test ./tests -run TestWorkflowValidation -v

# Run specific test
JFLOW_TEST_NAME=hello_world go test ./tests -run TestIndividualWorkflow -v

# Check all test files exist
go test ./tests -run TestWorkflowFilesExist -v

# Run benchmarks
go test ./tests -bench=. -benchmem
```

## Understanding Test Output

### Successful Test

```
=== RUN   TestAllWorkflows/hello_world
    workflow_test.go:171: Testing: hello_world - Basic hello world test
    workflow_test.go:123: Running workflow: hello_world_test
    workflow_test.go:158: Workflow completed successfully: hello_world_test
--- PASS: TestAllWorkflows/hello_world (2.34s)
```

### Failed Test

```
=== RUN   TestAllWorkflows/hello_world
    workflow_test.go:171: Testing: hello_world - Basic hello world test
    workflow_test.go:173: Test failed: workflow execution failed: ...
--- FAIL: TestAllWorkflows/hello_world (2.34s)
```

## Test Examples

### Test 01: Hello World
Most basic test - good for quick verification.
```bash
./run_tests.sh hello_world
```

### Test 10: Complex Pipeline
Most comprehensive test - tests multi-step pipelines.
```bash
./run_tests.sh complex_pipeline
```

### Validation Only
Fastest way to check if all workflows are syntactically correct.
```bash
./run_tests.sh --validation
# Should complete in < 1 second
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Test jflow Workflows
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Run tests
        run: go test ./tests -v -timeout 10m
```

### GitLab CI

```yaml
test:
  image: golang:1.21
  services:
    - docker:dind
  script:
    - go test ./tests -v -timeout 10m
```

### Jenkins

```groovy
stage('Test') {
  steps {
    sh 'go test ./tests -v -timeout 10m'
  }
}
```

## Troubleshooting

### Tests Don't Run

**Problem:** Permission denied on `run_tests.sh`

**Solution:**
```bash
chmod +x run_tests.sh
```

**Problem:** Docker not running

**Solution:**
```bash
# Start Docker
sudo systemctl start docker

# Verify
docker ps
```

### Tests Timeout

**Problem:** Tests take too long

**Solution:** Increase timeout
```bash
go test ./tests -run TestAllWorkflows -v -timeout 20m
```

### Tests Fail

**Problem:** Individual test fails

**Solution:** Run with verbose output to see details
```bash
./run_tests.sh -v hello_world
```

## Performance Expectations

| Operation | Expected Time |
|-----------|---------------|
| Validation | < 1 second |
| Single simple test | 2-5 seconds |
| Single complex test | 5-15 seconds |
| Full test suite | 30-90 seconds |

## Next Steps

1. **Read the documentation:**
   - `tests/README.md` - Testing framework details
   - `test_examples/README.md` - Test catalog
   - `TESTING_FRAMEWORK_SUMMARY.md` - Complete overview

2. **Run your first test:**
   ```bash
   ./run_tests.sh hello_world
   ```

3. **Add your own test:**
   - Copy an existing test from `test_examples/`
   - Modify it for your use case
   - Add it to `tests/workflow_test.go`
   - Run `./run_tests.sh --validation` to verify

## Common Use Cases

### Before Committing Code
```bash
make test-validation  # Quick syntax check
make test            # Full test suite
```

### Debugging a Workflow
```bash
# Run specific test with verbose output
./run_tests.sh -v my_test_name
```

### Performance Testing
```bash
make bench
```

### Quick Smoke Test
```bash
./run_tests.sh hello_world
```

## Getting Help

- Check `./run_tests.sh --help` for command options
- Read `tests/README.md` for detailed documentation
- View test examples in `test_examples/` directory
- Check `TESTING_FRAMEWORK_SUMMARY.md` for complete overview

## Quick Reference Card

```
Command                          | What It Does
---------------------------------|---------------------------
./run_tests.sh                   | Run all tests
./run_tests.sh -v                | Run all tests (verbose)
./run_tests.sh test_name         | Run specific test
./run_tests.sh --validation      | Validate syntax only (fast)
./run_tests.sh --list            | List available tests
./run_tests.sh --help            | Show help
make test                        | Run all tests
make test-one NAME=test_name     | Run specific test
make list                        | List available tests
make clean                       | Clean up test outputs
```

---

**Happy Testing! 🧪**
