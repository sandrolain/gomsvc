#!/bin/bash

# Run tests with coverage and output to a file
go test $(go list ./... | grep -v "example" | grep -v "pkg/grpclib/test") -v -race -coverprofile=coverage.out

# Display coverage in the terminal
go tool cover -func=coverage.out

# Generate an HTML report
go tool cover -html=coverage.out -o coverage.html

# Open the HTML report in the default web browser
open coverage.html


