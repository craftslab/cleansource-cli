# CleanSource SCA CLI

English | [中文](./README_cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/cleansource-sca-cli)](https://goreportcard.com/report/github.com/craftslab/cleansource-sca-cli)
[![License](https://img.shields.io/github/license/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/tags)

A Go implementation of the CleanSource SCA build scanner.

## Overview

- **Source code scanning** with fingerprint generation
- **Dependency analysis** for multiple build tools (Maven, pip, Gradle, npm, Go, etc.)
- **Multi-threaded processing** for improved performance
- **REST API integration** with the CleanSource SCA platform
- **Cross-platform support** (Windows, Linux, macOS)

## Features

- ✅ Source code fingerprinting (WFP generation)
- ✅ Maven dependency scanning
- ✅ Python pip dependency scanning
- ✅ File compression and archiving
- ✅ REST API client for server communication
- ✅ Concurrent processing for large codebases
- 🚧 Gradle dependency scanning (in progress)
- 🚧 npm/Node.js dependency scanning (in progress)
- 🚧 Go modules dependency scanning (in progress)
- 🚧 Pipenv dependency scanning (in progress)

## Installation

### Prerequisites

- Go 1.21 or later
- Git

### Build from source

```bash
git clone https://github.com/craftslab/cleansource-sca-cli.git
cd cleansource-sca-cli
go mod download
go build -o cleansource-sca-cli main.go
```

### Cross-compilation

For Windows:
```bash
GOOS=windows GOARCH=amd64 go build -o cleansource-sca-cli.exe main.go
```

For Linux:
```bash
GOOS=linux GOARCH=amd64 go build -o cleansource-sca-cli main.go
```

For macOS:
```bash
GOOS=darwin GOARCH=amd64 go build -o cleansource-sca-cli main.go
```

## Usage

### Basic Usage

```bash
# Scan a source directory
./cleansource-sca-cli --server-url https://your-server.com \
    --username your-username \
    --password your-password \
    --task-dir /path/to/source/code

# Using token authentication
./cleansource-sca-cli --server-url https://your-server.com \
    --token your-auth-token \
    --task-dir /path/to/source/code
```

### Advanced Options

```bash
# Full scan with custom project information
./cleansource-sca-cli --server-url https://your-server.com \
    --token your-token \
    --task-dir /path/to/source \
    --custom-project "MyProject" \
    --custom-product "MyProduct" \
    --custom-version "1.0.0" \
    --license-name "MIT" \
    --notification-email "dev@company.com" \
    --thread-num 60
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--server-url` | Server URL for API communication | Required |
| `--username` | Username for authentication | Required if no token |
| `--password` | Password for authentication | Required if no token |
| `--token` | Authentication token | Required if no username/password |
| `--task-dir` | Directory to scan | Required |
| `--scan-type` | Type of scan (source, docker, binary) | source |
| `--to-path` | Output directory for results | Parent of task-dir |
| `--build-depend` | Build dependency tree | true |
| `--custom-project` | Custom project name | Auto-detected |
| `--custom-product` | Custom product name | Auto-detected |
| `--custom-version` | Custom version | Auto-detected |
| `--license-name` | License name | Auto-detected |
| `--notification-email` | Notification email | - |
| `--thread-num` | Number of threads (1-60) | 30 |
| `--log-level` | Log level (debug, info, warn, error) | info |

## Architecture

1. **CLI Layer** (`cmd/`): Command-line interface using Cobra
2. **Application Layer** (`internal/app/`): Main business logic
3. **Scanner Layer** (`internal/scanner/`): File fingerprinting
4. **Build Tools** (`pkg/buildtools/`): Build system integration
5. **Client Layer** (`pkg/client/`): Server communication
6. **Utils** (`internal/utils/`): Common utilities

## Supported Build Tools

| Build Tool | Status | Description |
|------------|--------|-------------|
| Maven | ✅ Complete | Full dependency tree analysis |
| pip | ✅ Complete | Requirements.txt and installed packages |
| Gradle | 🚧 Partial | Basic detection, scanning in progress |
| npm | 🚧 Partial | Basic detection, scanning in progress |
| Go Modules | 🚧 Partial | Basic detection, scanning in progress |
| Pipenv | 🚧 Partial | Basic detection, scanning in progress |

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Build for current platform
go build -o cleansource-sca-cli main.go

# Build with optimizations
go build -ldflags="-s -w" -o cleansource-sca-cli main.go
```

### Adding New Build Tools

To add support for a new build tool:

1. Create a new scanner in `pkg/buildtools/`
2. Implement the `Scannable` interface:
   - `ExeFind()`: Find the build tool executable
   - `FileFind()`: Check for required files
   - `ScanExecute()`: Execute the dependency scan
3. Add detection logic in `pkg/buildtools/scanner.go`
4. Test with sample projects

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request
