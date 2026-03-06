// Test 05: Resource Specification
// Tests: CPU and memory resource requirements

const workflow = jflow.Workflow("resource_spec_test");

const resourceTest = jflow.Process({
    name: "resource_test",
    description: "Test with specific resource requirements",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'Testing with 2 cores and 1GB RAM' > /tmp/resource_test.txt && sleep 1"
    }],
    outputs: [{
        name: "output",
        url: "file:///tmp/resource_test.txt",
        path: "/tmp/resource_test.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 2,
        ram_gb: 1.0,
        disk_gb: 2
    }
});

workflow.Add(resourceTest);
