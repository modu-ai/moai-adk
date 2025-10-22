# moai-domain-data-science - Technical Reference

_Last updated: 2025-10-22_

> Comprehensive technical reference for data science workflows, tool configurations, API documentation, and best practices using Pandas 2.2.0, NumPy 2.2.0, Jupyter 1.1.0, and the broader Python data ecosystem.

---

## Table of Contents

1. [Tool Version Matrix](#tool-version-matrix)
2. [Project Structure Standards](#project-structure-standards)
3. [Testing Framework Configuration](#testing-framework-configuration)
4. [Pandas API Patterns](#pandas-api-patterns)
5. [NumPy Performance Optimization](#numpy-performance-optimization)
6. [Jupyter Notebook Best Practices](#jupyter-notebook-best-practices)
7. [Visualization Libraries](#visualization-libraries)
8. [Statistical Modeling Tools](#statistical-modeling-tools)
9. [Data Quality and Validation](#data-quality-and-validation)
10. [Reproducible Research Patterns](#reproducible-research-patterns)
11. [Performance Profiling](#performance-profiling)
12. [Common Anti-Patterns](#common-anti-patterns)

---

## Tool Version Matrix

### Core Libraries (2025-10-22)

| Tool | Version | Purpose | Installation | Status |
|------|---------|---------|--------------|--------|
| **Pandas** | 2.2.0 | Data manipulation and analysis | `uv pip install pandas==2.2.0` | ✅ Current |
| **NumPy** | 2.2.0 | Numerical computing foundation | `uv pip install numpy==2.2.0` | ✅ Current |
| **Jupyter** | 1.1.0 | Interactive computing environment | `uv pip install jupyter==1.1.0` | ✅ Current |
| **Matplotlib** | 3.9.0 | Static visualization | `uv pip install matplotlib==3.9.0` | ✅ Current |
| **Seaborn** | 0.13.2 | Statistical visualization | `uv pip install seaborn==0.13.2` | ✅ Current |
| **Plotly** | 5.24.0 | Interactive visualization | `uv pip install plotly==5.24.0` | ✅ Current |
| **SciPy** | 1.14.0 | Scientific computing algorithms | `uv pip install scipy==1.14.0` | ✅ Current |
| **scikit-learn** | 1.5.0 | Machine learning library | `uv pip install scikit-learn==1.5.0` | ✅ Current |
| **statsmodels** | 0.14.2 | Statistical modeling | `uv pip install statsmodels==0.14.2` | ✅ Current |

### Testing and Quality Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **pytest** | 8.3.0 | Test framework | `uv pip install pytest==8.3.0` |
| **pytest-cov** | 5.0.0 | Coverage reporting | `uv pip install pytest-cov==5.0.0` |
| **hypothesis** | 6.112.0 | Property-based testing | `uv pip install hypothesis==6.112.0` |
| **mypy** | 1.11.0 | Static type checking | `uv pip install mypy==1.11.0` |
| **ruff** | 0.6.0 | Fast linter and formatter | `uv pip install ruff==0.6.0` |
| **black** | 24.8.0 | Code formatter | `uv pip install black==24.8.0` |
| **pandas-vet** | 2024.8.10 | Pandas-specific linting | `uv pip install pandas-vet==2024.8.10` |

### Reproducibility and Workflow Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **DVC** | 3.55.0 | Data version control | `uv pip install dvc==3.55.0` |
| **papermill** | 2.6.0 | Parameterized notebook execution | `uv pip install papermill==2.6.0` |
| **great-expectations** | 1.1.0 | Data validation | `uv pip install great-expectations==1.1.0` |

---

## Project Structure Standards

### Recommended Directory Layout

```
data-science-project/
├── .dvc/                          # DVC configuration
├── .venv/                         # Virtual environment (uv managed)
├── data/
│   ├── raw/                       # Original, immutable data
│   ├── interim/                   # Intermediate processed data
│   ├── processed/                 # Final canonical datasets
│   └── external/                  # Third-party data sources
├── notebooks/
│   ├── exploratory/               # EDA notebooks (numbered)
│   │   ├── 01_data_overview.ipynb
│   │   └── 02_feature_engineering.ipynb
│   ├── modeling/                  # Model development
│   │   └── 03_baseline_model.ipynb
│   └── reporting/                 # Final reports
│       └── 04_final_analysis.ipynb
├── src/
│   ├── __init__.py
│   ├── data/
│   │   ├── __init__.py
│   │   ├── loader.py              # Data loading utilities
│   │   └── preprocessor.py        # Data cleaning/transformation
│   ├── features/
│   │   ├── __init__.py
│   │   └── engineering.py         # Feature engineering logic
│   ├── models/
│   │   ├── __init__.py
│   │   ├── train.py               # Model training
│   │   └── predict.py             # Inference logic
│   └── visualization/
│       ├── __init__.py
│       └── plots.py               # Reusable plotting functions
├── tests/
│   ├── conftest.py                # Shared test fixtures
│   ├── test_data/                 # Test data fixtures
│   ├── test_loader.py
│   ├── test_preprocessor.py
│   └── test_models.py
├── models/                        # Trained model artifacts
│   └── .gitkeep
├── reports/                       # Generated reports
│   ├── figures/                   # Saved visualizations
│   └── metrics.json               # Model performance metrics
├── .gitignore
├── dvc.yaml                       # DVC pipeline definition
├── params.yaml                    # Hyperparameters
├── pyproject.toml                 # Project configuration (uv)
├── README.md
└── requirements.txt               # Dependencies
```

### File Size and Complexity Limits

| Constraint | Limit | Rationale |
|------------|-------|-----------|
| Python file | ≤ 300 LOC | Maintainability |
| Function | ≤ 50 LOC | Testability |
| Parameters | ≤ 5 | Cognitive load |
| Cyclomatic complexity | ≤ 10 | Code clarity |
| Notebook cells | ≤ 20 lines | Readability |

---

## Testing Framework Configuration

### pyproject.toml Configuration

```toml
[project]
name = "data-science-project"
version = "0.1.0"
requires-python = ">=3.11"
dependencies = [
    "pandas==2.2.0",
    "numpy==2.2.0",
    "jupyter==1.1.0",
    "matplotlib==3.9.0",
    "scikit-learn==1.5.0",
]

[project.optional-dependencies]
dev = [
    "pytest==8.3.0",
    "pytest-cov==5.0.0",
    "hypothesis==6.112.0",
    "mypy==1.11.0",
    "ruff==0.6.0",
    "black==24.8.0",
]

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_functions = ["test_*"]
addopts = [
    "--strict-markers",
    "--cov=src",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-fail-under=85",
]

[tool.mypy]
python_version = "3.11"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
plugins = ["numpy.typing.mypy_plugin"]

[tool.ruff]
line-length = 88
select = ["E", "F", "I", "N", "W", "PD"]  # PD = pandas-vet rules
ignore = ["E203", "E501"]

[tool.ruff.per-file-ignores]
"tests/*" = ["F401", "F811"]

[tool.black]
line-length = 88
target-version = ['py311']
```

### pytest Configuration

```python
# tests/conftest.py
import pytest
import pandas as pd
import numpy as np

@pytest.fixture
def sample_dataframe():
    """Provide sample DataFrame for testing."""
    return pd.DataFrame({
        'date': pd.date_range('2024-01-01', periods=100),
        'value': np.random.randn(100),
        'category': np.random.choice(['A', 'B', 'C'], 100)
    })

@pytest.fixture
def sample_series():
    """Provide sample Series for testing."""
    return pd.Series(np.random.randn(100), name='test_series')
```

---

## Pandas API Patterns

### DataFrame Creation and Loading

```python
# CSV with type inference
df = pd.read_csv(
    'data.csv',
    dtype={'id': 'int64', 'category': 'category'},  # Optimize memory
    parse_dates=['date'],
    na_values=['NA', 'null', '']
)

# Parquet (preferred for processed data)
df = pd.read_parquet('data.parquet', engine='pyarrow')

# Excel with multiple sheets
dfs = pd.read_excel('data.xlsx', sheet_name=None)  # Returns dict

# JSON with nested structures
df = pd.read_json('data.json', orient='records', lines=True)
```

### Efficient Data Operations

```python
# ✅ Vectorized operations (fast)
df['total'] = df['price'] * df['quantity']
df['is_premium'] = df['price'] > 100

# ❌ Avoid iterrows (slow)
# for i, row in df.iterrows():
#     df.loc[i, 'total'] = row['price'] * row['quantity']

# ✅ Apply with NumPy vectorization
df['log_value'] = np.log(df['value'])

# ✅ Method chaining (readable)
result = (
    df
    .query('status == "active"')
    .groupby('category')
    .agg({'value': ['sum', 'mean', 'std']})
    .round(2)
)

# ✅ Memory-efficient chunking
chunks = pd.read_csv('large_file.csv', chunksize=10_000)
processed = pd.concat([process(chunk) for chunk in chunks])
```

### GroupBy Patterns

```python
# Multiple aggregations
summary = df.groupby('category').agg({
    'value': ['sum', 'mean', 'std', 'count'],
    'quantity': 'sum',
    'date': ['min', 'max']
})

# Custom aggregation functions
def iqr(x):
    return x.quantile(0.75) - x.quantile(0.25)

df.groupby('category')['value'].agg([iqr, 'median'])

# Transform (keep original shape)
df['value_zscore'] = df.groupby('category')['value'].transform(
    lambda x: (x - x.mean()) / x.std()
)

# Filter groups
df_filtered = df.groupby('category').filter(
    lambda x: len(x) >= 10  # Keep groups with ≥10 rows
)
```

### Time Series Operations

```python
# Set datetime index
df = df.set_index('date')

# Resample (aggregate to different frequency)
monthly = df.resample('MS').agg({  # MS = Month Start
    'value': 'sum',
    'quantity': 'mean'
})

# Rolling windows
df['ma_7'] = df['value'].rolling(window=7).mean()
df['std_30'] = df['value'].rolling(window=30).std()

# Expanding windows (cumulative)
df['cumsum'] = df['value'].expanding().sum()

# Shift for lag features
df['value_lag1'] = df['value'].shift(1)
df['value_lead1'] = df['value'].shift(-1)
```

---

## NumPy Performance Optimization

### Array Creation Best Practices

```python
import numpy as np

# ✅ Pre-allocate arrays (fast)
result = np.empty((1000, 1000))
for i in range(1000):
    result[i] = compute_row(i)

# ❌ Avoid dynamic growth (slow)
# result = []
# for i in range(1000):
#     result.append(compute_row(i))

# ✅ Use specialized constructors
zeros = np.zeros((100, 100))
ones = np.ones((100, 100))
identity = np.eye(100)
linspace = np.linspace(0, 10, 100)
```

### Broadcasting and Vectorization

```python
# Broadcasting (implicit shape expansion)
a = np.array([[1], [2], [3]])  # (3, 1)
b = np.array([10, 20, 30])     # (3,)
c = a + b  # (3, 3) - broadcast automatically

# Vectorized conditionals
values = np.array([1, 2, 3, 4, 5])
result = np.where(values > 3, 'high', 'low')

# Advanced indexing
data = np.random.randn(100)
data[data > 0] *= 2  # In-place modification
```

### Performance Profiling

```python
import timeit

# Compare performance
def loop_version():
    result = []
    for i in range(1000):
        result.append(i ** 2)
    return result

def vectorized_version():
    return np.arange(1000) ** 2

# Benchmark
loop_time = timeit.timeit(loop_version, number=1000)
vec_time = timeit.timeit(vectorized_version, number=1000)

print(f"Speedup: {loop_time / vec_time:.2f}x")
```

---

## Jupyter Notebook Best Practices

### Cell Organization

```python
# Cell 1: Imports (always at top)
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

%matplotlib inline
%load_ext autoreload
%autoreload 2

# Cell 2: Configuration
pd.set_option('display.max_columns', None)
pd.set_option('display.precision', 2)
sns.set_theme(style='whitegrid')

# Cell 3: Data loading
df = pd.read_csv('data.csv')

# Cell 4-N: Analysis steps (one logical operation per cell)
```

### Magic Commands

```python
# Timing
%time df.sort_values('value')  # Single execution
%timeit df.sort_values('value')  # Multiple runs, average

# Profiling
%prun df.groupby('category').sum()  # Line-by-line profiler

# Memory usage
%memit large_dataframe  # Memory consumption

# Debug
%debug  # Drop into debugger on exception

# System commands
!ls -la data/
!pip list | grep pandas
```

### Parameterized Notebooks with Papermill

```python
# notebook.ipynb (cell tagged as "parameters")
input_file = "data/raw.csv"
output_file = "data/processed.parquet"
threshold = 0.5

# Execute from command line
# papermill notebook.ipynb output.ipynb \
#   -p input_file data/new_raw.csv \
#   -p threshold 0.7
```

---

## Visualization Libraries

### Matplotlib Static Plots

```python
import matplotlib.pyplot as plt

# Figure with subplots
fig, axes = plt.subplots(2, 2, figsize=(12, 10))

axes[0, 0].hist(df['value'], bins=30, edgecolor='black')
axes[0, 0].set_title('Distribution')

axes[0, 1].scatter(df['x'], df['y'], alpha=0.5)
axes[0, 1].set_title('Scatter')

axes[1, 0].boxplot([df['group1'], df['group2']])
axes[1, 0].set_title('Boxplot')

axes[1, 1].plot(df.index, df['value'], label='Actual')
axes[1, 1].plot(df.index, df['predicted'], label='Predicted')
axes[1, 1].legend()

plt.tight_layout()
plt.savefig('reports/figures/overview.png', dpi=300, bbox_inches='tight')
```

### Seaborn Statistical Plots

```python
import seaborn as sns

# Pair plot (correlation matrix)
sns.pairplot(df, hue='category', diag_kind='kde')

# Heatmap (correlation)
corr = df.corr()
sns.heatmap(corr, annot=True, cmap='coolwarm', center=0)

# Categorical plots
sns.boxplot(data=df, x='category', y='value')
sns.violinplot(data=df, x='category', y='value', split=True)
```

### Plotly Interactive Dashboards

```python
import plotly.express as px
import plotly.graph_objects as go

# Interactive scatter
fig = px.scatter(
    df,
    x='x',
    y='y',
    color='category',
    size='value',
    hover_data=['id', 'date'],
    title='Interactive Scatter'
)
fig.write_html('reports/interactive.html')

# Subplots with different types
from plotly.subplots import make_subplots

fig = make_subplots(
    rows=2, cols=2,
    specs=[
        [{'type': 'scatter'}, {'type': 'bar'}],
        [{'type': 'box'}, {'type': 'heatmap'}]
    ]
)

fig.add_trace(
    go.Scatter(x=df['date'], y=df['value'], mode='lines'),
    row=1, col=1
)

fig.update_layout(height=800, showlegend=True)
```

---

## Statistical Modeling Tools

### scikit-learn Pipelines

```python
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from sklearn.compose import ColumnTransformer
from sklearn.linear_model import LogisticRegression

# Define preprocessing
numeric_features = ['age', 'income']
categorical_features = ['category', 'region']

numeric_transformer = Pipeline(steps=[
    ('scaler', StandardScaler())
])

categorical_transformer = Pipeline(steps=[
    ('onehot', OneHotEncoder(drop='first', sparse_output=False))
])

preprocessor = ColumnTransformer(
    transformers=[
        ('num', numeric_transformer, numeric_features),
        ('cat', categorical_transformer, categorical_features)
    ]
)

# Full pipeline
model = Pipeline(steps=[
    ('preprocessor', preprocessor),
    ('classifier', LogisticRegression())
])

# Fit and predict
model.fit(X_train, y_train)
predictions = model.predict(X_test)
```

### statsmodels Statistical Tests

```python
import statsmodels.api as sm
from statsmodels.formula.api import ols

# Linear regression
model = ols('value ~ x1 + x2 + C(category)', data=df).fit()
print(model.summary())

# Time series (ARIMA)
from statsmodels.tsa.arima.model import ARIMA

model = ARIMA(df['value'], order=(1, 1, 1))
fitted = model.fit()
forecast = fitted.forecast(steps=30)

# Statistical tests
from scipy.stats import ttest_ind, chi2_contingency

# T-test
t_stat, p_value = ttest_ind(group1, group2)

# Chi-squared test
chi2, p_value, dof, expected = chi2_contingency(contingency_table)
```

---

## Data Quality and Validation

### Great Expectations Configuration

```python
import great_expectations as gx

# Initialize context
context = gx.get_context()

# Create expectation suite
suite = context.add_expectation_suite("sales_suite")

# Define expectations
validator = context.get_validator(
    batch_request=batch_request,
    expectation_suite_name="sales_suite"
)

validator.expect_column_values_to_not_be_null("revenue")
validator.expect_column_values_to_be_between("revenue", min_value=0)
validator.expect_column_values_to_be_in_set("status", ['active', 'inactive'])
validator.expect_column_values_to_match_regex("email", r"^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$")

# Save suite
validator.save_expectation_suite()

# Run validation
checkpoint = context.add_checkpoint(
    name="sales_checkpoint",
    validations=[{"batch_request": batch_request}]
)

result = checkpoint.run()
```

---

## Reproducible Research Patterns

### DVC Pipeline Definition

```yaml
# dvc.yaml
stages:
  load_data:
    cmd: python src/data/loader.py
    outs:
      - data/raw/dataset.csv

  preprocess:
    cmd: python src/data/preprocessor.py
    deps:
      - data/raw/dataset.csv
      - src/data/preprocessor.py
    params:
      - preprocess.threshold
      - preprocess.fill_method
    outs:
      - data/processed/clean.parquet

  train:
    cmd: python src/models/train.py
    deps:
      - data/processed/clean.parquet
      - src/models/train.py
    params:
      - train.model_type
      - train.learning_rate
    outs:
      - models/model.pkl
    metrics:
      - metrics/train.json:
          cache: false

  evaluate:
    cmd: python src/models/evaluate.py
    deps:
      - models/model.pkl
      - data/processed/clean.parquet
    metrics:
      - metrics/test.json:
          cache: false
```

### Parameter Management

```yaml
# params.yaml
preprocess:
  threshold: 1.5
  fill_method: 'ffill'
  remove_outliers: true

train:
  model_type: 'logistic_regression'
  learning_rate: 0.01
  max_iter: 1000
  random_state: 42
```

---

## Performance Profiling

### cProfile for CPU Profiling

```python
import cProfile
import pstats
import io

profiler = cProfile.Profile()
profiler.enable()

# Code to profile
process_large_dataframe(df)

profiler.disable()

# Print stats
s = io.StringIO()
ps = pstats.Stats(profiler, stream=s).sort_stats('cumulative')
ps.print_stats(20)  # Top 20 functions
print(s.getvalue())
```

### Memory Profiling

```python
from memory_profiler import profile

@profile
def memory_intensive_function(df):
    large_copy = df.copy()
    result = large_copy.groupby('category').sum()
    return result

# Run with: python -m memory_profiler script.py
```

---

## Common Anti-Patterns

### Anti-Pattern 1: Iterating over Rows
```python
# ❌ BAD (slow)
for i, row in df.iterrows():
    df.loc[i, 'new_col'] = row['a'] + row['b']

# ✅ GOOD (fast)
df['new_col'] = df['a'] + df['b']
```

### Anti-Pattern 2: Dynamic DataFrame Growth
```python
# ❌ BAD (memory intensive)
result = pd.DataFrame()
for chunk in data_chunks:
    result = pd.concat([result, process(chunk)])

# ✅ GOOD (efficient)
result = pd.concat([process(chunk) for chunk in data_chunks])
```

### Anti-Pattern 3: Chained Assignment
```python
# ❌ BAD (SettingWithCopyWarning)
df[df['value'] > 0]['new_col'] = 1

# ✅ GOOD (explicit copy)
df.loc[df['value'] > 0, 'new_col'] = 1
```

---

_For practical examples and workflows, see [examples.md](examples.md)_
