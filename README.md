# AuditStream

[![Version](https://img.shields.io/badge/version-3.0.0-blue.svg)](https://github.com/kainonly/auditstream/releases/tag/v3.0.0)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Node](https://img.shields.io/badge/node-%3E%3D18.0.0-brightgreen.svg)](https://nodejs.org)
[![Build Status](https://img.shields.io/badge/build-passing-success.svg)](https://github.com/kainonly/auditstream/actions)

企业级审计流管理系统 - 实时、高性能、可扩展的审计日志解决方案。

## 特性

- 🚀 **高性能**: 支持每秒处理 50K+ 审计事件
- 🔄 **流式处理**: 基于事件驱动的流式数据处理引擎
- 🧩 **插件系统**: 灵活的插件架构，支持自定义审计规则
- 📊 **实时监控**: 内置可视化仪表板，实时监控审计流
- 🔐 **安全合规**: 支持 SOC2、ISO27001 等合规性标准
- 🌐 **多数据源**: 统一管理数据库、文件系统、云存储等审计日志
- 📈 **可扩展**: 微服务架构，支持水平扩展
- 🛡️ **数据安全**: 内置数据脱敏、加密传输、防篡改机制

## 快速开始

### 环境要求

- Node.js >= 18.0.0
- PostgreSQL >= 12 或 MySQL >= 8.0
- Redis >= 6.0

### 安装

```bash
# 使用 npm
npm install auditstream@3.0.0

# 使用 yarn
yarn add auditstream@3.0.0

# 使用 Docker
docker pull auditstream/auditstream:3.0.0
```

### 基础配置

创建配置文件 `config/app.yml`:

```yaml
server:
  http:
    port: 3000
    host: 0.0.0.0

auditStream:
  engine:
    enabled: true
    logLevel: info
    workers: 4

  storage:
    type: postgresql
    connection:
      host: localhost
      port: 5432
      database: auditstream
      username: auditstream_user
      password: your_password

  cache:
    type: redis
    connection:
      host: localhost
      port: 6379
```

### 启动服务

```bash
# 开发模式
npm run dev

# 生产模式
npm run start

# 使用 Docker
docker run -p 3000:3000 \
  -v $(pwd)/config:/app/config \
  auditstream/auditstream:3.0.0
```

### 验证安装

```bash
curl http://localhost:3000/api/v3/health
```

## 使用示例

### 创建审计事件

```javascript
const { AuditClient } = require('auditstream');

const client = new AuditClient({
  endpoint: 'http://localhost:3000',
  apiKey: 'your-api-key'
});

// 记录审计事件
await client.audit({
  eventType: 'user_action',
  actor: {
    id: 'user123',
    type: 'user',
    name: 'John Doe'
  },
  action: 'login',
  resource: {
    type: 'system',
    id: 'web-app'
  },
  metadata: {
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0...'
  }
});
```

### 查询审计日志

```javascript
// 查询审计日志
const events = await client.query({
  eventType: 'user_action',
  startTime: '2026-01-01T00:00:00Z',
  endTime: '2026-01-19T23:59:59Z',
  limit: 100
});

console.log(events);
```

### 生成审计报告

```javascript
// 生成合规性报告
const report = await client.generateReport({
  type: 'compliance',
  standard: 'SOC2',
  period: {
    start: '2026-01-01',
    end: '2026-01-31'
  },
  format: 'pdf'
});

console.log('报告已生成:', report.downloadUrl);
```

## 文档

- 📚 [完整文档](https://docs.auditstream.io/v3/)
- 🚀 [快速开始指南](./docs/GETTING-STARTED.md)
- 📖 [API 参考](https://api.auditstream.io/v3/)
- 🔄 [迁移指南](./docs/MIGRATION-v3.md)
- 🔌 [插件开发](./docs/PLUGIN-DEVELOPMENT.md)
- ❓ [常见问题](./docs/FAQ-v3.md)

## 版本说明

当前版本: **v3.0.0** (2026-01-19)

### 主要更新

- ✨ 全新微服务架构
- ⚡ 性能提升 300%
- 🔐 增强的安全特性
- 📊 实时监控仪表板
- 🧩 插件系统

详细更新内容请查看：
- [发布说明](./docs/RELEASE-v3.0.0.md)
- [更新日志](./CHANGELOG.md)

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                    AuditStream v3.0                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │   API    │  │ GraphQL  │  │ Webhook  │   API Layer │
│  │  Server  │  │ Endpoint │  │ Handler  │             │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘             │
│       │             │             │                     │
│  ┌────┴─────────────┴─────────────┴─────┐             │
│  │        Event Processing Engine        │   Core      │
│  │    ┌───────────┐  ┌───────────┐      │             │
│  │    │  Stream   │  │  Plugin   │      │             │
│  │    │ Processor │  │  Manager  │      │             │
│  │    └───────────┘  └───────────┘      │             │
│  └────┬─────────────┬─────────────┬─────┘             │
│       │             │             │                     │
│  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐             │
│  │PostgreSQL│  │  Redis   │  │  Object  │   Storage   │
│  │  /MySQL  │  │  Cache   │  │ Storage  │             │
│  └──────────┘  └──────────┘  └──────────┘             │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 性能指标

| 指标 | v2.x | v3.0 |
|------|------|------|
| 事件处理速度 | 10K/s | 50K/s |
| 平均响应时间 | 150ms | 45ms |
| 并发请求 | 1K req/s | 4K req/s |
| 内存占用 | 512MB | 256MB |

## 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

详细信息请查看 [贡献指南](./CONTRIBUTING.md)。

## 社区

- 💬 [GitHub Discussions](https://github.com/kainonly/auditstream/discussions)
- 🐛 [问题反馈](https://github.com/kainonly/auditstream/issues)
- 📧 邮件支持: support@auditstream.io
- 🌐 官网: https://auditstream.io

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](./LICENSE) 文件。

## 致谢

感谢所有为 AuditStream 做出贡献的开发者和社区成员！

---

**Made with ❤️ by the AuditStream Team**
