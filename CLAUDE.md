# hyadmin-api

## Development Environment

- **Windows local**: Source code editing only. No Go runtime available.
- **GitHub**: `hysp/hyadmin-api`
- **Deploy**: Linux server pulls from GitHub and runs the binary.

## Project Structure

Go backend for HySP Admin Platform.

- Module path: `github.com/hysp/hyadmin-api`
- Entry point: `cmd/server/main.go` → `hyadmin-api serve`
- DB migration: `cmd/migrate/main.go`

## Tech Stack

- Go + uber-go/fx (DI) + Gin (HTTP) + Cobra/Viper (CLI/config) + zap (logging)
- PostgreSQL + GORM

## Key Patterns

- Dependency injection via `go.uber.org/fx`
- Config: `configs/config.yaml` + env vars (via Viper `AutomaticEnv`)
- Structured logging: zap + lumberjack file rotation
- Routes: `GET /api/v1/health`, `GET /api/v1/modules`, CRUD `/api/v1/tenants`
- Tenant isolation: `X-Tenant-ID` header (middleware enforced)

## On Linux server

```bash
cd /hysp/hyadmin-api
git pull
go build -o hyadmin-api ./cmd/server
sudo systemctl restart hyadmin-api
```
