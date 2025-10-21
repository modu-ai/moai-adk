---
name: moai-domain-cli-tool
description: CLI tool development with argument parsing, POSIX compliance, and user-friendly help messages
allowed-tools:
  - Read
  - Bash
tier: 4
auto-load: "false"
---

# CLI Tool Expert

## What it does

Provides expertise in developing command-line interface tools with proper argument parsing, POSIX compliance, intuitive help messages, and standard exit codes.

## When to use

- “CLI tool development”, “command line parsing”, “POSIX compatibility”
- Automatically invoked when working with CLI projects
- CLI tool SPEC implementation (`/alfred:2-run`)

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
User: "/alfred:2-run CLI-001"
Claude: (creates RED CLI test, GREEN implementation with click, REFACTOR)

### Example 2: POSIX compliance check
User: "Check POSIX compatibility"
Claude: (validates exit codes, option formats, stderr usage)

## Works well with

- alfred-trust-validation (CLI testing)
- shell-expert (shell integration)
- python-expert/typescript-expert (implementation)
