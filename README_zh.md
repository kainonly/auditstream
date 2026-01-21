# AuditStream

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/kainonly/auditstream/release.yml?label=release&style=flat-square)](https://github.com/kainonly/auditstream/actions/workflows/release.yml)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/kainonly/auditstream/testing.yml?label=testing&style=flat-square)](https://github.com/kainonly/auditstream/actions/workflows/testing.yml)
[![Release](https://img.shields.io/github/v/release/kainonly/auditstream.svg?style=flat-square&include_prereleases)](https://github.com/kainonly/auditstream/releases)
[![Coveralls github](https://img.shields.io/coveralls/github/kainonly/auditstream.svg?style=flat-square)](https://coveralls.io/github/kainonly/auditstream)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kainonly/auditstream?style=flat-square)](https://github.com/kainonly/auditstream)
[![Go Report Card](https://goreportcard.com/badge/github.com/kainonly/auditstream?style=flat-square)](https://goreportcard.com/report/github.com/kainonly/auditstream)
[![GitHub license](https://img.shields.io/github/license/kainonly/auditstream?style=flat-square)](https://raw.githubusercontent.com/kainonly/auditstream/v3/LICENSE)

Go 审计日志收集服务。从 NATS JetStream 队列消费审计事件，批量写入 VictoriaLogs。

## 概览

![架构图](docs/plan.png)

## 特性

- 基于 Push 模式从 NATS JetStream 消费消息
- 可配置缓冲区大小和写入间隔的批量写入
- 启动时自动创建 Stream 和 Consumer
- 优雅关闭，确保最后一批数据写入
- 一个 Pod 消费一个 Stream，支持水平扩展

## 依赖

- Go 1.24+
- NATS JetStream
- VictoriaLogs

## 配置

创建 `config/values.yml`：

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

| 字段 | 说明 |
|------|------|
| `mode` | 日志模式：`debug` 或 `release` |
| `namespace` | 应用命名空间，用于 Stream 命名 |
| `stream` | Stream 名称（完整名称：`{namespace}_{stream}`） |
| `nats_hosts` | NATS 服务器地址列表 |
| `nats_token` | NATS 认证令牌 |
| `victoria` | VictoriaLogs 端点 URL |
| `victoria_path` | VictoriaLogs API 路径及查询参数 |
| `batch_size` | 缓冲区达到此数量时触发写入 |
| `flush_interval` | 定时写入间隔 |

## 安装

```bash
go get github.com/kainonly/auditstream/v3
```

## 数据流

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  应用服务 A  │     │  应用服务 B  │     │  应用服务 C  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │ 写入事件
                           ▼
                  ┌─────────────────┐
                  │  Transfer SDK   │
                  │  推送 JSON      │
                  └────────┬────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     NATS JetStream                          │
│  Stream: {namespace}_{stream}                               │
│  Subject: {namespace}.{stream}                              │
│  Consumer: default (工作队列模式)                            │
└───────────────────────────┬─────────────────────────────────┘
                            │ Consume()
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   AuditStream Pod                           │
│                                                             │
│   消息 ──► 缓冲区 ──► 写入 ──► POST /insert/jsonline         │
│              │                                              │
│      (达到 batch_size 或 flush_interval)                    │
└───────────────────────────┬─────────────────────────────────┘
                            │ 成功: ACK / 失败: NAK
                            ▼
                  ┌─────────────────────────┐
                  │      VictoriaLogs       │
                  └─────────────────────────┘
```

## 写入逻辑

两个条件任一满足即触发写入：

1. **数量触发**：缓冲区达到 `batch_size`（如 100 条）
2. **定时触发**：每隔 `flush_interval`（如 5 秒）

```
push(msg):
    加锁 → 追加到缓冲区 → 解锁
    if len(buffer) >= batch_size:
        flush()

flushLoop():
    每隔 flush_interval:
        flush()
    收到停止信号:
        flush()  // 关闭前最后写入一次

flush():
    加锁 → 交换缓冲区 → 解锁
    if 空: return
    write() → 成功: ACK 全部 / 失败: NAK 全部
```

## Transfer SDK

用于发送审计事件的客户端 SDK。

### 基本用法

```go
import "github.com/kainonly/auditstream/v3/transfer"

// 创建客户端
t, err := transfer.New(nc, "namespace")

// 发送审计事件
event := transfer.NewAuditEvent("audits", "用户登录").
    WithMeta("admin", "user", "login", 123, 456).
    WithRequest("/api/login", map[string]any{"username": "test"}).
    WithResponse(200, map[string]any{"success": true}).
    WithClient("192.168.1.1", "Mozilla/5.0")

err = t.Publish(ctx, "audits", event)

// 异步发送
future, err := t.PublishAsync("audits", event)

// 发送原始字节（预序列化的 JSON）
data, _ := sonic.Marshal(event)
t.PublishRaw(ctx, "audits", data)
```

### AuditEvent 字段

| 字段 | JSON | 说明 |
|------|------|------|
| Time | `time` | 事件时间 |
| Stream | `stream` | 日志流标识 |
| Msg | `msg` | 消息内容 |
| Platform | `platform` | 平台标识（如 admin、api） |
| Resource | `resource` | 资源类型（如 user、order） |
| Action | `action` | 操作类型（如 create、update、delete） |
| ObjectId | `object_id` | 对象 ID（int、string 或 []int） |
| UserId | `user_id` | 用户 ID |
| Path | `path` | 请求路径 |
| IP | `ip` | 客户端 IP |
| Extra | `extra` | 扩展数据（为 nil 时省略） |

### 构建方法

| 方法 | 参数 | 说明 |
|------|------|------|
| `NewAuditEvent` | stream, msg | 创建事件，自动设置当前时间 |
| `WithMeta` | platform, resource, action, objectId, userId | 设置核心元数据 |
| `WithRequest` | path, body | 设置请求路径和请求体（body 存入 Extra） |
| `WithResponse` | code, response | 设置响应码和响应数据（存入 Extra） |
| `WithClient` | ip, agent | 设置客户端 IP 和 User Agent（agent 存入 Extra） |
| `WithExtra` | key, value | 添加自定义扩展字段 |

## 水平扩展

部署多个 Pod 消费不同的 Stream：

```yaml
# Pod A: 消费 alpha_logs
stream: logs

# Pod B: 消费 alpha_auth
stream: auth

# Pod C: 消费 alpha_payments
stream: payments
```

## Docker

```bash
# 编译静态二进制
CGO_ENABLED=0 GOOS=linux go build -o auditstream

# 构建镜像
docker build -t auditstream .

# 挂载配置运行
docker run -v ./config:/app/config auditstream
```

## 许可证

[BSD-3-Clause License](LICENSE)
