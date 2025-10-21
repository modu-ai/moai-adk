---

name: moai-lang-shell
description: Shell scripting best practices with bats, shellcheck, and POSIX compliance. Use when writing or reviewing Shell scripts code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Shell Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Shell code discussions, framework guidance, or file extensions such as .sh/.bash. |
| Tier | 3 |

## What it does

Provides shell scripting expertise for TDD development, including bats testing framework, shellcheck linting, and POSIX compliance for portable scripts.

## When to use

- Engages when the conversation references Shell work, frameworks, or files like .sh/.bash.
- "Writing shell scripts", "bats testing", "POSIX compatibility"
- Automatically invoked when working with shell script projects
- Shell SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **bats**: Bash Automated Testing System
- **shunit2**: xUnit-style shell testing
- **assert.sh**: Shell assertion library
- Test-driven shell development

**Code Quality**:
- **shellcheck**: Static analysis for shell scripts
- **shfmt**: Shell script formatting
- **bashate**: Style checker

**POSIX Compliance**:
- Portable shell features (sh vs bash)
- Avoid bashisms for portability
- Use `[ ]` instead of `[[ ]]` for POSIX
- Standard utilities (no GNU extensions)

**Shell Patterns**:
- **Error handling**: set -e, set -u, set -o pipefail
- **Exit codes**: Proper use of 0 (success) and non-zero
- **Quoting**: Always quote variables ("$var")
- **Functions**: Modular script organization

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Use `#!/bin/sh` for POSIX, `#!/bin/bash` for Bash
- Check command existence with `command -v`
- Use `$()` over backticks
- Validate input arguments

## Examples
```bash
bats tests && shellcheck scripts/*.sh
```

## Inputs
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- GNU. "Bash Reference Manual." https://www.gnu.org/software/bash/manual/bash.html (accessed 2025-03-29).
- koalaman. "ShellCheck." https://www.shellcheck.net/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Shell-specific review)
- devops-expert (Deployment scripts)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
