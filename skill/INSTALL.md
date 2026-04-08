# Installation Guide

This guide walks you through installing the Dong skill for Claude Code.

## Prerequisites

- Claude Code CLI installed
- Linux (kernel 3.x+) or Windows 10/11

**No Go environment required** - binaries are pre-compiled for both platforms.

## Quick Install

### Linux/macOS

```bash
# Navigate to skill directory
cd skill

# Run installer
./install.sh

# Restart Claude Code
```

### Windows

```powershell
# Navigate to skill directory
cd skill

# Run installer
.\install.bat

# Restart Claude Code
```

## Manual Install

If the install script doesn't work, you can manually copy the skill:

### Linux/macOS

```bash
# Create skills directory if it doesn't exist
mkdir -p ~/.claude/skills

# Copy skill directory
cp -r skill ~/.claude/skills/dong

# Or create symbolic link
ln -s $(pwd)/skill ~/.claude/skills/dong
```

### Windows

```powershell
# Create skills directory if it doesn't exist
if (!(Test-Path "$env:USERPROFILE\.claude\skills")) {
    New-Item -ItemType Directory -Path "$env:USERPROFILE\.claude\skills"
}

# Copy skill directory
Copy-Item -Recurse skill "$env:USERPROFILE\.claude\skills\dong"
```

## Verification

After installation, restart Claude Code and verify:

```
/dong -v
```

Expected output:
```
Dong v0.1.0 (Go go1.21.0)
```

## Test the Skill

Run a quick test:

```
/dong -cli -cpu -pretty
```

You should see CPU information in JSON format.

## Building from Source (Optional)

Binaries are pre-compiled for both Linux and Windows. You only need to rebuild if you modified the source code.

### Linux/macOS

```bash
cd skill
./build.sh
```

### Windows

```powershell
cd skill
.\build.bat
```

## Troubleshooting

### "Binary not found" Error

The skill includes pre-compiled binaries for both platforms:
- Linux: `bin/dong`
- Windows: `bin/dong.exe`

If binaries are missing, download the skill package again.

### Permission Denied (Linux)

Some features require root:

```bash
sudo ./dong -cli -advanced -deep-hw -pretty
```

### Skill Not Found in Claude Code

1. Ensure the skill is installed in `~/.claude/skills/dong` (Linux/macOS) or `%USERPROFILE%\.claude\skills\dong` (Windows)
2. Restart Claude Code
3. Check the directory structure:

```bash
ls -la ~/.claude/skills/dong/  # Linux
dir %USERPROFILE%\.claude\skills\dong\  # Windows
```

Should contain:
- `SKILL.md` (required)
- `bin/dong` or `bin/dong.exe`
- Other documentation files

### "Command not found" in Claude Code

The skill might not be properly registered:

1. Check SKILL.md exists and has valid frontmatter
2. Verify the binary is executable: `chmod +x bin/dong`
3. Restart Claude Code

### Outdated Binary

If you updated the source code, rebuild (requires Go 1.21+):

```bash
./build.sh   # Linux - overwrites old binary
.\build.bat  # Windows - overwrites old binary
```

## Uninstallation

To remove the skill:

### Linux/macOS

```bash
rm -rf ~/.claude/skills/dong
```

### Windows

```powershell
Remove-Item -Recurse -Force "$env:USERPROFILE\.claude\skills\dong"
```

## Next Steps

After installation:

1. Read `QUICKREF.md` for common commands
2. Check `EXAMPLES.md` for usage examples
3. Run `./test.sh` to verify all features

## Support

For issues or questions:

1. Check the troubleshooting section above
2. Review `README.md` and `SKILL.md`
3. Run the test suite: `./test.sh`
4. Check binary: `./bin/dong -v`
