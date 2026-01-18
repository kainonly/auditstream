# AuditStream

A lightweight service for collecting and persisting audit logs. It consumes audit events from NATS JetStream queues and batch writes them to VictoriaLogs for long-term storage and analysis.

## Features

- Consumes audit logs from NATS JetStream queues
- Batch writes to VictoriaLogs for efficient storage
- Dynamic subscription management via KV configuration
- Periodic polling with configurable intervals
- Automatic configuration hot-reload

## Prerequisites

- NATS JetStream cluster
- VictoriaLogs instance

## Configuration

Create `config/values.yml`:

```yaml
mode: debug
namespace: alpha
duration: 5s
batch: 1000
nats_hosts:
  - nats://127.0.0.1:4222
nats_token: your-token
victorialogs: http://localhost:9428
```

| Field | Description |
|-------|-------------|
| `mode` | `debug` or `release` |
| `namespace` | Application namespace for stream/KV naming |
| `duration` | Polling interval for batch consumption |
| `batch` | Maximum messages per batch |
| `nats_hosts` | NATS server addresses |
| `nats_token` | NATS authentication token |
| `victorialogs` | VictoriaLogs endpoint URL |

## License

[BSD-3-Clause License](LICENSE)
