# Changelog

All notable changes to the Dong skill will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-04-07

### Added
- Initial release
- Cross-platform support (Linux + Windows)
- Hardware detection
  - CPU: model, cores, frequency, architecture
  - Memory: capacity, usage, modules
  - Disk: physical disks, logical partitions, health
  - Network: interfaces, IP, MAC
  - OS: name, version, kernel, hostname
- Software detection
  - Go, Node.js, Python, Java, Git
  - Docker, kubectl, .NET
- Advanced diagnostics
  - Hardware health (battery, temperature, GPU, BIOS)
  - System diagnostics (uptime, users, services)
  - Network diagnostics (gateway, DNS, connectivity)
  - Driver diagnostics (kernel modules, errors)
  - Performance diagnostics (CPU, memory, processes)
  - Software inventory (installed packages)
- JSON output with pretty-printing option
- Fast mode for quick scans
- Deep hardware mode for SMART disk health
- Report generation with file output
- Install scripts for Linux and Windows
- Build scripts for compiling from source
- Test suite for verification

### Platform Support
- Linux: Deepin 25, Ubuntu 22.04+, Debian 12+, Fedora 39+, Arch Linux
- Windows: Windows 10, Windows 11

### Known Limitations
- Some features require root/admin privileges
- SMART disk health requires smartmontools on Linux
- Memory module details require dmidecode on Linux
- Some WMI queries require admin on Windows

[0.1.0]: https://github.com/yourusername/dong/releases/tag/v0.1.0
