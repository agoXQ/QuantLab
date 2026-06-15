#!/bin/bash
set -e

cd /Users/agoxq/Documents/Project/QuantLab

SERVICES=("user" "strategy" "portfolio" "billing" "community" "ranking" "notification" "ai")

echo "=== Generating pb files for all services ==="

for svc in "${SERVICES[@]}"; do
  PROTO_FILE="api/$svc/v1/$svc.proto"
  if [ -f "$PROTO_FILE" ]; then
    echo "  -> Generating pb for $svc..."
    mkdir -p "app/$svc/pb"
    protoc --proto_path=. \
      --go_out=. --go_opt=module=github.com/agoXQ/QuantLab \
      --go-grpc_out=. --go-grpc_opt=module=github.com/agoXQ/QuantLab \
      "$PROTO_FILE"
  else
    echo "  -> WARNING: $PROTO_FILE not found, skipping"
  fi
done

echo ""
echo "=== Creating service layer directories ==="

for svc in "${SERVICES[@]}"; do
  echo "  -> Creating $svc service layer..."
  mkdir -p "app/$svc/etc"
  mkdir -p "app/$svc/internal/config"
  mkdir -p "app/$svc/internal/logic"
  mkdir -p "app/$svc/internal/server"
  mkdir -p "app/$svc/internal/svc"
  mkdir -p "app/$svc/${svc}service"
done

echo ""
echo "=== All scaffolds created ==="
