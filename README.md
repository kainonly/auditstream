# AuditStream

A lightweight service for collecting and persisting audit logs. It consumes audit events from NATS JetStream queues and batch writes them to VictoriaLogs for long-term storage and analysis.

## Features

- Push-based message consumption from NATS JetStream
- Per-subscription buffering with batch writes to VictoriaLogs
- Dynamic subscription management via KV configuration
- Automatic configuration hot-reload
- Graceful shutdown with final buffer flush

## Prerequisites

- NATS JetStream cluster
- VictoriaLogs instance

## Configuration

Create `config/values.yml`:

```yaml
mode: debug
namespace: alpha
nats_hosts:
  - nats://127.0.0.1:4222
nats_token: your-token
victoria: http://localhost:9428
batch_size: 100
flush_interval: 5s
```

| Field | Description |
|-------|-------------|
| `mode` | `debug` or `release` |
| `namespace` | Application namespace for stream/KV naming |
| `nats_hosts` | NATS server addresses |
| `nats_token` | NATS authentication token |
| `victoria` | VictoriaLogs endpoint URL |
| `batch_size` | Flush buffer when reaching this count |
| `flush_interval` | Flush buffer at this interval |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 NATS JetStream                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │  Stream A   │  │  Stream B   │  │  Stream C   │      │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
└─────────┼────────────────┼────────────────┼─────────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                   AuditStream                           │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐     │
│  │Subscription A│ │Subscription B│ │Subscription C│     │
│  │ buffer + flush│ │ buffer + flush│ │ buffer + flush│   │
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘     │
└─────────┼────────────────┼────────────────┼─────────────┘
          │                │                │
          └────────────────┼────────────────┘
                           ▼
                 ┌───────────────────┐
                 │   VictoriaLogs    │
                 └───────────────────┘
```

## License

[BSD-3-Clause License](LICENSE)
