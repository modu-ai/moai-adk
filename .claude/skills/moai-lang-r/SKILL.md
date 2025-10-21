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
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- R Core Team. "R Language Definition." https://cran.r-project.org/manuals.html (accessed 2025-03-29).
- RStudio. "testthat Reference." https://testthat.r-lib.org/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (R-specific review)
- data-science-expert (statistical analysis)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
