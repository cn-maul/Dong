# 洞玄（Dong）

Windows 本机检测工具，支持：
- CLI 扫描（适合 AI 工具/自动化）
- Web 可视化（手动查看）

## 当前脚本（`scripts/`）
- `build-cli.bat`：编译 CLI 版 `dong.exe`
- `build-web.bat`：编译 Web 版 `dong-web.exe`（前端嵌入）
- `run-fast.bat`：快扫并输出 JSON 报告
- `run-full.bat`：全扫并输出 JSON 报告

## 典型用法
1. 给 AI 工具走 CLI
```powershell
scripts\build-cli.bat
.\dong.exe -cli -all -fast -pretty
```

2. 你自己手动看前端页面
```powershell
scripts\build-web.bat
.\dong-web.exe -all
```
打开浏览器：`http://127.0.0.1:18080`

3. 生成 JSON 报告
```powershell
scripts\run-fast.bat
scripts\run-full.bat
```
输出目录：`reports/`

## 参数说明
- `-all`：全量检测（默认行为）
- `-hardware` / `-software`：只跑某一类
- `-cpu` / `-memory` / `-disk` / `-network` / `-os`：硬件子项
- `-fast`：快速模式（减少慢项）
- `-advanced`：进阶诊断（`-all` 且非 `-fast` 时默认执行）
- `-deep-hw`：深度硬件健康检测
- `-o <name>`：写入报告文件（默认到 `reports/`）
- `-pretty`：格式化 JSON
- `-web`：启动 Web 服务
- `-web-addr`：Web 地址（默认 `127.0.0.1:18080`）
- `-cli`：强制 CLI 模式（对 Web 版二进制有效）
- `-v`：版本信息

## 控制台日志
扫描过程会在控制台输出中文阶段日志：
- 开始
- 硬件检测
- 软件检测
- 进阶诊断
- 完成与耗时

## Web 刷新机制
前端“刷新数据”按钮会触发后端重新扫描（真实刷新），不是读旧缓存。

## 当前输出结构（摘要）
- 顶层：`timestamp`、`hostname`、`go_version`、`runtime`、`hardware`、`software`、`advanced`
- `hardware`：
  - `cpu`（型号、逻辑/物理核心、频率）
  - `memory`（物理/虚拟内存 GB、内存条信息）
  - `disk`（`physical_disks` + `logical_partitions`）
  - `network` + `network_physical_adapters`
  - `os`
- `software`：`go/node/python/java/git/docker/kubectl/dotnet`（安装状态与版本）
- `advanced`：
  - `hardware_health`（battery/bios/baseboard/gpu）
  - `system_diagnostics`
  - `network_diagnostics`
  - `driver_diagnostics`
  - `performance_diagnostics`
  - `software_inventory`（已安装软件）

## 运行要求
- Windows 10 / Windows 11
- Go 1.26+
