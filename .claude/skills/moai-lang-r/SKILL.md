---

name: moai-lang-r
description: R best practices with testthat, lintr, and data analysis patterns. Use when writing or reviewing R code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# R Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | R code discussions, framework guidance, or file extensions such as .r. |
| Tier | 3 |

## What it does

Provides R-specific expertise for TDD development, including testthat testing framework, lintr code linting, and statistical data analysis patterns.

## When to use

- Engages when the conversation references R work, frameworks, or files like .r.
- “Writing R tests”, “How to use testthat”, “Data analysis patterns”
- Automatically invoked when working with R projects
- R SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **testthat**: Unit testing framework
- **covr**: Test coverage tool
- **mockery**: Mocking library
- Test coverage ≥85% enforcement

**Code Quality**:
- **lintr**: Static code analysis
- **styler**: Code formatting
- **goodpractice**: R package best practices

**Package Management**:
- **devtools**: Package development tools
- **usethis**: Workflow automation
- **CRAN**: Official package repository

**Data Analysis Patterns**:
- **tidyverse**: Data manipulation (dplyr, ggplot2)
- **data.table**: High-performance data manipulation
- **Vectorization** over loops
- **Pipes** (%>%) for readable code

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Document functions with roxygen2
- Use meaningful variable names
- Avoid global variables
- Prefer functional programming

## Examples
```bash
Rscript -e 'devtools::test()'
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
- R Core Team. "R Language Definition." https://cran.r-project.org/manuals.html (accessed 2025-03-29).
- RStudio. "testthat Reference." https://testthat.r-lib.org/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (R-specific review)
- data-science-expert (statistical analysis)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
