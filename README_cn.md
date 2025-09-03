# CleanSource SCA CLI

[English](./README.md) | 中文

[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/cleansource-sca-cli)](https://goreportcard.com/report/github.com/craftslab/cleansource-sca-cli)
[![License](https://img.shields.io/github/license/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/cleansource-sca-cli.svg)](https://github.com/craftslab/cleansource-sca-cli/tags)

A Go implementation of the CleanSource SCA build scanner.

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
- ✅ 文件压缩和归档
- ✅ REST API 客户端用于服务器通信
- ✅ 大型代码库并发处理
- 🚧 Gradle 依赖扫描 (开发中)
- 🚧 npm/Node.js 依赖扫描 (开发中)
- 🚧 Go 模块依赖扫描 (开发中)
- 🚧 Pipenv 依赖扫描 (开发中)

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
| Maven | ✅ 完成 | 完整的依赖树分析 |
| pip | ✅ 完成 | Requirements.txt 和已安装包 |
| Gradle | 🚧 部分 | 基本检测，扫描开发中 |
| npm | 🚧 部分 | 基本检测，扫描开发中 |
| Go Modules | 🚧 部分 | 基本检测，扫描开发中 |
| Pipenv | 🚧 部分 | 基本检测，扫描开发中 |

## 开发

### 运行测试

```bash
go test ./...
```

### 构建

```bash
# 为当前平台构建
go build -o cleansource-sca-cli main.go

# 带优化的构建
go build -ldflags="-s -w" -o cleansource-sca-cli main.go
```

### 添加新的构建工具

要添加对新构建工具的支持：

1. 在 `pkg/buildtools/` 中创建新的扫描器
2. 实现 `Scannable` 接口：
   - `ExeFind()`: 查找构建工具可执行文件
   - `FileFind()`: 检查所需文件
   - `ScanExecute()`: 执行依赖扫描
3. 在 `pkg/buildtools/scanner.go` 中添加检测逻辑
4. 使用示例项目进行测试

## 贡献

1. Fork 仓库
2. 创建功能分支
3. 进行修改
4. 添加测试
5. 提交 Pull Request
