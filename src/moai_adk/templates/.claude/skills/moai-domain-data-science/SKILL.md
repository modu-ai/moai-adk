---

name: moai-domain-data-science
description: Data analysis, visualization, statistical modeling, and reproducible research workflows. Use when working on data science workflows scenarios.
allowed-tools:
  - Read
  - Bash
---

# Data Science Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for analytics and DS work |
| Trigger cues | Notebook workflows, data pipelines, feature engineering, experimentation plans. |
| Tier | 4 |

## What it does

Provides expertise in data analysis workflows, statistical modeling, data visualization, and reproducible research practices using Python (pandas, scikit-learn) or R (tidyverse).

## When to use

- Engages when analytics, experimentation, or data science implementation is requested.
- “Data analysis”, “Visualization”, “Statistical modeling”, “Reproducible research”
- Automatically invoked when working with data science projects
- Data science SPEC implementation (`/alfred:2-run`)

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
```markdown
- Orchestrate data prep → training → evaluation steps.
- Export metrics (precision/recall) to the Quality Report.
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- Google. "Rules of Machine Learning." https://developers.google.com/machine-learning/guides/rules-of-ml (accessed 2025-03-29).
- Netflix. "Metaflow: Human-Centric Framework for Data Science." https://metaflow.org/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (analysis testing)
- python-expert/r-expert (implementation)
- ml-expert (advanced modeling)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
