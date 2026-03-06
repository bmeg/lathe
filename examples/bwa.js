// Example: Reusable BWA/Samtools templates in current jflow specification

const params = jflow.Params || {};
const wf = jflow.Workflow("bwa_templates_example");

const localPath = jflow.Path("local_file");
const remoteObject = jflow.Object("remote_object");

const bwaMem = jflow.Tool({
  name: "bwa_mem",
  commandLine: "bwa mem -t {{threads}} {{ref}} {{r1}} {{r2}} > {{bam}}",
  image: params.bwa_image || "quay.io/biocontainers/bwa:0.7.17--hed695b0_7",
  inputs: {
    ref: "reference FASTA",
    r1: "read 1 FASTQ",
    r2: "read 2 FASTQ"
  },
  outputs: {
    bam: "{{bam}}"
  },
  resources: {
    cpu_cores: params.threads || 4,
    ram_gb: params.ram_gb || 8
  }
});

const samtoolsIndex = jflow.Tool({
  name: "samtools_index",
  commandLine: "samtools index {{bam}} {{bai}}",
  image: params.samtools_image || "quay.io/biocontainers/samtools:1.20--h50ea8bc_0",
  inputs: {
    bam: "aligned BAM"
  },
  outputs: {
    bai: "{{bai}}"
  },
  resources: {
    cpu_cores: 1,
    ram_gb: 2
  }
});

const alignSample = bwaMem({
  name: "align_sample",
  ref: params.reference_s3
    ? remoteObject(params.reference_s3)
    : localPath(params.reference || "/data/reference.fa"),
  r1: localPath(params.r1 || "/data/sample_R1.fastq.gz"),
  r2: localPath(params.r2 || "/data/sample_R2.fastq.gz"),
  threads: params.threads || 4,
  bam: params.bam || "/work/sample.bam"
});

const indexSample = samtoolsIndex({
  name: "index_sample",
  bam: localPath(params.bam || "/work/sample.bam"),
  bai: params.bai || "/work/sample.bam.bai"
});

wf.Add(alignSample);
wf.Add(indexSample);