// Test 02: File Operations
// Tests: Input/output file handling, command chaining

const workflow = jflow.Workflow("file_operations_test");

// Create a test file
const create = jflow.Process({
    name: "create_file",
    description: "Create a test input file",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'test data line 1\ntest data line 2\ntest data line 3' > /tmp/input.txt"
    }],
    outputs: [{
        name: "input_file",
        url: "file:///tmp/input.txt",
        path: "/tmp/input.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

// Process the file
const process = jflow.Process({
    name: "process_file",
    description: "Process the input file",
    executors: [{
        image: "ubuntu:20.04",
        command: "cat /data/input.txt | wc -l > /data/output.txt"
    }],
    inputs: [{
        name: "input",
        url: "file:///tmp/input.txt",
        path: "/data/input.txt",
        type: "FILE"
    }],
    outputs: [{
        name: "result",
        url: "file:///tmp/output.txt",
        path: "/data/output.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(create);
workflow.Add(process);
