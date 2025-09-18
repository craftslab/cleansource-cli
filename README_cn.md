# CleanSource SCA CLI

[English](./README.md) | 中文

[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/cleansource-sca-cli)](https://goreportcard.com/report/github.com/craftslab/cleansource-sca-cli)
[![License](https://img.shields.io/github/license/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/tags)

一个 Go 实现的 CleanSource SCA 构建扫描器，具有全面的多构建工具支持和广泛的测试覆盖。

## 概述

- **源代码扫描** 支持指纹生成
- **依赖分析** 支持多种构建工具 (Maven, pip, Gradle, npm, Go 等)
- **多线程处理** 提升性能
- **REST API 集成** 与 CleanSource SCA 平台对接
- **跨平台支持** (Windows, Linux, macOS)

## 功能特性

- ✅ 源代码指纹识别 (WFP 生成)
- ✅ Maven 依赖扫描
- ✅ Python pip 依赖扫描
- ✅ Gradle 依赖扫描
- ✅ npm/Node.js 依赖扫描
- ✅ Go 模块依赖扫描
- ✅ Pipenv 依赖扫描
- ✅ 文件压缩和归档
- ✅ REST API 客户端用于服务器通信
- ✅ 大型代码库并发处理
- ✅ 全面的测试覆盖
- ✅ 跨平台支持

## 安装

### 环境要求

- Go 1.21 或更高版本
- Git

### 从源码构建

```bash
git clone https://github.com/craftslab/cleansource-sca-cli.git
cd cleansource-sca-cli
go mod download
go build -o cleansource-sca-cli main.go
```

### 交叉编译

Windows 平台:
```bash
GOOS=windows GOARCH=amd64 go build -o cleansource-sca-cli.exe main.go
```

Linux 平台:
```bash
GOOS=linux GOARCH=amd64 go build -o cleansource-sca-cli main.go
```

macOS 平台:
```bash
GOOS=darwin GOARCH=amd64 go build -o cleansource-sca-cli main.go
```

## 使用方法

### 基本用法

```bash
# 扫描源代码目录
./cleansource-sca-cli --server-url https://your-server.com \
    --username your-username \
    --password your-password \
    --task-dir /path/to/source/code

# 使用令牌认证
./cleansource-sca-cli --server-url https://your-server.com \
    --token your-auth-token \
    --task-dir /path/to/source/code
```

### 高级选项

```bash
# 带有自定义项目信息的完整扫描
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

### 命令行选项

| 选项 | 描述 | 默认值 |
|--------|-------------|---------|
| `--server-url` | API 通信的服务器 URL | 必填 |
| `--username` | 认证用户名 | 无令牌时必填 |
| `--password` | 认证密码 | 无令牌时必填 |
| `--token` | 认证令牌 | 无用户名/密码时必填 |
| `--task-dir` | 要扫描的目录 | 必填 |
| `--scan-type` | 扫描类型 (source, docker, binary) | source |
| `--to-path` | 结果输出目录 | task-dir 的父目录 |
| `--build-depend` | 构建依赖树 | true |
| `--custom-project` | 自定义项目名称 | 自动检测 |
| `--custom-product` | 自定义产品名称 | 自动检测 |
| `--custom-version` | 自定义版本号 | 自动检测 |
| `--license-name` | 许可证名称 | 自动检测 |
| `--notification-email` | 通知邮箱 | - |
| `--thread-num` | 线程数 (1-60) | 30 |
| `--log-level` | 日志级别 (debug, info, warn, error) | info |

## 架构

1. **CLI 层** (`cmd/`): 使用 Cobra 的命令行界面
2. **应用层** (`internal/app/`): 主要业务逻辑
3. **扫描器层** (`internal/scanner/`): 文件指纹识别
4. **构建工具** (`pkg/buildtools/`): 构建系统集成
5. **客户端层** (`pkg/client/`): 服务器通信
6. **工具包** (`internal/utils/`): 通用工具

## 支持的构建工具

| 构建工具 | 状态 | 描述 |
|------------|--------|-------------|
| Maven | ✅ 完成 | 完整的依赖树分析，支持 POM 解析 |
| pip | ✅ 完成 | Requirements.txt 和已安装包分析 |
| Gradle | ✅ 完成 | Build.gradle 解析，支持依赖提取 |
| npm | ✅ 完成 | Package.json 解析，支持所有依赖类型 |
| Go Modules | ✅ 完成 | go.mod 解析，支持模块依赖分析 |
| Pipenv | ✅ 完成 | Pipfile 解析，支持 pipenv 依赖解析 |

### 构建工具检测

CLI 基于特征文件的存在自动检测构建工具：

- **Maven**: `pom.xml`
- **Gradle**: `build.gradle`, `build.gradle.kts`
- **npm**: `package.json`
- **Go Modules**: `go.mod`
- **Pipenv**: `Pipfile`, `Pipfile.lock`
- **pip**: `requirements.txt`, `setup.py`, `pyproject.toml`

## 开发

### 运行测试

项目包含所有扫描器实现的全面测试覆盖：

```bash
# 运行所有测试
go test ./...

# 运行详细输出测试
go test -v ./...

# 运行特定包测试
go test -v ./pkg/buildtools/...

# 运行集成测试
go test -v -run "TestScannerIntegration" .

# 运行基准测试
go test -bench=. -benchmem ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 测试脚本

**Linux/macOS:**
```bash
./run-tests.sh
```

**Windows:**
```cmd
run-tests.bat
```

**高级选项:**
```bash
# Linux/macOS - 仅运行单元测试（跳过集成测试）
./run-tests.sh --unit-only

# Linux/macOS - 启用详细输出
./run-tests.sh --verbose

# Linux/macOS - 设置自定义覆盖率阈值
./run-tests.sh --coverage-threshold 80

# Linux/macOS - 跳过测试工件清理
./run-tests.sh --no-cleanup

# Linux/macOS - 显示帮助
./run-tests.sh --help
```

```cmd
REM Windows - 仅运行单元测试（跳过集成测试）
run-tests.bat --unit-only

REM Windows - 启用详细输出
run-tests.bat --verbose

REM Windows - 设置自定义覆盖率阈值
run-tests.bat --coverage-threshold 80

REM Windows - 跳过测试工件清理
run-tests.bat --no-cleanup

REM Windows - 显示帮助
run-tests.bat --help
```

### 构建

```bash
# 为当前平台构建
go build -o cleansource-sca-cli main.go

# 带优化的构建
go build -ldflags="-s -w" -o cleansource-sca-cli main.go

# 交叉编译示例
GOOS=windows GOARCH=amd64 go build -o cleansource-sca-cli.exe main.go
GOOS=linux GOARCH=amd64 go build -o cleansource-sca-cli main.go
GOOS=darwin GOARCH=amd64 go build -o cleansource-sca-cli main.go
```

## 扫描器实现

### Go 模块扫描器
- **检测**: `go.mod` 文件
- **功能**: 模块名称/版本提取，通过 `go list` 进行依赖分析
- **依赖**: 需要 Go 1.11+ 和模块支持

### NPM 扫描器
- **检测**: `package.json` 文件
- **功能**: 项目信息提取，依赖解析（运行时、开发、对等）
- **依赖**: 可选的 npm 可执行文件以增强功能

### Gradle 扫描器
- **检测**: `build.gradle`, `build.gradle.kts` 文件
- **功能**: 项目信息提取，带作用域检测的依赖解析
- **依赖**: 可选的 Gradle 可执行文件或包装器

### Pipenv 扫描器
- **检测**: `Pipfile`, `Pipfile.lock` 文件
- **功能**: 项目信息提取，通过 `pipenv run pip freeze` 进行依赖解析
- **依赖**: 需要 pipenv 和 Python 环境

### Maven 扫描器
- **检测**: `pom.xml` 文件
- **功能**: POM 解析，依赖树分析
- **依赖**: 可选的 Maven 可执行文件以增强功能

### Pip 扫描器
- **检测**: `requirements.txt`, `setup.py`, `pyproject.toml` 文件
- **功能**: 需求解析，已安装包分析
- **依赖**: 可选的 pip 可执行文件

### 添加新的构建工具

要添加对新构建工具的支持：

1. 在 `pkg/buildtools/` 中创建新的扫描器
2. 实现 `Scannable` 接口：
   - `ExeFind()`: 查找构建工具可执行文件
   - `FileFind()`: 检查所需文件
   - `ScanExecute()`: 执行依赖扫描
3. 在 `pkg/buildtools/scanner.go` 中添加检测逻辑
4. 在 `pkg/buildtools/scanners_test.go` 中添加全面测试
5. 在 `internal/model/types_test.go` 中更新模型测试
6. 使用示例项目进行测试

## 示例

### 多项目扫描

CLI 可以扫描具有多种构建工具的项目：

```bash
# 扫描包含 Go 模块和 npm 的项目
./cleansource-sca-cli --server-url https://your-server.com \
    --token your-token \
    --task-dir /path/to/multi-language-project
```

### 项目结构示例

**Go 项目:**
```
project/
├── go.mod
├── main.go
└── go.sum
```

**Node.js 项目:**
```
project/
├── package.json
├── package-lock.json
└── src/
```

**Gradle 项目:**
```
project/
├── build.gradle
├── settings.gradle
└── src/
```

**Python Pipenv 项目:**
```
project/
├── Pipfile
├── Pipfile.lock
└── src/
```

## 贡献

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 进行修改
4. 为新功能添加全面测试
5. 确保所有测试通过 (`go test ./...`)
6. 运行测试脚本 (`./run-tests.sh` 或 `run-tests.bat`)
7. 如需要更新文档
8. 提交更改 (`git commit -m 'Add amazing feature'`)
9. 推送到分支 (`git push origin feature/amazing-feature`)
10. 打开 Pull Request

### 开发指南

- 遵循 Go 编码标准和最佳实践
- 为所有新功能添加测试
- 为新功能更新文档
- 确保跨平台兼容性
- 使用有意义的提交消息
- 保持代码库清洁和文档完善
