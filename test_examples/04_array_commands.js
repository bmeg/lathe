// Test 04: Array Command Format
// Tests: Command as array of strings (non-shell mode)

const workflow = jflow.Workflow("array_commands_test");

const arrayCmd = jflow.Process({
    name: "array_command",
    description: "Test command as array of strings",
    executors: [{
        image: "ubuntu:20.04",
        command: ["/bin/sh", "-c", "echo 'Array command test' > /tmp/array_output.txt"]
    }],
    outputs: [{
        name: "output",
        url: "file:///tmp/array_output.txt",
        path: "/tmp/array_output.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(arrayCmd);
