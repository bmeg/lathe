// Test 11: Callable Tool Template
// Tests: jflow.Tool() returning callable function, template substitution, Path factory

const workflow = jflow.Workflow("tool_template_callable_test");

// Producer step: create an input file used by the tool template instance
const prepareInput = jflow.Process({
    name: "prepare_tool_input",
    description: "Create input for callable tool template",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'matrix-data' > /tmp/tool_template_input.txt"
    }],
    outputs: [{
        name: "tool_input",
        url: "file:///tmp/tool_template_input.txt",
        path: "/tmp/tool_template_input.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

const inputMatrix = jflow.Path("input_matrix");

const matrixTool = jflow.Tool({
    name: "matrix_tool",
    commandLine: "cat {{input1}} > {{output1}} && echo {{param1}} >> {{output1}}",
    image: "ubuntu:20.04",
    inputs: {
        input1: "matrix input"
    },
    outputs: {
        output1: "{{output1}}"
    },
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

const runMatrixTool = matrixTool({
    name: "run_matrix_tool",
    input1: inputMatrix("/tmp/tool_template_input.txt"),
    output1: "/tmp/tool_template_output.txt",
    param1: 5
});

workflow.Add(prepareInput);
workflow.Add(runMatrixTool);
