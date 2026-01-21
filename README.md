# AuditStream

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/kainonly/auditstream/release.yml?label=release&style=flat-square)](https://github.com/kainonly/auditstream/actions/workflows/release.yml)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/kainonly/auditstream/testing.yml?label=testing&style=flat-square)](https://github.com/kainonly/auditstream/actions/workflows/testing.yml)
[![Release](https://img.shields.io/github/v/release/kainonly/auditstream.svg?style=flat-square&include_prereleases)](https://github.com/kainonly/auditstream/releases)
[![Coveralls github](https://img.shields.io/coveralls/github/kainonly/auditstream.svg?style=flat-square)](https://coveralls.io/github/kainonly/auditstream)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kainonly/auditstream?style=flat-square)](https://github.com/kainonly/auditstream)
[![Go Report Card](https://goreportcard.com/badge/github.com/kainonly/auditstream?style=flat-square)](https://goreportcard.com/report/github.com/kainonly/auditstream)
[![GitHub license](https://img.shields.io/github/license/kainonly/auditstream?style=flat-square)](https://raw.githubusercontent.com/kainonly/auditstream/v3/LICENSE)

A Go service for collecting and persisting audit logs. It consumes audit events from a NATS JetStream queue and batch writes them to VictoriaLogs.

## Overview

![Architecture](docs/plan.png)

## Features

- Push-based message consumption from NATS JetStream
- Batch writes to VictoriaLogs with configurable buffer size and flush interval
- Auto-create stream and consumer on startup
- Graceful shutdown with final buffer flush
- One stream per pod, horizontal scaling

## Requirements

- Go 1.24+
- NATS JetStream
- VictoriaLogs

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

## Installation

```bash
go get github.com/kainonly/auditstream/v3
```

## Transfer SDK

Client SDK for publishing audit events to NATS JetStream.

### Basic Usage

```go
import "github.com/kainonly/auditstream/v3/transfer"

// Create client
t, err := transfer.New(nc, "namespace")

// Publish audit event
event := transfer.NewAuditEvent("audits", "User logged in").
    WithMeta("admin", "user", "login", 123, 456).
    WithRequest("/api/login", map[string]any{"username": "test"}).
    WithResponse(200, map[string]any{"success": true}).
    WithClient("192.168.1.1", "Mozilla/5.0")

err = t.Publish(ctx, "audits", event)

// Async publish
future, err := t.PublishAsync("audits", event)

// Publish raw bytes (pre-serialized JSON)
data, _ := sonic.Marshal(event)
t.PublishRaw(ctx, "audits", data)
```

### AuditEvent Fields

| Field | JSON | Description |
|-------|------|-------------|
| Time | `time` | Event timestamp |
| Stream | `stream` | Log stream identifier |
| Msg | `msg` | Message content |
| Platform | `platform` | Platform identifier (e.g., admin, api) |
| Resource | `resource` | Resource type (e.g., user, order) |
| Action | `action` | Operation type (e.g., create, update, delete) |
| ObjectId | `object_id` | Object identifier (int, string, or []int) |
| UserId | `user_id` | User ID |
| Path | `path` | Request path |
| IP | `ip` | Client IP address |
| Extra | `extra` | Additional data (omitted if nil) |

### Builder Methods

| Method | Parameters | Description |
|--------|------------|-------------|
| `NewAuditEvent` | stream, msg | Create event with current timestamp |
| `WithMeta` | platform, resource, action, objectId, userId | Set core metadata |
| `WithRequest` | path, body | Set request path and body (body stored in Extra) |
| `WithResponse` | code, response | Set response code and data (stored in Extra) |
| `WithClient` | ip, agent | Set client IP and user agent (agent stored in Extra) |
| `WithExtra` | key, value | Add custom field to Extra |

## Flush Logic

Two conditions trigger a flush:

1. **Batch size**: Buffer reaches `batch_size` (e.g., 100 messages)
2. **Interval**: Every `flush_interval` (e.g., 5 seconds)

```
push(msg):
    lock → append to buffer → unlock
    if len(buffer) >= batch_size:
        flush()

flushLoop():
    every flush_interval:
        flush()
    on stop signal:
        flush()  // final flush before shutdown

flush():
    lock → swap buffer → unlock
    if empty: return
    write() → success: ACK all / failure: NAK all
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

## Docker

```bash
# Build static binary
CGO_ENABLED=0 GOOS=linux go build -o auditstream

# Build image
docker build -t auditstream .

# Run with config volume
docker run -v ./config:/app/config auditstream
```

## License

[BSD-3-Clause License](LICENSE)