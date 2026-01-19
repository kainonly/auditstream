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

# Tidy dependencies
go mod tidy
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
batch_size: 100          # Flush buffer when reaching this count
flush_interval: 5s       # Flush buffer at this interval
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     NATS JetStream                          │
│  ┌─────────────────┐                                        │
│  │ Stream: {namespace}_{stream}                             │
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
│   │  push() │───►│ buffer []Msg     │───►│ writeBatch()  │  │
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
                  │  /insert/jsonline       │
                  └─────────────────────────┘
```

## Key Source Files

- **main.go** - Entry point, signal handling, graceful shutdown
- **bootstrap/bootstrap.go** - Initialization (Zap logger, NATS, JetStream)
- **app/app.go** - Core logic: Consume, Buffer, Flush, WriteBatch
- **common/common.go** - Configuration struct (Values) and global logger

## Data Flow

1. **Initialization**: main.go → bootstrap → app.New() → app.Run()
2. **Consume**: JetStream pushes messages via Consume() callback
3. **Buffer**: push() adds message to buffer, triggers flush if batch_size reached
4. **Flush Loop**: Ticker triggers flush() every flush_interval
5. **Write**: writeBatch() POSTs JSONL to VictoriaLogs, then ACK/NAK messages

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
    writeBatch() → success: ACK all / failure: NAK all
```

## Key Libraries

- **nats.io/nats.go** - NATS client with JetStream support
- **go.uber.org/zap** - Structured logging

## Scaling

One pod consumes one stream. Deploy multiple pods for multiple streams:

```yaml
# Pod A
stream: audits

# Pod B
stream: events
```
