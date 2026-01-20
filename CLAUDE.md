# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AuditStream is a lightweight Go service for collecting and persisting audit logs. It consumes audit events from a NATS JetStream queue and batch writes them to VictoriaLogs for long-term storage and analysis.

## Build and Development Commands

```bash
# Build the application
go build -o auditstream

# Run the application (requires config/values.yml)
./auditstream

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Tidy dependencies
go mod tidy

# Build Docker image (build binary first with CGO_ENABLED=0)
CGO_ENABLED=0 GOOS=linux go build -o auditstream
docker build -t auditstream .
```

## Configuration

Configuration is loaded from `config/values.yml`. Copy `config/values.example.yml` to create it:

```yaml
mode: debug|release      # Logging mode
namespace: string        # Application namespace (used for stream naming)
stream: string           # Stream name to consume (full: {namespace}_{stream})
nats_hosts:              # NATS server addresses
  - nats://127.0.0.1:4222
nats_token: string       # NATS authentication token
victoria: string         # VictoriaLogs endpoint URL
victoria_path: string    # VictoriaLogs API path with query params
batch_size: 100          # Flush buffer when reaching this count
flush_interval: 5s       # Flush buffer at this interval
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     NATS JetStream                          │
│  ┌─────────────────┐                                        │
│  │ Stream: {namespace}_{stream}                             │
│  │ Subject: {namespace}.{stream}                            │
│  │ Consumer: default (work queue mode)                      │
│  └────────┬────────┘                                        │
└───────────┼─────────────────────────────────────────────────┘
            │
            │ Consume() - push-based
            ▼
┌─────────────────────────────────────────────────────────────┐
│                     AuditStream Pod                         │
│                                                             │
│   ┌─────────┐    ┌──────────────────┐    ┌───────────────┐  │
│   │  push() │───►│ buffer []Msg     │───►│ write()       │  │
│   │         │    │ (mutex protected)│    │ POST jsonline │  │
│   └─────────┘    └──────────────────┘    └───────────────┘  │
│                          │                                  │
│              Flush triggers:                                │
│              - len(buffer) >= batch_size                    │
│              - ticker every flush_interval                  │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
                  ┌─────────────────────────┐
                  │      VictoriaLogs       │
                  │  {victoria}{victoria_path}
                  └─────────────────────────┘
```

## Key Source Files

- **main.go** - Entry point, signal handling, graceful shutdown
- **bootstrap/bootstrap.go** - Initialization (Zap logger, NATS, JetStream)
- **app/app.go** - Core logic: Stream init, Consume, Buffer, Flush, Write
- **common/common.go** - Configuration struct (Values) and global logger
- **transfer/transfer.go** - Client SDK for publishing audit events

## Data Flow

1. **Initialization**: main.go → bootstrap → app.New() → app.Run()
2. **Stream Setup**: CreateOrUpdateStream + CreateOrUpdateConsumer (auto-create on startup)
3. **Consume**: JetStream pushes messages via Consume() callback
4. **Buffer**: push() adds message to buffer, triggers flush if batch_size reached
5. **Flush Loop**: Ticker triggers flush() every flush_interval
6. **Write**: write() POSTs JSONL to VictoriaLogs, then ACK/NAK messages

## Stream Initialization

On startup, AuditStream auto-creates stream and consumer if not exist:

```go
// Stream config
StreamConfig{
    Name:        "{namespace}_{stream}",
    Subjects:    ["{namespace}.{stream}"],
    Retention:   WorkQueuePolicy,  // 消息被消费后删除
    Storage:     FileStorage,
    Compression: S2Compression,
}

// Consumer config
ConsumerConfig{
    Name:      "default",
    AckPolicy: AckExplicitPolicy,  // 显式 ACK
}
```

## Flush Logic

```
push(msg):
    lock → append to buffer → unlock
    if len(buffer) >= batch_size:
        flush()

flushLoop():
    every flush_interval:
        flush()
    on stop signal:
        flush() // final flush before shutdown

flush():
    lock → swap buffer → unlock
    if empty: return
    write() → success: ACK all / failure: NAK all
```

## Key Libraries

- **nats.io/nats.go** - NATS client with JetStream support
- **go.uber.org/zap** - Structured logging
- **github.com/bytedance/sonic** - High-performance JSON
- **gopkg.in/yaml.v3** - YAML configuration parsing

## Transfer SDK

Client SDK for publishing audit events to NATS JetStream.

### Usage

```go
import "github.com/kainonly/auditstream/v3/transfer"

// Create transfer client
t, err := transfer.New(nc, "namespace")

// Publish with AuditEvent (recommended)
event := transfer.NewAuditEvent("audits", "用户登录").
    WithMeta("admin", "user", "login", 123, 456).        // platform, resource, action, objectId, userId
    WithRequest("/api/login", map[string]any{"username": "test"}).  // path, body
    WithResponse(200, map[string]any{"success": true}).  // code, response
    WithClient("192.168.1.1", "Mozilla/5.0")             // ip, agent
t.Publish(ctx, "audits", event)

// Publish custom data
t.Publish(ctx, "audits", map[string]any{
    "time":   time.Now(),
    "stream": "custom",
    "msg":    "Custom message",
})

// Async publish (fire and forget)
future, _ := t.PublishAsync("audits", event)

// Publish raw bytes (pre-serialized JSON)
data, _ := sonic.Marshal(event)
t.PublishRaw(ctx, "audits", data)
t.PublishRawAsync("audits", data)
```

### AuditEvent Fields

| Field | JSON | Description |
|-------|------|-------------|
| Time | `time` | Event timestamp (_time_field) |
| Stream | `stream` | Log stream identifier (_stream_fields) |
| Msg | `msg` | Message content (_msg_field) |
| Platform | `platform` | Platform identifier (e.g., admin, api) |
| Resource | `resource` | Resource type (e.g., user, order) |
| Action | `action` | Operation type (e.g., create, update, delete) |
| ObjectId | `object_id` | Object identifier (int, string, or []int for batch) |
| UserId | `user_id` | User ID (int) |
| Path | `path` | Request path |
| IP | `ip` | Client IP address |
| Extra | `extra` | Additional data (omitted if nil) |

### Builder Methods

| Method | Parameters | Description |
|--------|------------|-------------|
| `NewAuditEvent` | stream, msg | Create new event with timestamp |
| `WithMeta` | platform, resource, action, objectId, userId | Set core metadata |
| `WithRequest` | path, body | Set request info (body stored in Extra) |
| `WithResponse` | code, response | Set response info (stored in Extra) |
| `WithClient` | ip, agent | Set client info (agent stored in Extra if not empty) |
| `WithExtra` | key, value | Add custom field to Extra |

## Scaling

One pod consumes one stream. Deploy multiple pods for multiple streams:

```yaml
# Pod A
stream: audits

# Pod B
stream: events
```

## Docker

Dockerfile uses Alpine Linux base image:

```dockerfile
FROM alpine:edge
RUN apk add tzdata
WORKDIR /app
ADD auditstream /app/
CMD [ "./auditstream" ]
```

Build and run:

```bash
# Build static binary
CGO_ENABLED=0 GOOS=linux go build -o auditstream

# Build image
docker build -t auditstream .

# Run with config volume
docker run -v ./config:/app/config auditstream
```

## Testing

Integration tests require NATS connection. Set environment variables or create `.env` file:

```bash
NATS_NAMESPACE=test
NATS_HOST=nats://127.0.0.1:4222
NATS_TOKEN=s3cr3t
```

Run tests:

```bash
# Unit tests only
go test -v ./transfer -run "^Test.*Event"

# Integration tests (requires NATS)
go test -v ./transfer

# Bulk/stress tests
go test -v ./transfer -run "Bulk|Continuous"
```
