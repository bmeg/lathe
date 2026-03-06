// Test 01: Basic Hello World
// Tests: Simple command execution, output file creation

const workflow = jflow.Workflow("hello_world_test");

const hello = jflow.Process({
    name: "hello",
    description: "Basic hello world test",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'Hello from jflow test!' > /tmp/hello.txt"
    }],
    outputs: [{
        name: "greeting",
        url: "file:///tmp/hello.txt",
        path: "/tmp/hello.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(hello);
