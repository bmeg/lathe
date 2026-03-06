# Examples

This directory contains longer-form example workflows that demonstrate the current `jflow` specification.

## Files

- `bwa.js` - Compact example showing reusable callable `jflow.Tool` templates.
- `bwa_workflow.js` - Larger multi-sample alignment/index workflow using the same template model.

## API Features Demonstrated

- `jflow.Workflow(name)`
- `jflow.Tool(spec)` returning a callable process factory
- `jflow.Path(name)` and `jflow.Object(name)` file constructors
- `jflow.Params` parameterized workflow configuration

## Run in Dry-Run Mode

Dry-run validates parsing and DAG construction without executing commands:

```bash
lathe run --dry-run examples/bwa.js
lathe run --dry-run examples/bwa_workflow.js
```

(Equivalent in this repo: `go run . run --dry-run ...`)

## Required Parameters

Both examples use default fallback paths under `/data` and `/work`. Provide your own paths via parameters to avoid missing-file checks.

### `examples/bwa.js`

```bash
lathe run --dry-run examples/bwa.js \
  --params reference=/abs/path/ref.fa \
  --params r1=/abs/path/sample_R1.fastq.gz \
  --params r2=/abs/path/sample_R2.fastq.gz \
  --params bam=/abs/path/output/sample.bam \
  --params bai=/abs/path/output/sample.bam.bai \
  --params threads=8
```

Optional image overrides:

- `bwa_image`
- `samtools_image`

You can also use a remote reference object instead of local `reference`:

```bash
lathe run --dry-run examples/bwa.js \
  --params reference_s3=s3://your-bucket/path/ref.fa \
  --params r1=/abs/path/sample_R1.fastq.gz \
  --params r2=/abs/path/sample_R2.fastq.gz
```

### `examples/bwa_workflow.js`

```bash
lathe run --dry-run examples/bwa_workflow.js \
  --params ref_genome=/abs/path/ref.fa \
  --params sample1_r1=/abs/path/sample1_R1.fastq.gz \
  --params sample1_r2=/abs/path/sample1_R2.fastq.gz \
  --params sample2_r1=/abs/path/sample2_R1.fastq.gz \
  --params sample2_r2=/abs/path/sample2_R2.fastq.gz \
  --params cpu_align=8 \
  --params ram_gb=16
```

Optional remote reference override:

```bash
lathe run --dry-run examples/bwa_workflow.js \
  --params ref_genome_s3=s3://your-bucket/path/ref.fa \
  --params sample1_r1=/abs/path/sample1_R1.fastq.gz \
  --params sample1_r2=/abs/path/sample1_R2.fastq.gz \
  --params sample2_r1=/abs/path/sample2_R1.fastq.gz \
  --params sample2_r2=/abs/path/sample2_R2.fastq.gz
```

## Notes

- These examples are intentionally not part of the quick test suite because they represent larger, realistic workflows.
- If Docker execution is enabled in your environment, remove `--dry-run` to execute tasks.
