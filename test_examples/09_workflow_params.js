// Test 09: Workflow Parameters
// Tests: Runtime parameter passing to workflows

const workflow = jflow.Workflow("params_test");

// Use parameters with defaults
const message = jflow.Params.message || "default message";
const outputPath = jflow.Params.output || "/tmp/params_output.txt";

const paramTask = jflow.Process({
    name: "param_task",
    description: "Task using workflow parameters",
    executors: [{
        image: "ubuntu:20.04",
        command: `echo '${message}' > /tmp/param_out.txt`
    }],
    outputs: [{
        name: "output",
        url: `file://${outputPath}`,
        path: "/tmp/param_out.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

workflow.Add(paramTask);
