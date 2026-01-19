# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AuditStream is a lightweight Go service for collecting and persisting audit logs. It consumes audit events from NATS JetStream queues and batch writes them to VictoriaLogs for long-term storage and analysis.

## Build and Development Commands

```bash
# Build the application
go build -o auditstream

# Run the application (requires config/values.yml)
./auditstream

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download
```

## Configuration

Configuration is loaded from `config/values.yml`. Copy `config/values.example.yml` to create it:

```yaml
mode: debug|release        # Logging mode
namespace: string          # Application namespace (used for stream/KV naming)
duration: 5s               # Polling interval
batch: 1000                # Maximum messages per batch
nats_hosts:                # NATS server addresses
  - nats://127.0.0.1:4222
nats_token: string         # NATS authentication token
victorialogs: string       # VictoriaLogs endpoint URL
```

## Architecture

```
┌─────────────────────────────────────────────┐
│           NATS JetStream Cluster            │
│  ┌─────────────────────────────────────┐    │
│  │ KV Store (namespace)                │    │
│  │ - Subscription configurations       │    │
│  │ - Watched for hot-reload            │    │
│  └─────────────────────────────────────┘    │
│  ┌─────────────────────────────────────┐    │
│  │ Streams (namespace_key)             │    │
│  │ - Work-queue streams                │    │
│  │ - "default" consumer each           │    │
│  └─────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│         AuditStream Application             │
│  main.go → bootstrap → app.Run()            │
│  - Scheduler polls consumers periodically   │
│  - KV watcher enables hot-reload            │
│  - State queries via namespace.states       │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│            VictoriaLogs                     │
│  - Audit log batch storage                  │
└─────────────────────────────────────────────┘
```

## Key Source Files

- **main.go** - Entry point, NATS connection setup, graceful shutdown
- **bootstrap/bootstrap.go** - Initialization functions (Zap logger, NATS, JetStream, KV, Scheduler)
- **app/app.go** - App struct and initialization
- **app/subscribe.go** - Subscription management, Task() for batch fetching
- **app/state.go** - State tracking via request-reply pattern
- **common/common.go** - Shared structs (Values config) and globals (Logger)
- **transfer/transfer.go** - Transfer struct for managing subscriptions and streams

## Data Flow

1. **Initialization**: main.go → bootstrap functions → app.New() → app.Run()
2. **Dynamic Config**: KV.WatchAll() monitors for Put/Delete/Purge events
3. **Message Processing**: Scheduler triggers Task() → FetchNoWait() from consumer → Acknowledge
4. **State Queries**: Request-reply on `namespace.states` subject

## Key Libraries

- **nats.io/nats.go** - NATS client with JetStream and KV support
- **go.uber.org/zap** - Structured logging
- **github.com/bytedance/sonic** - High-performance JSON via goJson alias
- **github.com/go-co-op/gocron/v2** - Job scheduling for periodic polling
