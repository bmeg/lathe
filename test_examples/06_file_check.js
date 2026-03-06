// Test 06: File Check
// Tests: File existence checking functionality

const workflow = jflow.Workflow("file_check_test");

// Create a file
const create = jflow.Process({
    name: "create_file",
    description: "Create a file to check",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'File exists!' > /tmp/checkme.txt"
    }],
    outputs: [{
        name: "file_to_check",
        url: "file:///tmp/checkme.txt",
        path: "/tmp/checkme.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

// Check the file
const checkFile = jflow.FileCheck({
    file: jflow.File({
        path: "/tmp/checkme.txt"
    })
});

workflow.Add(create);
workflow.Add(checkFile);
