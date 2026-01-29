---
paths:
  - "**/*.R"
  - "**/*.Rmd"
  - "**/DESCRIPTION"
---

# R Rules

Version: R 4.4+

## Tooling

- Linting: lintr
- Formatting: styler
- Testing: testthat with coverage >= 85%
- Documentation: roxygen2

## Best Practices (2026)

- Use tidyverse for data manipulation
- Use ggplot2 for visualization
- Prefer tibbles over data.frames
- Use data.table::fread for large CSV files
- Load data to global scope for interactive analysis

## Large Data Handling

```r
# Use fread for fast CSV loading (10-100x faster)
library(data.table)
dt <- fread("large_file.csv", nThread = 4)

# Convert to tibble if needed
library(dplyr)
df <- as_tibble(dt)

# Use arrow for Parquet files
library(arrow)
df <- read_parquet("data.parquet")
```

## Shiny Patterns

```r
# Reactive programming
server <- function(input, output, session) {
  # Reactive value
  data <- reactive({
    req(input$file)
    fread(input$file$datapath)
  })

  # Render output
  output$plot <- renderPlot({
    ggplot(data(), aes(x, y)) + geom_point()
  })
}
```

## Testing with testthat

```r
# tests/testthat/test-analysis.R
test_that("analysis returns expected structure", {
  result <- run_analysis(sample_data)
  expect_s3_class(result, "tbl_df")
  expect_equal(ncol(result), 5)
})
```

## Performance

- Use `parallel::mclapply` for parallel processing
- Use `memoise` for caching expensive computations
- Profile with `profvis`
- Use `bench` for micro-benchmarking

## MoAI Integration

- Use Skill("moai-lang-r") for detailed patterns
- Follow TRUST 5 quality gates
