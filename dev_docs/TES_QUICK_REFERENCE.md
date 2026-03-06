# TES Format Quick Reference

## Basic Task Structure

```javascript
jflow.Process({
  name: "task_name",
  description: "What this task does",
  executors: [{ image: "...", command: "..." }],
  inputs: [{ url: "...", path: "..." }],
  outputs: [{ url: "...", path: "..." }],
  resources: { cpu_cores: 4, ram_gb: 8 },
  volumes: ["/data"],
  tags: { key: "value" }
})
```

## Command Formats

### String (Shell Features)
```javascript
executors: [{
  image: "ubuntu:20.04",
  command: "cat file.txt | grep pattern > output.txt"
}]
```

### Array (Precise Control)
```javascript
executors: [{
  image: "ubuntu:20.04",
  command: ["grep", "pattern", "file.txt"]
}]
```

## Executors

### Single Executor
```javascript
executors: [{
  image: "python:3.11",
  command: "python script.py",
  workdir: "/workspace",
  env: { "VAR": "value" },
  stdout: "/logs/out.log",
  stderr: "/logs/err.log"
}]
```

### Multiple Executors (Sequential)
```javascript
executors: [
  { image: "python:3.11", command: "prep.py" },
  { image: "r-base:4.2", command: "analyze.R" },
  { image: "python:3.11", command: "report.py" }
]
```

## Inputs

### From URL
```javascript
inputs: [{
  name: "input_data",
  url: "s3://bucket/data.csv",
  path: "/data/input.csv",
  type: "FILE"
}]
```

### Inline Content
```javascript
inputs: [{
  name: "config",
  content: "key=value\nflag=true",
  path: "/config/settings.conf"
}]
```

### Multiple Inputs
```javascript
inputs: [
  { url: "s3://bucket/ref.fa", path: "/ref/genome.fa" },
  { url: "s3://bucket/reads.fq", path: "/data/reads.fq" },
  { content: "settings", path: "/config/app.conf" }
]
```

## Outputs

### Single Output
```javascript
outputs: [{
  name: "result",
  url: "s3://bucket/output.txt",
  path: "/data/output.txt",
  type: "FILE"
}]
```

### Wildcard Outputs
```javascript
outputs: [{
  name: "logs",
  url: "s3://bucket/logs/",
  path: "/output/*.log",
  path_prefix: "/output/"
}]
```

### Multiple Outputs
```javascript
outputs: [
  { url: "s3://bucket/result.csv", path: "/data/result.csv" },
  { url: "s3://bucket/report.html", path: "/data/report.html" },
  { url: "s3://bucket/logs/", path: "/logs/*" }
]
```

## Resources

### Basic
```javascript
resources: {
  cpu_cores: 4,
  ram_gb: 8,
  disk_gb: 100
}
```

### Advanced
```javascript
resources: {
  cpu_cores: 16,
  ram_gb: 32,
  disk_gb: 500,
  preemptible: true,
  zones: ["us-west-1", "us-west-2"],
  backend_parameters: {
    "instance_type": "c5.4xlarge",
    "spot_price": "0.10"
  },
  backend_parameters_strict: false
}
```

## Volumes

```javascript
volumes: ["/data", "/workspace", "/tmp"]
```

Shared across all executors in the task.

## Tags

```javascript
tags: {
  "project": "genomics",
  "sample": "sample-001",
  "version": "1.0",
  "cost_center": "research"
}
```

## URL Schemes

| Scheme | Example | Use Case |
|--------|---------|----------|
| `s3://` | `s3://bucket/file.txt` | AWS S3 |
| `gs://` | `gs://bucket/file.txt` | Google Cloud Storage |
| `file://` | `file:///path/to/file.txt` | Local filesystem |
| `http://` | `http://example.com/data.csv` | HTTP download |
| `https://` | `https://example.com/data.csv` | HTTPS download |

## File Types

```javascript
type: "FILE"       // Single file
type: "DIRECTORY"  // Directory
```

## Complete Example

```javascript
const task = jflow.Process({
  name: "genomic_analysis",
  description: "Variant calling pipeline",
  
  executors: [
    {
      image: "broadinstitute/gatk:latest",
      command: "gatk HaplotypeCaller -R /ref/genome.fa -I /data/input.bam -O /data/variants.vcf",
      env: {
        "JAVA_OPTS": "-Xmx8g"
      }
    },
    {
      image: "biocontainers/bcftools:latest",
      command: ["bcftools", "filter", "-i", "QUAL>30", "/data/variants.vcf", "-o", "/data/filtered.vcf"]
    }
  ],
  
  inputs: [
    {
      name: "reference",
      url: "s3://genomics/hg38.fa",
      path: "/ref/genome.fa",
      type: "FILE"
    },
    {
      name: "alignment",
      url: "s3://samples/sample-001.bam",
      path: "/data/input.bam",
      type: "FILE"
    }
  ],
  
  outputs: [
    {
      name: "variants",
      url: "s3://results/sample-001.vcf",
      path: "/data/filtered.vcf",
      type: "FILE"
    },
    {
      name: "logs",
      url: "s3://logs/sample-001/",
      path: "/data/*.log",
      path_prefix: "/data/"
    }
  ],
  
  volumes: ["/data", "/ref"],
  
  resources: {
    cpu_cores: 8,
    ram_gb: 16,
    disk_gb: 200,
    preemptible: false
  },
  
  tags: {
    "project": "exome_seq",
    "sample": "sample-001",
    "stage": "variant_calling",
    "version": "2.0"
  }
});
```

## Legacy Format (Still Supported)

```javascript
jflow.Process({
  name: "legacy_task",
  commandLine: "python script.py input.txt output.txt",
  image: "python:3.11",
  inputs: { input: "input.txt" },
  outputs: { output: "output.txt" },
  cpus: 4,
  memoryMB: 8192
})
```

## Common Patterns

### Python Script
```javascript
executors: [{
  image: "python:3.11-slim",
  command: "python /app/script.py --input /data/in.csv --output /data/out.csv",
  env: { "PYTHONPATH": "/app" }
}]
```

### R Script
```javascript
executors: [{
  image: "r-base:4.2",
  command: "Rscript /app/analyze.R /data/input.csv /data/output.csv"
}]
```

### Shell Script
```javascript
executors: [{
  image: "ubuntu:20.04",
  command: "bash /scripts/process.sh /data/input /data/output"
}]
```

### Bioinformatics Pipeline
```javascript
executors: [
  {
    image: "biocontainers/fastqc:v0.11.9",
    command: ["fastqc", "/data/reads.fq", "-o", "/data/qc"]
  },
  {
    image: "biocontainers/trimmomatic:0.39",
    command: "trimmomatic SE /data/reads.fq /data/trimmed.fq LEADING:20"
  },
  {
    image: "biocontainers/bwa:0.7.17",
    command: ["bwa", "mem", "/ref/genome.fa", "/data/trimmed.fq"]
  }
]
```

### Data Pipeline
```javascript
executors: [
  {
    image: "python:3.11",
    command: "python extract.py /raw/data.json /staged/data.csv"
  },
  {
    image: "python:3.11",
    command: "python transform.py /staged/data.csv /processed/data.csv"
  },
  {
    image: "python:3.11",
    command: "python load.py /processed/data.csv /output/final.csv"
  }
]
```

## Tips

1. **Use absolute paths** in containers (e.g., `/data/file.txt`)
2. **Share volumes** between executors for multi-step tasks
3. **Use string commands** when you need shell features (pipes, redirects)
4. **Use array commands** for precise argument control
5. **Tag your tasks** for easy filtering and organization
6. **Specify resources** to optimize execution
7. **Use wildcards** for multiple output files
8. **Leverage inline content** for configuration files
9. **Enable preemptible** for cost savings on non-critical tasks
10. **Test with small data** before scaling up

## References

- [TES Alignment Guide](TES_ALIGNMENT_GUIDE.md) - Complete guide
- [JFLOW Standard](JFLOW_STANDARD.md) - Full specification
- [GA4GH TES API](https://ga4gh.github.io/task-execution-schemas/docs/) - Official spec
