// Example: Full BWA workflow in current jflow specification

const params = jflow.GetParams({
    ref_genome: "File",
    sample1_r1: "File",
    sample1_r2: "File",
    sample2_r1: "File",
    sample2_r2: "File",
    cpu_align: "Number"
});
const optionalParams = jflow.Params || {};
const wf = jflow.Workflow("bwa_workflow_example");
const tools = jflow.Import("./bwa.js");

const path = jflow.Path("path");
const obj = jflow.Object("object");

const bwaMem = tools.bwaMem;
const samtoolsIndex = tools.samtoolsIndex;

const reference = optionalParams.ref_genome_s3
    ? obj(optionalParams.ref_genome_s3)
    : path(params.ref_genome);

const alignSample1 = bwaMem({
    name: "align_sample_1",
    ref: reference,
    r1: path(params.sample1_r1),
    r2: path(params.sample1_r2),
    threads: params.cpu_align,
    bam: "/work/sample1.bam"
});

const alignSample2 = bwaMem({
    name: "align_sample_2",
    ref: reference,
    r1: path(params.sample2_r1),
    r2: path(params.sample2_r2),
    threads: params.cpu_align,
    bam: "/work/sample2.bam"
});

const indexSample1 = samtoolsIndex({
    name: "index_sample_1",
    bam: path("/work/sample1.bam"),
    bai: "/work/sample1.bam.bai"
});

const indexSample2 = samtoolsIndex({
    name: "index_sample_2",
    bam: path("/work/sample2.bam"),
    bai: "/work/sample2.bam.bai"
});

wf.Add(alignSample1);
wf.Add(alignSample2);
wf.Add(indexSample1);
wf.Add(indexSample2);
