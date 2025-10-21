---

name: moai-domain-cli-tool
description: CLI tool development with argument parsing, POSIX compliance, and user-friendly help messages. Use when working on command-line tooling scenarios.
allowed-tools:
  - Read
  - Bash
---

# CLI Tool Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for CLI design requests |
| Trigger cues | Command-line UX, packaging, distribution, and automation workflows. |
| Tier | 4 |

## What it does

Provides expertise in developing command-line interface tools with proper argument parsing, POSIX compliance, intuitive help messages, and standard exit codes.

## When to use

- Engages when building or enhancing command-line tools.
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
```bash
$ tool --help
$ tool run --config config.yml
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- Microsoft. "Command Line Interface Guidelines." https://learn.microsoft.com/windows/console/ (accessed 2025-03-29).
- Python Packaging Authority. "Command-line Interface Guidelines." https://packaging.python.org/en/latest/guides/distributing-packages-using-setuptools/#entry-points (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (CLI testing)
- shell-expert (shell integration)
- python-expert/typescript-expert (implementation)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
