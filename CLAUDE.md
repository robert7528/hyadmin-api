# hyadmin-api

## Development Environment

- **Windows local**: Source code editing only. No Go runtime available.
- **GitHub**: `robert7528/hyadmin-api`
- **CI/CD**: GitHub Actions → build & push `ghcr.io/robert7528/hyadmin-api:latest`
- **Deploy**: Linux API server (`/hysp/hyadmin-api/`) via Podman Quadlet + systemctl.

## Infrastructure Topology

```
[ Windows PC ]  →  git push  →  [ GitHub Actions ]
                                       │ build & push image
                                       ▼
                               ghcr.io/robert7528/hyadmin-api:latest
                                       │ podman pull (deploy.sh)
                                       ▼
[ Linux API Server ]  ←─── DATABASE_DSN ───→  [ Linux DB Server ]
  Podman + Quadlet                                  PostgreSQL
  /etc/hyadmin/api.env
```

- **API Server**：跑 hyadmin-api 容器，不裝 PostgreSQL
- **DB Server**：獨立 Linux 主機，跑 PostgreSQL
- **連線設定**：`DATABASE_DSN` 直接指向 DB Server IP/hostname（非 host-gateway）
- **Quadlet 的 `AddHost=host-gateway`**：僅供未來有需要連 API host 本機服務時使用，目前對 DB 連線無作用

## Project Structure

```
hyadmin-api/
├── cmd/
│   ├── server/main.go          # Cobra CLI → app.Run()
│   └── migrate/main.go         # Atlas migration runner (admin/tenant/all-tenants)
├── internal/
│   ├── app/app.go              # fx DI 組裝
│   ├── config/config.go        # Viper（env vars: DATABASE_DSN, SERVER_PORT...）
│   ├── logger/logger.go        # zap + lumberjack
│   ├── server/server.go        # Gin routes + fx Lifecycle
│   ├── middleware/
│   │   ├── tenant.go           # X-Tenant-ID → c.Set("tenant_id")
│   │   ├── tenant_db.go        # TenantDBMiddleware + GetTenantDB(c)
│   │   ├── auth.go             # 認證 stub
│   │   └── recovery.go         # panic recovery
│   ├── module/                 # 模組 registry（GET /api/v1/modules）
│   ├── tenant/                 # Tenant CRUD（admin DB）
│   ├── health/                 # GET /api/v1/health
│   └── database/
│       ├── database.go         # Connect()（admin DB）
│       ├── tenant_db_config.go # TenantDBConfig model
│       ├── manager.go          # DBManager（多租戶連線池 + dbresolver）
│       ├── migrate.go          # MigrateAdmin / MigrateTenant（Atlas runner）
│       └── loader/main.go      # Atlas CLI 用的 GORM schema loader
├── migrations/
│   ├── admin/                  # admin DB migration SQL（Atlas 自動生成）
│   └── tenant/                 # 租戶 DB/schema migration SQL
├── configs/config.yaml
├── atlas.hcl                   # Atlas CLI 設定
├── deployment/
│   ├── hyadmin-api.container   # Podman Quadlet
│   ├── api.env.example         # /etc/hyadmin/api.env 範本
│   ├── nginx-hyadmin-api.conf  # nginx location config
│   └── deploy.sh               # 完整部署腳本
└── Containerfile
```

## Tech Stack

- **Runtime**: Go + uber-go/fx (DI) + Gin (HTTP) + Cobra/Viper (CLI/config) + zap + lumberjack
- **DB**: PostgreSQL + GORM + gorm/dbresolver（讀寫分離）
- **DB 版控**: Atlas (`ariga.io/atlas` + `atlas-provider-gorm`)
- **部署**: Podman Quadlet + systemctl + nginx

## Key Patterns

### Config
- `configs/config.yaml` + env vars（Viper `AutomaticEnv` + `SetEnvKeyReplacer`）
- Env key 對應：`DATABASE_DSN` → `database.dsn`，`SERVER_PORT` → `server.port`
- 生產 env file：`/etc/hyadmin/api.env`

### Routes
| Method | Path | 說明 |
|--------|------|------|
| GET | `/api/v1/health` | 健康檢查 |
| GET | `/api/v1/modules` | 已註冊的 micro-frontend 模組清單 |
| CRUD | `/api/v1/tenants` | 租戶管理（admin DB） |
| * | `/api/v1/data/*` | 業務資料（套 TenantDBMiddleware） |

### 多租戶 DB（DBManager + dbresolver）
- `TenantDBConfig` 存在 admin DB，記錄每個 tenant 的連線設定
- `mode=database`：不同 database，各自 DSN
- `mode=schema`：同一 PostgreSQL，不同 schema（自動加 `search_path`）
- 讀寫分離：`ReplicaDSNs`（JSON array）→ dbresolver `RandomPolicy`
- Handler 取 tenant DB：`middleware.GetTenantDB(c)`
- Config 異動後呼叫：`mgr.InvalidateCache(tenantCode)`

### Atlas DB 版控
```bash
# 開發：從 GORM models 生成 migration SQL
atlas migrate diff <name> --env local
atlas migrate status --env local

# 部署：套用 pending migrations
go run ./cmd/migrate admin               # admin DB
go run ./cmd/migrate tenant --code thu   # 單一租戶
go run ./cmd/migrate all-tenants         # 全部租戶
```

### nginx
- 路徑：`/hyadmin-api/` → `http://127.0.0.1:8080/`
- **trailing slash**：nginx 剝離前綴，Gin 收到 `/api/v1/...`

## Deploy

```bash
# 第一次
git clone https://github.com/robert7528/hyadmin-api.git /hysp/hyadmin-api
sudo bash /hysp/hyadmin-api/deployment/deploy.sh
# → 會建立 /etc/hyadmin/api.env，填入 DB 密碼後再跑一次

# 更新
cd /hysp/hyadmin-api
sudo bash deployment/deploy.sh
# 步驟：git pull → env check → podman pull image
#        → Quadlet 安裝 → systemctl restart → nginx reload
# migrate 由 entrypoint.sh 在容器啟動時自動執行

# go mod tidy（第一次或新增依賴後）
go mod tidy
```
