# Dong Skill - System Detection for Claude Code

洞玄 (Dong) is a cross-platform system information detection skill for Claude Code. It provides comprehensive hardware, software, and system diagnostics.

## Features

- **Cross-platform**: Works on Linux (Debian/Ubuntu/Fedora/Arch/Deepin) and Windows 10/11
- **Pre-compiled binaries**: No compilation or Go environment required
- **Hardware Detection**: CPU, memory, disk, network, OS details
- **Software Detection**: Go, Node.js, Python, Java, Git, Docker, kubectl, .NET versions
- **Advanced Diagnostics**: Hardware health, system diagnostics, network tests, performance metrics
- **JSON Output**: Clean, parseable JSON format with optional pretty-printing

## Installation

### Option 1: Copy to Claude Code Skills Directory (Recommended)

**Linux/macOS:**
```bash
cp -r skill ~/.claude/skills/dong
```

**Windows (PowerShell):**
```powershell
Copy-Item -Recurse skill "$env:USERPROFILE\.claude\skills\dong"
```

### Option 2: Use Install Script

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

### Option 3: Symbolic Link

**Linux/macOS:**
```bash
ln -s $(pwd)/skill ~/.claude/skills/dong
```

**Windows (PowerShell, requires Admin):**
```powershell
New-Item -ItemType SymbolicLink -Path "$env:USERPROFILE\.claude\skills\dong" -Target "$(Get-Location)\skill"
```

## Verification

After installation, restart Claude Code and test:

```
/dong -v
```

You should see:
```
Dong v0.1.0 (Go go1.21.0)
```

## Quick Start

Both Linux and Windows binaries are pre-compiled and ready to use - no Go environment required.

Once installed, use the skill in Claude Code:

```
# Ask Claude to check your system
User: What's my CPU?
Claude: [runs ./dong -cli -cpu -pretty]

User: Check my memory and disk
Claude: [runs ./dong -cli -memory -disk -pretty]

User: Generate a system report
Claude: [runs ./dong -cli -all -o report.json -pretty]
```

## Building from Source (Optional)

The skill includes pre-compiled binaries for both Linux and Windows. You only need to rebuild if you modified the source code.

**Linux/macOS:**
```bash
cd skill
./build.sh
```

**Windows:**
```powershell
cd skill
.\build.bat
```

The compiled binary will be placed in `bin/`.

## Usage Examples

### Hardware Information

```bash
# CPU only
./dong -cli -cpu -pretty

# Memory only
./dong -cli -memory -pretty

# Disk information
./dong -cli -disk -pretty

# Network interfaces
./dong -cli -network -pretty

# All hardware
./dong -cli -hardware -pretty
```

### Software Information

```bash
# Check installed development tools
./dong -cli -software -pretty
```

### System Diagnostics

```bash
# Full scan with advanced diagnostics
./dong -cli -all -pretty

# Fast scan (skip advanced)
./dong -cli -all -fast -pretty

# Deep hardware health (requires root/admin)
sudo ./dong -cli -advanced -deep-hw -pretty
```

### Save Reports

```bash
# Save full report
./dong -cli -all -o system_report.json -pretty

# Quick scan to file
./dong -cli -all -fast -o quick_scan.json -pretty
```

## Output Format

All output is JSON. Example:

```json
{
  "timestamp": 1234567890,
  "hostname": "my-computer",
  "go_version": "go1.21.0",
  "runtime": "linux/amd64",
  "hardware": {
    "cpu": {
      "model": "11th Gen Intel(R) Core(TM) i5-11500 @ 2.70GHz",
      "cores_logical": 12,
      "cores_physical": 6,
      "frequency_mhz": 4100
    },
    "memory": {
      "total_gb": 15.5,
      "available_gb": 8.2,
      "used_gb": 7.3,
      "memory_load_percent": 47
    },
    ...
  },
  "software": {
    "go": { "installed": true, "version": "go version go1.21.0 linux/amd64" },
    "node": { "installed": true, "version": "v18.17.0" },
    ...
  },
  "advanced": {
    ...
  }
}
```

## Platform Support

### Linux
- **Tested on**: Deepin 25, Ubuntu 22.04, Debian 12, Fedora 39, Arch Linux
- **Requirements**: `/proc` filesystem, standard Linux utilities
- **Root required for**: SMART disk health, memory module details

### Windows
- **Tested on**: Windows 10, Windows 11
- **Requirements**: PowerShell 5.1+
- **Admin required for**: SMART disk health, some WMI queries

## Troubleshooting

### Binary not found

```
Error: bin/dong not found
```

Solution: Build the binary first:
```bash
./build.sh   # Linux
.\build.bat  # Windows
```

### Permission denied

```
Error: permission denied
```

Solution: Some features require root/admin:
```bash
sudo ./dong -cli -advanced -deep-hw -pretty
```

### Command not found in Claude Code

Solution: Restart Claude Code after installation.

## File Structure

```
skill/
├── SKILL.md           # Skill definition for Claude Code
├── README.md          # This file
├── LICENSE.txt        # MIT License
├── install.sh         # Linux/macOS installer
├── install.bat        # Windows installer
├── build.sh           # Build script (Linux/macOS)
├── build.bat          # Build script (Windows)
└── bin/
    ├── dong           # Linux binary
    └── dong.exe       # Windows binary
```

## Version History

- **v0.1.0**: Initial release
  - Cross-platform support (Linux + Windows)
  - Hardware detection
  - Software detection
  - Advanced diagnostics

## License

MIT License - see LICENSE.txt

## Contributing

Source code available at the Dong project repository.

## Support

For issues or feature requests, please use the project's issue tracker.
