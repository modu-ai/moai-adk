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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- Google. "Rules of Machine Learning." https://developers.google.com/machine-learning/guides/rules-of-ml (accessed 2025-03-29).
- Netflix. "Metaflow: Human-Centric Framework for Data Science." https://metaflow.org/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (analysis testing)
- python-expert/r-expert (implementation)
- ml-expert (advanced modeling)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
