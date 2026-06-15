#!/bin/bash
set -e

cd /Users/agoxq/Documents/Project/QuantLab

SERVICE=$1
PROTO_FILE="api/$SERVICE/v1/$SERVICE.proto"

if [ ! -f "$PROTO_FILE" ]; then
  echo "ERROR: $PROTO_FILE not found"
  exit 1
fi

echo "=== Generating pb for $SERVICE ==="
mkdir -p "app/$SERVICE/pb"

protoc --proto_path=. \
  --go_out=. --go_opt=module=github.com/agoXQ/QuantLab \
  --go-grpc_out=. --go-grpc_opt=module=github.com/agoXQ/QuantLab \
  "$PROTO_FILE"

echo "=== $SERVICE pb generated ==="
