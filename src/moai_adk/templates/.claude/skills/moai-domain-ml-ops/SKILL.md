---
name: moai-domain-ml-ops
description: Enterprise MLOps orchestration platform with MLflow 3.6.0 model tracking & registry, DVC 3.x data versioning & pipelines, Ray Serve 2.51.x model serving, Kubeflow Pipelines 1.10 ML workflows, Seldon Core 2.9.x Kubernetes deployment, Feast 0.56.0 feature store, Optuna 4.6.0 hyperparameter tuning, Evidently AI model monitoring & drift detection, GitHub Actions ML CI/CD, Prometheus 3.7.x metrics, Grafana 11.3+ dashboards; activates for ML pipeline orchestration, experiment tracking, model registry, feature engineering, model serving, monitoring, hyperparameter optimization, and production ML infrastructure.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Enterprise MLOps Platform — v4.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Updated** | 2025-11-12 |
| **Stable Stack** | MLflow 3.6.0, DVC 3.x, Ray Serve 2.51.1, Kubeflow 1.10, Seldon Core 2.9.1, Feast 0.56.0, Optuna 4.6.0, Evidently 0.2.2+, Prometheus 3.7.3, Grafana 11.3+ |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | On-demand for MLOps, ML pipelines, model serving, experiment tracking, feature engineering, monitoring requests |
| **Trigger cues** | MLflow, DVC, Ray Serve, Kubeflow, Seldon, Feast, Optuna, Evidently, MLOps, experiment tracking, model registry, data versioning, feature store, model serving, hyperparameter tuning, ML monitoring, CI/CD for ML |

---

## Stable 2025 MLOps Stack

### Core Technologies (November 2025 Stable)

**Experiment Tracking & Model Registry**:
- MLflow 3.6.0 (latest, Nov 2025) with AI Observability
- MLflow Model Registry for version management
- MLflow Runs API for experiment tracking
- MLflow Gateway for model endpoints

**Data Versioning & Pipelines**:
- DVC 3.x with cloud compute & cloud versioning
- DVC Pipeline stages (dvc.yaml)
- DVC remote storage integration (S3, GCS, Azure)
- Model Registry in DVC

**Model Serving**:
- Ray Serve 2.51.1 with model multiplexing
- Seldon Core 2.9.1 with KServe integration
- KServe 0.13.x for standardized serving
- Triton Inference Server 2.45.x

**ML Pipelines & Orchestration**:
- Kubeflow Pipelines 1.10 with native Python SDK
- Kubeflow Model Registry (UI + API)
- Argo Workflows 3.4.x backend
- KFP v2 compiler with container steps

**Feature Management**:
- Feast 0.56.0 with online/offline stores
- Feast Python SDK v0.56.x
- Feature Registry & feature views
- Historical feature retrieval

**Hyperparameter Optimization**:
- Optuna 4.6.0 with distributed tuning
- Optuna Pruners & Samplers
- Ray Tune 2.51.1integration for distributed HPO
- Trial optimization with multi-objective support

**Model Monitoring**:
- Evidently 0.2.2+ with 100+ metrics
- Data drift detection (PSI, K-L divergence, Wasserstein)
- Model performance tracking
- Test suites & Report generation

**Metrics & Observability**:
- Prometheus 3.7.3 (latest, Oct 2025)
- Prometheus 2.55.x (legacy support)
- Grafana 11.3+ with Scenes dashboards
- Alertmanager 0.27.x for alert routing

**Infrastructure & CI/CD**:
- GitHub Actions for ML pipeline automation
- Kubernetes 1.30+ for container orchestration
- Docker 27.x for containerization
- MLflow Tracking Server (production deployment)

---

## 1. MLflow 3.6.0: Experiment Tracking & Model Registry

### MLflow Architecture: Tracking + Model Registry + Gateway

**Components**:
- **MLflow Tracking Server**: Centralized experiment logging
- **MLflow Model Registry**: Version management & promotion
- **MLflow Gateway**: Production endpoint serving
- **MLflow Client**: Python SDK for logging

### Setup & Configuration

```python
# 1. Install MLflow 3.6.0
pip install mlflow==3.6.0 python-dotenv

# 2. Start MLflow Tracking Server
mlflow server \
  --backend-store-uri postgresql://user:pass@localhost/mlflow \
  --default-artifact-root s3://bucket/mlflow-artifacts \
  --host 0.0.0.0 \
  --port 5000

# 3. Set environment variables
export MLFLOW_TRACKING_URI=http://localhost:5000
export MLFLOW_EXPERIMENT_NAME=ml-ops-experiments
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret
```

### Experiment Tracking with MLflow

```python
import mlflow
from mlflow import log_metric, log_param, log_artifact
import sklearn
from sklearn.ensemble import RandomForestClassifier
from sklearn.datasets import load_breast_cancer
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, precision_score, recall_score

# Set experiment
mlflow.set_experiment("breast-cancer-classification")

# Start run
with mlflow.start_run(run_name="rf-baseline-v1"):
    # Load data
    X, y = load_breast_cancer(return_X_y=True)
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42
    )
    
    # Log parameters
    n_estimators = 100
    max_depth = 10
    log_param("n_estimators", n_estimators)
    log_param("max_depth", max_depth)
    
    # Train model
    model = RandomForestClassifier(
        n_estimators=n_estimators,
        max_depth=max_depth,
        random_state=42
    )
    model.fit(X_train, y_train)
    
    # Evaluate
    y_pred = model.predict(X_test)
    accuracy = accuracy_score(y_test, y_pred)
    precision = precision_score(y_test, y_pred)
    recall = recall_score(y_test, y_pred)
    
    # Log metrics
    log_metric("accuracy", accuracy)
    log_metric("precision", precision)
    log_metric("recall", recall)
    
    # Log model
    mlflow.sklearn.log_model(
        model,
        artifact_path="models",
        input_example=X_test[:5],
        registered_model_name="breast-cancer-classifier"
    )
    
    print(f"Accuracy: {accuracy:.4f}")
```

### Model Registry & Promotion

```python
from mlflow.tracking import MlflowClient

client = MlflowClient(tracking_uri="http://localhost:5000")

# Register model version
model_uri = "runs:/abc123/models"
mv = mlflow.register_model(model_uri, "breast-cancer-classifier")

# Transition to Staging
client.transition_model_version_stage(
    name="breast-cancer-classifier",
    version=mv.version,
    stage="Staging"
)

# Run tests in Staging, then promote to Production
client.transition_model_version_stage(
    name="breast-cancer-classifier",
    version=mv.version,
    stage="Production"
)

# Get latest production model
production_version = client.get_latest_versions(
    "breast-cancer-classifier",
    stages=["Production"]
)[0]

print(f"Production model: {production_version.name} v{production_version.version}")
```

---

## 2. DVC 3.x: Data Versioning & Pipelines

### DVC Initialization & Configuration

```bash
# Install DVC
pip install dvc==3.63.0 dvc-s3

# Initialize DVC
dvc init

# Add remote storage
dvc remote add myremote s3://my-bucket/dvc-storage
dvc remote default myremote

# Configure credentials
dvc config core.autostage true
```

### Data Versioning Workflow

```bash
# Track dataset
dvc add datasets/raw_data.csv
git add datasets/raw_data.csv.dvc .gitignore
git commit -m "Add raw dataset v1"

# Download/sync data
dvc pull

# Update dataset and create new version
dvc add datasets/raw_data.csv
git add datasets/raw_data.csv.dvc
git commit -m "Update dataset to v2"

# Access specific version via git
git checkout v1.0  # Previous version
dvc checkout       # Restore corresponding data
```

### DVC Pipelines (dvc.yaml)

```yaml
# dvc.yaml
stages:
  prepare:
    cmd: python src/prepare.py --input data/raw --output data/prepared
    deps:
      - data/raw
      - src/prepare.py
    outs:
      - data/prepared
    
  featurize:
    cmd: python src/featurize.py --input data/prepared --output data/features
    deps:
      - data/prepared
      - src/featurize.py
    outs:
      - data/features
    
  train:
    cmd: python src/train.py --features data/features --model model.pkl
    deps:
      - data/features
      - src/train.py
    outs:
      - model.pkl
    metrics:
      - metrics.json:
          cache: false
    plots:
      - training_history.csv:
          x: epoch
          y: loss
```

### Execute Pipeline

```bash
# Run pipeline
dvc repro

# Show pipeline DAG
dvc dag

# Get pipeline metrics
dvc metrics show

# Generate pipeline plots
dvc plots show
```

---

## 3. Kubeflow Pipelines 1.10: ML Workflows on Kubernetes

### KFP v2 Python SDK Installation

```bash
pip install kfp==2.7.0 google-cloud-aiplatform
```

### Define ML Pipeline

```python
from kfp import dsl, compiler
from kfp.dsl import Dataset, Model, Artifact
import json

@dsl.component(
    base_image="python:3.11-slim",
    packages_to_install=["pandas", "scikit-learn"]
)
def prepare_data(
    raw_data: dsl.Input[Dataset],
    prepared_data: dsl.Output[Dataset]
):
    """Prepare raw data for training."""
    import pandas as pd
    
    df = pd.read_csv(raw_data.path)
    df = df.dropna()
    df.to_csv(prepared_data.path, index=False)

@dsl.component(
    base_image="python:3.11-slim",
    packages_to_install=["pandas", "scikit-learn"]
)
def train_model(
    training_data: dsl.Input[Dataset],
    model: dsl.Output[Model],
    metrics: dsl.Output[Artifact]
):
    """Train ML model."""
    import pandas as pd
    from sklearn.ensemble import RandomForestClassifier
    from sklearn.model_selection import train_test_split
    from sklearn.metrics import accuracy_score
    import pickle
    import json
    
    df = pd.read_csv(training_data.path)
    X = df.drop('target', axis=1)
    y = df['target']
    
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)
    
    clf = RandomForestClassifier(n_estimators=100)
    clf.fit(X_train, y_train)
    
    accuracy = accuracy_score(y_test, clf.predict(X_test))
    
    with open(model.path, 'wb') as f:
        pickle.dump(clf, f)
    
    with open(metrics.path, 'w') as f:
        json.dump({'accuracy': float(accuracy)}, f)

@dsl.pipeline(
    name="ml-training-pipeline",
    description="End-to-end ML training pipeline"
)
def ml_pipeline(
    raw_data_uri: str
):
    """Define the complete ML pipeline."""
    
    raw_data = dsl.Input[Dataset](raw_data_uri)
    
    prepare_task = prepare_data(raw_data=raw_data)
    
    train_task = train_model(
        training_data=prepare_task.outputs['prepared_data']
    )
    
    return train_task.outputs['model']

# Compile pipeline
compiler.Compiler().compile(
    pipeline_func=ml_pipeline,
    package_path="ml_pipeline.yaml"
)
```

### Deploy to Kubeflow

```bash
# Set Kubeflow namespace
export KUBEFLOW_NAMESPACE=kubeflow

# Submit pipeline run
kfp run create \
  --experiment-name ml-experiments \
  --run-name training-run-001 \
  --package-file ml_pipeline.yaml

# Monitor via Kubeflow UI
# Open: http://kubeflow-cluster/pipeline/
```

---

## 4. Ray Serve 2.51.x: Multi-Model Serving

### Ray Serve Installation & Setup

```bash
pip install ray[serve]==2.51.1 torch torchvision transformers
```

### Single Model Deployment

```python
from ray import serve
import torch
from transformers import pipeline

# Initialize Ray Serve
serve.start()

@serve.deployment(
    num_replicas=2,
    user_config={
        "model_id": "bert-base-uncased"
    }
)
class TextClassificationModel:
    def __init__(self):
        self.config = serve.get_deployment_config()
        model_id = self.config.user_config["model_id"]
        self.model = pipeline(
            "text-classification",
            model=model_id,
            device=0 if torch.cuda.is_available() else -1
        )
    
    async def __call__(self, request):
        text = request.query_params.get("text")
        result = self.model(text)
        return {"text": text, "classification": result}

# Deploy
TextClassificationModel.deploy()

# Query endpoint
import requests
response = requests.get(
    "http://localhost:8000/TextClassificationModel",
    params={"text": "This movie is great!"}
)
print(response.json())
```

### Multi-Model Serving with Model Multiplexing

```python
from ray import serve
from typing import Dict, List
import torch
from transformers import AutoModelForSequenceClassification, AutoTokenizer

serve.start()

@serve.deployment(
    num_replicas=1,
    ray_actor_options={"num_gpus": 1}
)
class MultiModelEndpoint:
    def __init__(self):
        self.models = {}
        self.tokenizers = {}
    
    def load_model(self, model_name: str):
        """Lazy load model on demand."""
        if model_name not in self.models:
            self.models[model_name] = AutoModelForSequenceClassification.from_pretrained(
                model_name
            )
            self.tokenizers[model_name] = AutoTokenizer.from_pretrained(model_name)
    
    async def __call__(self, request):
        model_name = request.query_params.get("model", "bert-base-uncased")
        text = request.query_params.get("text")
        
        self.load_model(model_name)
        
        inputs = self.tokenizers[model_name](text, return_tensors="pt")
        with torch.no_grad():
            outputs = self.models[model_name](**inputs)
        
        return {
            "model": model_name,
            "logits": outputs.logits.tolist()
        }

MultiModelEndpoint.deploy()
```

### A/B Testing with Ray Serve

```python
from ray import serve

serve.start()

@serve.deployment
class ModelV1:
    def __call__(self, request):
        return {"version": "v1", "prediction": "model_v1_output"}

@serve.deployment
class ModelV2:
    def __call__(self, request):
        return {"version": "v2", "prediction": "model_v2_output"}

@serve.deployment
class ABTestRouter:
    def __init__(self, model_v1_handle, model_v2_handle):
        self.model_v1 = model_v1_handle
        self.model_v2 = model_v2_handle
    
    async def __call__(self, request):
        import random
        if random.random() < 0.5:
            return await self.model_v1.remote(request)
        else:
            return await self.model_v2.remote(request)

# Deploy models
ModelV1.deploy()
ModelV2.deploy()

# Deploy router with references
ABTestRouter.deploy(
    ModelV1.get_handle(),
    ModelV2.get_handle()
)
```

---

## 5. Seldon Core 2.9.x: Kubernetes Model Serving

### Seldon Core Installation

```bash
kubectl create namespace seldon-system
helm repo add seldon https://storage.googleapis.com/seldon-charts
helm install seldon-core seldon/seldon-core-operator \
  --namespace seldon-system \
  --set usageMetrics.enabled=true
```

### Deploy Model with SeldonDeployment

```yaml
# seldon-model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris-model
  namespace: default
spec:
  name: iris-classifier
  predictors:
    - name: default
      replicas: 2
      componentSpecs:
        - spec:
            containers:
              - name: classifier
                image: gcr.io/my-project/iris-model:v1
                ports:
                  - containerPort: 5000
                resources:
                  requests:
                    memory: "256Mi"
                    cpu: "100m"
                  limits:
                    memory: "512Mi"
                    cpu: "500m"
      graph:
        name: classifier
        type: MODEL
```

### Canary Deployment (Gradual Rollout)

```yaml
# canary-deployment.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris-canary
spec:
  name: iris-classifier
  predictors:
    - name: default
      replicas: 1
      componentSpecs:
        - spec:
            containers:
              - name: classifier
                image: gcr.io/my-project/iris-model:v1
      graph:
        name: classifier
    
    - name: canary
      replicas: 1
      traffic: 10  # Route 10% traffic to canary
      componentSpecs:
        - spec:
            containers:
              - name: classifier
                image: gcr.io/my-project/iris-model:v2-canary
      graph:
        name: classifier
```

---

## 6. Feast 0.56.0: Feature Store

### Feast Project Setup

```bash
pip install feast==0.56.0 feast[postgres] boto3

feast init feature_repo
cd feature_repo
```

### Define Feature Views

```python
# features/features.py
from datetime import timedelta
from feast import Entity, FeatureView, Field, FileSource, PushSource
from feast.types import Float32, Int32, String
from feast.on_demand_feature_view import on_demand_feature_view
import pandas as pd

# Define entity
customer = Entity(
    name="customer_id",
    join_key="customer_id"
)

# Define data source
customer_data_source = FileSource(
    path="data/customers.csv",
    timestamp_column="created_at"
)

# Define feature view
customer_features = FeatureView(
    name="customer_features",
    entities=[customer],
    ttl=timedelta(days=30),
    features=[
        Field(name="age", dtype=Int32),
        Field(name="city", dtype=String),
        Field(name="account_balance", dtype=Float32),
    ],
    online=True,
    source=customer_data_source
)

# On-demand feature view for transformations
@on_demand_feature_view(
    sources=[customer_features],
    schema={
        "age_group": String,
    },
)
def customer_age_group(features_df: pd.DataFrame) -> pd.DataFrame:
    features_df["age_group"] = pd.cut(
        features_df["age"],
        bins=[0, 18, 35, 50, 100],
        labels=["child", "young", "middle", "senior"]
    )
    return features_df[["age_group"]]
```

### Offline Feature Retrieval

```python
from feast import FeatureStore

fs = FeatureStore(repo_path=".")

# Get historical features for training
training_entity_df = pd.DataFrame({
    "customer_id": [1, 2, 3],
    "event_timestamp": [
        "2023-01-01",
        "2023-01-02",
        "2023-01-03"
    ]
})

training_df = fs.get_historical_features(
    entity_df=training_entity_df,
    features=[
        "customer_features:age",
        "customer_features:city",
        "customer_age_group:age_group"
    ]
).to_df()
```

### Online Feature Retrieval

```python
# Get real-time features for inference
feature_vector = fs.get_online_features(
    features=[
        "customer_features:age",
        "customer_features:account_balance",
        "customer_age_group:age_group"
    ],
    entity_rows=[{"customer_id": 1}]
).to_dict()
```

---

## 7. Optuna 4.6.0: Hyperparameter Tuning at Scale

### Optuna Installation

```bash
pip install optuna==4.6.0 optuna-dashboard ray[tune]==2.51.1
```

### Single-Machine Distributed Tuning

```python
import optuna
from optuna.pruners import MedianPruner
from optuna.samplers import TPESampler
from sklearn.datasets import load_breast_cancer
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import cross_val_score

def objective(trial):
    # Suggest hyperparameters
    n_estimators = trial.suggest_int("n_estimators", 50, 300)
    max_depth = trial.suggest_int("max_depth", 5, 30)
    min_samples_split = trial.suggest_int("min_samples_split", 2, 20)
    
    # Load data
    X, y = load_breast_cancer(return_X_y=True)
    
    # Train model
    clf = RandomForestClassifier(
        n_estimators=n_estimators,
        max_depth=max_depth,
        min_samples_split=min_samples_split,
        random_state=42,
        n_jobs=-1
    )
    
    # Evaluate with cross-validation
    scores = cross_val_score(clf, X, y, cv=5, scoring="accuracy")
    
    return scores.mean()

# Create study with pruning
sampler = TPESampler(seed=42)
pruner = MedianPruner(n_startup_trials=5, n_warmup_steps=10)

study = optuna.create_study(
    direction="maximize",
    sampler=sampler,
    pruner=pruner
)

# Optimize
study.optimize(objective, n_trials=100, n_jobs=4)

# Best result
print(f"Best accuracy: {study.best_value:.4f}")
print(f"Best params: {study.best_params}")
```

### Multi-Objective Optimization

```python
def multi_objective(trial):
    """Optimize for accuracy and model size."""
    n_estimators = trial.suggest_int("n_estimators", 50, 500)
    max_depth = trial.suggest_int("max_depth", 5, 30)
    
    X, y = load_breast_cancer(return_X_y=True)
    
    clf = RandomForestClassifier(
        n_estimators=n_estimators,
        max_depth=max_depth,
        random_state=42
    )
    
    accuracy = cross_val_score(clf, X, y, cv=3, scoring="accuracy").mean()
    
    # Model size as secondary objective (minimize)
    model_size = n_estimators * max_depth
    
    return accuracy, -model_size

study = optuna.create_study(
    directions=["maximize", "maximize"],
    sampler=TPESampler(seed=42)
)

study.optimize(multi_objective, n_trials=50)

# Pareto frontier
print(f"Trials: {len(study.best_trials)}")
for trial in study.best_trials[:5]:
    print(f"Accuracy: {trial.values[0]:.4f}, Model size: {-trial.values[1]}")
```

---

## 8. Evidently AI: Model Monitoring & Drift Detection

### Evidently Setup

```bash
pip install evidently==0.4.24 pandas scikit-learn
```

### Data Drift Detection

```python
from evidently.report import Report
from evidently.metric_preset import DataDriftPreset
from evidently.metrics import ColumnDriftMetric
import pandas as pd
from sklearn.datasets import load_breast_cancer

# Load reference and production data
X, y = load_breast_cancer(return_X_y=True)
df = pd.DataFrame(X, columns=[f"feature_{i}" for i in range(X.shape[1])])

# Split into reference and production
reference_data = df.iloc[:300]
production_data = df.iloc[300:400]

# Create drift report
report = Report(metrics=[DataDriftPreset()])
report.run(reference_data=reference_data, current_data=production_data)

# Save report
report.save_html("drift_report.html")

# Access results
drift_results = report.as_dict()
print(f"Drift detected: {drift_results['metrics'][0]['result']['drift_detected']}")
```

### Model Performance Monitoring

```python
from evidently.report import Report
from evidently.metrics import (
    ClassificationDummyMetric,
    PrecisionByClassMetric,
    RecallByClassMetric,
    F1ByClassMetric
)

# Simulate predictions and actuals
y_pred = [0, 1, 1, 0, 1, 1, 0, 0, 1, 1]
y_true = [0, 1, 1, 1, 1, 1, 0, 0, 1, 0]

current_data = pd.DataFrame({
    "prediction": y_pred,
    "target": y_true
})

# Create performance report
report = Report(metrics=[
    ClassificationDummyMetric(),
    PrecisionByClassMetric(),
    RecallByClassMetric(),
    F1ByClassMetric()
])

report.run(current_data=current_data)
report.save_html("performance_report.html")
```

---

## 9. GitHub Actions for ML CI/CD

### ML Pipeline Automation

```yaml
# .github/workflows/ml-pipeline.yml
name: ML Training Pipeline

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  train:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
      
      - name: Run DVC pipeline
        run: |
          dvc pull
          dvc repro
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
      
      - name: Run tests
        run: pytest tests/ --cov=src --cov-report=term
      
      - name: Log metrics to MLflow
        run: python scripts/log_metrics.py
        env:
          MLFLOW_TRACKING_URI: ${{ secrets.MLFLOW_TRACKING_URI }}
          MLFLOW_TRACKING_USERNAME: ${{ secrets.MLFLOW_TRACKING_USERNAME }}
          MLFLOW_TRACKING_PASSWORD: ${{ secrets.MLFLOW_TRACKING_PASSWORD }}
      
      - name: Push updated model
        if: success()
        run: |
          git config user.name "ML Pipeline Bot"
          git config user.email "ml-bot@example.com"
          git add -A
          git commit -m "Update models and metrics"
          git push
```

### Model Deployment on PR

```yaml
# .github/workflows/deploy-model.yml
name: Deploy Model

on:
  push:
    branches: [ main ]
    paths:
      - 'models/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
      
      - name: Build Docker image
        run: |
          docker build -t model-serving:latest .
          docker tag model-serving:latest ${{ secrets.ECR_REGISTRY }}/model-serving:latest
      
      - name: Push to ECR
        run: |
          aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ${{ secrets.ECR_REGISTRY }}
          docker push ${{ secrets.ECR_REGISTRY }}/model-serving:latest
      
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/model-server \
            model-server=${{ secrets.ECR_REGISTRY }}/model-serving:latest \
            --record
          kubectl rollout status deployment/model-server
```

---

## 10. Production ML Architecture: Online/Batch/Streaming Inference

### Online Inference with MLflow Gateway

```python
import requests
import json

# Call MLflow Gateway endpoint
def online_predict(payload: dict) -> dict:
    response = requests.post(
        "http://mlflow-gateway:7002/invocations",
        json=payload,
        headers={"Content-Type": "application/json"}
    )
    return response.json()

# Example usage
result = online_predict({
    "inputs": {
        "feature_1": [1.0],
        "feature_2": [2.0]
    }
})
```

### Batch Inference with DVC

```python
import mlflow
import pandas as pd
from pathlib import Path

# Load model from registry
client = mlflow.tracking.MlflowClient()
model_uri = "models:/iris-classifier/Production"
model = mlflow.pyfunc.load_model(model_uri)

# Load batch data
batch_data = pd.read_csv("data/batch_data.csv")

# Predict
predictions = model.predict(batch_data)

# Save results
results_df = batch_data.copy()
results_df['prediction'] = predictions
results_df.to_csv("results/batch_predictions.csv", index=False)
```

### Streaming Inference with Kafka

```python
from kafka import KafkaConsumer, KafkaProducer
import mlflow
import json

model = mlflow.pyfunc.load_model("models:/iris-classifier/Production")

consumer = KafkaConsumer(
    'input-topic',
    bootstrap_servers=['kafka:9092'],
    value_deserializer=lambda m: json.loads(m.decode('utf-8'))
)

producer = KafkaProducer(
    bootstrap_servers=['kafka:9092'],
    value_serializer=lambda m: json.dumps(m).encode('utf-8')
)

for message in consumer:
    features = message.value
    prediction = model.predict([features])[0]
    
    producer.send('output-topic', {
        'input': features,
        'prediction': prediction,
        'timestamp': message.timestamp
    })
```

---

## 11. ML Observability: Prometheus + Grafana

### Export Model Metrics to Prometheus

```python
from prometheus_client import Counter, Histogram, Gauge, generate_latest
import time
from flask import Flask, Response

app = Flask(__name__)

# Define metrics
predictions_total = Counter(
    'model_predictions_total',
    'Total predictions',
    ['model_name', 'prediction_class']
)

prediction_latency = Histogram(
    'model_prediction_latency_seconds',
    'Prediction latency',
    ['model_name']
)

model_accuracy = Gauge(
    'model_accuracy',
    'Model accuracy',
    ['model_name', 'dataset']
)

@app.route('/predict', methods=['POST'])
def predict():
    start = time.time()
    
    # Model prediction
    prediction = model.predict(request.json)
    
    # Record metrics
    predictions_total.labels(
        model_name='iris-classifier',
        prediction_class=prediction
    ).inc()
    
    prediction_latency.labels(
        model_name='iris-classifier'
    ).observe(time.time() - start)
    
    return {'prediction': prediction}

@app.route('/metrics')
def metrics():
    return Response(generate_latest(), mimetype='text/plain')
```

### Grafana Dashboard (JSON)

```json
{
  "dashboard": {
    "title": "ML Model Monitoring",
    "panels": [
      {
        "title": "Prediction Rate",
        "targets": [
          {
            "expr": "rate(model_predictions_total[1m])"
          }
        ]
      },
      {
        "title": "Prediction Latency (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, model_prediction_latency_seconds_bucket)"
          }
        ]
      },
      {
        "title": "Model Accuracy",
        "targets": [
          {
            "expr": "model_accuracy"
          }
        ]
      }
    ]
  }
}
```

---

## Best Practices for Production MLOps

### 1. Experiment Tracking Discipline
- Log all hyperparameters before training
- Track metrics continuously during training
- Store artifacts (model, plots, metrics)
- Use meaningful experiment names and tags

### 2. Data Versioning Strategy
- Version all datasets with DVC
- Link data versions to model versions
- Track lineage from raw data to trained model
- Enable reproducibility via git + DVC checkout

### 3. Pipeline as Code
- Define pipelines in YAML (DVC) or Python (KFP)
- Version pipeline definitions in git
- Implement stage dependencies explicitly
- Monitor pipeline execution and logs

### 4. Model Registry Workflow
- Register trained models immediately
- Use semantic versioning (MAJOR.MINOR.PATCH)
- Require tests before production promotion
- Track model ancestry and lineage

### 5. Continuous Monitoring
- Export model metrics to Prometheus
- Set up Grafana dashboards for key metrics
- Implement drift detection with Evidently
- Create alerts for anomalies and performance drops

### 6. Hyperparameter Optimization
- Use distributed tuning (Optuna + Ray)
- Implement early stopping with pruning
- Track all trials and their results
- Archive best hyperparameters with models

### 7. Feature Store Governance
- Define features in code (Feast)
- Separate feature engineering from model training
- Maintain feature documentation
- Version feature views alongside models

### 8. CI/CD for ML
- Automate model training and testing
- Run data validation before training
- Execute model tests on new versions
- Deploy to staging before production

### 9. Containerization
- Package models with dependencies
- Use multi-stage Docker builds
- Tag images with model version
- Scan for vulnerabilities in supply chain

### 10. Infrastructure as Code
- Define Kubernetes manifests in git
- Use Helm for deployment templating
- Version all infrastructure changes
- Automate rollbacks on failures

---

## Troubleshooting & Common Issues

### MLflow Connection Issues
```python
# Verify MLflow server is running
import requests
try:
    response = requests.get("http://localhost:5000")
    print("MLflow server is up")
except:
    print("MLflow server is down")
```

### DVC Pipeline Failures
```bash
# Debug DVC pipeline
dvc dag  # Show pipeline structure
dvc repro --dry-run  # Dry run to check for issues
dvc status  # Check file status
dvc pull  # Ensure data is available
```

### Ray Serve Deployment Issues
```bash
# Check Ray Serve status
ray status  # Cluster status
ray dashboard  # Access dashboard at localhost:8265
ray logs actor  # Check actor logs
```

### Kubeflow Pod Crashes
```bash
# Debug Kubeflow pipelines
kubectl get pods -n kubeflow
kubectl describe pod <pod-name> -n kubeflow
kubectl logs <pod-name> -n kubeflow
```

---

## References & Official Documentation

- MLflow 3.6.0: https://mlflow.org/docs/latest/
- DVC 3.x: https://dvc.org/doc/
- Ray Serve 2.51.x: https://docs.ray.io/en/latest/serve/
- Kubeflow Pipelines 1.10: https://www.kubeflow.org/docs/
- Seldon Core 2.9.x: https://docs.seldon.io/
- Feast 0.56.0: https://docs.feast.dev/
- Optuna 4.6.0: https://optuna.readthedocs.io/
- Evidently: https://docs.evidentlyai.com/
- Prometheus 3.7.3: https://prometheus.io/docs/
- Grafana 11.3+: https://grafana.com/docs/

---

## Skill Update History

| Version | Date | Changes |
| ------- | ---- | ------- |
| 4.0.0 | 2025-11-12 | Updated to latest November 2025 stable versions (MLflow 3.6.0, Optuna 4.6.0, Kubeflow 1.10, Ray 2.51.1) |
| 3.0.0 | 2025-06-15 | Added Evidently AI, improved examples |
| 2.0.0 | 2025-01-01 | Initial Enterprise v2.0 release |

