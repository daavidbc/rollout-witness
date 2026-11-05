# Rollout Witness

Rollout Witness is a Go CLI for measuring HTTP availability and latency during application deployments and Kubernetes rollouts. It sends repeated requests to a target endpoint, summarizes the results, and writes one JSON record per request to a JSONL evidence file.

## Features

- Configurable target URL, request count, timeout, and delay between requests.
- Optional HTTP `Host` header for host-based Ingress routing.
- Classification of HTTP 2xx responses as successes, non-2xx responses as HTTP failures, and connection errors as separate failures.
- Summary statistics: request counts, failure rate, average latency, average successful-request latency, minimum, maximum, and nearest-rank p95 latency.
- Per-request JSONL evidence containing a timestamp, HTTP status code, duration in milliseconds, and network error details.
- Nonzero exit status for invalid configuration, evidence file errors, or unsuccessful HTTP requests.

## Requirements

- Go with support for the standard-library `slices` package (Go 1.21 or newer).
- Network access to the endpoint you want to monitor.

No external Go dependencies are required by the current CLI.

## Run

From the repository root:

```bash
go run ./cmd/rollout-witness \
  --url http://localhost:8080/health \
  --requests 10 \
  --interval 1s \
  --timeout 3s \
  --output evidence.jsonl
```

Or build a binary:

```bash
go build -o rollout-witness ./cmd/rollout-witness
./rollout-witness --url http://localhost:8080/health --requests 10
```

## CLI flags

| Flag | Default | Description |
| --- | --- | --- |
| `--url` | `http://podinfo.lab.test/readyz` | HTTP endpoint to monitor |
| `--requests` | `10` | Number of requests; must be greater than zero |
| `--timeout` | `3s` | HTTP client timeout; must be greater than zero |
| `--interval` | `1s` | Delay between requests; must not be negative (`0s` is allowed) |
| `--host-header` | empty | Optional HTTP `Host` header override |
| `--output` | `evidence.jsonl` | JSONL evidence file path |

Use `--help` to view the flags supported by your build.

## Example: Monitor Podinfo through Tailscale

My homelab runs Podinfo on K3s behind Traefik. When connecting to the Raspberry Pi through its Tailscale IP, the `--host-header` flag preserves the hostname expected by the Ingress rule.

Replace `PI_TAILSCALE_IP` with the Pi's address:

```bash
go run ./cmd/rollout-witness \
  --url http://PI_TAILSCALE_IP/readyz \
  --host-header podinfo.lab.test \
  --requests 100 \
  --interval 1s \
  --timeout 3s \
  --output rollout-evidence.jsonl
```

The CLI sends the HTTP request to the address in `--url`, while the `Host` header is set to `podinfo.lab.test`. Rollout Witness uses the host operating system's networking and does not manage Tailscale connections itself.

## Example summary

```text
Rollout Witness Summary
Total Requests: 3
Successful Requests: 3
HTTP failures: 0
Connection errors: 0
Total failures: 0
Failure Rate: 0 %
Average latency: 68.643108ms
Total successful duration: 205.929326ms
Average successful latency: 68.643108ms
Minimum latency: 65.276444ms
Maximum latency: 71.059921ms
P95 latency: 71.059921ms
```

The p95 uses the nearest-rank method across all request durations, including failed connection attempts. With only three samples, p95 equals the maximum.

## JSONL evidence

Rollout Witness writes one JSON object per attempted request:

```jsonl
{"timestamp":"2026-11-05T22:34:05.66335635+01:00","status_code":200,"duration_ms":81.1,"error":""}
{"timestamp":"2026-11-05T22:34:06.7456216+01:00","status_code":200,"duration_ms":73.551,"error":""}
{"timestamp":"2026-11-05T22:34:07.819415697+01:00","status_code":200,"duration_ms":73.264,"error":""}
```

For network failures, `status_code` is `0`, and `error` contains the Go HTTP client's error message. For HTTP responses such as 404 or 503, the status code is recorded and `error` is empty, because an HTTP response was received.

Inspect an evidence file with:

```bash
cat evidence.jsonl
wc -l evidence.jsonl
```

**Important:** Each run creates or truncates the file specified by `--output`. Choose a new filename if you want to keep previous evidence. The current implementation writes the collected records after the request loop finishes, rather than streaming records to disk during monitoring.

## Exit codes

- `0`: All requests returned HTTP 2xx and the program completed successfully.
- `1`: Invalid input, a file/encoding error, or one or more HTTP or connection failures.
- `2`: Invalid command-line flag syntax or an unknown flag (handled by Go's `flag` package).

