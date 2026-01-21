# Pupload

<!-- ![Pupload Logo](docs/assets/pupload-logo.svg) -->

File upload orchestration for distributed processing pipelines.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pupload/pupload)](go.mod)

[Docs](https://docs.pupload.com) · [Discord](https://discord.gg/pupload) · [Examples](https://github.com/pupload/examples)

---

## What is this?

Pupload is a workflow engine for file processing.

You upload a file. It gets processed through multiple steps (resize, transcode, analyze, whatever). Each step runs in an isolated container with the resources it needs. Results flow to the next step automatically.

Define your pipeline in YAML. Deploy workers with different capabilities (CPU-only boxes, GPU machines, high-memory instances). The controller routes work to appropriate workers based on what each step needs.

No vendor lock-in. Runs on your infrastructure. Works with any S3-compatible storage.

Built with Go. Designed for teams that process files at scale but don't want to manage the plumbing.

## Quick Start

Requires Go 1.25+, Docker, and Redis.

```bash
# Install
go install github.com/pupload/pupload/cmd/pup@latest

# Initialize
mkdir my-project && cd my-project
pup init

# Run
pup dev
```

Example (`flows/resize.yaml`):

```yaml
Name: image-resize

Stores:
  - Name: uploads
    Type: s3
    Params:
      BucketName: my-uploads

DataWells:
  - Edge: input-image
    Store: uploads
    Source: upload
  - Edge: resized-image
    Store: uploads

Nodes:
  - ID: resize
    Uses: pupload/ffmpeg
    Inputs:
      - Name: VideoStream
        Edge: input-image
    Outputs:
      - Name: VideoOut
        Edge: resized-image
    Command: Encode
```

Test it:

```bash
pup test image-resize
```

## Install

```bash
# Go
go install github.com/pupload/pupload/cmd/pup@latest

# Source
git clone https://github.com/pupload/pupload.git
cd pupload
go build -o pup ./cmd/pup

# Docker
docker pull pupload/pupload:latest
```

[Releases](https://github.com/pupload/pupload/releases) for binaries.

## Why Pupload?

Most workflow engines are built for generic task orchestration. Pupload is built specifically for files.

**When to use Pupload:**
- You're processing user uploads (images, videos, documents)
- You need different processing tiers (cheap transcoding vs expensive ML models)
- Your pipeline has multiple steps (upload → resize → watermark → store)
- You want to scale workers independently based on GPU/CPU availability

**vs. rolling your own:** You don't have to build presigned URL handling, container orchestration, retry logic, resource scheduling, or distributed state management.

## Comparison

|  | Pupload | Temporal | Airflow | AWS Lambda |
|---|---|---|---|---|
| **File processing focus** | ✓ Built-in presigned URLs, storage | DIY | DIY | DIY |
| **Multi-step workflows** | ✓ DAG-based YAML | ✓ Code-based | ✓ Python DAGs | Requires Step Functions |
| **Container execution** | ✓ Docker native | DIY | Via operators | Via images |
| **GPU support** | ✓ NVIDIA, AMD, Intel, Apple | DIY routing | DIY operators | Limited (A10G only) |
| **Resource-based scheduling** | ✓ Built-in tiers | Manual | Manual | Fixed configs |
| **MIME type validation** | ✓ Validates types through pipeline | None | None | None |
| **Local testing** | ✓ `pup dev` runs everything | ✓ Local clusters | ✓ Standalone mode | SAM/LocalStack |
| **Observability** | OpenTelemetry | Built-in UI + tracing | Web UI | CloudWatch |
| **Version control** | ✓ YAML in git | ✓ Code in git | ✓ Code in git | JSON/YAML in git |
| **Vendor lock-in** | None | None | None | AWS only |
| **Best for** | File uploads → processing | General workflows | Batch/data pipelines | Simple functions |

## Use Cases

**Image/Video Processing**
```yaml
upload → thumbnail → watermark → multiple resolutions → CDN
```

**Document Pipeline**
```yaml
PDF upload → extract text → OCR scanned pages → index → store
```

**ML Inference**
```yaml
upload → preprocess → GPU inference → postprocess → results
```

**Archive Processing**
```yaml
zip upload → extract → virus scan → process each file → store
```

Mix CPU and GPU nodes. Scale GPU workers separately. Run expensive steps on dedicated hardware.

## How it works

**Flows** are YAML pipelines. They contain nodes (steps), edges (connections), datawells (input/output), and stores (storage backends).

**Nodes** run in Docker. Each gets presigned URLs for inputs/outputs. On completion, uploads results and triggers downstream nodes.

**Edges** connect nodes in a DAG. Validation prevents cycles.

**Resource tiers**:
- `c-small/medium/large` - CPU
- `m-small/medium/large` - Memory
- `g-small/medium/large` - GPU (NVIDIA, AMD, Intel, Apple)
- `gn-*/ga-*/gi-*` - Vendor-specific

Workers subscribe to resource queues. Controller schedules, workers execute.

## CLI

```bash
pup init                  # New project
pup dev                   # Start controller + worker
pup test <flow>           # Test flow
pup controller list       # List controllers
pup controller add <url>  # Add controller
```

## Configuration

Controller config (`config/controller.yaml`):

```yaml
controller:
  port: 8080
  redis:
    addr: localhost:6379
  storage:
    type: s3
    bucket: uploads
    endpoint: s3.amazonaws.com
```

Worker config (`config/worker.yaml`):

```yaml
worker:
  redis:
    addr: localhost:6379
  docker:
    enabled: true
  resources:
    - c-medium    # Subscribe to medium CPU tasks
    - g-small     # Subscribe to small GPU tasks
```

Workers only pick up tasks matching their subscribed resource tiers. Run cheap workers for CPU tasks, expensive GPU workers only when needed.

## Development

```bash
# Clone
git clone https://github.com/pupload/pupload.git
cd pupload

# Install deps
go mod download

# Run tests
go test ./...

# Run specific package tests
go test ./internal/validation/...

# Build
go build -o pup ./cmd/pup

# Run locally
./pup dev
```

### Project Structure

```
pupload/
├── cmd/pup/          # CLI entry point
├── internal/
│   ├── controller/   # Flow orchestration, scheduling
│   ├── worker/       # Task execution, container runtime
│   ├── validation/   # Flow validation (check flows before running)
│   ├── models/       # Data models (Flow, Node, Edge, etc.)
│   └── sync/         # Redis sync layer (queues, locks)
├── pkg/              # Public APIs (if you want to embed Pupload)
└── flows/            # Example flows
```

### Adding a Node Type

Create a Docker image that:
1. Accepts presigned URLs as env vars (`INPUT_*`, `OUTPUT_*`)
2. Downloads inputs, processes them, uploads outputs
3. Exits 0 on success

Register it in your flow YAML:

```yaml
Nodes:
  - ID: my-custom-node
    Uses: myregistry/my-processor:latest
    Inputs:
      - Name: InputFile
        Edge: some-edge
    # ...
```

### Running Integration Tests

```bash
# Start dependencies
docker-compose up -d redis minio

# Run tests
go test ./internal/integration/... -v

# Clean up
docker-compose down
```

## Contributing

1. Fork the repo
2. Create branch (`git checkout -b feature/thing`)
3. Commit (`git commit -am 'Add thing'`)
4. Push (`git push origin feature/thing`)
5. Open PR

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Community

- [Discord](https://discord.gg/pupload) - Questions and discussion
- [GitHub Discussions](https://github.com/pupload/pupload/discussions) - Ideas and feedback
- [Twitter](https://twitter.com/pupload) - Updates

## License

MIT © Pupload Contributors

---

For teams that process user uploads and don't want to reinvent workflow orchestration.