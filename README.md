# hyadmin-api

HySP Admin Platform — Go backend API.

## Tech Stack

- Go + uber-go/fx + Gin + Cobra/Viper + zap
- PostgreSQL + GORM

## Quick Start

```bash
# with Podman Compose
podman-compose up -d

# or Docker Compose
docker compose up -d
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/modules` | List registered micro-frontend modules |
| GET | `/api/v1/tenants` | List tenants |
| POST | `/api/v1/tenants` | Create tenant |
| GET | `/api/v1/tenants/:id` | Get tenant |
| PUT | `/api/v1/tenants/:id` | Update tenant |
| DELETE | `/api/v1/tenants/:id` | Delete tenant |

All routes require `X-Tenant-ID` header.

## Configuration

Edit `configs/config.yaml` or override via environment variables:

```yaml
server:
  port: "8080"
  mode: "debug"   # debug | release

database:
  dsn: "host=localhost user=hyadmin password=hyadmin dbname=hyadmin port=5432 sslmode=disable"

log:
  level: "debug"
  filename: "logs/hyadmin-api.log"
```

## Build

```bash
go mod tidy
go build -o hyadmin-api ./cmd/server
./hyadmin-api serve
```

## DB Migration

```bash
go run ./cmd/migrate
```
