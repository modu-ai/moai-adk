# moai-domain-data-science - Working Examples

_Last updated: 2025-10-22_

> This document provides real-world examples of data science workflows, visualization patterns, statistical modeling, and reproducible research using modern Python tools (Pandas 2.2.0, NumPy 2.2.0, Jupyter 1.1.0).

---

## Table of Contents

1. [Example 1: Exploratory Data Analysis Pipeline](#example-1-exploratory-data-analysis-pipeline)
2. [Example 2: Statistical Modeling with TDD](#example-2-statistical-modeling-with-tdd)
3. [Example 3: Interactive Visualization Dashboard](#example-3-interactive-visualization-dashboard)
4. [Example 4: Reproducible Research Workflow](#example-4-reproducible-research-workflow)
5. [Example 5: Time Series Analysis](#example-5-time-series-analysis)
6. [Example 6: Data Quality Validation Pipeline](#example-6-data-quality-validation-pipeline)

---

## Example 1: Exploratory Data Analysis Pipeline

### Scenario
Analyze sales data to identify trends, outliers, and correlations for business insights.

### Project Structure
```
sales-analysis/
├── data/
│   ├── raw/
│   │   └── sales_2024.csv
│   └── processed/
│       └── cleaned_sales.parquet
├── notebooks/
│   ├── 01_data_loading.ipynb
│   ├── 02_cleaning.ipynb
│   └── 03_analysis.ipynb
├── src/
│   ├── data_loader.py
│   ├── preprocessor.py
│   └── analyzer.py
├── tests/
│   ├── test_data_loader.py
│   ├── test_preprocessor.py
│   └── test_analyzer.py
├── requirements.txt
└── pyproject.toml
```

### Implementation (TDD Flow)

#### RED: Write Failing Test
```python
# tests/test_preprocessor.py
# @TEST:DATA-001 | SPEC: SPEC-DATA-001.md | CODE: src/preprocessor.py

import pytest
import pandas as pd
from src.preprocessor import SalesPreprocessor

def test_remove_outliers_using_iqr():
    """Should remove outliers beyond 1.5 * IQR."""
    data = pd.DataFrame({
        'revenue': [100, 150, 200, 250, 300, 10000]  # 10000 is outlier
    })

    processor = SalesPreprocessor()
    result = processor.remove_outliers(data, 'revenue')

    assert len(result) == 5
    assert result['revenue'].max() <= 300

def test_handle_missing_dates():
    """Should forward-fill missing dates with last known value."""
    data = pd.DataFrame({
        'date': ['2024-01-01', '2024-01-02', '2024-01-04'],
        'revenue': [100, 200, 400]
    })
    data['date'] = pd.to_datetime(data['date'])

    processor = SalesPreprocessor()
    result = processor.fill_missing_dates(data)

    assert len(result) == 4
    assert result.loc[2, 'revenue'] == 200  # forward-filled
```

#### GREEN: Implement Feature
```python
# src/preprocessor.py
# @CODE:DATA-001 | SPEC: SPEC-DATA-001.md | TEST: tests/test_preprocessor.py

import pandas as pd
import numpy as np
from typing import Optional

class SalesPreprocessor:
    """
    Preprocess sales data: outlier removal, missing value handling, date normalization.

    @TAG:DATA-001
    """

    def remove_outliers(
        self,
        df: pd.DataFrame,
        column: str,
        threshold: float = 1.5
    ) -> pd.DataFrame:
        """Remove outliers using IQR method."""
        Q1 = df[column].quantile(0.25)
        Q3 = df[column].quantile(0.75)
        IQR = Q3 - Q1

        lower_bound = Q1 - threshold * IQR
        upper_bound = Q3 + threshold * IQR

        return df[
            (df[column] >= lower_bound) &
            (df[column] <= upper_bound)
        ].copy()

    def fill_missing_dates(
        self,
        df: pd.DataFrame,
        date_col: str = 'date'
    ) -> pd.DataFrame:
        """Forward-fill missing dates in time series."""
        df = df.set_index(date_col)

        # Create complete date range
        full_range = pd.date_range(
            start=df.index.min(),
            end=df.index.max(),
            freq='D'
        )

        # Reindex and forward-fill
        df = df.reindex(full_range, method='ffill')
        return df.reset_index()
```

#### REFACTOR: Add Profiling and Documentation
```python
# src/preprocessor.py (enhanced)

import pandas as pd
import numpy as np
from typing import Optional, Dict, Any
import logging

logger = logging.getLogger(__name__)

class SalesPreprocessor:
    """
    Preprocess sales data with comprehensive logging and profiling.

    Features:
    - Outlier detection (IQR method)
    - Missing value imputation
    - Date normalization
    - Data profiling

    @TAG:DATA-001
    """

    def __init__(self, log_level: str = 'INFO'):
        logging.basicConfig(level=log_level)
        self.processing_stats: Dict[str, Any] = {}

    def remove_outliers(
        self,
        df: pd.DataFrame,
        column: str,
        threshold: float = 1.5
    ) -> pd.DataFrame:
        """
        Remove outliers using IQR method.

        Args:
            df: Input DataFrame
            column: Column to analyze
            threshold: IQR multiplier (default: 1.5)

        Returns:
            DataFrame with outliers removed
        """
        original_count = len(df)

        Q1 = df[column].quantile(0.25)
        Q3 = df[column].quantile(0.75)
        IQR = Q3 - Q1

        lower_bound = Q1 - threshold * IQR
        upper_bound = Q3 + threshold * IQR

        result = df[
            (df[column] >= lower_bound) &
            (df[column] <= upper_bound)
        ].copy()

        removed = original_count - len(result)
        logger.info(f"Removed {removed} outliers from {column}")

        self.processing_stats['outliers_removed'] = removed
        return result

    def profile_data(self, df: pd.DataFrame) -> Dict[str, Any]:
        """Generate comprehensive data profile."""
        return {
            'shape': df.shape,
            'missing_values': df.isnull().sum().to_dict(),
            'dtypes': df.dtypes.to_dict(),
            'numeric_summary': df.describe().to_dict(),
            'memory_usage': df.memory_usage(deep=True).sum() / 1024**2  # MB
        }
```

#### Running the Pipeline
```bash
# Install dependencies
uv pip install pandas==2.2.0 numpy==2.2.0 matplotlib==3.9.0

# Run tests (pytest)
pytest tests/test_preprocessor.py -v --cov=src --cov-report=term-missing

# Execute notebook pipeline
jupyter notebook notebooks/01_data_loading.ipynb

# Run full pipeline script
python -m src.main --input data/raw/sales_2024.csv --output data/processed/
```

### Expected Output
```
=== Data Profiling Report ===
Shape: (1000, 8)
Missing values: {'revenue': 5, 'quantity': 0, 'date': 0}
Outliers removed: 12
Memory usage: 0.08 MB

=== Processing Complete ===
Cleaned data saved to: data/processed/cleaned_sales.parquet
```

---

## Example 2: Statistical Modeling with TDD

### Scenario
Build a linear regression model to predict customer churn based on behavior metrics.

### Implementation

#### RED: Test Model Training
```python
# tests/test_churn_model.py
# @TEST:DATA-002

import pytest
import pandas as pd
from sklearn.model_selection import train_test_split
from src.churn_model import ChurnPredictor

def test_model_training_improves_accuracy():
    """Model should achieve >80% accuracy on test set."""
    # Load synthetic data
    X = pd.DataFrame({
        'tenure': [1, 5, 10, 15, 20],
        'monthly_charges': [20, 50, 80, 100, 120],
        'total_charges': [20, 250, 800, 1500, 2400]
    })
    y = pd.Series([1, 1, 0, 0, 0])  # churn labels

    model = ChurnPredictor()
    model.fit(X, y)

    accuracy = model.evaluate(X, y)
    assert accuracy >= 0.80

def test_feature_importance_ranking():
    """Should identify most predictive features."""
    X = pd.DataFrame({...})  # training data
    y = pd.Series([...])

    model = ChurnPredictor()
    model.fit(X, y)

    importance = model.get_feature_importance()
    assert importance['tenure'] > 0.3  # tenure is key predictor
```

#### GREEN: Implement Model
```python
# src/churn_model.py
# @CODE:DATA-002

import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import accuracy_score, classification_report
from typing import Dict

class ChurnPredictor:
    """
    Logistic regression model for customer churn prediction.

    @TAG:DATA-002
    """

    def __init__(self):
        self.model = LogisticRegression(random_state=42)
        self.scaler = StandardScaler()
        self.feature_names = None

    def fit(self, X: pd.DataFrame, y: pd.Series):
        """Train model with feature scaling."""
        self.feature_names = X.columns.tolist()

        X_scaled = self.scaler.fit_transform(X)
        self.model.fit(X_scaled, y)

    def predict(self, X: pd.DataFrame) -> pd.Series:
        """Predict churn probability."""
        X_scaled = self.scaler.transform(X)
        return self.model.predict(X_scaled)

    def evaluate(self, X: pd.DataFrame, y: pd.Series) -> float:
        """Calculate accuracy score."""
        predictions = self.predict(X)
        return accuracy_score(y, predictions)

    def get_feature_importance(self) -> Dict[str, float]:
        """Extract feature coefficients."""
        coefficients = self.model.coef_[0]
        return dict(zip(self.feature_names, abs(coefficients)))
```

#### Running Model Training
```bash
# Train model
python -m src.train_churn_model --data data/churn.csv --output models/

# Evaluate on test set
python -m src.evaluate_model --model models/churn_v1.pkl --test-data data/test.csv

# Expected output:
# Accuracy: 0.847
# Precision: 0.82
# Recall: 0.79
# F1-Score: 0.80
```

---

## Example 3: Interactive Visualization Dashboard

### Scenario
Create an interactive dashboard to explore sales trends using Plotly.

### Implementation
```python
# src/dashboard.py
# @CODE:DATA-003

import pandas as pd
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots

class SalesDashboard:
    """
    Interactive sales visualization dashboard.

    @TAG:DATA-003
    """

    def __init__(self, data: pd.DataFrame):
        self.data = data

    def plot_revenue_trend(self) -> go.Figure:
        """Line chart showing revenue over time."""
        fig = px.line(
            self.data,
            x='date',
            y='revenue',
            title='Revenue Trend (2024)',
            labels={'revenue': 'Revenue ($)', 'date': 'Date'}
        )

        # Add moving average
        self.data['ma_30'] = self.data['revenue'].rolling(30).mean()
        fig.add_scatter(
            x=self.data['date'],
            y=self.data['ma_30'],
            mode='lines',
            name='30-day MA',
            line=dict(dash='dash')
        )

        return fig

    def plot_category_distribution(self) -> go.Figure:
        """Pie chart of sales by category."""
        category_totals = self.data.groupby('category')['revenue'].sum()

        fig = px.pie(
            values=category_totals.values,
            names=category_totals.index,
            title='Revenue by Category'
        )
        return fig

    def create_full_dashboard(self) -> go.Figure:
        """Multi-panel dashboard."""
        fig = make_subplots(
            rows=2, cols=2,
            subplot_titles=(
                'Revenue Trend',
                'Top Products',
                'Regional Performance',
                'Category Distribution'
            ),
            specs=[
                [{'type': 'scatter'}, {'type': 'bar'}],
                [{'type': 'choropleth'}, {'type': 'pie'}]
            ]
        )

        # Add traces for each subplot
        # ... (implementation details)

        fig.update_layout(height=800, showlegend=True)
        return fig

# Usage
dashboard = SalesDashboard(sales_data)
fig = dashboard.create_full_dashboard()
fig.write_html('reports/sales_dashboard.html')
```

---

## Example 4: Reproducible Research Workflow

### Scenario
Ensure research analysis is reproducible across environments using DVC and Jupyter.

### Project Structure
```
research-project/
├── .dvc/
│   └── config
├── data/
│   ├── raw.csv.dvc
│   └── processed.parquet.dvc
├── notebooks/
│   ├── 01_exploration.ipynb
│   └── 02_modeling.ipynb
├── src/
│   └── analysis.py
├── dvc.yaml           # Pipeline definition
├── params.yaml        # Hyperparameters
└── requirements.txt
```

### DVC Pipeline Configuration
```yaml
# dvc.yaml
stages:
  preprocess:
    cmd: python src/preprocess.py
    deps:
      - data/raw.csv
      - src/preprocess.py
    params:
      - preprocess.threshold
    outs:
      - data/processed.parquet

  train:
    cmd: python src/train.py
    deps:
      - data/processed.parquet
      - src/train.py
    params:
      - train.learning_rate
      - train.epochs
    outs:
      - models/model.pkl
    metrics:
      - metrics/train.json:
          cache: false
```

### Parameters File
```yaml
# params.yaml
preprocess:
  threshold: 1.5
  fill_method: 'ffill'

train:
  learning_rate: 0.01
  epochs: 100
  batch_size: 32
```

### Running Reproducible Pipeline
```bash
# Initialize DVC
dvc init

# Track data with DVC
dvc add data/raw.csv

# Run full pipeline
dvc repro

# Compare experiments
dvc params diff

# View metrics
dvc metrics show

# Version control
git add dvc.yaml params.yaml data/raw.csv.dvc
git commit -m "feat: add reproducible pipeline"
```

---

## Example 5: Time Series Analysis

### Scenario
Forecast monthly sales using ARIMA and Prophet models.

### Implementation
```python
# src/forecaster.py
# @CODE:DATA-005

import pandas as pd
from statsmodels.tsa.arima.model import ARIMA
from prophet import Prophet
from typing import Tuple

class SalesForecaster:
    """
    Time series forecasting with multiple models.

    @TAG:DATA-005
    """

    def __init__(self, data: pd.DataFrame):
        self.data = data
        self.arima_model = None
        self.prophet_model = None

    def fit_arima(self, order: Tuple[int, int, int] = (1, 1, 1)):
        """Fit ARIMA model."""
        self.arima_model = ARIMA(
            self.data['revenue'],
            order=order
        )
        self.arima_fitted = self.arima_model.fit()

    def fit_prophet(self):
        """Fit Prophet model."""
        prophet_data = self.data[['date', 'revenue']].rename(
            columns={'date': 'ds', 'revenue': 'y'}
        )

        self.prophet_model = Prophet(
            yearly_seasonality=True,
            weekly_seasonality=False,
            daily_seasonality=False
        )
        self.prophet_model.fit(prophet_data)

    def forecast(self, periods: int = 30) -> pd.DataFrame:
        """Generate forecast for next N periods."""
        # ARIMA forecast
        arima_forecast = self.arima_fitted.forecast(steps=periods)

        # Prophet forecast
        future = self.prophet_model.make_future_dataframe(periods=periods)
        prophet_forecast = self.prophet_model.predict(future)

        # Combine forecasts
        result = pd.DataFrame({
            'date': pd.date_range(
                start=self.data['date'].max() + pd.Timedelta(days=1),
                periods=periods
            ),
            'arima_forecast': arima_forecast,
            'prophet_forecast': prophet_forecast['yhat'].tail(periods).values
        })

        return result
```

---

## Example 6: Data Quality Validation Pipeline

### Scenario
Implement automated data quality checks using Great Expectations.

### Implementation
```python
# src/data_validator.py
# @CODE:DATA-006

import great_expectations as gx
from great_expectations.core.batch import RuntimeBatchRequest

class DataQualityValidator:
    """
    Automated data quality validation.

    @TAG:DATA-006
    """

    def __init__(self, context_root_dir: str = './gx'):
        self.context = gx.get_context(context_root_dir=context_root_dir)

    def create_expectation_suite(self, suite_name: str):
        """Define data quality expectations."""
        suite = self.context.add_expectation_suite(suite_name)

        validator = self.context.get_validator(
            batch_request=RuntimeBatchRequest(
                datasource_name="pandas_datasource",
                data_connector_name="runtime_data_connector",
                data_asset_name="sales_data",
                runtime_parameters={"batch_data": self.data},
                batch_identifiers={"default_identifier_name": "default_identifier"}
            ),
            expectation_suite_name=suite_name
        )

        # Add expectations
        validator.expect_column_values_to_not_be_null("revenue")
        validator.expect_column_values_to_be_between("revenue", min_value=0)
        validator.expect_column_values_to_be_in_set("category", ['A', 'B', 'C'])
        validator.expect_table_row_count_to_be_between(min_value=100)

        validator.save_expectation_suite()

    def validate_data(self, data, suite_name: str):
        """Run validation checkpoint."""
        checkpoint = self.context.add_checkpoint(
            name="sales_checkpoint",
            config_version=1.0,
            validations=[
                {
                    "batch_request": RuntimeBatchRequest(...),
                    "expectation_suite_name": suite_name
                }
            ]
        )

        result = checkpoint.run()
        return result.success
```

### Running Validation
```bash
# Initialize Great Expectations
great_expectations init

# Run validation
python -m src.validate_data --input data/sales.csv

# Generate data docs
great_expectations docs build

# Output:
# ✅ Validation passed: 15/15 expectations met
# 📊 Data Docs: file://./gx/uncommitted/data_docs/local_site/index.html
```

---

## Testing Strategy

### Unit Tests
```python
# tests/test_forecaster.py
import pytest
from src.forecaster import SalesForecaster

def test_arima_forecast_shape():
    """ARIMA forecast should return correct number of periods."""
    forecaster = SalesForecaster(mock_data)
    forecaster.fit_arima()

    result = forecaster.forecast(periods=30)
    assert len(result) == 30

def test_prophet_includes_confidence_intervals():
    """Prophet forecast should include upper/lower bounds."""
    forecaster = SalesForecaster(mock_data)
    forecaster.fit_prophet()

    result = forecaster.forecast(periods=30)
    assert 'yhat_lower' in result.columns
    assert 'yhat_upper' in result.columns
```

### Integration Tests
```bash
# Run full test suite
pytest tests/ -v --cov=src --cov-report=html

# Coverage target: ≥85%
# Expected output:
# tests/test_preprocessor.py::test_remove_outliers PASSED
# tests/test_churn_model.py::test_model_training PASSED
# Coverage: 87%
```

---

## Quality Gates (TRUST 5)

| Principle | Implementation | Tool |
|-----------|----------------|------|
| **Test** | Unit + integration tests, ≥85% coverage | pytest, pytest-cov |
| **Readable** | Type hints, docstrings, PEP 8 compliance | mypy, ruff, black |
| **Unified** | Consistent pandas/numpy idioms | pandas-vet |
| **Secured** | No hardcoded credentials, input validation | bandit |
| **Trackable** | @TAG references, changelog | grep, rg |

---

## Common Patterns

### Pattern 1: Pandas Idioms (Vectorized Operations)
```python
# ❌ Avoid loops
for i in range(len(df)):
    df.loc[i, 'total'] = df.loc[i, 'price'] * df.loc[i, 'quantity']

# ✅ Use vectorization
df['total'] = df['price'] * df['quantity']
```

### Pattern 2: Memory-Efficient Data Loading
```python
# ✅ Load data in chunks for large files
chunks = pd.read_csv('large_file.csv', chunksize=10000)
result = pd.concat([process_chunk(chunk) for chunk in chunks])
```

### Pattern 3: Profiling Performance
```python
import cProfile
import pstats

profiler = cProfile.Profile()
profiler.enable()

# Your code here
process_data(df)

profiler.disable()
stats = pstats.Stats(profiler)
stats.sort_stats('cumulative')
stats.print_stats(10)
```

---

_For detailed API references and tool configurations, see [reference.md](reference.md)_
