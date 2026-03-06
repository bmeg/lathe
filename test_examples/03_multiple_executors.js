// Test 03: Multiple Executors
// Tests: Sequential executor execution, data passing between executors

const workflow = jflow.Workflow("multiple_executors_test");

const multiStep = jflow.Process({
    name: "multi_step_process",
    description: "Process with multiple sequential executors",
    executors: [
        {
            image: "ubuntu:20.04",
            command: "echo 'Step 1 complete' > /data/step1.txt"
        },
        {
            image: "ubuntu:20.04",
            command: "cat /data/step1.txt && echo 'Step 2 complete' > /data/step2.txt"
        },
        {
            image: "ubuntu:20.04",
            command: "cat /data/step2.txt && echo 'All steps complete' > /data/final.txt"
        }
    ],
    outputs: [{
        name: "final_output",
        url: "file:///tmp/multi_exec_final.txt",
        path: "/data/final.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(multiStep);
