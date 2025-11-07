# Language Detection Guide

## Overview

MoAI-ADK automatically detects project language from configuration files and package managers.

## Supported Languages

- Python (pyproject.toml, setup.py, requirements.txt)
- JavaScript/TypeScript (package.json)
- Go (go.mod)
- Rust (Cargo.toml)
- Ruby (Gemfile)
- PHP (composer.json)
- Java (pom.xml, build.gradle)

## Detection Mechanism

The language detection works by scanning the project root directory for standard configuration files in this order:

1. `pyproject.toml` → Python
2. `package.json` → JavaScript/TypeScript
3. `go.mod` → Go
4. `Cargo.toml` → Rust
5. `Gemfile` or `*.gemspec` → Ruby
6. `composer.json` → PHP
7. `pom.xml` or `build.gradle` → Java

## Configuration

Set the detected language in `.moai/config.json`:

```json
{
  "project": {
    "language": "python"
  }
}
```

## Best Practices

- Keep language config in sync with actual project files
- Use language-specific tools for validation and testing
- Enable language-specific hooks and skills

