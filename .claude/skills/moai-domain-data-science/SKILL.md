---
name: moai-domain-data-science
description: Data analysis, visualization, statistical modeling, and reproducible
  research workflows
allowed-tools:
- Read
- Bash
---

# Data Science Expert

## What it does

Provides expertise in data analysis workflows, statistical modeling, data visualization, and reproducible research practices using Python (pandas, scikit-learn) or R (tidyverse).

## When to use

- "데이터 분석", "시각화", "통계 모델링", "재현 가능한 연구"
- Automatically invoked when working with data science projects
- Data science SPEC implementation (`/alfred:2-build`)

## How it works

**Data Analysis (Python)**:
- **pandas**: Data manipulation (DataFrames, groupby, merge)
- **numpy**: Numerical computing
- **scipy**: Scientific computing, statistics
- **statsmodels**: Statistical modeling

**Data Analysis (R)**:
- **tidyverse**: dplyr, ggplot2, tidyr
- **data.table**: High-performance data manipulation
- **caret**: Machine learning framework

**Visualization**:
- **matplotlib/seaborn**: Python plotting
- **plotly**: Interactive visualizations
- **ggplot2**: R grammar of graphics
- **D3.js**: Web-based visualizations

**Statistical Modeling**:
- **Hypothesis testing**: t-tests, ANOVA, chi-square
- **Regression**: Linear, logistic, polynomial
- **Time series**: ARIMA, seasonal decomposition
- **Bayesian inference**: PyMC3, Stan

**Reproducible Research**:
- **Jupyter notebooks**: Interactive analysis
- **R Markdown**: Literate programming
- **Version control**: Git for notebooks (nbstripout)
- **Environment management**: conda, renv

## Examples

### Example 1: Exploratory data analysis
User: "/alfred:2-build ANALYSIS-001"
Claude: (creates RED analysis test, GREEN pandas implementation, REFACTOR with visualizations)

### Example 2: Statistical modeling
User: "회귀 분석 모델 구축"
Claude: (implements linear regression with hypothesis testing)

## Works well with

- alfred-trust-validation (analysis testing)
- python-expert/r-expert (implementation)
- ml-expert (advanced modeling)
