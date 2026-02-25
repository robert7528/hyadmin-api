#!/usr/bin/env bash
# HySP Admin API - Deployment Script (Podman)
# Usage: sudo bash /hysp/hyadmin-api/deployment/deploy.sh

set -euo pipefail

APP_DIR="/hysp/hyadmin-api"
IMAGE="localhost/hyadmin-api:latest"
QUADLET_SRC="$APP_DIR/deployment/hyadmin-api.container"
QUADLET_DEST="/etc/containers/systemd/hyadmin-api.container"
NGINX_SRC="$APP_DIR/deployment/nginx-hyadmin-api.conf"
NGINX_DEST="/etc/nginx/conf.d/service/hyadmin-api.conf"
ENV_DEST="/etc/hyadmin/api.env"

echo "=== [1/7] Pull latest code ==="
cd "$APP_DIR"
git pull

echo "=== [2/7] Setup env file (if not exists) ==="
if [ ! -f "$ENV_DEST" ]; then
    mkdir -p /etc/hyadmin
    cp "$APP_DIR/deployment/api.env.example" "$ENV_DEST"
    chmod 600 "$ENV_DEST"
    echo ""
    echo "  !! 請編輯 $ENV_DEST 填入正確的 DB 密碼 !!"
    echo "  !! 然後重新執行此 script !!"
    echo ""
    exit 1
fi

echo "=== [3/7] Run admin DB migrations ==="
if ! command -v atlas &>/dev/null; then
    echo "Installing Atlas CLI..."
    curl -sSf https://atlasgo.sh | sh
fi
export ADMIN_DATABASE_URL=$(grep '^DATABASE_DSN=' "$ENV_DEST" | cut -d= -f2-)
go run ./cmd/migrate admin

echo "=== [4/7] Build container image ==="
podman build -t "$IMAGE" .

echo "=== [5/7] Install Quadlet file ==="
cp "$QUADLET_SRC" "$QUADLET_DEST"
systemctl daemon-reload

echo "=== [6/7] Enable & restart service ==="
systemctl enable hyadmin-api
systemctl restart hyadmin-api
systemctl status hyadmin-api --no-pager

echo "=== [7/7] Install nginx config ==="
mkdir -p "$(dirname "$NGINX_DEST")"
cp "$NGINX_SRC" "$NGINX_DEST"
nginx -t && systemctl reload nginx

echo ""
echo "Done."
echo "  API:   http://127.0.0.1:8080/api/v1/health"
echo "  Nginx: https://your-domain/hyadmin-api/"
