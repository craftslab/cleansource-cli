# CleanSource SCA CLI

English | [中文](./README_cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/cleansource-sca-cli)](https://goreportcard.com/report/github.com/craftslab/cleansource-sca-cli)
[![License](https://img.shields.io/github/license/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/tags)

A Go implementation of the CleanSource SCA build scanner with comprehensive support for multiple build tools and extensive test coverage.

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
- ✅ Gradle dependency scanning
- ✅ npm/Node.js dependency scanning
- ✅ Go modules dependency scanning
- ✅ Pipenv dependency scanning
- ✅ File compression and archiving
- ✅ REST API client for server communication
- ✅ Concurrent processing for large codebases
- ✅ Comprehensive test coverage
- ✅ Cross-platform support

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
| Maven | ✅ Complete | Full dependency tree analysis with POM parsing |
| pip | ✅ Complete | Requirements.txt and installed packages analysis |
| Gradle | ✅ Complete | Build.gradle parsing with dependency extraction |
| npm | ✅ Complete | Package.json parsing with all dependency types |
| Go Modules | ✅ Complete | go.mod parsing with module dependency analysis |
| Pipenv | ✅ Complete | Pipfile parsing with pipenv dependency resolution |

### Build Tool Detection

The CLI automatically detects build tools based on the presence of characteristic files:

- **Maven**: `pom.xml`
- **Gradle**: `build.gradle`, `build.gradle.kts`
- **npm**: `package.json`
- **Go Modules**: `go.mod`
- **Pipenv**: `Pipfile`, `Pipfile.lock`
- **pip**: `requirements.txt`, `setup.py`, `pyproject.toml`

## Development

### Running Tests

The project includes comprehensive test coverage for all scanner implementations:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./pkg/buildtools/...

# Run integration tests
go test -v -run "TestScannerIntegration" .

# Run benchmarks
go test -bench=. -benchmem ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Test Scripts

**Linux/macOS:**
```bash
./run-tests.sh
```

**Windows:**
```cmd
run-tests.bat
```

**Advanced Options:**
```bash
# Linux/macOS - Run only unit tests (skip integration tests)
./run-tests.sh --unit-only

# Linux/macOS - Enable verbose output
./run-tests.sh --verbose

# Linux/macOS - Set custom coverage threshold
./run-tests.sh --coverage-threshold 80

# Linux/macOS - Skip cleanup of test artifacts
./run-tests.sh --no-cleanup

# Linux/macOS - Show help
./run-tests.sh --help
```

```cmd
REM Windows - Run only unit tests (skip integration tests)
run-tests.bat --unit-only

REM Windows - Enable verbose output
run-tests.bat --verbose

REM Windows - Set custom coverage threshold
run-tests.bat --coverage-threshold 80

REM Windows - Skip cleanup of test artifacts
run-tests.bat --no-cleanup

REM Windows - Show help
run-tests.bat --help
```

### Building

```bash
# Build for current platform
go build -o cleansource-sca-cli main.go

# Build with optimizations
go build -ldflags="-s -w" -o cleansource-sca-cli main.go

# Cross-compilation examples
GOOS=windows GOARCH=amd64 go build -o cleansource-sca-cli.exe main.go
GOOS=linux GOARCH=amd64 go build -o cleansource-sca-cli main.go
GOOS=darwin GOARCH=amd64 go build -o cleansource-sca-cli main.go
```

## Scanner Implementations

### Go Modules Scanner
- **Detection**: `go.mod` files
- **Features**: Module name/version extraction, dependency analysis via `go list`
- **Dependencies**: Requires Go 1.11+ with modules support

### NPM Scanner
- **Detection**: `package.json` files
- **Features**: Project info extraction, dependency parsing (runtime, dev, peer)
- **Dependencies**: Optional npm executable for enhanced functionality

### Gradle Scanner
- **Detection**: `build.gradle`, `build.gradle.kts` files
- **Features**: Project info extraction, dependency parsing with scope detection
- **Dependencies**: Optional Gradle executable or wrapper

### Pipenv Scanner
- **Detection**: `Pipfile`, `Pipfile.lock` files
- **Features**: Project info extraction, dependency resolution via `pipenv run pip freeze`
- **Dependencies**: Requires pipenv and Python environment

### Maven Scanner
- **Detection**: `pom.xml` files
- **Features**: POM parsing, dependency tree analysis
- **Dependencies**: Optional Maven executable for enhanced functionality

### Pip Scanner
- **Detection**: `requirements.txt`, `setup.py`, `pyproject.toml` files
- **Features**: Requirements parsing, installed package analysis
- **Dependencies**: Optional pip executable

### Adding New Build Tools

To add support for a new build tool:

1. Create a new scanner in `pkg/buildtools/`
2. Implement the `Scannable` interface:
   - `ExeFind()`: Find the build tool executable
   - `FileFind()`: Check for required files
   - `ScanExecute()`: Execute the dependency scan
3. Add detection logic in `pkg/buildtools/scanner.go`
4. Add comprehensive tests in `pkg/buildtools/scanners_test.go`
5. Update model tests in `internal/model/types_test.go`
6. Test with sample projects

## Examples

### Multi-Project Scanning

The CLI can scan projects with multiple build tools:

```bash
# Scan a project with Go modules and npm
./cleansource-sca-cli --server-url https://your-server.com \
    --token your-token \
    --task-dir /path/to/multi-language-project
```

### Project Structure Examples

**Go Project:**
```
project/
├── go.mod
├── main.go
└── go.sum
```

**Node.js Project:**
```
project/
├── package.json
├── package-lock.json
└── src/
```

**Gradle Project:**
```
project/
├── build.gradle
├── settings.gradle
└── src/
```

**Python Pipenv Project:**
```
project/
├── Pipfile
├── Pipfile.lock
└── src/
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add comprehensive tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Run the test scripts (`./run-tests.sh` or `run-tests.bat`)
7. Update documentation if needed
8. Commit your changes (`git commit -m 'Add amazing feature'`)
9. Push to the branch (`git push origin feature/amazing-feature`)
10. Open a Pull Request

### Development Guidelines

- Follow Go coding standards and best practices
- Add tests for all new functionality
- Update documentation for new features
- Ensure cross-platform compatibility
- Use meaningful commit messages
- Keep the codebase clean and well-documented

