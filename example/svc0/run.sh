#!/bin/bash

export PORT=3000
export REDIS_ADDR="localhost:6379"
export REDIS_PWD="mypassword"
export LOG_LEVEL="debug"

go run ./main.go