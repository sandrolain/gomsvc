#!/bin/bash

# Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "gosec is not installed. Installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Run gosec with standard configuration
echo "Running gosec security scanner..."
gosec -fmt=json -out=security-report.json ./...

# Check if the scan found any issues
if [ $? -eq 1 ]; then
    echo "Security issues found. Check security-report.json for details"
    exit 1
elif [ $? -eq 2 ]; then
    echo "Scan failed. Please check your code and try again"
    exit 2
else
    echo "No security issues found!"
fi
