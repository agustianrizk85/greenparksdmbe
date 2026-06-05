#!/usr/bin/env bash
# Deploy backend SDM: tarik kode terbaru, build, jalankan/-ulang via PM2.
set -euo pipefail
cd "$(dirname "$0")"

echo "==> git pull"
git pull --ff-only

echo "==> go build"
export PATH="$PATH:/usr/local/go/bin"
CGO_ENABLED=0 go build -trimpath -o sdm-server ./cmd/server

set -a; [ -f /opt/apps/sdm.env ] && . /opt/apps/sdm.env; set +a

echo "==> (re)start PM2: sdm-be"
pm2 restart sdm-be --update-env 2>/dev/null || pm2 start ./sdm-server --name sdm-be --update-env
pm2 save
echo "==> selesai. status:"
pm2 status sdm-be
