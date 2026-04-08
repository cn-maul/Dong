# 🚀 快速开始 - Dong Skill

欢迎使用 Dong 系统检测技能！

## ⚡ 立即开始

### 1. 安装 (2分钟)

**Linux/macOS:**
```bash
cd skill
./install.sh
```

**Windows:**
```powershell
cd skill
.\install.bat
```

### 2. 重启 Claude Code

安装后重启 Claude Code 以加载技能。

### 3. 测试

```
/dong -v
/dong -cli -cpu -pretty
```

## 📦 包含内容

✅ **预编译二进制文件** - 无需编译，开箱即用
- Linux: `bin/dong` (6.1MB)
- Windows: `bin/dong.exe` (6.3MB)

✅ **完整文档**
- `README.md` - 主要文档
- `QUICKREF.md` - 快速参考
- `EXAMPLES.md` - 使用示例
- `INSTALL.md` - 安装指南

✅ **自动化脚本**
- `install.sh` / `install.bat` - 安装脚本
- `test.sh` / `test.bat` - 测试脚本

## 🎯 常用命令

```bash
# 查看版本
./dong -v

# CPU 信息
./dong -cli -cpu -pretty

# 内存信息
./dong -cli -memory -pretty

# 磁盘信息
./dong -cli -disk -pretty

# 全部硬件
./dong -cli -hardware -pretty

# 已安装软件
./dong -cli -software -pretty

# 完整扫描
./dong -cli -all -pretty

# 快速扫描
./dong -cli -all -fast -pretty

# 保存报告
./dong -cli -all -o report.json -pretty
```

## 📚 详细文档

- **安装问题？** 查看 `INSTALL.md`
- **命令参考？** 查看 `QUICKREF.md`
- **使用示例？** 查看 `EXAMPLES.md`
- **文件说明？** 查看 `FILES.md`

## 🌐 平台支持

| 平台 | 二进制 | 状态 |
|------|--------|------|
| Linux (x64) | `bin/dong` | ✅ 已包含 |
| Windows (x64) | `bin/dong.exe` | ✅ 已包含 |
| macOS (x64) | `bin/dong` | ✅ 已包含 |

## ❓ 需要帮助？

1. 运行测试: `./test.sh` (Linux) 或 `.\test.bat` (Windows)
2. 查看文档: `README.md`, `INSTALL.md`, `EXAMPLES.md`
3. 验证二进制: `./bin/dong -v`

## 📄 许可证

MIT License - 详见 `LICENSE.txt`

---

**无需 Go 环境，无需编译，直接使用！**
