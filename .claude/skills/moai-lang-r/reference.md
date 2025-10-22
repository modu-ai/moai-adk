# R 4.4+ CLI Reference — Tool Command Matrix

**Framework**: R 4.4+ + testthat 3.2.3+ + tidyverse 2.0+ + lintr 3.1+

---

## R Runtime Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `R --version` | Check R version | `R --version` → `R version 4.4.2` |
| `R` | Start R interactive session | `R` |
| `Rscript script.R` | Run R script | `Rscript script.R` |
| `Rscript -e "code"` | Execute R code inline | `Rscript -e "print(R.version)"` |
| `R CMD BATCH script.R` | Run script in batch mode | `R CMD BATCH script.R` |
| `R CMD check package/` | Check R package | `R CMD check mypackage/` |
| `R CMD build package/` | Build R package | `R CMD build mypackage/` |
| `R CMD INSTALL package.tar.gz` | Install package from source | `R CMD INSTALL mypackage.tar.gz` |
| `R --help` | Show R help | `R --help` |
| `R --vanilla` | Start R with no saved data | `R --vanilla` |
| `R -q` | Start R quietly (no banner) | `R -q` |

---

## Package Management

| Command | Purpose | Example |
|---------|---------|---------|
| `install.packages("package")` | Install CRAN package | `install.packages("tidyverse")` |
| `remove.packages("package")` | Remove package | `remove.packages("tidyverse")` |
| `update.packages()` | Update all packages | `update.packages()` |
| `library(package)` | Load package | `library(tidyverse)` |
| `installed.packages()` | List installed packages | `installed.packages()` |
| `packageVersion("package")` | Check package version | `packageVersion("tidyverse")` |

### devtools / remotes

| Command | Purpose | Example |
|---------|---------|---------|
| `devtools::install_github("user/repo")` | Install from GitHub | `devtools::install_github("tidyverse/dplyr")` |
| `devtools::load_all()` | Load package in development | `devtools::load_all()` |
| `devtools::test()` | Run package tests | `devtools::test()` |
| `devtools::check()` | Check package | `devtools::check()` |
| `devtools::document()` | Generate documentation | `devtools::document()` |

---

## Testing Framework (testthat 3.2.3+)

| Command | Purpose | Example |
|---------|---------|---------|
| `testthat::test_dir("tests/")` | Run all tests in directory | `testthat::test_dir("tests/")` |
| `testthat::test_file("test-calc.R")` | Run specific test file | `testthat::test_file("tests/testthat/test-calc.R")` |
| `devtools::test()` | Run package tests | `devtools::test()` |
| `testthat::auto_test()` | Watch and run tests on change | `testthat::auto_test()` |

---

## Code Coverage (covr)

| Command | Purpose | Example |
|---------|---------|---------|
| `covr::package_coverage()` | Calculate package coverage | `covr::package_coverage()` |
| `covr::report()` | Generate coverage report | `covr::report()` |

---

## Linting (lintr 3.1+)

| Command | Purpose | Example |
|---------|---------|---------|
| `lintr::lint("file.R")` | Lint specific file | `lintr::lint("R/calculator.R")` |
| `lintr::lint_dir("directory")` | Lint directory | `lintr::lint_dir("R/")` |
| `lintr::lint_package()` | Lint entire package | `lintr::lint_package()` |

---

## Code Formatting (styler)

| Command | Purpose | Example |
|---------|---------|---------|
| `styler::style_file("file.R")` | Format file | `styler::style_file("R/calculator.R")` |
| `styler::style_dir("directory")` | Format directory | `styler::style_dir("R/")` |
| `styler::style_pkg()` | Format package | `styler::style_pkg()` |

---

## tidyverse Workflow

**Core tidyverse Packages**:
```r
library(tidyverse)
# Loads: dplyr, ggplot2, tidyr, readr, purrr, tibble, stringr, forcats
```

**Common Operations**:
```r
# Data manipulation (dplyr)
mtcars %>%
  filter(mpg > 20) %>%
  select(mpg, cyl, hp) %>%
  mutate(hp_per_cyl = hp / cyl) %>%
  arrange(desc(mpg))

# Data reading (readr)
read_csv("data/data.csv")
write_csv(data, "output/result.csv")

# Visualization (ggplot2)
ggplot(mtcars, aes(x = mpg, y = hp)) +
  geom_point() +
  geom_smooth(method = "lm") +
  theme_minimal()
```

---

## Combined Workflow (Quality Gate)

**Before Commit** (all must pass):

```bash
#!/bin/bash
set -e

echo "Running R quality gate checks..."

# 1. Run tests
echo "1. Running tests..."
Rscript -e "devtools::test()"

# 2. Check coverage
echo "2. Checking coverage..."
Rscript -e "cov <- covr::package_coverage(); print(covr::percent_coverage(cov)); if (covr::percent_coverage(cov) < 85) quit(status = 1)"

# 3. Lint code
echo "3. Linting code..."
Rscript -e "lintr::lint_package()"

# 4. Format code
echo "4. Formatting code..."
Rscript -e "styler::style_pkg()"

echo "✅ All quality gates passed!"
```

---

## TRUST 5 Principles Integration

### T - Test First (testthat 3.2.3+)
```r
devtools::test()
```

### R - Readable (lintr 3.1+ + styler)
```r
lintr::lint_package()
styler::style_pkg()
```

### U - Unified Types (Runtime Validation)
```r
assertthat::assert_that(is.numeric(x))
checkmate::assertNumber(x, lower = 0, upper = 100)
```

### S - Security
```r
devtools::check()
```

### T - Trackable (@TAG)
```bash
rg '@(CODE|TEST|SPEC):' -n R/ tests/ --type r
```

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: R 4.4+ CLI Tools Reference
