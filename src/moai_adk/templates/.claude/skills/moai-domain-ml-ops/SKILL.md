---
name: moai-domain-ml-ops
version: 4.0.0
created: 2025-11-18
updated: '2025-11-18'
status: stable
stability: stable
---


# Enterprise MLOps Platform - 

**Complete ML lifecycle orchestration with 2025's most stable stack**

> **Primary Agent**: backend-expert
> **Stack**: MLflow 3.6.0, DVC 3.x, Ray Serve 2.51.x, Kubeflow 1.10, Seldon Core 2.9.x, Feast 0.56.0, Optuna 4.6.0, Evidently AI, Prometheus 3.7.x, Grafana 11.3+

## Level 1: Quick Reference

### Core Technology Stack

**Experiment Tracking & Model Registry**:
- **MLflow 3.6.0**: Latest with AI Observability
- Model Registry for version management
- MLflow Gateway for model endpoints

**Data Versioning & Pipelines**:
- **DVC 3.x**: Cloud compute & versioning
- Pipeline stages (dvc.yaml)
- Remote storage integration (S3, GCS, Azure)

**Model Serving**:
- **Ray Serve 2.51.1**: Model multiplexing
- **Seldon Core 2.9.1**: KServe integration
- **Triton Inference Server 2.45.x**

**ML Orchestration**:
- **Kubeflow Pipelines 1.10**: Native Python SDK
- **Argo Workflows 3.4.x**: Backend orchestration

**Feature Management**:
- **Feast 0.56.0**: Online/offline stores
- Feature Registry & historical retrieval

**Optimization & Monitoring**:
- **Optuna 4.6.0**: Distributed hyperparameter tuning
- **Evidently 0.2.2+**: 100+ monitoring metrics
- **Prometheus 3.7.3** + **Grafana 11.3+**: Observability stack

### When to Use This Skill

- ✅ Setting up ML experiment tracking
- ✅ Implementing data versioning
- ✅ Deploying models at scale
- ✅ Building ML workflows
- ✅ Managing models on Kubernetes
- ✅ Creating feature stores
- ✅ Optimizing hyperparameters
- ✅ Monitoring model drift
- ✅ Automating ML CI/CD
- ✅ Building comprehensive observability

## Level 2: Practical Implementation

### MLflow 3.6.0 - Experiment Tracking

```python
# Setup
pip install mlflow==3.6.0 python-dotenv

# Start server
mlflow server --backend-store-uri postgresql://user:pass@localhost/mlflow \
              --default-artifact-root s3://bucket/mlflow-artifacts \
              --host 0.0.0.0 --port 5000
```

**Experiment Tracking**

```python
import mlflow
from sklearn.ensemble import RandomForestClassifier
from sklearn.datasets import load_breast_cancer
from sklearn.model_selection import train_test_split

mlflow.set_experiment("breast-cancer-classification")

with mlflow.start_run(run_name="rf-baseline-v1"):
    # Load and split data
    X, y = load_breast_cancer(return_X_y=True)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

    # Log parameters
    n_estimators = 100
    mlflow.log_param("n_estimators", n_estimators)
    mlflow.log_param("max_depth", 10)

    # Train and evaluate
    model = RandomForestClassifier(n_estimators=n_estimators, max_depth=10)
    model.fit(X_train, y_train)
    accuracy = model.score(X_test, y_test)

    # Log metrics and model
    mlflow.log_metric("accuracy", accuracy)
    mlflow.sklearn.log_model(model, "model",
                           registered_model_name="breast-cancer-classifier")
```

### DVC 3.x - Data Versioning

```bash
# Setup
pip install dvc==3.63.0 dvc-s3
dvc init
dvc remote add myremote s3://my-bucket/dvc-storage
dvc remote default myremote
```

**Pipeline Configuration (dvc.yaml)**

```yaml
stages:
  prepare:
    cmd: python src/prepare.py --input data/raw --output data/prepared
    deps: [data/raw, src/prepare.py]
    outs: [data/prepared]

  train:
    cmd: python src/train.py --features data/features --model model.pkl
    deps: [data/features, src/train.py]
    outs: [model.pkl]
    metrics: [metrics.json]
```

**Execute Pipeline**
```bash
dvc repro  # Run pipeline
dvc dag    # Show DAG
dvc metrics show  # View results
```

### Ray Serve 2.51.x - Model Serving

```python
# Setup
pip install ray[serve]==2.51.1

from ray import serve
import torch

serve.start()

@serve.deployment(num_replicas=2, autoscaling_config=min_replicas=1)
class ModelDeployment:
    def __init__(self, model_path: str):
        self.model = torch.load(model_path)
        self.model.eval()

    async def __call__(self, request):
        data = await request.json()
        prediction = self.model.predict(data["features"])
        return {"prediction": prediction.tolist()}

```

### Kubeflow Pipelines 1.10 - ML Workflows

```python
# Setup
pip install kfp==2.7.0 google-cloud-aiplatform

from kfp import dsl, compiler
from kfp.dsl import Dataset, Model, Artifact

@dsl.component(
    base_image="python:3.11-slim",
    packages_to_install=["pandas", "scikit-learn", "mlflow"]
)
def train_model(
    input_data: Dataset,
    model_output: Model,
    metrics_output: Output[dict]
):
    import pandas as pd
    from sklearn.ensemble import RandomForestClassifier
    from sklearn.model_selection import train_test_split
    import mlflow

    # Load data
    df = pd.read_csv(input_data.path)
    X = df.drop('target', axis=1)
    y = df['target']

    # Train model
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)
    model = RandomForestClassifier(n_estimators=100)
    model.fit(X_train, y_train)

    # Save model and metrics
    import joblib
    joblib.dump(model, model_output.path)

    metrics_output.update({
        "accuracy": model.score(X_test, y_test),
        "n_features": X.shape[1]
    })

@dsl.pipeline(name="ml-training-pipeline")
def ml_pipeline():
    data_task = load_data()
    train_task = train_model(input_data=data_task.outputs["data"])
```

### Feast 0.56.0 - Feature Store

```python
# Setup
pip install feast==0.56.0

# feature_store.yaml
project: my_project
registry: data/registry.db
provider: local
offline_store:
  type: file
online_store:
  type: redis
  connection_string: localhost:6379
```

```python
from feast import FeatureStore, Entity, FeatureView
from feast.types import Float32, Int64
from feast.data_source import PushSource

# Define entities and features
customer = Entity(name="customer", join_keys=["customer_id"])

customer_features = FeatureView(
    name="customer_features",
    entities=["customer"],
    schema=[
        Field(name="customer_id", dtype=Int64),
        Field(name="monthly_spend", dtype=Float32),
        Field(name="transaction_count", dtype=Int64)
    ],
    source=PushSource(name="customer_source")
)

# Initialize store
store = FeatureStore(repo_path=".")

# Get features for training
feature_service = store.get_feature_service("customer_features_v1")
training_data = store.get_historical_features(
    entity_df=customer_df,
    features=feature_service
).to_df()
```

## Level 3: Advanced Integration

### Evidently AI - Model Monitoring

```python
# Setup
pip install evidently==0.2.2

from evidently.report import Report
from evidently.metrics import DatasetDriftMetric, ClassificationPerformanceMetric

# Data drift detection
drift_report = Report(metrics=[
    DatasetDriftMetric(),
    ClassificationPerformanceMetric()
])

drift_report.run(
    reference_data=reference_df,
    current_data=current_data,
    column_mapping=column_mapping
)

# Save report
drift_report.save_html("drift_report.html")
```

### Optuna 4.6.0 - Hyperparameter Optimization

```python
# Setup
pip install optuna==4.6.0

import optuna

def objective(trial):
    # Suggest hyperparameters
    n_estimators = trial.suggest_int('n_estimators', 50, 300)
    max_depth = trial.suggest_int('max_depth', 3, 10)
    learning_rate = trial.suggest_float('learning_rate', 0.01, 0.3)

    # Train model
    model = XGBClassifier(
        n_estimators=n_estimators,
        max_depth=max_depth,
        learning_rate=learning_rate
    )

    # Cross-validation
    score = cross_val_score(model, X, y, cv=5).mean()
    return score

# Optimize
study = optuna.create_study(direction='maximize')
study.optimize(objective, n_trials=100)
```

### Prometheus + Grafana - Observability

```python
# Model metrics export
from prometheus_client import Counter, Histogram, Gauge, start_http_server

# Define metrics
model_predictions = Counter('model_predictions_total', 'Total model predictions')
model_latency = Histogram('model_inference_seconds', 'Model inference latency')
model_accuracy = Gauge('model_accuracy', 'Current model accuracy')

# Use in model serving
@model_latency.time()
def predict(features):
    model_predictions.inc()
    return model.predict(features)
```

## Installation Commands

```bash
# Core MLOps stack
pip install mlflow==3.6.0
pip install dvc==3.63.0 dvc-s3
pip install ray[serve]==2.51.1
pip install kfp==2.7.0
pip install feast==0.56.0
pip install optuna==4.6.0
pip install evidently==0.2.2

# Monitoring
pip install prometheus-client
pip install grafana-api

# Additional tools
pip install wandb  # Weights & Biases
pip install tensorboard
pip install mlflow-skinet  # MLflow UI enhancements
```

## Quick Architecture Patterns

### 1. Experiment → Registry → Production
```
MLflow Tracking → MLflow Registry → Ray Serve → Monitoring
     ↓               ↓              ↓         ↓
  Experiments      Model Versions  Deployments  Metrics
```

### 2. Data → Features → Model → Serving
```
DVC Versioning → Feast Features → Training → Ray Serve
     ↓              ↓              ↓         ↓
  Data Pipeline   Feature Store    Model     API Endpoint
```

### 3. CI/CD Pipeline
```
GitHub Actions → DVC Pipeline → MLflow → Registry → Production
       ↓              ↓           ↓          ↓         ↓
    Test & Build   Data Version  Experiments  Models  Deployment
```

## Best Practices

1. **Version Everything**: Data, models, code, experiments
2. **Automate Deployments**: Use CI/CD for model promotion
3. **Monitor Continuously**: Track drift, performance, latency
4. **Use Feature Stores**: Prevent training/serving skew
5. **Document Experiments**: MLflow tracking with proper tags
6. **Test Thoroughly**: Model validation before production
7. **Scale Gracefully**: Autoscaling and load balancing

---

**Version**: 4.0.0 Enterprise
**Last Updated**: 2025-11-18
**Status**: Production Ready
**Enterprise Grade**: ✅ Full Enterprise Support
