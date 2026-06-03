# looter

Custom OpenTelemetry Collector distro for experimentation.

## Prerequisites

- Go 1.23+ (or whatever is required by the selected collector version)

## Build

```bash
make build
```

This generates the custom collector binary at:

- `dist/lootercol`

## Run

```bash
make run
```

Collector config:

- `config/collector.yaml`

## Default ports

- OTLP gRPC: `4317`
- OTLP HTTP: `4318`
- Health check: `13133`
- pprof: `1777`
- zPages: `55679`
