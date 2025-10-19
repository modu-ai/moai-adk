---
name: moai-domain-data-science
description: Data analysis, visualization, statistical modeling, and reproducible research workflows
allowed-tools:
  - Read
  - Bash
tier: 2
auto-load: "true"
---

# Data Science Expert

## What it does

Provides expertise in data analysis workflows, statistical modeling, data visualization, and reproducible research practices using Python (pandas, scikit-learn) or R (tidyverse).

## When to use

- "데이터 분석", "시각화", "통계 모델링", "재현 가능한 연구", "EDA", "회귀 분석", "시계열 분석", "가설 검정", "pandas", "scikit-learn"
- "Data analysis", "Visualization", "Statistical modeling", "Exploratory data analysis", "Regression", "Time series"
- Automatically invoked when working with data science projects
- Data science SPEC implementation (`/alfred:2-run`)

- "데이터 분석", "시각화", "통계 모델링", "재현 가능한 연구"
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

### Example 1: Exploratory Data Analysis (EDA)

```python
# @CODE:EDA-001: 데이터 탐색
import pandas as pd
import numpy as np
import seaborn as sns
import matplotlib.pyplot as plt

# 데이터 로드
df = pd.read_csv('sales.csv')

# 기본 통계
print(df.describe())
print(df.info())

# 결측값 확인
print(df.isnull().sum())

# 분포 시각화
plt.figure(figsize=(12, 4))

plt.subplot(1, 3, 1)
df['price'].hist(bins=50)
plt.title('Price Distribution')

plt.subplot(1, 3, 2)
df['quantity'].boxplot()
plt.title('Quantity Boxplot')

plt.subplot(1, 3, 3)
df.groupby('category')['price'].mean().plot(kind='bar')
plt.title('Average Price by Category')

plt.tight_layout()
plt.show()

# 상관관계 분석
correlation = df.corr()
sns.heatmap(correlation, annot=True)
```

### Example 2: Statistical Modeling

**Linear Regression**:
```python
# @CODE:REGRESSION-001: 선형 회귀
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import train_test_split
from sklearn.metrics import r2_score, mean_squared_error

# 데이터 준비
X = df[['feature1', 'feature2', 'feature3']]
y = df['target']

X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42
)

# 모델 학습
model = LinearRegression()
model.fit(X_train, y_train)

# 예측 및 평가
y_pred = model.predict(X_test)
r2 = r2_score(y_test, y_pred)
rmse = np.sqrt(mean_squared_error(y_test, y_pred))

print(f"R²: {r2:.4f}")
print(f"RMSE: {rmse:.4f}")

# 계수 해석
for feature, coeff in zip(X.columns, model.coef_):
    print(f"{feature}: {coeff:.4f}")
```

### Example 3: Time Series Analysis

```python
# @CODE:TIMESERIES-001: 시계열 분석
import pandas as pd
from statsmodels.tsa.seasonal import seasonal_decompose
from statsmodels.tsa.arima.model import ARIMA

# 시계열 데이터
ts = df.set_index('date')['value']

# 분해
decomposition = seasonal_decompose(ts, model='additive', period=12)
decomposition.plot()

# ARIMA 모델
model = ARIMA(ts, order=(1, 1, 1))
fitted_model = model.fit()

# 예측
forecast = fitted_model.get_forecast(steps=12)
print(forecast.summary_table())

# 시각화
plt.figure(figsize=(12, 6))
ts.plot(label='Original')
forecast.predicted_mean.plot(label='Forecast')
plt.fill_between(forecast.conf_int().index,
                 forecast.conf_int().iloc[:, 0],
                 forecast.conf_int().iloc[:, 1],
                 alpha=0.3)
plt.legend()
plt.show()
```

### Example 4: Reproducible Research

```python
# @CODE:REPRODUCIBLE-001: 재현 가능한 연구
# environment.yml (conda)
name: data-analysis
channels:
  - conda-forge
dependencies:
  - python=3.11
  - pandas=2.0
  - numpy=1.24
  - scikit-learn=1.2
  - matplotlib=3.6
  - jupyter

# 사용:
# conda env create -f environment.yml
# conda activate data-analysis
```

## Keywords

"데이터 분석", "시각화", "통계 모델링", "EDA", "시계열 분석", "회귀 분석", "가설 검정", "pandas", "scikit-learn", "reproducible research", "statistical inference"

## Reference

- Data analysis patterns: `.moai/memory/development-guide.md#데이터-분석`
- Statistical methods: CLAUDE.md#통계-분석-패턴
- Visualization guide: `.moai/memory/development-guide.md#데이터-시각화`

## Works well with

- moai-foundation-trust (분석 테스트)
- moai-domain-ml (고급 모델링)
- moai-lang-python (Python 구현)
