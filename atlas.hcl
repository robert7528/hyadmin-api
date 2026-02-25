# Atlas project configuration
# Docs: https://atlasgo.io/atlas-schema/projects
#
# Usage:
#   atlas migrate diff --env local        # generate migration from GORM models
#   atlas migrate apply --env local       # apply pending migrations
#   atlas migrate status --env local      # check migration status
#   atlas migrate lint --env local        # lint migration files

# GORM model schema — Atlas CLI compiles & runs the loader to extract schema
data "external_schema" "gorm" {
  program = [
    "go", "run", "-mod=mod",
    "./internal/database/loader",
  ]
}

# ── Admin DB (Tenant, TenantDBConfig tables) ──────────────────────────────────
env "local" {
  src = data.external_schema.gorm.url
  url = "postgres://hyadmin:hyadmin@localhost:5432/hyadmin?sslmode=disable"
  migration {
    dir = "file://migrations/admin"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  src = data.external_schema.gorm.url
  url = getenv("ADMIN_DATABASE_URL")
  migration {
    dir = "file://migrations/admin"
  }
}

# ── Tenant DB / Schema ────────────────────────────────────────────────────────
# For tenant migrations, use --url flag directly:
#   atlas migrate apply --dir file://migrations/tenant \
#     --url "postgres://user:pass@host/db?search_path=tenant_code"
