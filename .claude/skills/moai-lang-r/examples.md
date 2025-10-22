# moai-lang-r - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with testthat & lintr

```r
# Create new project structure
usethis::create_package("mypackage")

# Setup testing infrastructure
usethis::use_testthat()

# Setup code quality tools
usethis::use_lint()

# Create .lintr configuration
writeLines(c(
  "linters: linters_with_defaults(",
  "  line_length_linter(120),",
  "  object_name_linter = NULL",
  ")"
), ".lintr")

# Install dependencies
install.packages(c("testthat", "lintr", "devtools"))
```

**DESCRIPTION file**:
```
Package: mypackage
Title: My Analysis Package
Version: 0.1.0
Authors@R: person("First", "Last", email = "first.last@example.com", role = c("aut", "cre"))
Description: Package for data analysis.
License: MIT + file LICENSE
Encoding: UTF-8
LazyData: true
Suggests:
    testthat (>= 3.2.0),
    lintr (>= 3.2.0)
RoxygenNote: 7.3.2
```

## Example 2: TDD Workflow with testthat

**RED: Write failing test**
```r
# tests/testthat/test-calculator.R
test_that("add() works with positive numbers", {
  result <- add(2, 3)
  expect_equal(result, 5)
})

test_that("add() handles negative numbers", {
  result <- add(-1, -2)
  expect_equal(result, -3)
})

test_that("add() handles zero", {
  result <- add(0, 5)
  expect_equal(result, 5)
})

test_that("add() requires numeric inputs", {
  expect_error(add("a", "b"))
  expect_error(add(1, "b"))
})
```

**GREEN: Implement feature**
```r
# R/calculator.R
#' Add two numbers
#'
#' @param a Numeric value
#' @param b Numeric value
#' @return Sum of a and b
#' @export
add <- function(a, b) {
  if (!is.numeric(a) || !is.numeric(b)) {
    stop("Both arguments must be numeric")
  }
  return(a + b)
}
```

**REFACTOR: Improve code quality**
```r
# R/calculator.R
#' Add two numbers
#'
#' This function adds two numeric values together with input validation.
#'
#' @param a A numeric value (first operand)
#' @param b A numeric value (second operand)
#' @return A numeric value representing the sum of a and b
#' @examples
#' add(2, 3)  # returns 5
#' add(-1, 1) # returns 0
#' @export
add <- function(a, b) {
  # Validate inputs
  if (!is.numeric(a)) {
    stop("Argument 'a' must be numeric, got: ", class(a))
  }
  if (!is.numeric(b)) {
    stop("Argument 'b' must be numeric, got: ", class(b))
  }

  # Perform addition
  result <- a + b
  return(result)
}
```

## Example 3: Tidyverse Data Analysis with Tests

**Test file: tests/testthat/test-data-analysis.R**
```r
library(dplyr)
library(testthat)

test_that("filter_active_users() returns only active users", {
  test_data <- tibble::tibble(
    id = 1:5,
    name = c("Alice", "Bob", "Charlie", "David", "Eve"),
    status = c("active", "inactive", "active", "active", "inactive")
  )

  result <- filter_active_users(test_data)

  expect_equal(nrow(result), 3)
  expect_equal(result$name, c("Alice", "Charlie", "David"))
  expect_true(all(result$status == "active"))
})

test_that("summarize_by_group() calculates correct statistics", {
  test_data <- tibble::tibble(
    group = c("A", "A", "B", "B", "B"),
    value = c(10, 20, 15, 25, 30)
  )

  result <- summarize_by_group(test_data, group, value)

  expect_equal(nrow(result), 2)
  expect_equal(result$mean_value, c(15, 23.33), tolerance = 0.01)
  expect_equal(result$count, c(2, 3))
})
```

**Implementation: R/data-analysis.R**
```r
#' Filter active users from dataset
#'
#' @param data A data frame with 'status' column
#' @return Filtered data frame containing only active users
#' @importFrom dplyr filter
#' @export
filter_active_users <- function(data) {
  if (!"status" %in% names(data)) {
    stop("Data must contain 'status' column")
  }

  data %>%
    dplyr::filter(status == "active")
}

#' Summarize data by group
#'
#' @param data A data frame
#' @param group_var Grouping variable (unquoted)
#' @param value_var Value variable (unquoted)
#' @return Summary data frame with mean and count
#' @importFrom dplyr group_by summarise n
#' @importFrom rlang enquo
#' @export
summarize_by_group <- function(data, group_var, value_var) {
  group_var <- rlang::enquo(group_var)
  value_var <- rlang::enquo(value_var)

  data %>%
    dplyr::group_by(!!group_var) %>%
    dplyr::summarise(
      mean_value = mean(!!value_var, na.rm = TRUE),
      count = dplyr::n(),
      .groups = "drop"
    )
}
```

## Example 4: Code Quality with lintr

**.lintr configuration**:
```r
linters: linters_with_defaults(
  line_length_linter(120),
  object_name_linter = NULL,
  cyclocomp_linter(25),
  indentation_linter(4)
)
exclusions: list(
  "tests/testthat.R",
  "data-raw/"
)
```

**Run linting**:
```r
# Lint all R files
lintr::lint_package()

# Lint specific file
lintr::lint("R/calculator.R")

# Lint with custom linters
lintr::lint_package(
  linters = lintr::linters_with_defaults(
    line_length_linter = lintr::line_length_linter(100)
  )
)
```

## Example 5: Test Coverage with covr

```r
# Install covr package
install.packages("covr")

# Calculate package coverage
coverage <- covr::package_coverage()

# View coverage report
covr::report(coverage)

# Check if coverage meets threshold (85%)
covr::percent_coverage(coverage)

# Generate coverage report for CI
covr::codecov(
  coverage = coverage,
  token = Sys.getenv("CODECOV_TOKEN")
)
```

**Add to tests/testthat.R**:
```r
library(testthat)
library(mypackage)

# Run all tests
test_check("mypackage")
```

## Example 6: Data Pipeline Testing

**Test file: tests/testthat/test-pipeline.R**
```r
test_that("data pipeline processes correctly", {
  # Arrange: Create test data
  raw_data <- tibble::tibble(
    id = 1:10,
    value = rnorm(10, mean = 50, sd = 10),
    category = rep(c("A", "B"), 5),
    date = seq.Date(as.Date("2025-01-01"), by = "day", length.out = 10)
  )

  # Act: Run pipeline
  result <- run_data_pipeline(raw_data)

  # Assert: Verify transformations
  expect_s3_class(result, "data.frame")
  expect_equal(nrow(result), 10)
  expect_true("normalized_value" %in% names(result))
  expect_true(all(result$normalized_value >= 0 & result$normalized_value <= 1))
})

test_that("pipeline handles missing values", {
  data_with_na <- tibble::tibble(
    id = 1:5,
    value = c(1, NA, 3, NA, 5),
    category = c("A", "A", "B", "B", NA)
  )

  result <- run_data_pipeline(data_with_na, remove_na = TRUE)

  expect_equal(nrow(result), 2)  # Only complete cases
  expect_true(all(!is.na(result)))
})
```

---

_For complete CLI reference and configuration options, see reference.md_
