# AuditStream

A lightweight service for collecting and persisting audit logs. It consumes audit events from a NATS JetStream queue and batch writes them to VictoriaLogs.

## Features

- Push-based message consumption from NATS JetStream
- Batch writes to VictoriaLogs with configurable buffer
- Graceful shutdown with final buffer flush
- Cloud-native design: one stream per pod, scale horizontally

## Prerequisites

- NATS JetStream cluster
- VictoriaLogs instance

## Configuration

Create `config/values.yml`:

```yaml
mode: debug
namespace: alpha
stream: logs
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
| `namespace` | Application namespace for stream naming |
| `stream` | Stream name to consume (full name: `{namespace}_{stream}`) |
| `nats_hosts` | NATS server addresses |
| `nats_token` | NATS authentication token |
| `victoria` | VictoriaLogs endpoint URL |
| `batch_size` | Flush buffer when reaching this count |
| `flush_interval` | Flush buffer at this interval |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     NATS JetStream                          │
│  ┌─────────────────┐                                        │
│  │ Stream: alpha_logs                                       │
│  │ Consumer: default (work queue)                           │
│  └────────┬────────┘                                        │
└───────────┼─────────────────────────────────────────────────┘
            │
            │ Consume()
            ▼
┌─────────────────────────────────────────────────────────────┐
│                     AuditStream Pod                         │
│                                                             │
│   ┌─────────┐    ┌──────────────────┐    ┌───────────────┐  │
│   │ Message │───►│ Buffer           │───►│ writeBatch()  │  │
│   │ Handler │    │ (batch_size:100) │    │ POST /insert  │  │
│   └─────────┘    └──────────────────┘    └───────┬───────┘  │
│                          │                       │          │
│                   flush_interval: 5s             │          │
└──────────────────────────┼───────────────────────┼──────────┘
                           │                       │
                           │         ┌─────────────┘
                           ▼         ▼
                  ┌─────────────────────────┐
                  │      VictoriaLogs       │
                  │  /insert/jsonline       │
                  └─────────────────────────┘
```

## Scaling

Deploy multiple pods to consume from different streams:

```yaml
# Pod A: consumes alpha_logs
stream: logs

# Pod B: consumes alpha_auth
stream: auth

# Pod C: consumes alpha_payments
stream: payments
```

## License

[BSD-3-Clause License](LICENSE)
