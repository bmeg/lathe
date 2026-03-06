// Example module: reusable BWA/Samtools tool templates for jflow.Import

const rawParams = jflow.Params || {};

const optionalSchema = {};
if (rawParams.bwa_image !== undefined) {
  optionalSchema.bwa_image = "String";
}
if (rawParams.samtools_image !== undefined) {
  optionalSchema.samtools_image = "String";
}
if (rawParams.threads !== undefined) {
  optionalSchema.threads = "Number";
}
if (rawParams.ram_gb !== undefined) {
  optionalSchema.ram_gb = "Number";
}

const typedParams = Object.keys(optionalSchema).length > 0
  ? jflow.GetParams(optionalSchema)
  : {};

const params = {
  bwa_image: "quay.io/biocontainers/bwa:0.7.17--hed695b0_7",
  samtools_image: "quay.io/biocontainers/samtools:1.20--h50ea8bc_0",
  threads: 4,
  ram_gb: 8,
  ...typedParams
};

const bwaMem = jflow.Tool({
  name: "bwa_mem",
  commandLine: "bwa mem -t {{threads}} {{ref}} {{r1}} {{r2}} > {{bam}}",
  image: params.bwa_image,
  inputs: {
    ref: "File",
    r1: "File",
    r2: "File",
    threads: "Value",
    bam: "Value"
  },
  outputs: {
    bam: "{{bam}}",
    bamMatches: "{{bam}}*"
  },
  resources: {
    cpu_cores: params.threads,
    ram_gb: params.ram_gb
  }
});

const samtoolsIndex = jflow.Tool({
  name: "samtools_index",
  commandLine: "samtools index {{bam}} {{bai}}",
  image: params.samtools_image,
  inputs: {
    bam: "File",
    bai: "Value"
  },
  outputs: {
    bai: "{{bai}}"
  },
  resources: {
    cpu_cores: 1,
    ram_gb: 2
  }
});

export { bwaMem, samtoolsIndex };