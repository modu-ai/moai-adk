---
name: moai-domain-cli-tool
description: CLI tool development with argument parsing, POSIX compliance, and user-friendly
  help messages
allowed-tools:
- Read
- Bash
---

# CLI Tool Expert

## What it does

Provides expertise in developing command-line interface tools with proper argument parsing, POSIX compliance, intuitive help messages, and standard exit codes.

## When to use

- "CLI 도구 개발", "명령줄 파싱", "POSIX 호환성"
- Automatically invoked when working with CLI projects
- CLI tool SPEC implementation (`/alfred:2-build`)

## How it works

**Argument Parsing**:
- **Python**: argparse, click, typer
- **Node.js**: commander, yargs, oclif
- **Rust**: clap, structopt
- **Go**: cobra, flag
- **Subcommands**: git-style commands (tool add, tool remove)

**POSIX Compliance**:
- **Short options**: -h, -v
- **Long options**: --help, --version
- **Option arguments**: -o file, --output=file
- **Standard streams**: stdin, stdout, stderr
- **Exit codes**: 0 (success), 1-255 (errors)

**User Experience**:
- **Help messages**: Comprehensive usage documentation
- **Auto-completion**: Shell completion (bash, zsh, fish)
- **Progress indicators**: Spinners, progress bars
- **Color output**: ANSI colors for readability
- **Interactive prompts**: Confirmation dialogs

**Configuration**:
- **Config files**: YAML, JSON, TOML (e.g., ~/.toolrc)
- **Environment variables**: Fallback configuration
- **Precedence**: CLI args > env vars > config file > defaults

## Examples

### Example 1: CLI tool with subcommands
User: "/alfred:2-build CLI-001"
Claude: (creates RED CLI test, GREEN implementation with click, REFACTOR)

### Example 2: POSIX compliance check
User: "POSIX 호환성 확인"
Claude: (validates exit codes, option formats, stderr usage)

## Works well with

- alfred-trust-validation (CLI testing)
- shell-expert (shell integration)
- python-expert/typescript-expert (implementation)
