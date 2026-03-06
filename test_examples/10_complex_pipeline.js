// Test 10: Complex Pipeline
// Tests: Multi-step pipeline with data dependencies

const workflow = jflow.Workflow("complex_pipeline_test");

// Step 1: Generate data
const generate = jflow.Process({
    name: "generate_data",
    description: "Generate test data",
    executors: [{
        image: "ubuntu:20.04",
        command: "for i in 1 2 3 4 5; do echo 'Data line '$i; done > /tmp/data.txt"
    }],
    outputs: [{
        name: "raw_data",
        url: "file:///tmp/pipeline_data.txt",
        path: "/tmp/data.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

// Step 2: Process data
const process = jflow.Process({
    name: "process_data",
    description: "Process the data",
    executors: [{
        image: "ubuntu:20.04",
        command: "cat /data/input.txt | grep 'Data' | wc -l > /data/count.txt"
    }],
    inputs: [{
        name: "input",
        url: "file:///tmp/pipeline_data.txt",
        path: "/data/input.txt",
        type: "FILE"
    }],
    outputs: [{
        name: "processed",
        url: "file:///tmp/pipeline_count.txt",
        path: "/data/count.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

// Step 3: Summarize
const summarize = jflow.Process({
    name: "summarize",
    description: "Create summary",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'Line count:' && cat /data/count.txt > /data/summary.txt"
    }],
    inputs: [{
        name: "count",
        url: "file:///tmp/pipeline_count.txt",
        path: "/data/count.txt",
        type: "FILE"
    }],
    outputs: [{
        name: "summary",
        url: "file:///tmp/pipeline_summary.txt",
        path: "/data/summary.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(generate);
workflow.Add(process);
workflow.Add(summarize);
