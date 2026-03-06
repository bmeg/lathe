// Test 07: Metadata and Tags
// Tests: TES tags for task metadata

const workflow = jflow.Workflow("metadata_tags_test");

const taggedTask = jflow.Process({
    name: "tagged_task",
    description: "Task with metadata tags",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'Task with tags' > /tmp/tagged_output.txt"
    }],
    outputs: [{
        name: "output",
        url: "file:///tmp/tagged_output.txt",
        path: "/tmp/tagged_output.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    },
    tags: {
        "test_type": "metadata",
        "priority": "high",
        "version": "1.0"
    }
});

workflow.Add(taggedTask);
