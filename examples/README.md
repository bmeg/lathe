# Examples

This directory contains longer-form example workflows that demonstrate the current `jflow` specification.

## Files

- `bwa.js` - Importable module exporting reusable callable `jflow.Tool` templates.
- `bwa_workflow.js` - Larger multi-sample alignment/index workflow that imports `bwa.js`.
- `min_ref.fa`, `min_R1.fastq`, `min_R2.fastq` - Minimal local test inputs.
- `min_ref.fa.gz`, `min_R1.fastq.gz`, `min_R2.fastq.gz` - Compressed variants of minimal test inputs.
- `bwa_params.yaml` - Parameter file for running `bwa_workflow.js` with minimal local test inputs.

## API Features Demonstrated

- `jflow.Workflow(name)`
- `jflow.Tool(spec)` returning a callable process factory
- `jflow.Path(name)` and `jflow.Object(name)` file constructors
- `jflow.Params` parameterized workflow configuration
- `jflow.GetParams(schema)` typed parameter validation and file normalization
- `jflow.Import(path, paramsOverride?)` for module-style tool sharing with optional parameter overrides
- `export` syntax in imported sub-scripts
- `Tool.inputs` typed as `{ variableName: "File" | "Value" }`
- `Tool.outputs` typed as `{ outputName: "glob/template" }`

### Optional Typed Module Parameters

Imported modules can validate only parameters that are provided by the caller by building a dynamic schema and passing it to `jflow.GetParams`.

```javascript
const rawParams = jflow.Params || {};
const optionalSchema = {};
if (rawParams.threads !== undefined) optionalSchema.threads = "Number";
if (rawParams.bwa_image !== undefined) optionalSchema.bwa_image = "String";

const typedParams = Object.keys(optionalSchema).length > 0
  ? jflow.GetParams(optionalSchema)
  : {};
```

This pattern keeps defaults in the module while still enforcing types for user-provided overrides.

## Run in Dry-Run Mode

Dry-run validates parsing and DAG construction without executing commands:

```bash
lathe run --dry-run --params-file examples/bwa_params.yaml examples/bwa_workflow.js
```

(Equivalent in this repo: `go run . run --dry-run ...`)

## Required Parameters

Both examples use default fallback paths under `/data` and `/work`. Provide your own paths via parameters to avoid missing-file checks.

### `examples/bwa.js`

This file is a module imported by `bwa_workflow.js` and is not intended to be run directly.
It exports:

- `bwaMem`
- `samtoolsIndex`

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

Minimal local fixture run (using YAML parameter file):

```bash
lathe run --dry-run --params-file examples/bwa_params.yaml examples/bwa_workflow.js
```

## Notes

- These examples are intentionally not part of the quick test suite because they represent larger, realistic workflows.
- If Docker execution is enabled in your environment, remove `--dry-run` to execute tasks.
