#!/bin/bash

SCRIPT=$(realpath "$0")
WD=$(dirname "$SCRIPT")

export LOG_LEVEL="debug"
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=$(kubectl get secret --namespace redis redis -o jsonpath="{.data.redis-password}" | base64 -d)

go run "$WD/main.go"
