# 洞玄（Dong）

跨平台本机检测工具，支持：
- CLI 扫描（适合 AI 工具/自动化）
- Web 可视化（手动查看）

## 支持平台

- **Linux** (Deepin/Ubuntu/Debian/Fedora/Arch 等)
- **Windows** (Windows 10/11)

## 快速开始

### Linux

```bash
# 编译
scripts/build-cli.sh

# 快速扫描
./dong -cli -all -fast -pretty

# 完整扫描
./dong -cli -all -pretty

# 保存报告
./dong -cli -all -o report.json -pretty
```

### Windows

```powershell
# 编译
scripts\build-cli.bat

# 快速扫描
.\dong.exe -cli -all -fast -pretty

# 完整扫描
.\dong.exe -cli -all -pretty

# 保存报告
.\dong.exe -cli -all -o report.json -pretty
```

## 脚本说明

### Linux (`scripts/`)

| 脚本 | 说明 |
|------|------|
| `build-cli.sh` | 编译 CLI 版本 `dong` |
| `build-web.sh` | 编译 Web 版本 `dong-web` |
| `run-fast.sh` | 快速扫描并输出 JSON |
| `run-full.sh` | 完整扫描并输出 JSON |

### Windows (`scripts/`)

| 脚本 | 说明 |
|------|------|
| `build-cli.bat` | 编译 CLI 版本 `dong.exe` |
| `build-web.bat` | 编译 Web 版本 `dong-web.exe` |
| `run-fast.bat` | 快速扫描并输出 JSON |
| `run-full.bat` | 完整扫描并输出 JSON |

## 参数说明

| 参数 | 说明 |
|------|------|
| `-all` | 全量检测（默认行为） |
| `-hardware` / `-software` | 只跑某一类 |
| `-cpu` / `-memory` / `-disk` / `-network` / `-os` | 硬件子项 |
| `-fast` | 快速模式（减少慢项） |
| `-advanced` | 进阶诊断（`-all` 且非 `-fast` 时默认执行） |
| `-deep-hw` | 深度硬件健康检测 |
| `-o <name>` | 写入报告文件（默认到 `reports/`） |
| `-pretty` | 格式化 JSON |
| `-web` | 启动 Web 服务 |
| `-web-addr` | Web 地址（默认 `127.0.0.1:18080`） |
| `-cli` | 强制 CLI 模式（对 Web 版二进制有效） |
| `-v` | 版本信息 |

## 输出结构

### 顶层字段

- `timestamp`: Unix 时间戳
- `hostname`: 主机名
- `go_version`: Go 运行时版本
- `runtime`: 平台 (linux/amd64, windows/amd64)

### hardware

| 字段 | 说明 |
|------|------|
| `cpu` | 型号、逻辑/物理核心、频率 |
| `memory` | 物理/虚拟内存、内存条信息 |
| `disk` | 物理磁盘 + 逻辑分区 |
| `network` | 网卡、IP、MAC |
| `os` | 系统信息 |

### software

已安装开发工具版本：Go、Node、Python、Java、Git、Docker、Kubectl、Dotnet

### advanced

| 字段 | 说明 |
|------|------|
| `hardware_health` | 电池、温度、GPU、主板、BIOS |
| `system_diagnostics` | 启动时间、登录用户、失败服务 |
| `network_diagnostics` | 网关、DNS、网络连通性 |
| `driver_diagnostics` | 内核模块、驱动错误 |
| `performance_diagnostics` | CPU/内存使用、资源占用进程 |
| `software_inventory` | 已安装软件包 |

## 运行要求

- Linux (内核 3.x+) 或 Windows 10/11
- Go 1.21+ (编译时)
- 部分检测需要 root/管理员权限 (SMART、dmidecode 等)

## 项目结构

```
Dong/
├── cmd/
│   ├── main.go              # 主入口
│   └── web/                 # Web 前端
├── detector/
│   ├── hardware/
│   │   ├── linux.go         # Linux 硬件检测
│   │   └── windows.go       # Windows 硬件检测
│   ├── software/
│   │   ├── linux.go         # Linux 软件检测
│   │   └── windows.go       # Windows 软件检测
│   └── advanced/
│       ├── linux.go         # Linux 进阶诊断
│       └── windows.go       # Windows 进阶诊断
├── scripts/                 # 构建脚本
└── skill/                   # Claude Code Skill
```

## 许可证

MIT License
