#!/usr/bin/env bash
# Generate OpenAPI schema for AI tooling / docs
# Usage: ./scripts/generate_ai_schema.sh

set -euo pipefail

echo "Generating OpenAPI schema..."
# TODO: integrate swaggo/swag or similar
# go run github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go
echo "Not yet implemented — add swaggo/swag or oapi-codegen as needed."
