---
source: https://opencode.ai/docs/formatters/
fetched: 2026-01-08
title: Formatters Configuration
---

# OpenCode Formatters Documentation

## Overview

OpenCode automatically applies language-specific code formatting after files are written or edited, ensuring generated code aligns with project style standards.

## Built-in Formatters

| Formatter | Languages/Files |
|-----------|-----------------|
| Prettier | JavaScript, TypeScript, HTML, CSS, Markdown, JSON, YAML |
| Biome | JavaScript, TypeScript, JSON |
| Rustfmt | Rust (.rs) |
| Gofmt | Go (.go) |
| Ruff | Python |
| uv | Python |

Each formatter requires either a command-line tool or specific configuration file to function.

## Automatic Formatting Process

OpenCode automatically:
1. Matches file extensions to enabled formatters
2. Executes the appropriate command
3. Applies changes without user intervention

## Configuration

Configure formatters via the `formatter` section in `opencode.json`.

### Basic Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "formatter": {
    "prettier": {
      "disabled": false,
      "command": ["prettier", "--write", "$FILE"]
    }
  }
}
```

### Disable All Formatters

```json
{
  "$schema": "https://opencode.ai/config.json",
  "formatter": false
}
```

### Disable Specific Formatter

```json
{
  "$schema": "https://opencode.ai/config.json",
  "formatter": {
    "prettier": {
      "disabled": true
    }
  }
}
```

## Custom Formatter Example

Define custom formatters with:
- Command array with `$FILE` placeholder
- Environment variables
- Target file extensions

### Deno Markdown Formatter

```json
{
  "$schema": "https://opencode.ai/config.json",
  "formatter": {
    "deno-md": {
      "command": ["deno", "fmt", "$FILE"],
      "extensions": [".md", ".markdown"],
      "env": {
        "DENO_NO_UPDATE_CHECK": "1"
      }
    }
  }
}
```

### Override Prettier

```json
{
  "$schema": "https://opencode.ai/config.json",
  "formatter": {
    "prettier": {
      "command": ["npx", "prettier", "--write", "--config", ".prettierrc", "$FILE"],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

## Formatter Options

| Option | Description |
|--------|-------------|
| `command` | Array of command and arguments, use `$FILE` for file path |
| `extensions` | Array of file extensions to match |
| `disabled` | Set to true to disable |
| `env` | Environment variables for the formatter process |

## Examples

### Python with Black

```json
{
  "formatter": {
    "black": {
      "command": ["black", "$FILE"],
      "extensions": [".py"]
    }
  }
}
```

### Multiple Formatters

```json
{
  "$schema": "https://opencode.ai/config.json",
  "formatter": {
    "prettier": {
      "command": ["prettier", "--write", "$FILE"],
      "extensions": [".js", ".ts", ".jsx", ".tsx", ".json"]
    },
    "rustfmt": {
      "command": ["rustfmt", "$FILE"],
      "extensions": [".rs"]
    },
    "gofmt": {
      "command": ["gofmt", "-w", "$FILE"],
      "extensions": [".go"]
    }
  }
}
```

## Best Practices

1. **Match project config**: Use the same formatter settings as your project
2. **Test locally**: Ensure formatters work before enabling in OpenCode
3. **Use project formatters**: Prefer project-installed formatters over global
4. **Check extensions**: Verify extension mappings match your file types
