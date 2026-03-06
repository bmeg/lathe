// Test 08: Legacy Format Support
// Tests: Backward compatibility with legacy jflow format

const workflow = jflow.Workflow("legacy_format_test");

const legacyTask = jflow.Process({
    name: "legacy_task",
    description: "Task using legacy format",
    commandLine: "echo 'Legacy format works!' > /tmp/legacy_output.txt",
    image: "ubuntu:20.04",
    outputs: {
        "legacy_out": "/tmp/legacy_output.txt"
    },
    cpus: 1,
    memoryMB: 512
});

workflow.Add(legacyTask);
