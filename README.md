# AuditStream

A lightweight service for collecting and persisting audit logs. It consumes audit events from a NATS JetStream queue and batch writes them to VictoriaLogs.

## Overview

![Architecture](docs/plan.png)

## Features

- Push-based message consumption from NATS JetStream
- Batch writes to VictoriaLogs with configurable buffer
- Auto-create stream and consumer on startup
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
victoria_path: /insert/jsonline?_stream_fields=stream&_msg_field=msg&_time_field=time
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
| `victoria_path` | VictoriaLogs API path with query params |
| `batch_size` | Flush buffer when reaching this count |
| `flush_interval` | Flush buffer at this interval |

## Data Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ App Service │     │ App Service │     │ App Service │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │ Write Event
                           ▼
                  ┌─────────────────┐
                  │  Transfer SDK   │
                  │  Push JSON      │
                  └────────┬────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     NATS JetStream                          │
│  Stream: {namespace}_{stream}                               │
│  Subject: {namespace}.{stream}                              │
│  Consumer: default (WorkQueue)                              │
└───────────────────────────┬─────────────────────────────────┘
                            │ Consume()
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   AuditStream Pod                           │
│                                                             │
│   Message ──► Buffer ──► Flush ──► POST /insert/jsonline    │
│                 │                                           │
│         (batch_size OR flush_interval)                      │
└───────────────────────────┬─────────────────────────────────┘
                            │ Success: ACK / Fail: NAK
                            ▼
                  ┌─────────────────────────┐
                  │      VictoriaLogs       │
                  └─────────────────────────┘
```

## Transfer SDK

Client SDK for publishing audit events.

```go
import "github.com/kainonly/auditstream/v3/transfer"

// Create client
t, err := transfer.New(nc, "namespace")

// Publish audit event
event := transfer.NewAuditEvent("user-actions", "User logged in").
    WithAction("login").
    WithUser("user123", "192.168.1.1")
t.Publish(ctx, "audits", event)

// Async publish
t.PublishAsync("audits", event)
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
