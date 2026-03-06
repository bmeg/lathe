// Example: Full BWA workflow in current jflow specification

const params = jflow.Params || {};
const wf = jflow.Workflow("bwa_workflow_example");

const path = jflow.Path("path");
const obj = jflow.Object("object");

const bwaMem = jflow.Tool({
    name: "bwa_mem",
    commandLine: "bwa mem -t {{threads}} {{ref}} {{r1}} {{r2}} > {{bam}}",
    image: params.bwa_image || "quay.io/biocontainers/bwa:0.7.17--hed695b0_7",
    inputs: { ref: "reference", r1: "read1", r2: "read2" },
    outputs: { bam: "{{bam}}" },
    resources: {
        cpu_cores: params.cpu_align || 4,
        ram_gb: params.ram_gb || 8
    }
});

const samtoolsIndex = jflow.Tool({
    name: "samtools_index",
    commandLine: "samtools index {{bam}} {{bai}}",
    image: params.samtools_image || "quay.io/biocontainers/samtools:1.20--h50ea8bc_0",
    inputs: { bam: "bam" },
    outputs: { bai: "{{bai}}" },
    resources: {
        cpu_cores: 1,
        ram_gb: 2
    }
});

const reference = params.ref_genome_s3
    ? obj(params.ref_genome_s3)
    : path(params.ref_genome || "/data/ref.fa");

const alignSample1 = bwaMem({
    name: "align_sample_1",
    ref: reference,
    r1: path(params.sample1_r1 || "/data/sample1_R1.fastq.gz"),
    r2: path(params.sample1_r2 || "/data/sample1_R2.fastq.gz"),
    threads: params.cpu_align || 4,
    bam: "/work/sample1.bam"
});

const alignSample2 = bwaMem({
    name: "align_sample_2",
    ref: reference,
    r1: path(params.sample2_r1 || "/data/sample2_R1.fastq.gz"),
    r2: path(params.sample2_r2 || "/data/sample2_R2.fastq.gz"),
    threads: params.cpu_align || 4,
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
