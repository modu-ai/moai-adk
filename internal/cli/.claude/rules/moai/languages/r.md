---
paths:
  - "**/*.R"
  - "**/*.Rmd"
  - "**/DESCRIPTION"
  - "**/NAMESPACE"
---

# R Rules

Version: R 4.4+

## Tooling

- IDE: RStudio or VS Code
- Linting: lintr
- Formatting: styler
- Testing: testthat
- Package management: renv

## MUST

- Use tidyverse style for data manipulation
- Use ggplot2 for visualization
- Use roxygen2 for documentation
- Use renv for dependency management
- Vectorize operations over loops
- Handle NA values explicitly

## MUST NOT

- Use attach() for data frames
- Use setwd() in scripts
- Use T/F instead of TRUE/FALSE
- Leave hardcoded file paths
- Suppress warnings without reason
- Use <<- for global assignment

## File Conventions

- test-*.R for test files in tests/testthat/
- Use snake_case for functions and variables
- R/ for package source code
- Use .Rproj for project settings
- Keep scripts under 500 lines

## Testing

- Use testthat with expect_* assertions
- Use snapshot tests for complex output
- Use withr for temporary state changes
- Test edge cases and NA handling

## MoAI Integration

- Use Skill("moai-lang-r") for detailed patterns
- Follow TRUST 5 quality gates
