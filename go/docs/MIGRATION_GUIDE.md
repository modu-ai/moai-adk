# Python to Go Migration Guide

This guide helps you migrate your MoAI-ADK installation from Python to Go implementation.

## Overview

MoAI-ADK now provides a native Go implementation that offers better performance and easier distribution. This guide walks you through the migration process.

## Prerequisites

- Go 1.23+ installed (if building from source)
- Existing Python-based MoAI-ADK installation
- `.claude/settings.json` file in your project

## Installation Methods

### Method 1: go install (Recommended for Go users)

```bash
go install github.com/anthropics/moai-adk-go@latest
```

### Method 2: GitHub Releases

Download the appropriate binary for your platform from [GitHub Releases](https://github.com/anthropics/moai-adk-go/releases/latest).

```bash
# macOS ARM64 (Apple Silicon)
curl -sSL https://github.com/anthropics/moai-adk-go/releases/latest/download/moai-adk-darwin-arm64 -o moai-adk
chmod +x moai-adk
sudo mv moai-adk /usr/local/bin/
```

### Method 3: Install Script

```bash
curl -sSL https://raw.githubusercontent.com/anthropics/moai-adk-go/main/scripts/install.sh | bash
```

### Method 4: Homebrew (macOS)

```bash
brew tap anthropics/tap
brew install moai-adk
```

## Migration Process

### Step 1: Verify Installation

```bash
moai-adk version
```

Expected output:
```
moai-adk version 1.0.0
commit: <commit-hash>
built at: <build-date>
```

### Step 2: Check Current Configuration

```bash
moai-adk migrate --dry-run
```

This shows what changes will be made without applying them.

### Step 3: Perform Migration

```bash
moai-adk migrate
```

This command:
- Creates a backup of your `.claude/settings.json`
- Converts all hook commands to Go format
- Preserves all non-hook settings
- Creates a log at `.moai/logs/hook-implementation.log`

### Step 4: Verify Migration

Test your hooks to ensure they work correctly:
```bash
# Run a test command
moai-adk hook session-start --project-dir "$CLAUDE_PROJECT_DIR"
```

### Step 5: Rollback (if needed)

If you encounter issues:
```bash
moai-adk migrate --rollback
```

## Dual-Mode Operation

During the transition period, both Python and Go implementations can coexist:

### Automatic Detection

When you run `moai-adk init`, it automatically detects which implementation to use:

1. **Go binary found**: Uses Go implementation
2. **Go binary not found**: Falls back to Python

### Manual Override

Force a specific implementation:

```bash
# Force Go implementation
moai-adk init --force-go --go-binary-path /path/to/moai-adk

# Force Python implementation
moai-adk init --force-python
```

## Compatibility Verification

The `verify-compatibility` tool compares Python and Go implementations:

```bash
go run cmd/verify-compatibility/main.go session_start
```

This verifies:
- JSON schema compatibility
- Exit code compatibility
- Error message compatibility

## Logging

Hook executions are logged at `.moai/logs/hook-implementation.log`:

```json
{"hook":"session_start","implementation":"go","binary_path":"/usr/local/bin/moai-adk","version":"1.0.0","timestamp":"2026-01-29T12:00:00Z","duration":"50ms","success":true}
```

Log rotation:
- Daily log files
- 7-day retention
- Automatic cleanup

## Troubleshooting

### Issue: Go binary not found

**Solution**: Ensure Go binary is in your PATH or specify full path:
```bash
export PATH="$PATH:$HOME/.local/bin"
```

### Issue: Hooks not executing

**Solution**: Check hook implementation log:
```bash
cat .moai/logs/hook-implementation.log
```

### Issue: Performance degradation

**Solution**: Verify you're using Go implementation:
```bash
grep "implementation" .moai/logs/hook-implementation.log | tail -1
```

## Python Deprecation Timeline

- **Go 1.0 release**: Python enters maintenance-only mode
- **6 months after Go 1.0**: Python deprecated
- **12 months after Go 1.0**: Python support removed

## Benefits of Go Implementation

1. **Performance**: 10-50x faster hook execution
2. **Distribution**: Single static binary, no dependencies
3. **Cross-platform**: Native support for macOS, Linux, Windows
4. **Memory**: Lower memory footprint
5. **Updates**: Built-in self-update command

## Support

For issues or questions:
- GitHub Issues: https://github.com/anthropics/moai-adk-go/issues
- Documentation: https://github.com/anthropics/moai-adk-go/blob/main/README.md

## Comparison: Python vs Go

| Feature | Python | Go |
|---------|--------|-----|
| Cold start | ~500ms | ~50ms |
| Hook execution | ~200ms | ~20ms |
| Dependencies | Python, uv, packages | None (static binary) |
| Distribution | pip install | Single binary |
| Memory | ~100MB | ~20MB |
| Platform support | Python platforms | Native binaries |

## Migration Checklist

- [ ] Install Go binary
- [ ] Verify installation with `moai-adk version`
- [ ] Run `moai-adk migrate --dry-run` to preview changes
- [ ] Run `moai-adk migrate` to apply migration
- [ ] Test hooks execution
- [ ] Verify logs at `.moai/logs/hook-implementation.log`
- [ ] Remove Python MoAI-ADK (optional, after verification)

## Advanced Usage

### Custom Go Binary Path

If Go binary is in a non-standard location:

```bash
moai-adk migrate --force-go --go-binary-path /custom/path/moai-adk
```

### Batch Migration

For multiple projects:

```bash
for project in project1 project2 project3; do
    cd $project
    moai-adk migrate
    cd ..
done
```

### Verification Script

Automated verification script:

```bash
#!/bin/bash
for hook in session_start pre_tool post_tool session_end; do
    echo "Testing $hook..."
    moai-adk hook $hook --project-dir "$CLAUDE_PROJECT_DIR"
done
```

---

**Last Updated**: 2026-01-29
**MoAI-ADK Version**: 1.0.0 (Go)
