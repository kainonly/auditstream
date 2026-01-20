# AuditStream v3.0 迁移指南

本指南将帮助您从 AuditStream v2.x 平滑迁移到 v3.0.0。

## 目录

- [迁移前准备](#迁移前准备)
- [环境要求](#环境要求)
- [数据备份](#数据备份)
- [迁移步骤](#迁移步骤)
- [配置迁移](#配置迁移)
- [API 迁移](#api迁移)
- [常见问题](#常见问题)

## 迁移前准备

### 评估影响范围

在开始迁移之前，请评估以下内容：

1. **当前版本**: 确认您当前运行的 AuditStream 版本
2. **依赖服务**: 检查所有依赖 AuditStream 的服务和应用
3. **自定义配置**: 整理所有自定义配置和规则
4. **数据量**: 评估当前审计数据的规模

### 迁移时间窗口

建议在以下时间段进行迁移：
- 业务低峰期
- 有足够的回滚时间
- 技术团队可以全程支持

## 环境要求

### 软件版本要求

| 组件 | v2.x 要求 | v3.0 要求 | 说明 |
|------|-----------|-----------|------|
| Node.js | 16.x+ | 18.x+ | 必须升级 |
| PostgreSQL | 10+ | 12+ | 建议升级 |
| MySQL | 5.7+ | 8.0+ | 建议升级 |
| Redis | 5.0+ | 6.0+ | 必须升级 |
| Elasticsearch (可选) | 7.x | 8.x | 建议升级 |

### 硬件要求

| 资源 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 4 核 | 8 核 |
| 内存 | 8 GB | 16 GB |
| 磁盘 | 100 GB | 500 GB SSD |
| 网络 | 1 Gbps | 10 Gbps |

## 数据备份

### 1. 备份数据库

#### PostgreSQL
```bash
# 备份整个数据库
pg_dump -U auditstream_user -h localhost auditstream_db > backup_v2_$(date +%Y%m%d).sql

# 验证备份
pg_restore --list backup_v2_$(date +%Y%m%d).sql
```

#### MySQL
```bash
# 备份整个数据库
mysqldump -u auditstream_user -p auditstream_db > backup_v2_$(date +%Y%m%d).sql

# 验证备份
mysql -u auditstream_user -p -e "source backup_v2_$(date +%Y%m%d).sql" test_restore_db
```

### 2. 备份配置文件

```bash
# 创建配置备份目录
mkdir -p ./backups/config_v2_$(date +%Y%m%d)

# 备份所有配置文件
cp -r ./config/* ./backups/config_v2_$(date +%Y%m%d)/
cp .env ./backups/config_v2_$(date +%Y%m%d)/
```

### 3. 备份审计日志

```bash
# 备份审计日志文件
tar -czf audit_logs_backup_$(date +%Y%m%d).tar.gz ./logs/
```

## 迁移步骤

### 第一步：环境升级

#### 1. 升级 Node.js

```bash
# 使用 nvm 升级
nvm install 18
nvm use 18
nvm alias default 18

# 验证版本
node --version  # 应该显示 v18.x.x
```

#### 2. 升级数据库

参考对应数据库的官方升级文档：
- [PostgreSQL 升级指南](https://www.postgresql.org/docs/current/upgrading.html)
- [MySQL 升级指南](https://dev.mysql.com/doc/refman/8.0/en/upgrading.html)

#### 3. 升级 Redis

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install redis-server=6:6.*

# 验证版本
redis-server --version
```

### 第二步：安装 AuditStream v3.0

#### 停止旧版本服务

```bash
# 停止 v2.x 服务
npm run stop

# 或使用 pm2
pm2 stop auditstream
```

#### 安装新版本

```bash
# 方式一：NPM 安装
npm install auditstream@3.0.0

# 方式二：从源码安装
git clone https://github.com/kainonly/auditstream.git
cd auditstream
git checkout v3.0.0
npm install
npm run build
```

### 第三步：配置迁移

#### 自动迁移（推荐）

```bash
# 运行配置迁移工具
npx auditstream migrate:config --from=v2 --to=v3 --config=./config/app.yml

# 验证迁移结果
npx auditstream validate:config --version=3
```

#### 手动迁移

如果自动迁移失败，请参考以下配置映射：

**v2.x 配置示例**:
```yaml
# config/app.yml (v2.x)
server:
  port: 3000
  host: localhost

audit:
  enabled: true
  level: info
  storage: postgresql

database:
  host: localhost
  port: 5432
  database: auditstream
  username: user
  password: pass

redis:
  host: localhost
  port: 6379
```

**v3.0 配置示例**:
```yaml
# config/app.yml (v3.0)
server:
  http:
    port: 3000
    host: localhost

auditStream:
  engine:
    enabled: true
    logLevel: info
    processors:
      - type: default
        storage: postgresql

  storage:
    type: postgresql
    connection:
      host: localhost
      port: 5432
      database: auditstream
      username: user
      password: pass

  cache:
    type: redis
    connection:
      host: localhost
      port: 6379
      db: 0
```

### 第四步：数据库迁移

#### 自动迁移

```bash
# 运行数据库迁移脚本
npx auditstream migrate:db --version=3.0.0

# 验证数据完整性
npx auditstream verify:db
```

#### 迁移脚本详情

迁移脚本会执行以下操作：

1. **表结构更新**
   - 添加新的系统表
   - 更新现有表结构
   - 创建新的索引

2. **数据转换**
   - 审计规则格式转换
   - 用户权限数据迁移
   - 历史审计日志格式更新

3. **清理工作**
   - 删除废弃的表和字段
   - 清理冗余索引

#### 回滚点

迁移过程会自动创建回滚点：

```bash
# 如需回滚
npx auditstream rollback:db --to=v2.x
```

### 第五步：启动服务

#### 启动 v3.0 服务

```bash
# 开发环境
npm run dev

# 生产环境
npm run start

# 使用 PM2
pm2 start ecosystem.config.js
```

#### 验证服务状态

```bash
# 检查服务健康状态
curl http://localhost:3000/api/v3/health

# 预期响应
{
  "status": "healthy",
  "version": "3.0.0",
  "uptime": 123,
  "timestamp": "2026-01-19T00:00:00.000Z"
}
```

### 第六步：功能验证

#### 1. 基础功能测试

```bash
# 运行测试套件
npm run test

# 运行集成测试
npm run test:integration
```

#### 2. 审计功能验证

- 创建测试审计规则
- 生成测试审计日志
- 验证审计数据采集
- 检查审计报告生成

#### 3. 性能测试

```bash
# 运行性能测试
npm run test:performance

# 或使用压测工具
ab -n 10000 -c 100 http://localhost:3000/api/v3/audit/events
```

## 配置迁移

### 环境变量变更

| v2.x | v3.0 | 说明 |
|------|------|------|
| `AUDIT_LEVEL` | `AUDIT_LOG_LEVEL` | 重命名 |
| `DB_HOST` | `DATABASE_HOST` | 重命名 |
| `CACHE_ENABLED` | `REDIS_ENABLED` | 重命名 |
| - | `STREAM_WORKERS` | 新增：流处理工作线程数 |
| - | `PLUGIN_PATH` | 新增：插件目录路径 |

### 配置文件迁移对照表

详细的配置映射关系：

```javascript
// v2.x -> v3.0 配置映射
{
  "audit.enabled": "auditStream.engine.enabled",
  "audit.level": "auditStream.engine.logLevel",
  "audit.storage": "auditStream.storage.type",
  "database.*": "auditStream.storage.connection.*",
  "redis.*": "auditStream.cache.connection.*"
}
```

## API 迁移

### 端点变更

| v2.x 端点 | v3.0 端点 | 变更说明 |
|-----------|-----------|----------|
| `POST /api/v1/audit` | `POST /api/v3/audit/events` | 路径和格式调整 |
| `GET /api/v1/audit/:id` | `GET /api/v3/audit/events/:id` | 路径调整 |
| `GET /api/v1/rules` | `GET /api/v3/audit/rules` | 路径调整 |
| `POST /api/v1/reports` | `POST /api/v3/reports/generate` | 路径和功能增强 |

### 请求/响应格式变更

#### 创建审计事件

**v2.x 请求**:
```json
{
  "type": "user_action",
  "user": "user123",
  "action": "login",
  "timestamp": 1234567890
}
```

**v3.0 请求**:
```json
{
  "eventType": "user_action",
  "actor": {
    "id": "user123",
    "type": "user"
  },
  "action": "login",
  "timestamp": "2026-01-19T00:00:00.000Z",
  "metadata": {}
}
```

### 认证方式变更

#### v2.x (API Key)

```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:3000/api/v1/audit
```

#### v3.0 (OAuth2.0)

```bash
# 获取访问令牌
curl -X POST http://localhost:3000/oauth/token \
  -d "grant_type=client_credentials" \
  -d "client_id=your-client-id" \
  -d "client_secret=your-client-secret"

# 使用访问令牌
curl -H "Authorization: Bearer your-access-token" \
  http://localhost:3000/api/v3/audit/events
```

## 常见问题

### Q1: 迁移过程中服务必须停机吗？

A: 对于生产环境，建议采用蓝绿部署或滚动升级策略，最小化停机时间。具体方案请参考 [高可用迁移指南](./HA-MIGRATION.md)。

### Q2: 迁移失败如何回滚？

A:
```bash
# 停止 v3.0 服务
pm2 stop auditstream

# 恢复数据库
psql -U user auditstream < backup_v2_20260119.sql

# 恢复配置
cp -r ./backups/config_v2_20260119/* ./config/

# 启动 v2.x 服务
pm2 start auditstream-v2
```

### Q3: 数据迁移需要多长时间？

A: 取决于数据量：
- 小型部署 (< 1GB): 5-10 分钟
- 中型部署 (1-10GB): 30-60 分钟
- 大型部署 (> 10GB): 1-4 小时

建议先在测试环境进行评估。

### Q4: v2.x API 还能继续使用吗？

A: v3.0 提供了 v2 API 的兼容层，但性能较差且功能受限。建议在 3 个月内完成 API 升级，v2 API 将在 v4.0 中完全移除。

### Q5: 插件系统如何迁移？

A: v2.x 的自定义规则需要重写为 v3.0 插件格式。请参考 [插件开发指南](./PLUGIN-DEVELOPMENT.md)。

## 获取帮助

如果在迁移过程中遇到问题：

1. 查看 [FAQ](./FAQ-v3.md)
2. 搜索 [GitHub Issues](https://github.com/kainonly/auditstream/issues)
3. 加入社区讨论 [GitHub Discussions](https://github.com/kainonly/auditstream/discussions)
4. 联系技术支持 support@auditstream.io

## 相关资源

- [v3.0.0 发布说明](./RELEASE-v3.0.0.md)
- [完整更新日志](../CHANGELOG.md)
- [API 文档](https://api.auditstream.io/v3/)
- [插件开发指南](./PLUGIN-DEVELOPMENT.md)
