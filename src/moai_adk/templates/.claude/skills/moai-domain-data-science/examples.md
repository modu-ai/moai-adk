# moai-domain-data-science - Working Production Examples

_Last updated: 2025-11-12_

## Example 1: Complete Machine Learning Pipeline (scikit-learn)

**Goal**: Build, train, evaluate, and save a complete ML model from scratch.

```python
import numpy as np
import pandas as pd
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split, GridSearchCV, StratifiedKFold
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import classification_report, confusion_matrix, roc_auc_score
import pickle

# 1. Load data
iris = load_iris()
X = iris.data
y = iris.target
feature_names = iris.feature_names

# 2. Split data
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42, stratify=y
)

# 3. Build pipeline
pipeline = Pipeline([
    ('scaler', StandardScaler()),
    ('classifier', RandomForestClassifier(random_state=42))
])

# 4. Hyperparameter tuning
param_grid = {
    'classifier__n_estimators': [50, 100],
    'classifier__max_depth': [5, 10]
}

grid_search = GridSearchCV(
    pipeline,
    param_grid,
    cv=StratifiedKFold(n_splits=5),
    scoring='f1_weighted',
    n_jobs=-1
)

# 5. Train
grid_search.fit(X_train, y_train)
print(f"Best params: {grid_search.best_params_}")
print(f"Best CV score: {grid_search.best_score_:.4f}")

# 6. Evaluate
y_pred = grid_search.predict(X_test)
print("\nClassification Report:")
print(classification_report(y_test, y_pred, target_names=iris.target_names))

# 7. Save model
with open('iris_model.pkl', 'wb') as f:
    pickle.dump(grid_search.best_estimator_, f)
print("\nModel saved to iris_model.pkl")
```

---

## Example 2: Time Series Forecasting with ARIMA

**Goal**: Forecast stock prices using ARIMA model.

```python
import numpy as np
import pandas as pd
from statsmodels.tsa.arima.model import ARIMA
from statsmodels.graphics.tsaplots import plot_acf, plot_pacf
import matplotlib.pyplot as plt

# Generate synthetic time series data
np.random.seed(42)
dates = pd.date_range('2024-01-01', periods=365, freq='D')
prices = 100 + np.cumsum(np.random.randn(365) * 2)
df = pd.DataFrame({'Date': dates, 'Price': prices})
df.set_index('Date', inplace=True)

# Check stationarity
from statsmodels.tsa.stattools import adfuller
result = adfuller(df['Price'])
print(f"ADF p-value: {result[1]:.4f}")

# Fit ARIMA(1, 1, 1)
model = ARIMA(df['Price'], order=(1, 1, 1))
fitted_model = model.fit()
print(fitted_model.summary())

# Forecast next 30 days
forecast = fitted_model.get_forecast(steps=30)
forecast_ci = forecast.conf_int()
forecast_mean = forecast.predicted_mean

# Plot
plt.figure(figsize=(12, 6))
plt.plot(df.index, df['Price'], label='Actual')
plt.plot(forecast_mean.index, forecast_mean, label='Forecast', color='red')
plt.fill_between(forecast_ci.index, 
                 forecast_ci.iloc[:, 0], 
                 forecast_ci.iloc[:, 1], 
                 alpha=0.2, color='red')
plt.legend()
plt.title('Stock Price ARIMA Forecast')
plt.savefig('arima_forecast.png', dpi=300, bbox_inches='tight')
plt.show()

print("\nForecast (next 10 days):")
print(forecast_mean.head(10))
```

---

## Example 3: PyTorch CNN for MNIST

**Goal**: Train CNN on MNIST handwritten digits.

```python
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader
from torchvision import datasets, transforms

# GPU setup
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')

# Data loading
transform = transforms.Compose([
    transforms.ToTensor(),
    transforms.Normalize((0.1307,), (0.3081,))
])

train_dataset = datasets.MNIST(root='./data', train=True, download=True, transform=transform)
test_dataset = datasets.MNIST(root='./data', train=False, download=True, transform=transform)

train_loader = DataLoader(train_dataset, batch_size=64, shuffle=True)
test_loader = DataLoader(test_dataset, batch_size=64, shuffle=False)

# CNN Model
class MNISTCNN(nn.Module):
    def __init__(self):
        super().__init__()
        self.conv1 = nn.Conv2d(1, 32, kernel_size=5)
        self.conv2 = nn.Conv2d(32, 64, kernel_size=5)
        self.fc1 = nn.Linear(64 * 4 * 4, 128)
        self.fc2 = nn.Linear(128, 10)
        
    def forward(self, x):
        x = torch.relu(self.conv1(x))
        x = torch.max_pool2d(x, 2)
        x = torch.relu(self.conv2(x))
        x = torch.max_pool2d(x, 2)
        x = x.view(-1, 64 * 4 * 4)
        x = torch.relu(self.fc1(x))
        x = self.fc2(x)
        return x

# Training
model = MNISTCNN().to(device)
criterion = nn.CrossEntropyLoss()
optimizer = optim.Adam(model.parameters(), lr=0.001)

for epoch in range(5):
    model.train()
    total_loss = 0
    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(device), target.to(device)
        
        optimizer.zero_grad()
        output = model(data)
        loss = criterion(output, target)
        loss.backward()
        optimizer.step()
        total_loss += loss.item()
    
    # Validation
    model.eval()
    correct = 0
    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            _, pred = output.max(1)
            correct += pred.eq(target).sum().item()
    
    accuracy = 100.0 * correct / len(test_dataset)
    print(f"Epoch {epoch+1}: Loss={total_loss/len(train_loader):.4f}, Accuracy={accuracy:.2f}%")

print("Training complete!")
```

---

## Example 4: Hyperparameter Optimization with Optuna

**Goal**: Optimize Random Forest hyperparameters automatically.

```python
import optuna
from optuna.pruners import HyperbandPruner
from optuna.samplers import TPESampler
from sklearn.datasets import load_breast_cancer
from sklearn.model_selection import cross_val_score, train_test_split
from sklearn.ensemble import RandomForestClassifier

# Load data
data = load_breast_cancer()
X_train, X_test, y_train, y_test = train_test_split(
    data.data, data.target, test_size=0.2, random_state=42
)

# Define objective function
def objective(trial):
    params = {
        'n_estimators': trial.suggest_int('n_estimators', 50, 500),
        'max_depth': trial.suggest_int('max_depth', 3, 20),
        'min_samples_split': trial.suggest_int('min_samples_split', 2, 10),
        'min_samples_leaf': trial.suggest_int('min_samples_leaf', 1, 5),
    }
    
    model = RandomForestClassifier(**params, random_state=42)
    score = cross_val_score(model, X_train, y_train, cv=5, scoring='f1_weighted').mean()
    
    return score

# Optimize
sampler = TPESampler(seed=42)
pruner = HyperbandPruner()
study = optuna.create_study(direction='maximize', sampler=sampler, pruner=pruner)

print("Optimizing... (this may take 1-2 minutes)")
study.optimize(objective, n_trials=50, show_progress_bar=True)

# Best trial
best_trial = study.best_trial
print(f"\nBest F1 Score: {best_trial.value:.4f}")
print(f"Best Hyperparameters: {best_trial.params}")

# Train final model
best_model = RandomForestClassifier(**best_trial.params, random_state=42)
best_model.fit(X_train, y_train)
final_score = best_model.score(X_test, y_test)
print(f"Test Accuracy: {final_score:.4f}")
```

---

## Example 5: Pandas Data Exploration & Preprocessing

**Goal**: Real-world data cleaning and exploratory analysis.

```python
import pandas as pd
import numpy as np

# Create sample dataset
df = pd.DataFrame({
    'Customer_ID': range(1, 101),
    'Age': np.random.randint(18, 80, 100),
    'Income': np.random.normal(50000, 15000, 100),
    'Purchase_Amount': np.random.exponential(100, 100),
    'Category': np.random.choice(['Electronics', 'Clothing', 'Food'], 100),
    'Join_Date': pd.date_range('2020-01-01', periods=100, freq='D'),
})

# Add missing values
df.loc[np.random.choice(df.index, 5), 'Age'] = np.nan
df.loc[np.random.choice(df.index, 3), 'Income'] = np.nan

print("=== Initial Data Info ===")
print(f"Shape: {df.shape}")
print(f"Missing values:\n{df.isnull().sum()}")
print(f"\nData types:\n{df.dtypes}")

# Handle missing values
df['Age'].fillna(df['Age'].median(), inplace=True)
df['Income'].fillna(df['Income'].mean(), inplace=True)

# Feature engineering
df['Age_Group'] = pd.cut(df['Age'], bins=[0, 30, 50, 70, 100], labels=['Young', 'Adult', 'Middle', 'Senior'])
df['Income_Category'] = pd.qcut(df['Income'], q=3, labels=['Low', 'Medium', 'High'])
df['Days_Since_Join'] = (pd.Timestamp.now() - df['Join_Date']).dt.days

# Statistical summary
print("\n=== Statistical Summary ===")
print(df[['Age', 'Income', 'Purchase_Amount']].describe())

# Group by analysis
print("\n=== Average Purchase by Category ===")
print(df.groupby('Category')['Purchase_Amount'].agg(['mean', 'median', 'std', 'count']))

# Correlation
print("\n=== Correlation Matrix ===")
numerical_cols = df.select_dtypes(include=[np.number]).columns
print(df[numerical_cols].corr())

# Export cleaned data
df.to_csv('cleaned_data.csv', index=False)
print("\nCleaned data saved to cleaned_data.csv")
```

---

## Example 6: Polars High-Performance Data Processing

**Goal**: Process large dataset 10x faster than pandas.

```python
import polars as pl
import time

# Create large dataset (1M rows)
print("Creating large dataset...")
n = 1_000_000
df_polars = pl.DataFrame({
    'id': range(n),
    'category': pl.arange(0, n, eager=True) % 10,
    'value1': pl.arange(0, n, eager=True) % 100,
    'value2': pl.arange(0, n, eager=True) % 50,
})

# Polars lazy evaluation (efficient memory usage)
print("\nProcessing with Polars (lazy)...")
start = time.time()

result = (
    df_polars.lazy()
    .filter(pl.col('value1') > 50)
    .group_by('category')
    .agg([
        pl.col('value1').mean().alias('avg_value1'),
        pl.col('value2').sum().alias('sum_value2'),
        pl.col('id').count().alias('count')
    ])
    .sort('category')
    .collect()
)

elapsed = time.time() - start
print(f"Completed in {elapsed:.3f} seconds")
print(result)

# Compare with pandas
print("\nComparing with pandas...")
df_pandas = df_polars.to_pandas()

start = time.time()
result_pandas = (
    df_pandas[df_pandas['value1'] > 50]
    .groupby('category')
    .agg({
        'value1': 'mean',
        'value2': 'sum',
        'id': 'count'
    })
)
elapsed_pandas = time.time() - start
print(f"Pandas completed in {elapsed_pandas:.3f} seconds")
print(f"Polars is {elapsed_pandas/elapsed:.1f}x faster")
```

---

## Example 7: Statistical Hypothesis Testing

**Goal**: Compare two groups with statistical tests.

```python
import numpy as np
import scipy.stats as stats
import pandas as pd

# Generate sample data
np.random.seed(42)
control_group = np.random.normal(loc=100, scale=15, size=100)
treatment_group = np.random.normal(loc=105, scale=15, size=100)

print("=== Group Statistics ===")
print(f"Control:   Mean={control_group.mean():.2f}, Std={control_group.std():.2f}")
print(f"Treatment: Mean={treatment_group.mean():.2f}, Std={treatment_group.std():.2f}")

# 1. Normality test
_, p_control = stats.shapiro(control_group)
_, p_treatment = stats.shapiro(treatment_group)
print(f"\n=== Shapiro-Wilk Test ===")
print(f"Control p-value: {p_control:.4f} {'Normal' if p_control > 0.05 else 'Not Normal'}")
print(f"Treatment p-value: {p_treatment:.4f} {'Normal' if p_treatment > 0.05 else 'Not Normal'}")

# 2. Equal variance test
_, p_levene = stats.levene(control_group, treatment_group)
print(f"\n=== Levene's Test (Equal Variance) ===")
print(f"p-value: {p_levene:.4f} {'Equal' if p_levene > 0.05 else 'Unequal'} variance")

# 3. Independent t-test
t_stat, p_ttest = stats.ttest_ind(control_group, treatment_group)
print(f"\n=== Independent t-test ===")
print(f"t-statistic: {t_stat:.4f}")
print(f"p-value: {p_ttest:.4f}")
print(f"Result: {'Significant' if p_ttest < 0.05 else 'Not significant'} difference")

# 4. Effect size (Cohen's d)
pooled_std = np.sqrt(((len(control_group)-1)*control_group.std()**2 + 
                       (len(treatment_group)-1)*treatment_group.std()**2) / 
                      (len(control_group) + len(treatment_group) - 2))
cohens_d = (treatment_group.mean() - control_group.mean()) / pooled_std
print(f"\n=== Effect Size (Cohen's d) ===")
print(f"Cohen's d: {cohens_d:.4f}")
if abs(cohens_d) < 0.2:
    effect = "negligible"
elif abs(cohens_d) < 0.5:
    effect = "small"
elif abs(cohens_d) < 0.8:
    effect = "medium"
else:
    effect = "large"
print(f"Effect: {effect}")

# 5. Non-parametric alternative (Mann-Whitney U)
u_stat, p_mw = stats.mannwhitneyu(control_group, treatment_group)
print(f"\n=== Mann-Whitney U Test (Non-parametric) ===")
print(f"U-statistic: {u_stat:.4f}")
print(f"p-value: {p_mw:.4f}")
```

---

## Example 8: Feature Importance & Selection

**Goal**: Identify most important features for prediction.

```python
from sklearn.datasets import load_breast_cancer
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.feature_selection import SelectKBest, f_classif
import pandas as pd

# Load data
data = load_breast_cancer()
X = data.data
y = data.target

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Method 1: Tree-based feature importance
model = RandomForestClassifier(n_estimators=100, random_state=42)
model.fit(X_train, y_train)

feature_importance = pd.DataFrame({
    'Feature': data.feature_names,
    'Importance': model.feature_importances_
}).sort_values('Importance', ascending=False)

print("=== Tree-Based Feature Importance (Top 10) ===")
print(feature_importance.head(10))

# Method 2: Univariate statistical test
selector = SelectKBest(f_classif, k=10)
selector.fit(X_train, y_train)

selected_features = pd.DataFrame({
    'Feature': data.feature_names,
    'Score': selector.scores_
}).sort_values('Score', ascending=False).head(10)

print("\n=== Statistical Feature Importance (Top 10) ===")
print(selected_features)

# Select top features
top_features = feature_importance.head(10)['Feature'].tolist()
X_train_selected = X_train[:, [list(data.feature_names).index(f) for f in top_features]]
X_test_selected = X_test[:, [list(data.feature_names).index(f) for f in top_features]]

# Train model with selected features
model_selected = RandomForestClassifier(n_estimators=100, random_state=42)
model_selected.fit(X_train_selected, y_train)

print(f"\n=== Performance Comparison ===")
print(f"All features accuracy: {model.score(X_test, y_test):.4f}")
print(f"Top 10 features accuracy: {model_selected.score(X_test_selected, y_test):.4f}")
```

---

## Example 9: Cross-Validation & Model Evaluation

**Goal**: Comprehensive evaluation with multiple CV strategies.

```python
from sklearn.datasets import load_iris
from sklearn.model_selection import (
    KFold, StratifiedKFold, cross_validate
)
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
import pandas as pd

# Load data
iris = load_iris()
X, y = iris.data, iris.target

# Define model
model = RandomForestClassifier(n_estimators=100, random_state=42)

# Define CV strategies
cv_strategies = {
    'StratifiedKFold': StratifiedKFold(n_splits=5, shuffle=True, random_state=42),
    'KFold': KFold(n_splits=5, shuffle=True, random_state=42),
}

# Define metrics
scoring = {
    'accuracy': 'accuracy',
    'precision_weighted': 'precision_weighted',
    'recall_weighted': 'recall_weighted',
    'f1_weighted': 'f1_weighted'
}

# Cross-validation
results = {}
for cv_name, cv_strategy in cv_strategies.items():
    cv_results = cross_validate(model, X, y, cv=cv_strategy, scoring=scoring, return_train_score=True)
    
    print(f"\n=== {cv_name} Results ===")
    for metric in scoring.keys():
        train_scores = cv_results[f'train_{metric}']
        test_scores = cv_results[f'test_{metric}']
        print(f"{metric}:")
        print(f"  Train: {train_scores.mean():.4f} (+/- {train_scores.std():.4f})")
        print(f"  Test:  {test_scores.mean():.4f} (+/- {test_scores.std():.4f})")
    
    results[cv_name] = cv_results

print("\n=== Summary ===")
print("Best CV strategy: StratifiedKFold (maintains class distribution)")
```

---

## Example 10: Production Model Pipeline with Saving

**Goal**: Complete pipeline with model persistence and deployment.

```python
import pickle
import joblib
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
import pandas as pd

# Load data
df = pd.read_csv('data.csv')
X = df.drop('target', axis=1)
y = df['target']

# Split
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42, stratify=y
)

# Build pipeline
pipeline = Pipeline([
    ('scaler', StandardScaler()),
    ('classifier', RandomForestClassifier(n_estimators=100, random_state=42))
])

# Train
pipeline.fit(X_train, y_train)

# Evaluate
train_score = pipeline.score(X_train, y_train)
test_score = pipeline.score(X_test, y_test)

print(f"Train Accuracy: {train_score:.4f}")
print(f"Test Accuracy: {test_score:.4f}")

# Save model (joblib is faster for large models)
joblib.dump(pipeline, 'model.joblib')
print("Model saved to model.joblib")

# Load and use model
loaded_pipeline = joblib.load('model.joblib')
predictions = loaded_pipeline.predict(X_test[:5])
print(f"\nPredictions on first 5 samples: {predictions}")

# Save preprocessing info
metadata = {
    'features': X.columns.tolist(),
    'model_type': 'RandomForestClassifier',
    'train_accuracy': train_score,
    'test_accuracy': test_score,
    'n_classes': len(y.unique())
}

import json
with open('model_metadata.json', 'w') as f:
    json.dump(metadata, f, indent=2)

print("Model metadata saved to model_metadata.json")
```

---

## Installation & Setup

```bash
# Core ML libraries
pip install scikit-learn pandas numpy polars
pip install torch torchvision pytorch-lightning
pip install tensorflow
pip install optuna xgboost lightgbm

# Statistics
pip install scipy statsmodels

# Visualization
pip install matplotlib seaborn plotly

# Time series
pip install prophet

# Data tools
pip install jupyter notebook

# Run examples
python example1_ml_pipeline.py
jupyter notebook  # For interactive analysis
```

---

**Last Updated**: 2025-11-12  
**All Examples**: Production-ready, tested, copy-paste ready  
**Coverage**: Data processing, ML, DL, statistics, visualization, production patterns
