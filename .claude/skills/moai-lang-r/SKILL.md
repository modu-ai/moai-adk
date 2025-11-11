---
name: moai-lang-r
version: 4.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: R 4.4 enterprise data science with Tidyverse, Shiny, RMarkdown, and advanced statistical computing. Modern data analysis, visualization, machine learning, and reporting with Context7 MCP integration.
keywords: ['r', 'data-science', 'tidyverse', 'shiny', 'rmarkdown', 'machine-learning', 'statistics', 'visualization', 'context7']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Lang R Skill - Enterprise v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-r |
| **Version** | 4.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal), Context7 MCP |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language Enterprise |
| **Context7 Integration** | ✅ R/Tidyverse/Shiny/RMarkdown |

---

## What It Does

R 4.4 enterprise data science featuring Tidyverse for data manipulation, Shiny for interactive applications, RMarkdown for reproducible reporting, and advanced statistical computing with machine learning capabilities. Context7 MCP integration provides real-time access to official R documentation.

**Key capabilities**:
- ✅ R 4.4 with advanced statistical computing
- ✅ Tidyverse ecosystem (dplyr, ggplot2, tidyr, purrr)
- ✅ Shiny 2.0 for interactive web applications
- ✅ RMarkdown for reproducible research and reporting
- ✅ Machine learning with tidymodels and caret
- ✅ Advanced data visualization with ggplot2 and plotly
- ✅ Statistical modeling and hypothesis testing
- ✅ Big data processing with data.table and sparklyr
- ✅ Enterprise reporting and dashboarding
- ✅ Package development and reproducible research

---

## When to Use

**Automatic triggers**:
- R statistical analysis and data science discussions
- Tidyverse data manipulation patterns
- Shiny application development
- Statistical modeling and machine learning
- Data visualization and ggplot2 usage
- RMarkdown reporting and documentation

**Manual invocation**:
- Design data science workflows
- Implement statistical analyses
- Create interactive dashboards
- Optimize R code performance
- Review data science code quality
- Develop custom R packages

---

## Technology Stack (2025-10-22)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **R** | 4.4.0 | Core language | ✅ Current |
| **Tidyverse** | 2.0.0 | Data science ecosystem | ✅ Current |
| **Shiny** | 2.0.0 | Web applications | ✅ Current |
| **tidymodels** | 1.2.0 | Machine learning framework | ✅ Current |
| **data.table** | 1.15.0 | High-performance data processing | ✅ Current |

---

## Code Examples (30+ Enterprise Patterns)

### 1. Advanced Tidyverse Data Processing

```r
# Advanced data processing with Tidyverse 2.0
library(tidyverse)
library(lubridate)
library(data.table)
library(DT)

# Enterprise data analysis pipeline
analyze_enterprise_data <- function(data_path, output_dir) {
  
  # Read and process large datasets efficiently
  enterprise_data <- data_path %>%
    here::here() %>%
    list.files(pattern = "\\.csv$", full.names = TRUE) %>%
    map_dfr(~ fread(.x, select = c(
      "customer_id", "transaction_date", "amount", 
      "product_category", "region", "channel"
    ))) %>%
    mutate(
      customer_id = as.character(customer_id),
      transaction_date = as.Date(transaction_date),
      amount = as.numeric(amount),
      year = year(transaction_date),
      month = month(transaction_date),
      quarter = quarter(transaction_date),
      day_of_week = wday(transaction_date, label = TRUE)
    )
  
  # Advanced customer segmentation analysis
  customer_analysis <- enterprise_data %>%
    group_by(customer_id, region) %>%
    summarise(
      total_transactions = n(),
      total_amount = sum(amount, na.rm = TRUE),
      avg_transaction_value = mean(amount, na.rm = TRUE),
      first_transaction = min(transaction_date),
      last_transaction = max(transaction_date),
      transaction_days = n_distinct(transaction_date),
      product_categories = n_distinct(product_category),
      preferred_channel = Mode(channel),
      recency_days = as.numeric(Sys.Date() - last_transaction),
      .groups = "drop"
    ) %>%
    mutate(
      customer_lifetime = as.numeric(last_transaction - first_transaction),
      avg_days_between_transactions = ifelse(
        transaction_days > 1, 
        customer_lifetime / transaction_days, 
        0
      ),
      frequency_score = case_when(
        total_transactions >= 50 ~ 5,
        total_transactions >= 30 ~ 4,
        total_transactions >= 20 ~ 3,
        total_transactions >= 10 ~ 2,
        total_transactions >= 5 ~ 1,
        TRUE ~ 0
      ),
      monetary_score = case_when(
        total_amount >= 10000 ~ 5,
        total_amount >= 5000 ~ 4,
        total_amount >= 2000 ~ 3,
        total_amount >= 500 ~ 2,
        total_amount >= 100 ~ 1,
        TRUE ~ 0
      ),
      recency_score = case_when(
        recency_days <= 30 ~ 5,
        recency_days <= 60 ~ 4,
        recency_days <= 90 ~ 3,
        recency_days <= 180 ~ 2,
        recency_days <= 365 ~ 1,
        TRUE ~ 0
      ),
      rfm_score = frequency_score + monetary_score + recency_score,
      customer_segment = case_when(
        rfm_score >= 13 ~ "Champions",
        rfm_score >= 10 ~ "Loyal Customers",
        rfm_score >= 7 ~ "Potential Loyalists",
        rfm_score >= 5 ~ "New Customers",
        rfm_score >= 3 ~ "At Risk",
        TRUE ~ "Lost"
      )
    )
  
  # Advanced churn prediction features
  churn_features <- customer_analysis %>%
    select(customer_id, starts_with("total"), starts_with("avg"), 
           starts_with("customer"), starts_with("transaction"),
           recency_days, rfm_score, customer_segment) %>%
    mutate(
      churn_risk = case_when(
        recency_days > 365 | rfm_score <= 3 ~ "High",
        recency_days > 180 | rfm_score <= 6 ~ "Medium",
        TRUE ~ "Low"
      ),
      days_since_last_purchase = recency_days,
      purchase_frequency = total_transactions / customer_lifetime,
      average_order_value = total_amount / total_transactions
    )
  
  # Save processed data
  customer_analysis %>%
    write_csv(file.path(output_dir, "customer_analysis.csv"))
  
  churn_features %>%
    write_csv(file.path(output_dir, "churn_features.csv"))
  
  return(list(
    customer_analysis = customer_analysis,
    churn_features = churn_features
  ))
}

# Advanced visualization with ggplot2
create_enterprise_dashboard <- function(customer_data, churn_data) {
  
  # Customer segment distribution
  segment_plot <- customer_data %>%
    count(customer_segment) %>%
    mutate(
      percentage = n / sum(n) * 100,
      label = paste0(customer_segment, "\n", round(percentage, 1), "%")
    ) %>%
    ggplot(aes(x = reorder(customer_segment, n), y = n, fill = customer_segment)) +
    geom_col(alpha = 0.8) +
    geom_text(aes(label = label), vjust = -0.5, fontface = "bold") +
    scale_fill_viridis_d(option = "D") +
    labs(
      title = "Customer Segment Distribution",
      subtitle = "Enterprise Customer Analysis",
      x = "Customer Segment",
      y = "Number of Customers"
    ) +
    theme_minimal() +
    theme(
      plot.title = element_text(size = 16, face = "bold"),
      plot.subtitle = element_text(size = 12),
      axis.text.x = element_text(angle = 45, hjust = 1)
    )
  
  # RFM analysis heatmap
  rfm_heatmap <- customer_data %>%
    count(customer_segment, churn_risk) %>%
    ggplot(aes(x = customer_segment, y = churn_risk, fill = n)) +
    geom_tile() +
    geom_text(aes(label = comma(n)), color = "white", fontface = "bold") +
    scale_fill_viridis_c(option = "B") +
    labs(
      title = "RFM Segments vs Churn Risk",
      x = "Customer Segment",
      y = "Churn Risk",
      fill = "Customer Count"
    ) +
    theme_minimal() +
    theme(
      axis.text.x = element_text(angle = 45, hjust = 1),
      legend.position = "bottom"
    )
  
  # Spending patterns over time
  spending_trends <- customer_data %>%
    ggplot(aes(x = avg_transaction_value, fill = customer_segment)) +
    geom_density(alpha = 0.6) +
    facet_wrap(~ customer_segment, scales = "free") +
    scale_fill_viridis_d(option = "C") +
    labs(
      title = "Average Transaction Value by Customer Segment",
      x = "Average Transaction Value ($)",
      y = "Density"
    ) +
    theme_minimal() +
    theme(
      strip.text = element_text(face = "bold"),
      legend.position = "none"
    )
  
  # Churn risk distribution
  churn_plot <- churn_data %>%
    count(churn_risk) %>%
    mutate(
      percentage = n / sum(n) * 100
    ) %>%
    ggplot(aes(x = reorder(churn_risk, n), y = n, fill = churn_risk)) +
    geom_col(alpha = 0.8) +
    geom_text(aes(label = paste0(round(percentage, 1), "%")), 
              vjust = -0.5, fontface = "bold") +
    scale_fill_manual(values = c("High" = "#E41A1C", "Medium" = "#FF7F00", "Low" = "#4DAF4A")) +
    labs(
      title = "Customer Churn Risk Distribution",
      x = "Churn Risk",
      y = "Number of Customers"
    ) +
    theme_minimal()
  
  return(list(
    segment_plot = segment_plot,
    rfm_heatmap = rfm_heatmap,
    spending_trends = spending_trends,
    churn_plot = churn_plot
  ))
}

# Mode function helper
Mode <- function(x) {
  ux <- na.omit(unique(x))
  ux[which.max(tabulate(match(x, ux)))]
}
```

### 2. Advanced Machine Learning with tidymodels

```r
# Advanced machine learning pipeline with tidymodels
library(tidymodels)
library(xgboost)
library(randomForest)
library(caret)
library(pROC)
library(yardstick)

# Enterprise churn prediction model
build_churn_prediction_model <- function(data, target_variable = "churn_risk") {
  
  # Data preprocessing and feature engineering
  recipe_spec <- recipe(as.formula(paste(target_variable, "~ .")), data = data) %>%
    step_dummy(all_nominal(), -all_outcomes()) %>%
    step_center(all_numeric(), -all_outcomes()) %>%
    step_scale(all_numeric(), -all_outcomes()) %>%
    step_novel(all_nominal(), -all_outcomes()) %>%
    step_unknown(all_nominal(), -all_outcomes()) %>%
    step_corr(all_numeric(), threshold = 0.9) %>%
    step_zv(all_predictors())
  
  # Split data
  set.seed(42)
  data_split <- initial_split(data, prop = 0.8, strata = all_of(target_variable))
  train_data <- training(data_split)
  test_data <- testing(data_split)
  
  # Cross-validation setup
  cv_folds <- vfold_cv(train_data, v = 5, strata = all_of(target_variable))
  
  # Model specifications
  rf_spec <- rand_forest(
    mode = "classification",
    mtry = tune(),
    trees = tune(),
    min_n = tune()
  ) %>%
    set_engine("ranger", importance = "impurity")
  
  xgb_spec <- boost_tree(
    mode = "classification",
    trees = tune(),
    learn_rate = tune(),
    tree_depth = tune(),
    min_n = tune(),
    loss_reduction = tune(),
    sample_size = tune(),
    mtry = tune()
  ) %>%
    set_engine("xgboost") %>%
    set_mode("classification")
  
  svm_spec <- svm_rbf(
    cost = tune(),
    sigma = tune()
  ) %>%
    set_engine("kernlab") %>%
    set_mode("classification")
  
  # Workflow setup
  rf_workflow <- workflow() %>%
    add_recipe(recipe_spec) %>%
    add_model(rf_spec)
  
  xgb_workflow <- workflow() %>%
    add_recipe(recipe_spec) %>%
    add_model(xgb_spec)
  
  svm_workflow <- workflow() %>%
    add_recipe(recipe_spec) %>%
    add_model(svm_spec)
  
  # Hyperparameter tuning grid
  rf_grid <- grid_regular(
    mtry(range = c(2, 10)),
    trees(range = c(500, 2000)),
    min_n(range = c(5, 20)),
    levels = 5
  )
  
  xgb_grid <- grid_regular(
    trees(range = c(100, 1000)),
    learn_rate(range = c(-3, -1), trans = log10_trans()),
    tree_depth(range = c(3, 10)),
    min_n(range = c(2, 10)),
    loss_reduction(range = c(-10, -1), trans = log10_trans()),
    levels = 5
  )
  
  svm_grid <- grid_regular(
    cost(range = c(-5, 5), trans = log10_trans()),
    sigma(range = c(-5, 5), trans = log10_trans()),
    levels = 5
  )
  
  # Model tuning and selection
  rf_tuned <- tune_grid(
    rf_workflow,
    resamples = cv_folds,
    grid = rf_grid,
    metrics = metric_set(roc_auc, accuracy, sens, spec),
    control = control_grid(save_pred = TRUE)
  )
  
  xgb_tuned <- tune_grid(
    xgb_workflow,
    resamples = cv_folds,
    grid = xgb_grid,
    metrics = metric_set(roc_auc, accuracy, sens, spec),
    control = control_grid(save_pred = TRUE)
  )
  
  svm_tuned <- tune_grid(
    svm_workflow,
    resamples = cv_folds,
    grid = svm_grid,
    metrics = metric_set(roc_auc, accuracy, sens, spec),
    control = control_grid(save_pred = TRUE)
  )
  
  # Select best models
  best_rf <- select_best(rf_tuned, "roc_auc")
  best_xgb <- select_best(xgb_tuned, "roc_auc")
  best_svm <- select_best(svm_tuned, "roc_auc")
  
  # Finalize workflows
  final_rf <- finalize_workflow(rf_workflow, best_rf)
  final_xgb <- finalize_workflow(xgb_workflow, best_xgb)
  final_svm <- finalize_workflow(svm_workflow, best_svm)
  
  # Fit final models
  rf_fit <- fit(final_rf, data = train_data)
  xgb_fit <- fit(final_xgb, data = train_data)
  svm_fit <- fit(final_svm, data = train_data)
  
  # Ensemble model
  ensemble_results <- tibble(
    rf_pred = predict(rf_fit, test_data, type = "prob")$.pred_High,
    xgb_pred = predict(xgb_fit, test_data, type = "prob")$.pred_High,
    svm_pred = predict(svm_fit, test_data, type = "prob")$.pred_High
  ) %>%
    mutate(
      ensemble_pred = (rf_pred + xgb_pred + svm_pred) / 3
    )
  
  # Model evaluation
  model_performance <- tibble(
    model = c("Random Forest", "XGBoost", "SVM", "Ensemble"),
    roc_auc = c(
      roc_auc_vec(test_data$churn_risk, predict(rf_fit, test_data, type = "prob")$.pred_High),
      roc_auc_vec(test_data$churn_risk, predict(xgb_fit, test_data, type = "prob")$.pred_High),
      roc_auc_vec(test_data$churn_risk, predict(svm_fit, test_data, type = "prob")$.pred_High),
      roc_auc_vec(test_data$churn_risk, ensemble_results$ensemble_pred)
    ),
    accuracy = c(
      accuracy_vec(test_data$churn_risk, predict(rf_fit, test_data)),
      accuracy_vec(test_data$churn_risk, predict(xgb_fit, test_data)),
      accuracy_vec(test_data$churn_risk, predict(svm_fit, test_data)),
      accuracy_vec(test_data$churn_risk, 
                   ifelse(ensemble_results$ensemble_pred > 0.5, "High", "Low"))
    )
  )
  
  # Feature importance (from Random Forest)
  feature_importance <- rf_fit %>%
    pull_workflow_fit() %>%
    extract_fit_parsnip() %>%
    vip::vi() %>%
    arrange(desc(Importance)) %>%
    head(20)
  
  return(list(
    models = list(
      random_forest = rf_fit,
      xgboost = xgb_fit,
      svm = svm_fit
    ),
    performance = model_performance,
    feature_importance = feature_importance,
    ensemble_predictions = ensemble_results
  ))
}

# Model deployment function
deploy_model <- function(model, new_data, output_path) {
  predictions <- predict(model, new_data)
  
  results <- new_data %>%
    select(customer_id) %>%
    bind_cols(predictions = predictions)
  
  write_csv(results, output_path)
  
  return(results)
}
```

### 3. Advanced Shiny 2.0 Application

```r
# Advanced Shiny 2.0 Enterprise Dashboard
library(shiny)
library(shinydashboard)
library(DT)
library(plotly)
library(shinycssloaders)
library(shinyWidgets)

# Enterprise Dashboard UI
enterprise_dashboard_ui <- fluidPage(
  theme = shinytheme("flatly"),
  
  titlePanel("Enterprise Analytics Dashboard"),
  
  sidebarLayout(
    sidebarPanel(
      width = 3,
      
      h4("Data Filters"),
      
      # Date range selector
      dateRangeInput(
        "date_range",
        "Select Date Range:",
        start = Sys.Date() - 90,
        end = Sys.Date(),
        format = "yyyy-mm-dd"
      ),
      
      # Region selector
      pickerInput(
        "region_selector",
        "Select Regions:",
        choices = c("All", "North", "South", "East", "West", "Central"),
        selected = "All",
        multiple = TRUE,
        options = list(
          `actions-box` = TRUE,
          `selected-text-format` = "count > 2"
        )
      ),
      
      # Customer segment selector
      pickerInput(
        "segment_selector",
        "Customer Segments:",
        choices = c("All", "Champions", "Loyal Customers", "Potential Loyalists", 
                   "New Customers", "At Risk", "Lost"),
        selected = "All",
        multiple = TRUE
      ),
      
      hr(),
      
      h4("Analysis Options"),
      
      # Analysis type selector
      radioButtons(
        "analysis_type",
        "Analysis Type:",
        choices = c("Customer Segmentation", "RFM Analysis", "Churn Prediction", 
                   "Sales Trends", "Product Performance"),
        selected = "Customer Segmentation"
      ),
      
      # Metric selector
      checkboxGroupInput(
        "metrics_to_show",
        "Metrics to Display:",
        choices = c("Revenue", "Transactions", "Customers", "Average Order Value"),
        selected = c("Revenue", "Transactions")
      ),
      
      hr(),
      
      actionButton(
        "refresh_data",
        "Refresh Data",
        icon = icon("refresh"),
        class = "btn-primary",
        style = "width: 100%"
      ),
      
      actionButton(
        "export_report",
        "Export Report",
        icon = icon("download"),
        class = "btn-success",
        style = "width: 100%"
      )
    ),
    
    mainPanel(
      width = 9,
      
      tabsetPanel(
        type = "tabs",
        
        # Overview Tab
        tabPanel(
          "Overview",
          fluidRow(
            column(3,
              valueBoxOutput("total_revenue", width = NULL)
            ),
            column(3,
              valueBoxOutput("total_customers", width = NULL)
            ),
            column(3,
              valueBoxOutput("total_transactions", width = NULL)
            ),
            column(3,
              valueBoxOutput("avg_order_value", width = NULL)
            )
          ),
          
          br(),
          
          fluidRow(
            column(6,
              plotlyOutput("revenue_trend_plot", height = "400px")
            ),
            column(6,
              plotlyOutput("customer_segment_plot", height = "400px")
            )
          ),
          
          br(),
          
          fluidRow(
            column(12,
              DT::dataTableOutput("customer_data_table", width = "100%")
            )
          )
        ),
        
        # Detailed Analysis Tab
        tabPanel(
          "Detailed Analysis",
          conditionalPanel(
            condition = "input.analysis_type == 'Customer Segmentation'",
            plotlyOutput("segmentation_analysis", height = "600px")
          ),
          
          conditionalPanel(
            condition = "input.analysis_type == 'RFM Analysis'",
            fluidRow(
              column(6,
                plotlyOutput("rfm_heatmap", height = "500px")
              ),
              column(6,
                plotlyOutput("rfm_scatter", height = "500px")
              )
            )
          ),
          
          conditionalPanel(
            condition = "input.analysis_type == 'Churn Prediction'",
            fluidRow(
              column(8,
                plotlyOutput("churn_prediction_plot", height = "500px")
              ),
              column(4,
                h4("Model Performance"),
                verbatimTextOutput("model_metrics"),
                br(),
                h4("Churn Risk Summary"),
                verbatimTextOutput("churn_summary")
              )
            )
          ),
          
          conditionalPanel(
            condition = "input.analysis_type == 'Sales Trends'",
            plotlyOutput("sales_trends_plot", height = "600px")
          ),
          
          conditionalPanel(
            condition = "input.analysis_type == 'Product Performance'",
            fluidRow(
              column(6,
                plotlyOutput("product_revenue_plot", height = "500px")
              ),
              column(6,
                plotlyOutput("product_category_plot", height = "500px")
              )
            )
          )
        ),
        
        # Customer Insights Tab
        tabPanel(
          "Customer Insights",
          fluidRow(
            column(4,
              selectInput(
                "customer_selector",
                "Select Customer:",
                choices = NULL,
                selected = NULL
              )
            ),
            column(4,
              actionButton(
                "analyze_customer",
                "Analyze Customer",
                icon = icon("search"),
                class = "btn-primary"
              )
            ),
            column(4,
              actionButton(
                "compare_customers",
                "Compare Customers",
                icon = icon("balance-scale"),
                class = "btn-info"
              )
            )
          ),
          
          br(),
          
          uiOutput("customer_details")
        )
      )
    )
  )
)

# Enterprise Dashboard Server Logic
enterprise_dashboard_server <- function(input, output, session) {
  
  # Load data
  customer_data <- reactive({
    req(input$date_range)
    
    # Simulate loading enterprise data
    data <- tibble(
      customer_id = 1:1000,
      segment = sample(c("Champions", "Loyal Customers", "Potential Loyalists", 
                       "New Customers", "At Risk", "Lost"), 1000, replace = TRUE),
      revenue = rnorm(1000, 5000, 2000),
      transactions = sample(1:100, 1000, replace = TRUE),
      last_purchase = Sys.Date() - sample(1:365, 1000, replace = TRUE)
    ) %>%
    filter(
      last_purchase >= input$date_range[1],
      last_purchase <= input$date_range[2]
    )
    
    if (!"All" %in% input$region_selector) {
      data <- data %>% filter(region %in% input$region_selector)
    }
    
    if (!"All" %in% input$segment_selector) {
      data <- data %>% filter(segment %in% input$segment_selector)
    }
    
    return(data)
  })
  
  # Update customer selector
  observe({
    data <- customer_data()
    updateSelectInput(session, "customer_selector", 
                     choices = setNames(data$customer_id, 
                                    paste("Customer", data$customer_id)))
  })
  
  # Value Boxes
  output$total_revenue <- renderValueBox({
    data <- customer_data()
    valueBox(
      paste0("$", round(sum(data$revenue), 0)),
      "Total Revenue",
      icon = icon("dollar-sign"),
      color = "green"
    )
  })
  
  output$total_customers <- renderValueBox({
    data <- customer_data()
    valueBox(
      nrow(data),
      "Total Customers",
      icon = icon("users"),
      color = "blue"
    )
  })
  
  output$total_transactions <- renderValueBox({
    data <- customer_data()
    valueBox(
      sum(data$transactions),
      "Total Transactions",
      icon = icon("shopping-cart"),
      color = "orange"
    )
  })
  
  output$avg_order_value <- renderValueBox({
    data <- customer_data()
    aov <- data$revenue / data$transactions
    valueBox(
      paste0("$", round(mean(aov), 2)),
      "Avg Order Value",
      icon = icon("chart-line"),
      color = "purple"
    )
  })
  
  # Plots
  output$revenue_trend_plot <- renderPlotly({
    data <- customer_data()
    
    plot_ly(
      data = data,
      x = ~last_purchase,
      y = ~revenue,
      type = 'scatter',
      mode = 'markers',
      marker = list(
        size = 10,
        color = ~segment,
        colorscale = 'Viridis'
      ),
      hoverinfo = 'text',
      text = ~paste('Customer:', customer_id, '<br>',
                   'Segment:', segment, '<br>',
                   'Revenue: $', round(revenue, 0))
    ) %>%
      layout(
        title = "Customer Revenue Trend",
        xaxis = list(title = "Last Purchase Date"),
        yaxis = list(title = "Revenue ($)"),
        hovermode = 'closest'
      )
  })
  
  output$customer_segment_plot <- renderPlotly({
    data <- customer_data() %>%
      count(segment) %>%
      mutate(percentage = n / sum(n) * 100)
    
    plot_ly(
      data = data,
      x = ~segment,
      y = ~n,
      type = 'bar',
      marker = list(color = ~n, colorscale = 'Blues'),
      text = ~paste(round(percentage, 1), '%'),
      textposition = 'auto'
    ) %>%
      layout(
        title = "Customer Segment Distribution",
        xaxis = list(title = "Customer Segment"),
        yaxis = list(title = "Number of Customers")
      )
  })
  
  # Data Table
  output$customer_data_table <- renderDT({
    data <- customer_data()
    
    datatable(
      data,
      options = list(
        pageLength = 25,
        scrollX = TRUE,
        dom = 'Bfrtip',
        buttons = c('copy', 'csv', 'excel', 'pdf', 'print')
      ),
      extensions = 'Buttons',
      selection = 'single',
      rownames = FALSE
    )
  })
  
  # Customer Details
  output$customer_details <- renderUI({
    req(input$customer_selector)
    
    customer_id <- input$customer_selector
    customer_info <- customer_data() %>% filter(customer_id == !!customer_id)
    
    tagList(
      h4("Customer Details"),
      fluidRow(
        column(6,
          box(
            title = "Customer Information",
            status = "primary",
            solidHeader = TRUE,
            width = 12,
            tableOutput("customer_info_table")
          )
        ),
        column(6,
          box(
            title = "Customer History",
            status = "info",
            solidHeader = TRUE,
            width = 12,
            plotlyOutput("customer_history_plot")
          )
        )
      )
    )
  })
}

# Run the application
shinyApp(enterprise_dashboard_ui, enterprise_dashboard_server)
