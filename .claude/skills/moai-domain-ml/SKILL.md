---
name: moai-domain-ml
description: Machine learning model training, evaluation, deployment, and MLOps workflows
allowed-tools:
- Read
- Bash
---

# ML Expert

## What it does

Provides expertise in machine learning model development, training, evaluation, hyperparameter tuning, deployment, and MLOps workflows for production ML systems.

## When to use

- "머신러닝 모델 개발", "모델 학습", "모델 배포", "MLOps"
- Automatically invoked when working with ML projects
- ML SPEC implementation (`/alfred:2-build`)

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

### Example 1: Model training with scikit-learn
User: "/alfred:2-build ML-001"
Claude: (creates RED model test, GREEN training pipeline, REFACTOR with cross-validation)

### Example 2: Model deployment
User: "모델 API 배포"
Claude: (creates FastAPI endpoint with model serving)

## Works well with

- alfred-trust-validation (model testing)
- python-expert (ML implementation)
- data-science-expert (data preparation)
