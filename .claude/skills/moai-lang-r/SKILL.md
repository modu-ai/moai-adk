---
name: moai-lang-r
description: R 4.4+ development specialist covering tidyverse, ggplot2, Shiny, and data science patterns. Use when developing data analysis pipelines, visualizations, or Shiny applications.
version: 1.0.0
updated: 2025-12-07
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

R 4.4+ Development Specialist - tidyverse, ggplot2, Shiny, renv, and modern R patterns.

Auto-Triggers: `.R` files, `.Rmd`, `.qmd`, `DESCRIPTION`, `renv.lock`, Shiny/ggplot2 discussions

Core Capabilities:
- R 4.4 Features: Native pipe |>, lambda syntax \(x), improved error messages
- Data Manipulation: dplyr, tidyr, purrr, stringr, forcats
- Visualization: ggplot2, plotly, scales, patchwork
- Web Applications: Shiny, reactivity, modules, bslib
- Testing: testthat 3.0, snapshot testing, mocking
- Package Management: renv, pak, DESCRIPTION
- Reproducible Reports: R Markdown, Quarto
- Database: DBI, dbplyr, pool

### Quick Patterns

dplyr Data Pipeline:
```r
library(tidyverse)

result <- data |>
  filter(year >= 2020) |>
  mutate(
    revenue_k = revenue / 1000,
    growth = (revenue - lag(revenue)) / lag(revenue)
  ) |>
  group_by(category) |>
  summarise(
    total_revenue = sum(revenue_k, na.rm = TRUE),
    avg_growth = mean(growth, na.rm = TRUE),
    .groups = "drop"
  )
```

ggplot2 Visualization:
```r
library(ggplot2)

ggplot(data, aes(x = date, y = value, color = category)) +
  geom_line(linewidth = 1) +
  geom_point(size = 2) +
  scale_color_viridis_d() +
  labs(
    title = "Trend Analysis",
    x = "Date", y = "Value",
    color = "Category"
  ) +
  theme_minimal()
```

Shiny Basic App:
```r
library(shiny)

ui <- fluidPage(
  selectInput("var", "Variable:", choices = names(mtcars)),
  plotOutput("plot")
)

server <- function(input, output, session) {
  output$plot <- renderPlot({
    ggplot(mtcars, aes(.data[[input$var]])) +
      geom_histogram()
  })
}

shinyApp(ui, server)
```

---

## Implementation Guide (5 minutes)

### R 4.4 Modern Features

Native Pipe Operator |>:
```r
# R 4.4+ native pipe (preferred over magrittr %>%)
result <- data |>
  filter(!is.na(value)) |>
  mutate(log_value = log(value)) |>
  summarise(mean_log = mean(log_value))

# Placeholder _ for non-first argument
data |>
  lm(formula = y ~ x, data = _)

# Chain with function calls
data |>
  (\(d) split(d, d$group))() |>
  lapply(\(x) mean(x$value))
```

Lambda Syntax with Backslash:
```r
# R 4.1+ shorthand lambda (replaces function(x))
map(data, \(x) x^2)
map2(list1, list2, \(x, y) x + y)
pmap(list(a = 1:3, b = 4:6), \(a, b) a * b)

# In dplyr contexts
data |>
  mutate(across(where(is.numeric), \(x) scale(x)[,1]))
```

Improved Error Messages:
```r
# R 4.4 provides clearer error traces
# Enable for debugging
options(error = rlang::entrace)
rlang::global_entrace()

# Informative conditions
rlang::abort(

  message = "Data validation failed",
  class = "validation_error",
  body = c(
    i = "Column 'value' contains negative numbers",
    x = "Expected all positive values"
  )
)
```

### tidyverse Data Manipulation

dplyr Core Verbs:
```r
library(dplyr)

# filter, select, mutate, summarise, arrange
processed <- raw_data |>
  filter(status == "active", amount > 0) |>
  select(id, date, amount, category) |>
  mutate(
    month = floor_date(date, "month"),
    amount_scaled = amount / max(amount)
  ) |>
  arrange(desc(date))

# group_by with summarise
summary <- processed |>
  group_by(category, month) |>
  summarise(
    n = n(),
    total = sum(amount),
    avg = mean(amount),
    .groups = "drop"
  )

# across for multiple columns
data |>
  mutate(across(starts_with("price"), \(x) round(x, 2))) |>
  summarise(across(where(is.numeric), list(
    mean = \(x) mean(x, na.rm = TRUE),
    sd = \(x) sd(x, na.rm = TRUE)
  )))
```

tidyr Reshaping:
```r
library(tidyr)

# pivot_longer (wide to long)
wide_data |>
  pivot_longer(
    cols = starts_with("year_"),
    names_to = "year",
    names_prefix = "year_",
    values_to = "value"
  )

# pivot_wider (long to wide)
long_data |>
  pivot_wider(
    names_from = category,
    values_from = value,
    values_fill = 0
  )

# nest/unnest for list columns
data |>
  group_by(id) |>
  nest() |>
  mutate(model = map(data, \(d) lm(y ~ x, data = d))) |>
  mutate(summary = map(model, broom::tidy)) |>
  unnest(summary)
```

purrr Functional Programming:
```r
library(purrr)

# map variants
files |> map(\(f) read_csv(f))
files |> map_dfr(\(f) read_csv(f), .id = "source")
values |> map_dbl(\(x) mean(x, na.rm = TRUE))
values |> map_chr(\(x) paste(x, collapse = ", "))

# safely for error handling
safe_read <- safely(read_csv)
results <- files |> map(safe_read)
successes <- results |> map("result") |> compact()
errors <- results |> map("error") |> compact()

# possibly with default value
safe_log <- possibly(log, otherwise = NA_real_)
values |> map_dbl(safe_log)

# walk for side effects
plots |> walk(\(p) ggsave(paste0(p$name, ".png"), p$plot))
```

stringr Text Processing:
```r
library(stringr)

# Common operations
str_detect(text, "pattern")
str_extract(text, "\\d+")
str_replace_all(text, "old", "new")
str_split(text, ",")

# Complex patterns with regex
data |>
  mutate(
    email_domain = str_extract(email, "(?<=@)[^@]+$"),
    first_word = str_extract(name, "^\\w+"),
    cleaned = str_squish(str_to_lower(text))
  )
```

### ggplot2 Visualization Patterns

Complete Plot Structure:
```r
library(ggplot2)
library(scales)

p <- ggplot(data, aes(x = x, y = y, color = group)) +
  # Geometries
  geom_point(alpha = 0.7, size = 3) +
  geom_smooth(method = "lm", se = TRUE) +
  # Scales
  scale_x_continuous(labels = comma) +
  scale_y_log10(labels = dollar) +
  scale_color_brewer(palette = "Set2") +
  # Facets
  facet_wrap(~ category, scales = "free_y") +
  # Labels
  labs(
    title = "Analysis Title",
    subtitle = "Descriptive subtitle",
    caption = "Source: Dataset",
    x = "X Axis Label",
    y = "Y Axis Label"
  ) +
  # Theme
  theme_minimal(base_size = 12) +
  theme(
    legend.position = "bottom",
    plot.title = element_text(face = "bold"),
    axis.text.x = element_text(angle = 45, hjust = 1)
  )

# Save with proper dimensions
ggsave("output.png", p, width = 10, height = 6, dpi = 300)
```

Multiple Plots with patchwork:
```r
library(patchwork)

p1 <- ggplot(data, aes(x)) + geom_histogram()
p2 <- ggplot(data, aes(x, y)) + geom_point()
p3 <- ggplot(data, aes(group, y)) + geom_boxplot()

# Combine plots
combined <- (p1 | p2) / p3 +
  plot_annotation(
    title = "Combined Analysis",
    tag_levels = "A"
  ) +
  plot_layout(heights = c(1, 2))
```

### Shiny Application Patterns

Modular Shiny App:
```r
# Module: data_filter_module.R
dataFilterUI <- function(id) {
  ns <- NS(id)
  tagList(
    selectInput(ns("category"), "Category:", choices = NULL),
    sliderInput(ns("range"), "Range:", min = 0, max = 100, value = c(0, 100))
  )
}

dataFilterServer <- function(id, data) {
  moduleServer(id, function(input, output, session) {
    observe({
      categories <- unique(data()$category)
      updateSelectInput(session, "category", choices = categories)
    })

    filtered_data <- reactive({
      req(input$category)
      data() |>
        filter(
          category == input$category,
          value >= input$range[1],
          value <= input$range[2]
        )
    })

    return(filtered_data)
  })
}
```

Reactive Patterns:
```r
server <- function(input, output, session) {
  # reactive: Cached computation
  processed_data <- reactive({
    raw_data() |>
      filter(year == input$year) |>
      mutate(calc = value * input$multiplier)
  })

  # reactiveVal: Mutable state
  counter <- reactiveVal(0)
  observeEvent(input$increment, {
    counter(counter() + 1)
  })

  # reactiveValues: Multiple mutable values
  rv <- reactiveValues(
    data = NULL,
    selected = NULL,
    history = list()
  )

  # eventReactive: Trigger on specific event
  analysis <- eventReactive(input$run_analysis, {
    req(processed_data())
    expensive_computation(processed_data())
  })

  # observe: Side effects
  observe({
    req(input$selection)
    rv$selected <- input$selection
  })

  # debounce for rapid inputs
  search_term <- reactive(input$search) |> debounce(300)
}
```

Shiny with bslib Modern UI:
```r
library(shiny)
library(bslib)

ui <- page_sidebar(
  title = "Data Dashboard",
  theme = bs_theme(
    bootswatch = "flatly",
    primary = "#0066CC"
  ),
  sidebar = sidebar(
    title = "Controls",
    selectInput("dataset", "Dataset:", choices = c("mtcars", "iris")),
    numericInput("n", "Rows:", value = 10, min = 1)
  ),
  navset_card_tab(
    nav_panel("Table", tableOutput("table")),
    nav_panel("Plot", plotOutput("plot")),
    nav_panel("Summary", verbatimTextOutput("summary"))
  )
)
```

### testthat Testing Framework

Test Structure:
```r
# tests/testthat/test-calculations.R
library(testthat)

test_that("calculate_growth returns correct values", {
  # Arrange
  data <- tibble(year = 2020:2022, value = c(100, 110, 121))

  # Act
  result <- calculate_growth(data)

  # Assert
  expect_equal(nrow(result), 3)
  expect_equal(result$growth[2], 0.1, tolerance = 0.001)
  expect_true(is.na(result$growth[1]))
})

test_that("calculate_growth handles edge cases", {
  expect_error(calculate_growth(NULL), "data cannot be NULL")
  expect_warning(calculate_growth(tibble()), "empty data")
})
```

Snapshot Testing:
```r
test_that("plot output is consistent", {
  p <- create_summary_plot(test_data)
  expect_snapshot_output(print(p))
})

test_that("table output matches snapshot", {
  result <- generate_report(test_data)
  expect_snapshot(result)
})
```

Mocking External Dependencies:
```r
test_that("fetch_data uses API correctly", {
  local_mocked_bindings(
    httr2_request = function(url) {
      list(status_code = 200, body = '{"data": [1,2,3]}')
    }
  )

  result <- fetch_data("https://api.example.com/data")
  expect_equal(result, c(1, 2, 3))
})
```

### renv Dependency Management

Project Setup:
```r
# Initialize renv in project
renv::init()

# Install packages
renv::install("tidyverse")
renv::install("shiny")
renv::install("user/package")  # GitHub

# Snapshot current state
renv::snapshot()

# Restore from lockfile
renv::restore()

# Update packages
renv::update()
```

renv.lock Structure:
```json
{
  "R": {
    "Version": "4.4.0",
    "Repositories": [
      {"Name": "CRAN", "URL": "https://cloud.r-project.org"}
    ]
  },
  "Packages": {
    "dplyr": {
      "Package": "dplyr",
      "Version": "1.1.4",
      "Source": "Repository"
    }
  }
}
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Advanced Shiny patterns (async, caching, deployment)
- Complex ggplot2 extensions and custom themes
- Database integration with dbplyr and pool
- R package development patterns
- Performance optimization techniques
- Production deployment (Docker, Posit Connect)

See:
- [reference.md](reference.md) - Complete reference documentation
- [examples.md](examples.md) - Production-ready code examples

---

## Context7 Library Mappings

```
/tidyverse/dplyr - Data manipulation verbs
/tidyverse/ggplot2 - Grammar of graphics visualization
/tidyverse/purrr - Functional programming toolkit
/tidyverse/tidyr - Data tidying functions
/tidyverse/stringr - String manipulation
/rstudio/shiny - Web application framework
/r-lib/testthat - Unit testing framework
/rstudio/renv - Dependency management
```

---

## Works Well With

- `moai-lang-python` - Python/R interoperability with reticulate
- `moai-domain-database` - SQL patterns and database optimization
- `moai-quality-testing` - TDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-foundation-trust` - TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

R Version Check:
```r
R.version.string  # Should be 4.4+
packageVersion("dplyr")
```

Native Pipe Not Working:
- Ensure R version is 4.1+ for |>
- Check RStudio settings: Tools > Global Options > Code > Use native pipe

renv Issues:
```r
# Clear cache and reinitialize
renv::clean()
renv::rebuild()

# Force snapshot
renv::snapshot(force = TRUE)
```

Shiny Reactivity Debug:
```r
# Enable reactive log
options(shiny.reactlog = TRUE)
reactlog::reactlog_enable()

# View reactive dependencies
shiny::reactlogShow()
```

ggplot2 Font Issues:
```r
# Install and configure fonts
library(showtext)
font_add_google("Roboto", "roboto")
showtext_auto()
```

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
