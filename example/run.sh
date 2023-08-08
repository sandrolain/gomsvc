#!/bin/bash

export PORT=3000
export REDIS_ADDR="localhost:6379"
export REDIS_PWD="mypassword"

go run ./main.go