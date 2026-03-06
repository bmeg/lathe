#!/bin/bash
# jflow Test Runner Script
# Provides convenient interface for running jflow workflow tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERBOSE=false
LIST_TESTS=false
TEST_NAME=""
TIMEOUT="10m"

# Print usage
usage() {
    echo "Usage: $0 [OPTIONS] [test_name]"
    echo ""
    echo "Run jflow workflow tests"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --verbose  Run tests with verbose output"
    echo "  -l, --list     List all available tests"
    echo "  -t, --timeout  Set test timeout (default: 10m)"
    echo "  --validation   Run validation only (no execution)"
    echo "  --bench        Run benchmarks"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all tests"
    echo "  $0 hello_world        # Run specific test"
    echo "  $0 -v                 # Run all tests with verbose output"
    echo "  $0 --list             # List all available tests"
    echo "  $0 --validation       # Validate all workflows"
    echo ""
    exit 0
}

# List available tests
list_tests() {
    echo -e "${BLUE}Available jflow tests:${NC}"
    echo ""
    
    if [ -d "test_examples" ]; then
        for file in test_examples/*.js; do
            if [ -f "$file" ]; then
                basename "$file" .js | sed 's/^/  - /'
            fi
        done
    else
        echo -e "${RED}Error: test_examples directory not found${NC}"
        exit 1
    fi
    
    echo ""
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -l|--list)
            LIST_TESTS=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --validation)
            echo -e "${BLUE}Running workflow validation tests...${NC}"
            go test ./tests -v -run TestWorkflowValidation -timeout "$TIMEOUT"
            exit $?
            ;;
        --bench)
            echo -e "${BLUE}Running benchmarks...${NC}"
            go test ./tests -bench=. -benchmem
            exit $?
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
        *)
            TEST_NAME="$1"
            shift
            ;;
    esac
done

# Handle --list flag
if [ "$LIST_TESTS" = true ]; then
    list_tests
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Must be run from the project root directory${NC}"
    exit 1
fi

# Check if test_examples directory exists
if [ ! -d "test_examples" ]; then
    echo -e "${RED}Error: test_examples directory not found${NC}"
    echo "Please ensure test examples are in place"
    exit 1
fi

# Build verbose flag
VERBOSE_FLAG=""
if [ "$VERBOSE" = true ]; then
    VERBOSE_FLAG="-v"
fi

# Run tests
if [ -z "$TEST_NAME" ]; then
    # Run all tests
    echo -e "${BLUE}Running all jflow workflow tests...${NC}"
    echo ""
    
    # First check if files exist
    echo -e "${YELLOW}Checking test files...${NC}"
    if ! go test ./tests $VERBOSE_FLAG -run TestWorkflowFilesExist -timeout 10s; then
        echo -e "${RED}Test file check failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ All test files found${NC}"
    echo ""
    
    # Run validation
    echo -e "${YELLOW}Validating workflows...${NC}"
    if ! go test ./tests $VERBOSE_FLAG -run TestWorkflowValidation -timeout 30s; then
        echo -e "${RED}Workflow validation failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ All workflows validated${NC}"
    echo ""
    
    # Run full test suite
    echo -e "${YELLOW}Running workflow execution tests...${NC}"
    if go test ./tests $VERBOSE_FLAG -run TestAllWorkflows -timeout "$TIMEOUT"; then
        echo ""
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo ""
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
else
    # Run specific test
    echo -e "${BLUE}Running test: ${TEST_NAME}${NC}"
    echo ""
    
    export JFLOW_TEST_NAME="$TEST_NAME"
    if go test ./tests $VERBOSE_FLAG -run TestIndividualWorkflow -timeout "$TIMEOUT"; then
        echo ""
        echo -e "${GREEN}✓ Test passed: ${TEST_NAME}${NC}"
        exit 0
    else
        echo ""
        echo -e "${RED}✗ Test failed: ${TEST_NAME}${NC}"
        exit 1
    fi
fi
