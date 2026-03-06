# Makefile for jflow test suite

.PHONY: help test test-verbose test-validation test-files test-one clean bench list

# Default target
help:
	@echo "jflow Test Suite Commands:"
	@echo ""
	@echo "  make test              - Run all tests"
	@echo "  make test-verbose      - Run all tests with verbose output"
	@echo "  make test-validation   - Run validation tests only (fast)"
	@echo "  make test-files        - Check that all test files exist"
	@echo "  make test-one NAME=x   - Run a specific test (e.g., make test-one NAME=hello_world)"
	@echo "  make bench             - Run benchmarks"
	@echo "  make list              - List all available tests"
	@echo "  make clean             - Clean up test outputs"
	@echo ""

# Run all tests
test:
	@echo "Running all jflow tests..."
	@./run_tests.sh

# Run tests with verbose output
test-verbose:
	@echo "Running all jflow tests (verbose)..."
	@./run_tests.sh -v

# Run validation only (no execution)
test-validation:
	@echo "Validating all workflows..."
	@./run_tests.sh --validation

# Check test files exist
test-files:
	@echo "Checking test files..."
	@go test ./tests -v -run TestWorkflowFilesExist -timeout 10s

# Run a specific test
test-one:
ifndef NAME
	@echo "Error: NAME not specified"
	@echo "Usage: make test-one NAME=hello_world"
	@exit 1
endif
	@./run_tests.sh $(NAME)

# Run benchmarks
bench:
	@./run_tests.sh --bench

# List available tests
list:
	@./run_tests.sh --list

# Clean up test outputs
clean:
	@echo "Cleaning up test outputs..."
	@rm -f /tmp/hello.txt /tmp/input.txt /tmp/output.txt
	@rm -f /tmp/multi_exec_*.txt /tmp/array_output.txt
	@rm -f /tmp/resource_test.txt /tmp/checkme.txt
	@rm -f /tmp/tagged_output.txt /tmp/legacy_output.txt
	@rm -f /tmp/params_output.txt /tmp/param_out.txt
	@rm -f /tmp/pipeline_*.txt
	@rm -f test.out
	@echo "Cleanup complete"
