---
name: dong
description: Windows/Linux system detection tool for hardware info (CPU, memory, disk, network), installed software versions (Go, Node, Python, Java, Git, Docker, etc.), system diagnostics, performance metrics, and comprehensive system reports. Works on Windows 10/11 and Linux (Debian/Ubuntu/Fedora/Arch/Deepin).
---

Dong (洞玄) is a cross-platform system information detection tool. It collects hardware details, software versions, system diagnostics, and performance metrics.

## Binary Location

Pre-compiled binaries are included for both platforms:
- Linux: `bin/dong`
- Windows: `bin/dong.exe`

No compilation required - both binaries are ready to use.

## Quick Commands

```bash
# Full scan
./dong -cli -all -pretty

# Fast scan (skip expensive checks)
./dong -cli -all -fast -pretty

# Hardware only
./dong -cli -hardware -pretty

# Software only
./dong -cli -software -pretty

# Advanced diagnostics
./dong -cli -advanced -pretty

# Save to file
./dong -cli -all -o report.json -pretty
```

## Detection Categories

### Hardware (`-hardware`)

**CPU** (`-cpu`)
- Model name, vendor
- Logical and physical cores
- Current frequency (MHz)
- Architecture

**Memory** (`-memory`)
- Total/used/available memory (GB)
- Virtual memory (swap)
- Memory load percentage
- Memory modules (manufacturer, capacity, speed)

**Disk** (`-disk`)
- Physical disks (model, type: SSD/HDD, size)
- Logical partitions (mount point, filesystem, usage)
- Disk health status (with `-deep-hw`)

**Network** (`-network`)
- Network interfaces
- IPv4 addresses, MAC addresses
- Link speed, status

**OS** (`-os`)
- OS name and version
- Kernel version (Linux)
- Architecture
- Hostname, username

### Software (`-software`)

Detects installed development tools:
- Go (version)
- Node.js (version)
- Python (version)
- Java (version)
- Git (version)
- Docker (version)
- kubectl (version, client info)
- .NET (version)

### Advanced Diagnostics (`-advanced`)

**Hardware Health** (`hardware_health`)
- Battery status and capacity (laptops)
- CPU temperature
- GPU information
- Motherboard vendor/name
- BIOS vendor/version
- Disk SMART data (with `-deep-hw`, requires root/admin)

**System Diagnostics** (`system_diagnostics`)
- System uptime
- Boot time
- Logged in users
- Failed services (systemd/Windows services)

**Network Diagnostics** (`network_diagnostics`)
- Default gateway
- DNS servers
- Gateway connectivity test
- External connectivity test
- DNS resolution test

**Driver Diagnostics** (`driver_diagnostics`)
- Loaded kernel modules (Linux)
- Problematic devices (Windows)
- Kernel errors (Linux, requires permissions)

**Performance Diagnostics** (`performance_diagnostics`)
- CPU usage percentage
- Memory usage percentage
- Load averages (Linux)
- Top CPU-consuming processes
- Top memory-consuming processes

**Software Inventory** (`software_inventory`)
- Installed packages (up to 300)
- Package manager detection (dpkg/rpm/pacman on Linux, registry on Windows)

## Command Flags

| Flag | Description |
|------|-------------|
| `-all` | Run all detection (default if no other flags) |
| `-hardware` | Detect hardware info only |
| `-software` | Detect software info only |
| `-cpu` | CPU info only |
| `-memory` | Memory info only |
| `-disk` | Disk info only |
| `-network` | Network info only |
| `-os` | OS info only |
| `-fast` | Fast mode (skip advanced diagnostics) |
| `-advanced` | Advanced diagnostics (auto-enabled in full scan unless `-fast`) |
| `-deep-hw` | Deep hardware health (SMART, requires root/admin) |
| `-o <name>` | Output to file (saved to `reports/` directory) |
| `-pretty` | Pretty print JSON output |
| `-cli` | Force CLI mode |
| `-v` | Show version |

## Usage Examples

### Hardware Queries

```
User: "What's my CPU?"
-> ./dong -cli -cpu -pretty

User: "How much memory do I have?"
-> ./dong -cli -memory -pretty

User: "Check my disk space"
-> ./dong -cli -disk -pretty

User: "What's my network configuration?"
-> ./dong -cli -network -pretty
```

### Software Queries

```
User: "Do I have Python installed?"
-> ./dong -cli -software -pretty

User: "What version of Go do I have?"
-> ./dong -cli -software -pretty | grep -A2 go

User: "Check my development tools"
-> ./dong -cli -software -pretty
```

### System Diagnostics

```
User: "Check my system health"
-> ./dong -cli -all -pretty

User: "Diagnose hardware issues"
-> ./dong -cli -advanced -deep-hw -pretty

User: "What's using my CPU?"
-> ./dong -cli -advanced -pretty

User: "Check network connectivity"
-> ./dong -cli -advanced -pretty
```

### Reports

```
User: "Generate a system report"
-> ./dong -cli -all -o system_report.json -pretty

User: "Quick system scan"
-> ./dong -cli -all -fast -o quick_scan.json -pretty
```

## Output Format

All output is JSON with the following structure:

```json
{
  "timestamp": 1234567890,
  "hostname": "computer-name",
  "go_version": "go1.21.0",
  "runtime": "linux/amd64",
  "hardware": {
    "cpu": { ... },
    "memory": { ... },
    "disk": { ... },
    "network": [ ... ],
    "os": { ... }
  },
  "software": {
    "go": { "installed": true, "version": "..." },
    "node": { ... },
    "python": { ... },
    ...
  },
  "advanced": {
    "hardware_health": { ... },
    "system_diagnostics": { ... },
    "network_diagnostics": { ... },
    "driver_diagnostics": { ... },
    "performance_diagnostics": { ... },
    "software_inventory": { ... }
  }
}
```

## Platform Notes

### Linux
- Binary: `bin/dong` (pre-compiled, ready to use)
- Requires `/proc` filesystem for most detections
- Some features need root (SMART, dmidecode)
- Supports dpkg/rpm/pacman package managers
- Uses commands: `lscpu`, `lsblk`, `df`, `ip`, `lspci`

### Windows
- Binary: `bin/dong.exe` (pre-compiled, ready to use)
- Requires Windows 10/11
- Uses WMI/CIM for hardware detection
- Uses PowerShell for advanced features
- Requires admin for SMART and some WMI queries

## Permissions

- **Standard user**: Most features work
- **Root/Admin required for**:
  - Disk SMART health (`-deep-hw`)
  - Memory module details (dmidecode on Linux)
  - Some WMI queries (Windows)

## File Structure

```
skill/
├── SKILL.md           # This file
├── README.md          # Installation guide
├── LICENSE.txt        # MIT License
├── install.sh         # Linux/macOS installer
├── install.bat        # Windows installer
└── bin/
    ├── dong           # Linux binary (pre-compiled)
    └── dong.exe       # Windows binary (pre-compiled)
```

Both binaries are included - no compilation or Go environment required.
