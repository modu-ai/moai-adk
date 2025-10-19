---
name: moai-lang-r
description: R best practices with testthat, lintr, and data analysis patterns
allowed-tools:
- Read
- Bash
---

# R Expert

## What it does

Provides R-specific expertise for TDD development, including testthat testing framework, lintr code linting, and statistical data analysis patterns.

## When to use

- "R 테스트 작성", "testthat 사용법", "데이터 분석 패턴"
- Automatically invoked when working with R projects
- R SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with testthat
User: "/alfred:2-build ANALYSIS-001"
Claude: (creates RED test with testthat, GREEN implementation, REFACTOR with tidyverse)

### Example 2: Linting check
User: "lintr 실행"
Claude: (runs lintr::lint_package() and reports style issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (R-specific review)
- data-science-expert (statistical analysis)
