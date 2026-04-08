# Quick Reference

## Common Commands

### Hardware
```bash
./dong -cli -cpu -pretty          # CPU info
./dong -cli -memory -pretty       # Memory info
./dong -cli -disk -pretty         # Disk info
./dong -cli -network -pretty      # Network info
./dong -cli -os -pretty           # OS info
./dong -cli -hardware -pretty     # All hardware
```

### Software
```bash
./dong -cli -software -pretty     # Installed dev tools
```

### Diagnostics
```bash
./dong -cli -advanced -pretty     # Advanced diagnostics
./dong -cli -all -pretty          # Full scan + diagnostics
./dong -cli -all -fast -pretty    # Fast scan (no diagnostics)
```

### Reports
```bash
./dong -cli -all -o report.json -pretty    # Save full report
./dong -cli -all -fast -o quick.json       # Save quick report
```

## Output Structure

```json
{
  "timestamp": 1234567890,
  "hostname": "computer-name",
  "go_version": "go1.21.0",
  "runtime": "linux/amd64",
  "hardware": {
    "cpu": { "model": "...", "cores_logical": 12, ... },
    "memory": { "total_gb": 15.5, ... },
    "disk": { "physical_disks": [...], "logical_partitions": [...] },
    "network": [{ "name": "...", "ipv4": "..." }],
    "os": { "display_name": "...", "kernel": "..." }
  },
  "software": {
    "go": { "installed": true, "version": "..." },
    "node": { ... },
    ...
  },
  "advanced": {
    "hardware_health": { ... },
    "system_diagnostics": { ... },
    "network_diagnostics": { ... },
    "performance_diagnostics": { ... },
    "software_inventory": { ... }
  }
}
```

## Flags Summary

| Flag | Purpose |
|------|---------|
| `-all` | Everything |
| `-fast` | Skip advanced |
| `-pretty` | Format JSON |
| `-o FILE` | Save to file |
| `-deep-hw` | Deep hardware (needs root) |
| `-v` | Version |

## Platform Specifics

### Linux
- Uses `/proc` filesystem
- Supports dpkg/rpm/pacman
- Root needed for SMART, dmidecode

### Windows
- Uses WMI/PowerShell
- Admin needed for SMART, some WMI

## Example Queries

**"What's my CPU?"**
```bash
./dong -cli -cpu -pretty
```

**"How much RAM do I have?"**
```bash
./dong -cli -memory -pretty
```

**"Check my disk space"**
```bash
./dong -cli -disk -pretty
```

**"What software is installed?"**
```bash
./dong -cli -software -pretty
```

**"System health check"**
```bash
./dong -cli -all -pretty
```

**"Diagnose network issues"**
```bash
./dong -cli -advanced -pretty
```
