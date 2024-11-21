# gomsvc

> Opinionated µservices framework for GoLang

[![Go Report Card](https://goreportcard.com/badge/github.com/sandrolain/gomsvc)](https://goreportcard.com/report/github.com/sandrolain/gomsvc)
[![GoDoc](https://godoc.org/github.com/sandrolain/gomsvc?status.svg)](https://godoc.org/github.com/sandrolain/gomsvc)

## Overview

`gomsvc` is a modern, opinionated microservices framework for Go that provides a robust foundation for building scalable and maintainable services. It includes utilities for HTTP clients, gRPC services, and common middleware patterns.

## Features

- **HTTP Client Library**
  - Type-safe generic methods
  - Comprehensive error handling
  - Flexible request configuration
  - Automatic retries and timeouts

- **gRPC Support**
  - Code generation utilities
  - Service templates
  - Common interceptors

- **Development Tools**
  - Task-based workflow
  - Code generation scripts
  - Test coverage reporting
  - Linting and formatting

## Installation

```bash
go get github.com/sandrolain/gomsvc
```

## Quick Start

### HTTP Client Example

```go
package main

import (
    "context"
    "time"
    "github.com/sandrolain/gomsvc/pkg/httplib/client"
)

type Response struct {
    Message string `json:"message"`
    Status  int    `json:"status"`
}

func main() {
    // Initialize request configuration
    init := client.Init{
        Headers: map[string]string{
            "Authorization": "Bearer token",
        },
        Timeout: 10 * time.Second,
        RetryCount: 3,
    }

    // Make a type-safe GET request
    ctx := context.Background()
    resp, err := client.GetJSON[Response](ctx, "https://api.example.com/data", init)
    if err != nil {
        panic(err)
    }

    // Use the strongly-typed response
    fmt.Printf("Status: %d, Message: %s\n", resp.Body.Status, resp.Body.Message)
}
```

## Development

This project uses [Task](https://taskfile.dev) for managing development tasks.

### Prerequisites

1. Install Go (1.18 or later)
   ```bash
   brew install go    # macOS
   ```

2. Install Task
   ```bash
   brew install go-task    # macOS
   ```

3. Install project dependencies
   ```bash
   task install-deps
   ```

### Development Commands

| Command | Description |
|---------|-------------|
| `task test` | Run tests with coverage report |
| `task gen-grpc` | Generate gRPC code |
| `task fmt` | Format Go code |
| `task lint` | Run linters |
| `task build` | Build the project |
| `task clean` | Clean build artifacts |
| `task check` | Run all checks (fmt, lint, test) |
| `task install-deps` | Install project dependencies |

### Project Structure

```
.
├── cmd/                    # Command line tools
├── pkg/                    # Public packages
│   ├── httplib/           # HTTP utilities
│   │   ├── client/        # HTTP client package
│   │   └── server/        # HTTP server package
│   └── grpclib/           # gRPC utilities
├── scripts/               # Development scripts
│   ├── test.sh           # Test execution
│   └── gen-grpc.sh       # gRPC generation
├── Taskfile.yml          # Task definitions
└── README.md             # This file
```

### Running Tests

The test suite includes unit tests and integration tests:

```bash
# Run all tests with coverage
task test

# Run specific package tests
go test ./pkg/httplib/client -v

# Run tests with race detection
go test -race ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-resty](https://github.com/go-resty/resty) - HTTP client library
- [Task](https://taskfile.dev) - Task runner
- [buf](https://buf.build) - Protocol buffer tooling