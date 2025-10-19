---
name: moai-domain-ml
description: Machine learning model training, evaluation, deployment, and MLOps workflows
allowed-tools:
  - Read
  - Bash
tier: 4
auto-load: "false"
---

# ML Expert

## What it does

Provides expertise in machine learning model development, training, evaluation, hyperparameter tuning, deployment, and MLOps workflows for production ML systems.

## When to use

- "머신러닝 모델 개발", "모델 학습", "모델 배포", "MLOps"
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

## TDD Workflow for ML Development

### Phase 1: RED (Test - Model Contract)
```python
# @TEST:ML-001 | SPEC: SPEC-ML-001.md
import pytest
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split

def test_iris_classifier_accuracy():
    """모델 정확도 테스트 (RED 단계)"""
    X, y = load_iris(return_X_y=True)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

    model = IrisClassifier()
    model.train(X_train, y_train)

    accuracy = model.evaluate(X_test, y_test)
    assert accuracy >= 0.95  # 모델이 95% 이상 정확도를 달성해야 함
```

### Phase 2: GREEN (Implementation)
```python
# @CODE:ML-001 | SPEC: SPEC-ML-001.md | TEST: tests/test_ml_001.py
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score

class IrisClassifier:
    def __init__(self):
        self.model = RandomForestClassifier(n_estimators=100, random_state=42)

    def train(self, X_train, y_train):
        """모델 학습"""
        self.model.fit(X_train, y_train)

    def evaluate(self, X_test, y_test):
        """모델 평가"""
        predictions = self.model.predict(X_test)
        return accuracy_score(y_test, predictions)
```

### Phase 3: REFACTOR (Optimization)
```python
# @CODE:ML-001:REFACTOR | 교차 검증 추가
from sklearn.model_selection import cross_val_score

def train_with_validation(X, y):
    """교차 검증을 통한 안정적인 모델 학습"""
    model = RandomForestClassifier(n_estimators=100)
    scores = cross_val_score(model, X, y, cv=5)

    print(f"CV Scores: {scores}")
    print(f"Mean: {scores.mean():.4f} (+/- {scores.std():.4f})")

    # 최종 모델 학습
    model.fit(X, y)
    return model
```

## Examples

### Example 1: Classification Model Training (scikit-learn)

**RED (Test)**:
```python
# @TEST:ML-001 | SPEC: SPEC-ML-001.md
def test_customer_churn_prediction():
    """고객 이탈 예측 모델 테스트"""
    X_train, X_test, y_train, y_test = load_churn_data()

    model = ChurnPredictor()
    model.fit(X_train, y_train)

    precision = model.precision(X_test, y_test)
    recall = model.recall(X_test, y_test)

    assert precision >= 0.85  # 85% 이상 정밀도
    assert recall >= 0.80     # 80% 이상 재현율
```

**GREEN (Implementation)**:
```python
# @CODE:ML-001 | TEST: tests/test_churn.py
from sklearn.ensemble import GradientBoostingClassifier
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import precision_score, recall_score

class ChurnPredictor:
    def __init__(self):
        self.scaler = StandardScaler()
        self.model = GradientBoostingClassifier(n_estimators=100, learning_rate=0.1)

    def fit(self, X_train, y_train):
        X_scaled = self.scaler.fit_transform(X_train)
        self.model.fit(X_scaled, y_train)

    def precision(self, X_test, y_test):
        X_scaled = self.scaler.transform(X_test)
        predictions = self.model.predict(X_scaled)
        return precision_score(y_test, predictions)

    def recall(self, X_test, y_test):
        X_scaled = self.scaler.transform(X_test)
        predictions = self.model.predict(X_scaled)
        return recall_score(y_test, predictions)
```

**REFACTOR (Hyperparameter Tuning)**:
```python
# @CODE:ML-001:REFACTOR | 하이퍼파라미터 최적화
from sklearn.model_selection import GridSearchCV

def optimize_churn_model(X_train, y_train):
    """GridSearchCV를 통한 최적화"""
    param_grid = {
        'n_estimators': [50, 100, 200],
        'learning_rate': [0.01, 0.1, 0.5],
        'max_depth': [3, 5, 7]
    }

    grid_search = GridSearchCV(
        GradientBoostingClassifier(),
        param_grid,
        cv=5,
        n_jobs=-1
    )

    grid_search.fit(X_train, y_train)
    print(f"Best params: {grid_search.best_params_}")
    print(f"Best CV score: {grid_search.best_score_:.4f}")

    return grid_search.best_estimator_

# 효과: 정확도 78% → 91% (13% 향상)
```

### Example 2: Deep Learning with TensorFlow/Keras

**Image Classification Model**:
```python
# @CODE:ML-002 | 이미지 분류 모델
import tensorflow as tf
from tensorflow.keras import layers

def build_cnn_model(input_shape=(28, 28, 1), num_classes=10):
    """CNN 모델 구축 (MNIST 분류)"""
    model = tf.keras.Sequential([
        layers.Conv2D(32, (3, 3), activation='relu', input_shape=input_shape),
        layers.MaxPooling2D((2, 2)),
        layers.Conv2D(64, (3, 3), activation='relu'),
        layers.MaxPooling2D((2, 2)),
        layers.Conv2D(64, (3, 3), activation='relu'),
        layers.Flatten(),
        layers.Dense(64, activation='relu'),
        layers.Dropout(0.5),
        layers.Dense(num_classes, activation='softmax')
    ])

    model.compile(
        optimizer='adam',
        loss='sparse_categorical_crossentropy',
        metrics=['accuracy']
    )

    return model

# 학습
model = build_cnn_model()
history = model.fit(X_train, y_train, epochs=10, validation_split=0.2)
# 정확도: ~99% (MNIST)
```

### Example 3: Model Evaluation with Cross-Validation

**Before (Simple train/test split)**:
```python
# ❌ 문제: 단 한 번의 평가로 모델 신뢰도 낮음
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)
model.fit(X_train, y_train)
accuracy = model.score(X_test, y_test)
print(f"Accuracy: {accuracy:.4f}")  # 0.92 - 우연일 수도?
```

**After (Stratified K-Fold Cross-Validation)**:
```python
# ✅ 개선: 5번의 반복 평가로 모델 신뢰도 증가
from sklearn.model_selection import StratifiedKFold
from sklearn.metrics import classification_report

skf = StratifiedKFold(n_splits=5, shuffle=True, random_state=42)
scores = []

for train_idx, test_idx in skf.split(X, y):
    X_train, X_test = X[train_idx], X[test_idx]
    y_train, y_test = y[train_idx], y[test_idx]

    model = RandomForestClassifier()
    model.fit(X_train, y_train)

    score = model.score(X_test, y_test)
    scores.append(score)

print(f"CV Scores: {scores}")
print(f"Mean: {np.mean(scores):.4f} (+/- {np.std(scores):.4f})")
# 결과: [0.92, 0.94, 0.93, 0.91, 0.95]
# Mean: 0.9300 (+/- 0.0142) - 더 안정적인 평가!
```

### Example 4: Model Deployment with FastAPI

**Model Serving**:
```python
# @CODE:ML-004 | 모델 배포
import joblib
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

# 학습된 모델 로드
model = joblib.load("trained_model.pkl")
scaler = joblib.load("scaler.pkl")

class PredictionInput(BaseModel):
    features: list

class PredictionOutput(BaseModel):
    prediction: float
    probability: float

@app.post("/predict")
async def predict(input_data: PredictionInput) -> PredictionOutput:
    """모델 예측 API"""
    # 입력 데이터 전처리
    X = scaler.transform([input_data.features])

    # 예측
    prediction = model.predict(X)[0]
    probability = model.predict_proba(X)[0].max()

    return PredictionOutput(
        prediction=prediction,
        probability=probability
    )

# 사용:
# curl -X POST "http://localhost:8000/predict" \
#   -H "Content-Type: application/json" \
#   -d '{"features": [1.0, 2.0, 3.0, 4.0]}'
```

**MLflow Model Registry**:
```python
# @CODE:ML-004:DEPLOY | MLflow를 통한 버전 관리
import mlflow
from mlflow.models import MLflowError

# 모델 로깅
mlflow.sklearn.log_model(model, "churn_model")

# 모델 등록
result = mlflow.register_model("runs:/<RUN_ID>/churn_model", "churn-production")

# 스테이징 → 프로덕션 전환
client = mlflow.tracking.MlflowClient()
client.transition_model_version_stage(
    "churn-production",
    version=1,
    stage="Production"
)

print("모델 배포 완료! 프로덕션 준비됨 ✅")
```

## Keywords

"머신러닝", "모델 학습", "모델 평가", "하이퍼파라미터 튜닝", "cross-validation", "scikit-learn", "TensorFlow", "PyTorch", "모델 배포", "MLOps", "모델 버전 관리", "gradient boosting", "deep learning"

## Reference

- ML model training: `.moai/memory/development-guide.md#머신러닝-패턴`
- Model evaluation metrics: CLAUDE.md#모델-평가-지표
- MLOps workflows: `.moai/memory/development-guide.md#MLOps-파이프라인`

## Works well with

- moai-foundation-trust (모델 테스트 및 검증)
- moai-lang-python (Python 구현)
- moai-domain-data-science (데이터 준비 및 전처리)
- moai-domain-backend (모델 서빙 및 API)
- moai-domain-devops (모델 배포 및 모니터링)
