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
- Google Cloud. "MLOps Continuous Delivery." https://cloud.google.com/architecture/mlops-continuous-delivery-and-automation-pipelines (accessed 2025-03-29).
- NVIDIA. "MLOps Best Practices." https://developer.nvidia.com/blog/category/ai/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (model testing)
- python-expert (ML implementation)
- data-science-expert (data preparation)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
