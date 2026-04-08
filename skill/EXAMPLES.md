# Dong Skill - Usage Examples

This document provides practical examples of using the Dong skill in Claude Code.

## Installation Verification

After installing, verify the skill works:

```bash
/dong -v
# Output: Dong v0.1.0 (Go go1.26.1)
```

## Hardware Queries

### CPU Information

```
User: What's my CPU?
Claude runs: ./dong -cli -cpu -pretty
```

Output shows:
- Model name (e.g., "11th Gen Intel Core i5-11500")
- Logical and physical cores
- Current frequency in MHz
- Vendor and architecture

### Memory Information

```
User: How much RAM do I have?
Claude runs: ./dong -cli -memory -pretty
```

Output shows:
- Total, used, and available memory in GB
- Swap/virtual memory
- Memory load percentage
- Memory module details (if available)

### Disk Information

```
User: Check my disk space
Claude runs: ./dong -cli -disk -pretty
```

Output shows:
- Physical disks with type (SSD/HDD)
- Logical partitions with mount points
- Capacity and usage for each partition
- Filesystem type

### Network Information

```
User: What's my IP address?
Claude runs: ./dong -cli -network -pretty
```

Output shows:
- Network interface names
- IPv4 addresses
- MAC addresses
- Interface status

### OS Information

```
User: What OS am I running?
Claude runs: ./dong -cli -os -pretty
```

Output shows:
- OS name and version
- Kernel version (Linux)
- Architecture
- Hostname

## Software Queries

### Development Tools

```
User: What development tools do I have installed?
Claude runs: ./dong -cli -software -pretty
```

Checks for:
- Go, Node.js, Python, Java
- Git, Docker, kubectl, .NET
- Shows version for each installed tool

### Specific Tool Version

```
User: What version of Python do I have?
Claude runs: ./dong -cli -software -pretty
# Then filters for Python
```

## System Diagnostics

### Quick System Check

```
User: Give me a quick system overview
Claude runs: ./dong -cli -all -fast -pretty
```

Scans hardware and software, skips advanced diagnostics.

### Full System Diagnosis

```
User: Diagnose my system
Claude runs: ./dong -cli -all -pretty
```

Includes:
- All hardware info
- All software info
- Advanced diagnostics (hardware health, performance, network)

### Hardware Health Check

```
User: Check my hardware health
Claude runs: ./dong -cli -advanced -pretty
```

Shows:
- Battery status (laptops)
- CPU temperature
- GPU information
- Motherboard/BIOS info
- Uptime and boot time

### Network Diagnostics

```
User: Test my network connectivity
Claude runs: ./dong -cli -advanced -pretty
# Focuses on network_diagnostics section
```

Tests:
- Gateway connectivity
- External connectivity
- DNS resolution
- Lists DNS servers

### Performance Analysis

```
User: What's using my system resources?
Claude runs: ./dong -cli -advanced -pretty
# Focuses on performance_diagnostics section
```

Shows:
- CPU and memory usage
- Load averages (Linux)
- Top CPU-consuming processes
- Top memory-consuming processes

## Report Generation

### Save Full System Report

```
User: Generate a system report
Claude runs: ./dong -cli -all -o system_report.json -pretty
```

Saves to `reports/system_report.json`

### Quick Scan Report

```
User: Create a quick system scan report
Claude runs: ./dong -cli -all -fast -o quick_scan.json -pretty
```

### Hardware Report Only

```
User: Generate a hardware report
Claude runs: ./dong -cli -hardware -o hardware.json -pretty
```

## Advanced Use Cases

### Deep Hardware Health (Requires Root)

```
User: Check my disk health with SMART data
Claude runs: sudo ./dong -cli -advanced -deep-hw -pretty
```

Shows:
- SMART disk health status
- Physical disk health
- Predicted failures

### Monitoring Script

Create a monitoring script:

```bash
#!/bin/bash
# Save as monitor.sh
while true; do
    echo "=== $(date) ==="
    ./dong -cli -cpu -memory -pretty | grep -A5 "cpu\|memory"
    sleep 60
done
```

### Compare Systems

Generate reports on multiple systems:

```bash
# System 1
./dong -cli -all -o system1.json -pretty

# System 2
./dong -cli -all -o system2.json -pretty

# Compare
diff system1.json system2.json
```

### Extract Specific Info

```bash
# Get just the CPU model
./dong -cli -cpu -pretty | jq '.hardware.cpu.model'

# Get memory usage percentage
./dong -cli -memory -pretty | jq '.hardware.memory.memory_load_percent'

# List all IP addresses
./dong -cli -network -pretty | jq '.hardware.network[].ipv4'

# Check if Docker is installed
./dong -cli -software -pretty | jq '.software.docker.installed'
```

## Integration with Other Tools

### With jq (JSON processor)

```bash
# Pretty print specific section
./dong -cli -all -pretty | jq '.hardware'

# Extract multiple values
./dong -cli -all -pretty | jq '{cpu: .hardware.cpu.model, memory: .hardware.memory.total_gb}'

# Find processes using most memory
./dong -cli -advanced -pretty | jq '.advanced.performance_diagnostics.top_memory_processes'
```

### With grep

```bash
# Search for specific hardware
./dong -cli -all -pretty | grep -i "intel"

# Find error messages
./dong -cli -advanced -pretty | grep -i "error\|failed"
```

### With Python

```python
import json
import subprocess

# Run dong and parse output
result = subprocess.run(['./dong', '-cli', '-all', '-pretty'], capture_output=True, text=True)
data = json.loads(result.stdout)

# Access specific info
print(f"CPU: {data['hardware']['cpu']['model']}")
print(f"Memory: {data['hardware']['memory']['total_gb']} GB")
print(f"Go installed: {data['software']['go']['installed']}")
```

## Troubleshooting

### Permission Denied

```bash
# For deep hardware scans
sudo ./dong -cli -advanced -deep-hw -pretty
```

### Missing Dependencies

```bash
# Install smartmontools for SMART data (Linux)
sudo apt install smartmontools  # Debian/Ubuntu
sudo dnf install smartmontools  # Fedora
```

### Binary Not Found

```bash
# Build from source
cd skill
./build.sh
```

## Tips

1. **Use `-fast` for quick checks**: Skip advanced diagnostics when you only need basic info
2. **Use `-o` for reports**: Save output to files for later analysis or comparison
3. **Combine with jq**: Use jq to filter and format specific information
4. **Root for deep scans**: Use sudo/admin for SMART data and hardware details
5. **Regular monitoring**: Set up cron jobs or scripts to monitor system health over time
