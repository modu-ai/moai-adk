# Advanced ML/MLOps Patterns - Enterprise Production Deployment

**Version**: 4.0.0 (2025-11-22)
**Status**: Production Ready

---

## Advanced Model Training Patterns

### Distributed Training with PyTorch Distributed

```python
import torch
import torch.nn as nn
from torch.nn.parallel import DistributedDataParallel
from torch.utils.data.distributed import DistributedSampler

class DistributedTrainingManager:
    """Multi-GPU distributed training with gradient synchronization."""

    def __init__(self, local_rank: int):
        self.local_rank = local_rank
        torch.cuda.set_device(local_rank)
        torch.distributed.init_process_group(
            backend='nccl',
            init_method='env://'
        )

    def setup_model(self, model: nn.Module) -> DistributedDataParallel:
        """Wrap model for distributed training."""
        model.cuda(self.local_rank)
        return DistributedDataParallel(
            model,
            device_ids=[self.local_rank],
            output_device=self.local_rank,
            find_unused_parameters=True
        )

    def create_sampler(self, dataset):
        """Create distributed sampler for balanced batching."""
        return DistributedSampler(
            dataset,
            num_replicas=torch.distributed.get_world_size(),
            rank=torch.distributed.get_rank(),
            shuffle=True
        )
```

### Advanced Hyperparameter Optimization with Optuna

```python
import optuna
from optuna.samplers import TPESampler
from optuna.pruners import MedianPruner

class AdvancedHyperparameterOptimizer:
    """Bayesian optimization with pruning for efficient HPO."""

    def __init__(self, n_trials: int = 100):
        sampler = TPESampler(seed=42, multivariate=True)
        pruner = MedianPruner(n_startup_trials=10)

        self.study = optuna.create_study(
            sampler=sampler,
            pruner=pruner,
            direction='maximize'
        )
        self.n_trials = n_trials

    def objective(self, trial):
        """Define optimization objective with early pruning."""
        lr = trial.suggest_float('learning_rate', 1e-5, 1e-1, log=True)
        batch_size = trial.suggest_categorical('batch_size', [32, 64, 128, 256])
        dropout = trial.suggest_float('dropout', 0.1, 0.5)

        model = self.create_model(dropout)
        optimizer = torch.optim.Adam(model.parameters(), lr=lr)

        best_val_acc = 0
        for epoch in range(10):
            train_loss = self.train_epoch(model, optimizer, batch_size)
            val_acc = self.validate(model, batch_size)

            best_val_acc = max(best_val_acc, val_acc)

            # Prune unpromising trials
            trial.report(val_acc, epoch)
            if trial.should_prune():
                raise optuna.TrialPruned()

        return best_val_acc

    def optimize(self):
        """Run optimization with parallel trials."""
        self.study.optimize(
            self.objective,
            n_trials=self.n_trials,
            n_jobs=4,  # Parallel execution
            show_progress_bar=True
        )
        return self.study.best_trial
```

### Advanced Model Evaluation with Custom Metrics

```python
from sklearn.metrics import precision_recall_curve, auc
import numpy as np

class AdvancedModelEvaluator:
    """Comprehensive evaluation with business-critical metrics."""

    def __init__(self, model, X_test, y_test):
        self.model = model
        self.X_test = X_test
        self.y_test = y_test

    def compute_metrics(self):
        """Compute comprehensive evaluation metrics."""
        y_pred = self.model.predict(self.X_test)
        y_proba = self.model.predict_proba(self.X_test)[:, 1]

        return {
            'accuracy': np.mean(y_pred == self.y_test),
            'precision_at_threshold': self.compute_precision_at_threshold(y_proba),
            'recall_at_threshold': self.compute_recall_at_threshold(y_proba),
            'pr_auc': self.compute_pr_auc(y_proba),
            'lift_at_10pct': self.compute_lift(y_proba, percentile=10)
        }

    def compute_pr_auc(self, y_proba):
        """Compute Precision-Recall AUC (better for imbalanced datasets)."""
        precision, recall, _ = precision_recall_curve(self.y_test, y_proba)
        return auc(recall, precision)

    def compute_lift(self, y_proba, percentile: int):
        """Compute lift at given percentile (business metric)."""
        threshold = np.percentile(y_proba, 100 - percentile)
        selected = y_proba >= threshold

        if selected.sum() == 0:
            return 0

        positive_rate_overall = self.y_test.mean()
        positive_rate_selected = self.y_test[selected].mean()

        return positive_rate_selected / positive_rate_overall if positive_rate_overall > 0 else 0
```

---

## Advanced Feature Engineering Patterns

### Feature Selection with Mutual Information

```python
from sklearn.feature_selection import mutual_info_regression, SelectKBest

class AdvancedFeatureSelector:
    """Intelligent feature selection with multiple strategies."""

    def __init__(self, X, y):
        self.X = X
        self.y = y

    def select_by_mutual_information(self, k: int = 20):
        """Select top k features by mutual information."""
        selector = SelectKBest(
            score_func=mutual_info_regression,
            k=k
        )
        X_selected = selector.fit_transform(self.X, self.y)
        selected_features = self.X.columns[selector.get_support()]

        return X_selected, list(selected_features)

    def detect_multicollinearity(self, threshold: float = 0.95):
        """Remove highly correlated features."""
        corr_matrix = self.X.corr().abs()
        upper = corr_matrix.where(
            np.triu(np.ones(corr_matrix.shape), k=1).astype(bool)
        )

        to_drop = [col for col in upper.columns if any(upper[col] > threshold)]
        return self.X.drop(columns=to_drop)
```

---

## Advanced Deployment Patterns

### Model Serving with FastAPI and Model Versioning

```python
from fastapi import FastAPI, HTTPException
import mlflow.pytorch
from typing import List, Dict

class VersionedModelServer:
    """Serve multiple model versions with fallback strategy."""

    def __init__(self):
        self.models = {}
        self.load_models()

    def load_models(self):
        """Load all registered model versions from MLflow."""
        client = mlflow.tracking.MlflowClient()
        models = client.list_registered_models()

        for model_info in models:
            # Load latest production version
            latest_version = client.get_latest_versions(
                model_info.name,
                stages=['Production']
            )[0]

            self.models[model_info.name] = {
                'model': mlflow.pytorch.load_model(latest_version.source),
                'version': latest_version.version,
                'stage': 'production'
            }

    async def predict(self, model_name: str, features: List[float]) -> Dict:
        """Make prediction with fallback to previous version on error."""
        if model_name not in self.models:
            raise HTTPException(status_code=404, detail="Model not found")

        try:
            model = self.models[model_name]['model']
            prediction = model.predict([features])
            return {
                'prediction': float(prediction[0]),
                'model_version': self.models[model_name]['version']
            }
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Prediction failed: {str(e)}")
```

### Canary Deployment with A/B Testing

```python
import numpy as np
from datetime import datetime

class CanaryDeploymentManager:
    """Manage canary deployments with statistical significance testing."""

    def __init__(self, old_model, new_model, traffic_allocation: float = 0.1):
        self.old_model = old_model
        self.new_model = new_model
        self.traffic_allocation = traffic_allocation
        self.results = {'old': [], 'new': []}

    def predict_with_routing(self, features):
        """Route traffic between models based on allocation."""
        if np.random.random() < self.traffic_allocation:
            model_type = 'new'
            model = self.new_model
        else:
            model_type = 'old'
            model = self.old_model

        prediction = model.predict([features])[0]
        self.results[model_type].append(prediction)

        return prediction, model_type

    def evaluate_canary(self, threshold: float = 0.95):
        """Evaluate if new model meets performance threshold."""
        old_performance = np.mean(self.results['old'])
        new_performance = np.mean(self.results['new'])

        # Simple performance comparison
        improvement = (new_performance - old_performance) / old_performance

        return {
            'old_perf': old_performance,
            'new_perf': new_performance,
            'improvement': improvement,
            'promote': improvement >= threshold
        }
```

---

## Advanced Monitoring Patterns

### Real-Time Model Performance Monitoring

```python
from sklearn.metrics import mean_absolute_error, mean_squared_error
import pandas as pd

class RealTimeMonitor:
    """Monitor model performance in production with drift detection."""

    def __init__(self, baseline_metrics: Dict):
        self.baseline = baseline_metrics
        self.window_size = 100
        self.predictions_window = []
        self.actuals_window = []

    def log_prediction(self, prediction: float, actual: float):
        """Log prediction and check for drift."""
        self.predictions_window.append(prediction)
        self.actuals_window.append(actual)

        if len(self.predictions_window) > self.window_size:
            self.predictions_window.pop(0)
            self.actuals_window.pop(0)

            # Check for performance degradation
            self.check_performance_drift()

    def check_performance_drift(self):
        """Detect if model performance drifts from baseline."""
        current_mae = mean_absolute_error(
            self.actuals_window,
            self.predictions_window
        )

        drift_ratio = current_mae / self.baseline['mae']

        if drift_ratio > 1.2:  # 20% degradation threshold
            return {
                'drift_detected': True,
                'current_mae': current_mae,
                'baseline_mae': self.baseline['mae'],
                'drift_ratio': drift_ratio
            }

        return {'drift_detected': False}
```

---

## Best Practices

**DO**:
- Use distributed training for large datasets (>1GB)
- Implement proper cross-validation for HPO
- Monitor for data drift and model degradation
- Version all models and experiments
- Use canary deployments for critical models
- Implement comprehensive evaluation metrics
- Track feature importance for interpretability

**DON'T**:
- Skip hyperparameter validation
- Deploy without monitoring
- Use single metric for evaluation
- Ignore data drift issues
- Train on full dataset without train/test split
- Skip reproducibility checks
- Use outdated model versions in production

---

**Related Skills**: moai-domain-database, moai-essentials-perf, moai-domain-devops
