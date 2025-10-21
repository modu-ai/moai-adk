---

name: moai-domain-ml
description: Machine learning model training, evaluation, deployment, and MLOps workflows. Use when working on machine learning pipelines scenarios.
allowed-tools:
  - Read
  - Bash
---

# ML Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for ML lifecycle |
| Trigger cues | Model training, evaluation, deployment, MLOps guardrails. |
| Tier | 4 |

## What it does

Provides expertise in machine learning model development, training, evaluation, hyperparameter tuning, deployment, and MLOps workflows for production ML systems.

## When to use

- Engages when machine learning workflows or model operations are discussed.
- “Machine learning model development”, “model training”, “model deployment”, “MLOps”
- Automatically invoked when working with ML projects
- ML SPEC implementation (`/alfred:2-run`)

## How it works

**Model Training**:
- **scikit-learn**: Classical ML (RandomForest, SVM, KNN)
- **TensorFlow/Keras**: Deep learning (CNN, RNN, Transformers)
- **PyTorch**: Research-oriented deep learning
- **XGBoost/LightGBM**: Gradient boosting

**Model Evaluation**:
- **Classification**: Accuracy, precision, recall, F1, ROC-AUC
- **Regression**: RMSE, MAE, R²
- **Cross-validation**: k-fold, stratified k-fold
- **Confusion matrix**: Error analysis

**Hyperparameter Tuning**:
- **Grid search**: Exhaustive search
- **Random search**: Stochastic search
- **Bayesian optimization**: Optuna, Hyperopt
- **AutoML**: Auto-sklearn, TPOT

**Model Deployment**:
- **Serialization**: pickle, joblib, ONNX
- **Serving**: FastAPI, TensorFlow Serving, TorchServe
- **Containerization**: Docker for reproducibility
- **Versioning**: MLflow, DVC

**MLOps Workflows**:
- **Experiment tracking**: MLflow, Weights & Biases
- **Feature store**: Feast, Tecton
- **Model registry**: Centralized model management
- **Monitoring**: Data drift detection, model performance

## Examples
```markdown
- Trigger model training pipeline (e.g., `dvc repro`).
- Register artifact path in Completion Report.
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
- Google Cloud. "MLOps Continuous Delivery." https://cloud.google.com/architecture/mlops-continuous-delivery-and-automation-pipelines (accessed 2025-03-29).
- NVIDIA. "MLOps Best Practices." https://developer.nvidia.com/blog/category/ai/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (model testing)
- python-expert (ML implementation)
- data-science-expert (data preparation)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
