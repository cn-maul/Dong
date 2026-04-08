# Dong Skill - File Reference

This document describes all files included in the Dong skill package.

## Core Files (Required)

| File | Purpose | Required |
|------|---------|----------|
| `SKILL.md` | Skill definition for Claude Code | ✓ Yes |
| `bin/dong` | Linux binary executable (pre-compiled) | ✓ Yes (Linux) |
| `bin/dong.exe` | Windows binary executable (pre-compiled) | ✓ Yes (Windows) |
| `LICENSE.txt` | MIT License | ✓ Yes |

## Installation Files

| File | Purpose |
|------|---------|
| `install.sh` | Linux/macOS installer script |
| `install.bat` | Windows installer script |
| `INSTALL.md` | Detailed installation guide |

## Build Files

| File | Purpose |
|------|---------|
| `build.sh` | Build script for Linux/macOS |
| `build.bat` | Build script for Windows |

## Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Main documentation and overview |
| `QUICKREF.md` | Quick reference for common commands |
| `EXAMPLES.md` | Detailed usage examples |
| `CHANGELOG.md` | Version history and changes |
| `FILES.md` | This file |

## Test Files

| File | Purpose |
|------|---------|
| `test.sh` | Test suite for Linux/macOS |
| `test.bat` | Test suite for Windows |

## File Structure

```
skill/
├── SKILL.md              # Skill definition (required by Claude Code)
├── README.md             # Main documentation
├── LICENSE.txt           # MIT License
├── INSTALL.md            # Installation guide
├── QUICKREF.md           # Quick reference
├── EXAMPLES.md           # Usage examples
├── CHANGELOG.md          # Version history
├── FILES.md              # This file
├── install.sh            # Linux installer
├── install.bat           # Windows installer
├── build.sh              # Linux build script
├── build.bat             # Windows build script
├── test.sh               # Linux test suite
├── test.bat              # Windows test suite
└── bin/
    ├── dong              # Linux binary (6.1M, pre-compiled)
    └── dong.exe          # Windows binary (6.3M, pre-compiled)
```

**Note**: Both binaries are included - no compilation required.

## Minimum Required Files

For the skill to function, you need at minimum:

```
skill/
├── SKILL.md              # Required by Claude Code
├── LICENSE.txt           # Required for legal
└── bin/
    └── dong              # The actual binary
```

## File Sizes

Approximate sizes:

- Linux binary: ~6MB
- Windows binary: ~6MB
- Documentation: ~50KB total
- Scripts: ~20KB total
- **Total package: ~13MB**

Both binaries are pre-compiled - no Go environment or compilation required.

## Platform-Specific Notes

### Linux
- Binary: `bin/dong`
- Install with: `./install.sh`
- Build with: `./build.sh`
- Test with: `./test.sh`

### Windows
- Binary: `bin/dong.exe`
- Install with: `.\install.bat`
- Build with: `.\build.bat`
- Test with: `.\test.bat`

### macOS
- Uses same files as Linux
- Binary: `bin/dong`
- Install with: `./install.sh`

## Generated Files

The following files are generated during runtime:

- `reports/*.json` - Generated when using `-o` flag
- `*.log` - Not currently used

## Removing Files

To reduce package size, you can remove platform-specific binary:

**Remove Linux binary (keep only Windows):**
```bash
rm bin/dong
```

**Remove Windows binary (keep only Linux):**
```bash
rm bin/dong.exe
```

You can also remove:

- `EXAMPLES.md` - Optional documentation
- `CHANGELOG.md` - Optional documentation
- `FILES.md` - This file
- `test.sh`, `test.bat` - Only needed for testing
- `build.sh`, `build.bat` - Only needed if modifying source

**Do NOT remove:**

- `SKILL.md` - Required for Claude Code
- `LICENSE.txt` - Required legally
- `bin/dong` or `bin/dong.exe` - Required binary
- `install.sh` or `install.bat` - Needed for installation
