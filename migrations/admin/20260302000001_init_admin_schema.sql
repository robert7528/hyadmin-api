-- Atlas migration: init admin schema
-- Generated: 2026-03-02
-- Note: All DDL uses IF NOT EXISTS for idempotency (NopRevisionReadWriter).

-- ─────────────────────────────────────────────
-- Tenants
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_tenants (
    id          BIGSERIAL    PRIMARY KEY,
    code        VARCHAR(100) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    enabled     BOOLEAN      NOT NULL DEFAULT true,
    infra_type  VARCHAR(50)  NOT NULL DEFAULT 'podman',
    infra_config TEXT,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uk_hyadmin_tenants_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_tenants_deleted_at ON hyadmin_tenants (deleted_at);

-- ─────────────────────────────────────────────
-- Tenant DB Configs
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_tenant_db_configs (
    id           BIGSERIAL    PRIMARY KEY,
    tenant_code  VARCHAR(100) NOT NULL,
    mode         VARCHAR(20)  NOT NULL DEFAULT 'database',
    primary_dsn  TEXT         NOT NULL,
    replica_dsns TEXT,
    schema       VARCHAR(100),
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    CONSTRAINT uk_hyadmin_tenant_db_configs_tenant_code UNIQUE (tenant_code)
);

-- ─────────────────────────────────────────────
-- Admin Users
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_users (
    id            BIGSERIAL    PRIMARY KEY,
    tenant_code   VARCHAR(100) NOT NULL,
    username      VARCHAR(255) NOT NULL,
    password_hash TEXT,
    display_name  TEXT,
    email         TEXT,
    provider      VARCHAR(50)  NOT NULL DEFAULT 'local',
    provider_id   VARCHAR(255),
    enabled       BOOLEAN      NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ,
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT uk_tenant_user UNIQUE (tenant_code, username)
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_users_tenant_code ON hyadmin_users (tenant_code);
CREATE INDEX IF NOT EXISTS idx_hyadmin_users_deleted_at  ON hyadmin_users (deleted_at);

-- ─────────────────────────────────────────────
-- Roles
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_roles (
    id          BIGSERIAL    PRIMARY KEY,
    tenant_code VARCHAR(100) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_roles_tenant_code ON hyadmin_roles (tenant_code);
CREATE INDEX IF NOT EXISTS idx_hyadmin_roles_deleted_at  ON hyadmin_roles (deleted_at);

-- ─────────────────────────────────────────────
-- User → Role assignments (mirrors Casbin g policy)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

-- ─────────────────────────────────────────────
-- Platform Modules (top-level nav tabs)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_modules (
    id           BIGSERIAL    PRIMARY KEY,
    name         VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    i18n         JSONB        NOT NULL DEFAULT '{}',
    icon         VARCHAR(100),
    route        VARCHAR(255) NOT NULL,
    url          TEXT,
    description  TEXT,
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    enabled      BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    CONSTRAINT uk_hyadmin_modules_name UNIQUE (name)
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_modules_deleted_at ON hyadmin_modules (deleted_at);

-- ─────────────────────────────────────────────
-- Features (sidebar menu items per module)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_features (
    id           BIGSERIAL    PRIMARY KEY,
    module_id    BIGINT       NOT NULL,
    name         VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    i18n         JSONB        NOT NULL DEFAULT '{}',
    icon         VARCHAR(100),
    path         VARCHAR(255) NOT NULL,
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    enabled      BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_features_module_id  ON hyadmin_features (module_id);
CREATE INDEX IF NOT EXISTS idx_hyadmin_features_deleted_at ON hyadmin_features (deleted_at);

-- ─────────────────────────────────────────────
-- Permissions (fine-grained access control)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_permissions (
    id          BIGSERIAL    PRIMARY KEY,
    feature_id  BIGINT       NOT NULL,
    code        VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    i18n        JSONB        NOT NULL DEFAULT '{}',
    description TEXT,
    type        VARCHAR(20)  NOT NULL DEFAULT 'button',
    sort_order  INTEGER      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uk_hyadmin_permissions_code UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_permissions_feature_id ON hyadmin_permissions (feature_id);
CREATE INDEX IF NOT EXISTS idx_hyadmin_permissions_deleted_at ON hyadmin_permissions (deleted_at);

-- ─────────────────────────────────────────────
-- Role → Permission assignments (mirrors Casbin p policy)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_role_permissions (
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- ─────────────────────────────────────────────
-- Audit Logs
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_audit_logs (
    id          BIGSERIAL    PRIMARY KEY,
    tenant_code VARCHAR(100) NOT NULL,
    user_id     BIGINT,
    username    VARCHAR(255),
    action      VARCHAR(50),
    resource    VARCHAR(100),
    resource_id VARCHAR(255),
    detail      TEXT,
    ip          VARCHAR(50),
    user_agent  TEXT,
    created_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_hyadmin_audit_logs_tenant_code ON hyadmin_audit_logs (tenant_code);
CREATE INDEX IF NOT EXISTS idx_hyadmin_audit_logs_user_id     ON hyadmin_audit_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_hyadmin_audit_logs_created_at  ON hyadmin_audit_logs (created_at);

-- ─────────────────────────────────────────────
-- Casbin Rules (managed by casbin/gorm-adapter)
-- Table name matches casbinx.NewEnforcer config.
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_casbin_rules (
    id    BIGSERIAL    PRIMARY KEY,
    ptype VARCHAR(100),
    v0    VARCHAR(100),
    v1    VARCHAR(100),
    v2    VARCHAR(100),
    v3    VARCHAR(100),
    v4    VARCHAR(100),
    v5    VARCHAR(100),
    CONSTRAINT uk_hyadmin_casbin_rules UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
);

-- ─────────────────────────────────────────────
-- Application Settings (hot-updatable via Admin UI)
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS hyadmin_settings (
    key         VARCHAR(255) PRIMARY KEY,
    value       TEXT         NOT NULL DEFAULT '',
    type        VARCHAR(20)  NOT NULL DEFAULT 'string',
    group_name  VARCHAR(100) NOT NULL DEFAULT 'general',
    description TEXT,
    is_public   BOOLEAN      NOT NULL DEFAULT false,
    updated_at  TIMESTAMPTZ,
    updated_by  VARCHAR(255)
);
