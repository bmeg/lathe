// Test 12: Path/Object Factories
// Tests: jflow.Path() and jflow.Object() constructor behavior in template values

const workflow = jflow.Workflow("path_object_factories_test");

const prepareInput = jflow.Process({
    name: "prepare_factory_input",
    description: "Create local file consumed via Path factory",
    executors: [{
        image: "ubuntu:20.04",
        command: "echo 'factory-input' > /tmp/factory_input.txt"
    }],
    outputs: [{
        name: "factory_input",
        url: "file:///tmp/factory_input.txt",
        path: "/tmp/factory_input.txt",
        type: "FILE"
    }],
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

const localPath = jflow.Path("local_input");
const remoteObject = jflow.Object("remote_reference");

const summarizeTool = jflow.Tool({
    name: "summarize_paths",
    commandLine: "echo local={{local}} > {{output}} && echo remote={{remote}} >> {{output}} && cat {{local}} >> {{output}}",
    image: "ubuntu:20.04",
    inputs: {
        local: "File",
        remote: "Value",
        output: "Value"
    },
    outputs: {
        output: "{{output}}"
    },
    resources: {
        cpu_cores: 1,
        ram_gb: 0.5
    }
});

const summarize = summarizeTool({
    name: "summarize_path_and_object",
    local: localPath("/tmp/factory_input.txt"),
    remote: remoteObject("s3://example-bucket/example-key"),
    output: "/tmp/path_object_summary.txt"
});

workflow.Add(prepareInput);
workflow.Add(summarize);
