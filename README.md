# hyadmin-api

HySP Admin Platform — Go backend API.

## Tech Stack

- Go + uber-go/fx (DI) + Gin (HTTP) + Cobra/Viper (CLI/config) + zap + lumberjack
- PostgreSQL + GORM + dbresolver（讀寫分離）
- Atlas（DB 版本控制）
- Podman Quadlet + systemctl（部署）

## Quick Start（本地開發）

```bash
# 啟動 PostgreSQL
podman-compose up -d db

# 複製設定
cp deployment/api.env.example .env.local

# 套用 DB migration
go run ./cmd/migrate admin

# 啟動 server
go run ./cmd/server serve
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/modules` | 已註冊的 micro-frontend 模組清單 |
| GET | `/api/v1/tenants` | List tenants |
| POST | `/api/v1/tenants` | Create tenant |
| GET | `/api/v1/tenants/:id` | Get tenant |
| PUT | `/api/v1/tenants/:id` | Update tenant |
| DELETE | `/api/v1/tenants/:id` | Delete tenant |
| * | `/api/v1/data/*` | 業務資料（需 TenantDBMiddleware） |

所有路由需要 `X-Tenant-ID` header。

## Multi-Tenant DB

每個 tenant 的 DB 連線設定存在 admin DB（`tenant_db_configs` 表）：

| 欄位 | 說明 |
|------|------|
| `mode` | `database`（各自 DSN）或 `schema`（同 PostgreSQL 不同 schema） |
| `primary_dsn` | Write 連線 |
| `replica_dsns` | Read 連線 JSON array（空 = 不分離） |
| `schema` | schema mode 時的 PostgreSQL schema 名稱 |

Handler 取 tenant DB：

```go
db := middleware.GetTenantDB(c)
```

## DB Migration（Atlas）

```bash
# 開發：從 GORM models 生成 migration SQL
atlas migrate diff <name> --env local
atlas migrate status --env local

# 套用 admin DB migrations
go run ./cmd/migrate admin

# 套用單一 tenant DB/schema migrations
go run ./cmd/migrate tenant --code <tenant_code>

# 套用全部 tenant
go run ./cmd/migrate all-tenants
```

## Configuration

`configs/config.yaml`，可用環境變數覆寫：

| 環境變數 | 對應設定 | 預設值 |
|----------|----------|--------|
| `SERVER_PORT` | `server.port` | `8080` |
| `SERVER_MODE` | `server.mode` | `debug` |
| `DATABASE_DSN` | `database.dsn` | — |
| `LOG_LEVEL` | `log.level` | `info` |

生產環境變數放 `/etc/hyadmin/api.env`（參考 `deployment/api.env.example`）。

## Deploy

```bash
# 第一次
git clone https://github.com/robert7528/hyadmin-api.git /hysp/hyadmin-api
sudo bash /hysp/hyadmin-api/deployment/deploy.sh

# 更新
cd /hysp/hyadmin-api && sudo bash deployment/deploy.sh
```

`deploy.sh` 執行步驟：git pull → env check → migrate admin → podman build → Quadlet 安裝 → systemctl restart → nginx reload
