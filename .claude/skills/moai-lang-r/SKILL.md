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
result <- data |>
  filter(!is.na(value)) |>
  mutate(log_value = log(value)) |>
  summarise(mean_log = mean(log_value))

# Placeholder _ for non-first argument
data |>
  lm(formula = y ~ x, data = _)
```

Lambda Syntax with Backslash:
```r
map(data, \(x) x^2)
map2(list1, list2, \(x, y) x + y)

# In dplyr contexts
data |>
  mutate(across(where(is.numeric), \(x) scale(x)[,1]))
```

### tidyverse Data Manipulation

dplyr Core Verbs:
```r
library(dplyr)

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
  mutate(across(starts_with("price"), \(x) round(x, 2)))
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
```

purrr Functional Programming:
```r
library(purrr)

files |> map(\(f) read_csv(f))
files |> map_dfr(\(f) read_csv(f), .id = "source")
values |> map_dbl(\(x) mean(x, na.rm = TRUE))

# safely for error handling
safe_read <- safely(read_csv)
results <- files |> map(safe_read)
successes <- results |> map("result") |> compact()
```

### ggplot2 Visualization Patterns

Complete Plot Structure:
```r
library(ggplot2)
library(scales)

p <- ggplot(data, aes(x = x, y = y, color = group)) +
  geom_point(alpha = 0.7, size = 3) +
  geom_smooth(method = "lm", se = TRUE) +
  scale_x_continuous(labels = comma) +
  scale_y_log10(labels = dollar) +
  scale_color_brewer(palette = "Set2") +
  facet_wrap(~ category, scales = "free_y") +
  labs(
    title = "Analysis Title",
    subtitle = "Descriptive subtitle",
    x = "X Axis Label",
    y = "Y Axis Label"
  ) +
  theme_minimal(base_size = 12) +
  theme(legend.position = "bottom")

ggsave("output.png", p, width = 10, height = 6, dpi = 300)
```

Multiple Plots with patchwork:
```r
library(patchwork)

p1 <- ggplot(data, aes(x)) + geom_histogram()
p2 <- ggplot(data, aes(x, y)) + geom_point()
p3 <- ggplot(data, aes(group, y)) + geom_boxplot()

combined <- (p1 | p2) / p3 +
  plot_annotation(title = "Combined Analysis", tag_levels = "A")
```

### Shiny Application Patterns

Modular Shiny App:
```r
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

    reactive({
      req(input$category)
      data() |>
        filter(
          category == input$category,
          value >= input$range[1],
          value <= input$range[2]
        )
    })
  })
}
```

Reactive Patterns:
```r
server <- function(input, output, session) {
  # reactive: Cached computation
  processed_data <- reactive({
    raw_data() |>
      filter(year == input$year)
  })

  # reactiveVal: Mutable state
  counter <- reactiveVal(0)
  observeEvent(input$increment, {
    counter(counter() + 1)
  })

  # eventReactive: Trigger on specific event
  analysis <- eventReactive(input$run_analysis, {
    expensive_computation(processed_data())
  })

  # debounce for rapid inputs
  search_term <- reactive(input$search) |> debounce(300)
}
```

### testthat Testing Framework

Test Structure:
```r
library(testthat)

test_that("calculate_growth returns correct values", {
  data <- tibble(year = 2020:2022, value = c(100, 110, 121))

  result <- calculate_growth(data)

  expect_equal(nrow(result), 3)
  expect_equal(result$growth[2], 0.1, tolerance = 0.001)
  expect_true(is.na(result$growth[1]))
})

test_that("calculate_growth handles edge cases", {
  expect_error(calculate_growth(NULL), "data cannot be NULL")
})
```

### renv Dependency Management

Project Setup:
```r
renv::init()
renv::install("tidyverse")
renv::install("shiny")
renv::snapshot()
renv::restore()
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
- [Advanced Patterns](modules/advanced-patterns.md) - Complete advanced patterns guide

---

## Context7 Library Mappings

```
/tidyverse/dplyr - Data manipulation verbs
/tidyverse/ggplot2 - Grammar of graphics visualization
/tidyverse/purrr - Functional programming toolkit
/tidyverse/tidyr - Data tidying functions
/rstudio/shiny - Web application framework
/r-lib/testthat - Unit testing framework
/rstudio/renv - Dependency management
```

---

## Works Well With

- `moai-lang-python` - Python/R interoperability with reticulate
- `moai-domain-database` - SQL patterns and database optimization
- `moai-workflow-testing` - TDD and testing strategies
- `moai-essentials-debug` - AI-powered debugging
- `moai-foundation-quality` - TRUST 5 quality principles

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
renv::clean()
renv::rebuild()
renv::snapshot(force = TRUE)
```

Shiny Reactivity Debug:
```r
options(shiny.reactlog = TRUE)
reactlog::reactlog_enable()
shiny::reactlogShow()
```

ggplot2 Font Issues:
```r
library(showtext)
font_add_google("Roboto", "roboto")
showtext_auto()
```

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
