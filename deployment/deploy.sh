#!/usr/bin/env bash
# HySP Admin API - Deployment Script (pull from ghcr.io)
# Usage: sudo bash /hysp/hyadmin-api/deployment/deploy.sh

set -euo pipefail

APP_DIR="/hysp/hyadmin-api"
IMAGE="ghcr.io/robert7528/hyadmin-api:latest"
QUADLET_SRC="$APP_DIR/deployment/hyadmin-api.container"
QUADLET_DEST="/etc/containers/systemd/hyadmin-api.container"
NGINX_SRC="$APP_DIR/deployment/nginx-hyadmin-api.conf"
NGINX_DEST="/etc/nginx/conf.d/service/hyadmin-api.conf"
ENV_FILE="/etc/hyadmin/api.env"

echo "=== [1/4] Pull latest source (configs / quadlet / nginx) ==="
cd "$APP_DIR"
git pull

echo "=== [2/4] Setup env file (if not exists) ==="
if [ ! -f "$ENV_FILE" ]; then
    mkdir -p /etc/hyadmin
    cp "$APP_DIR/deployment/api.env.example" "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    echo ""
    echo "  !! 請編輯 $ENV_FILE 填入正確設定後重新執行 !!"
    echo "  !! 必填：DATABASE_DSN, JWT_SECRET          !!"
    echo ""
    exit 1
fi

echo "=== [3/4] Pull & start container ==="
# 若 GHCR package 為私有，需先執行：
#   podman login ghcr.io -u <github_username> -p <PAT>
podman pull "$IMAGE"

cp "$QUADLET_SRC" "$QUADLET_DEST"
systemctl daemon-reload
systemctl enable hyadmin-api
systemctl restart hyadmin-api
systemctl status hyadmin-api --no-pager

echo "=== [4/4] Install nginx config ==="
mkdir -p "$(dirname "$NGINX_DEST")"
cp "$NGINX_SRC" "$NGINX_DEST"
nginx -t && systemctl reload nginx

echo ""
echo "Done.  (DB migrations 由 entrypoint 自動套用)"
echo "  API:  http://127.0.0.1:8080/api/v1/health"
echo "  Log:  journalctl -u hyadmin-api -f"
