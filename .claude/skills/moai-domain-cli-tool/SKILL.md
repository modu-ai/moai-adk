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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- Microsoft. "Command Line Interface Guidelines." https://learn.microsoft.com/windows/console/ (accessed 2025-03-29).
- Python Packaging Authority. "Command-line Interface Guidelines." https://packaging.python.org/en/latest/guides/distributing-packages-using-setuptools/#entry-points (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (CLI testing)
- shell-expert (shell integration)
- python-expert/typescript-expert (implementation)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
