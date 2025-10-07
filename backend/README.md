# Backend 服务

Go 登录服务在启动时会自动连接 PostgreSQL 并执行幂等建表，默认配置与 `docker-compose.yml` 中的数据库保持一致。

## 环境变量

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `9012` | HTTP 服务监听端口 |
| `DB_HOST` | `127.0.0.1` | 数据库主机地址 |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_USER` | `opsboard` | 数据库用户名 |
| `DB_PASSWORD` | `admin` | 数据库密码 |
| `DB_NAME` | `opsboard` | 连接的数据库名称 |
| `DB_SSLMODE` | `disable` | PostgreSQL `sslmode` 设置 |
| `DATABASE_URL` | – | 如设置将优先生效，格式示例：`postgres://user:pass@host:port/dbname?sslmode=disable` |

## 启动

```bash
GOCACHE=$(pwd)/.gocache \
DB_HOST=127.0.0.1 \
DB_PORT=5432 \
DB_USER=opsboard \
DB_PASSWORD=admin \
DB_NAME=opsboard \
go run ./cmd/server
```

首次运行将自动创建下列表：

- `users`
- `user_credentials`
- `user_identities`
- `user_roles`
- `user_audit_logs`

登录接口入口：`POST /api/login`。
