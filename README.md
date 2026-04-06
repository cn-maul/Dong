# Dong

一个用于 **Windows 10 / Windows 11** 的本地检测工具，快速输出当前机器的硬件配置与软件环境信息，适合维修、巡检、装机后核验等场景。

## 功能概览

- 硬件检测：CPU、内存、磁盘、网卡、系统版本
- 软件检测：Go、Node/NPM、Python、Java、Git、Docker、kubectl、.NET
- 环境信息：常用环境变量
- 进程信息：前 10 个进程（来自 `tasklist`）
- 性能模式：
  - 默认全量模式（更完整）
  - `-fast` 快扫模式（跳过较慢的软件深度检测）
  - `-advanced` 进阶诊断（系统故障、驱动、网络连通、软件清单等）

## 运行要求

- 操作系统：Windows 10 / Windows 11
- Go 版本：推荐 Go 1.26+

> 说明：本工具是本地执行，软件检测项依赖系统是否已安装对应命令（如 `go`、`node`、`docker` 等）。

## 快速开始

### 1) 开发环境直接运行

```powershell
go run ./cmd -all -pretty
```

### 2) 构建可执行文件

```powershell
go build -trimpath -ldflags "-s -w -buildid=" -o dong.exe ./cmd
```

### 3) 执行检测

```powershell
.\dong.exe -all -pretty
```

## 命令行参数

| 参数 | 说明 |
|---|---|
| `-all` | 执行全量检测（默认不指定参数时也会全量） |
| `-hardware` | 仅检测硬件信息 |
| `-software` | 仅检测软件环境 |
| `-cpu` | 仅检测 CPU |
| `-memory` | 仅检测内存 |
| `-disk` | 仅检测磁盘 |
| `-network` | 仅检测网络 |
| `-os` | 仅检测操作系统信息 |
| `-fast` | 快速模式：跳过较慢的软件深度检查（如 Docker 运行状态深度检测） |
| `-advanced` | 执行进阶诊断（默认在 `-all` 且未开启 `-fast` 时自动执行） |
| `-deep-hw` | 深度硬件健康检测（SMART/PhysicalDisk 健康） |
| `-o <文件名>` | 输出到文件（默认目录 `reports/`） |
| `-pretty` | 以格式化 JSON 输出 |
| `-v` | 显示版本 |

## 常用示例

### 全量检测（输出到终端）

```powershell
.\dong.exe -all -pretty
```

### 快速全量检测（现场快速判断）

```powershell
.\dong.exe -all -fast -pretty
```

### 全量 + 进阶诊断（默认全扫会自动包含）

```powershell
.\dong.exe -all -pretty
```

也可以显式指定：

```powershell
.\dong.exe -advanced -pretty
```

### 深度硬件健康（建议排查硬盘问题时开启）

```powershell
.\dong.exe -all -advanced -deep-hw -pretty
```

### 仅软件环境检测

```powershell
.\dong.exe -software -pretty
```

### 输出到文件

```powershell
.\dong.exe -all -o report.json -pretty
```

生成路径示例：

```text
reports/report.json
```

### 自定义报告目录

可通过环境变量 `DONG_REPORTS_DIR` 指定输出目录：

```powershell
$env:DONG_REPORTS_DIR="D:\DongReports"
.\dong.exe -all -o my-report.json -pretty
```

## 输出结构说明（JSON）

顶层字段：

- `timestamp`：Unix 时间戳
- `hostname`：主机名
- `go_version`：构建/运行时 Go 版本信息
- `runtime`：运行平台（如 `windows/amd64`）
- `hardware`：硬件检测结果
- `software`：软件环境检测结果
- `advanced`：进阶诊断结果（全扫自动启用）

`hardware` 常见子字段：

- `cpu`：逻辑核数、物理核数、CPU 型号
- `memory`：总内存、可用内存、已用内存、内存负载
- `disk`：盘符、容量、文件系统
- `network`：网卡名、MAC、IPv4
- `os`：Windows 版本、构建号、用户名、计算机名等

`software` 常见子字段：

- `go` / `node` / `python` / `java` / `git` / `docker` / `kubectl` / `dotnet`
- `env`：环境变量信息
- `processes`：进程列表（名称、内存占用文本）

## 性能建议

- 日常维修建议优先使用：
  - `.\dong.exe -all -fast`
- 需要完整软件细节时使用：
  - `.\dong.exe -all`
- 构建发布建议始终使用：
  - `go build -trimpath -ldflags "-s -w -buildid=" -o dong.exe ./cmd`

## 故障排查

- 提示某软件未安装：
  - 检查该命令是否可在 PowerShell 直接执行（例如 `go version`）
- 输出文件未生成：
  - 检查 `-o` 文件名是否正确
  - 检查 `DONG_REPORTS_DIR` 是否存在写权限
- 输出结果乱码/不完整：
  - 优先使用 `-pretty` 便于阅读
  - 用 `-o` 输出到文件再查看

## 版本信息

查看版本：

```powershell
.\dong.exe -v
```

