#!/usr/bin/env bash
#
# Redeploy the app on the VPS after a manual `git pull`.
# Run from the repo root:  ./deploy.sh
#
# Rebuilds the Go binary and the React bundle, then restarts the API service
# and reloads Apache. Requires: go, node/npm, and sudo rights for systemctl.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

echo "==> Pulling latest changes"
git pull --ff-only

echo "==> Building backend"
cd "$ROOT/backend"
go build -o server ./cmd/server

echo "==> Building frontend"
cd "$ROOT/frontend"
npm ci
npm run build

echo "==> Restarting services"
sudo systemctl restart expense-tracker
sudo systemctl reload httpd

echo "==> Done. Health check:"
sleep 1
curl -fsS http://127.0.0.1:8080/health && echo
