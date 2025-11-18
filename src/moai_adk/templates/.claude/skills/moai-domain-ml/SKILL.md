---
name: moai-domain-ml
version: 4.0.0
status: stable
description: "Enterprise Machine Learning specialist with TensorFlow 2.20.0, PyTorch\
  \ 2.9.0, \nScikit-learn 1.7.2 expertise. Master AutoML, neural architecture search,\
  \ MLOps \nautomation, and production ML deployment. Build scalable ML pipelines\
  \ with \ncomprehensive monitoring and experiment tracking.\n"
allowed-tools:
- Read
- Write
- Edit
- Bash
- Glob
- WebFetch
- WebSearch
tags:
- machine-learning
- tensorflow
- pytorch
- scikit-learn
- automl
- mlops
- deep-learning
updated: '2025-11-18'
stability: stable
---


# Enterprise Machine Learning

## Level 1: Quick Reference

### Core Capabilities
- **Deep Learning**: TensorFlow 2.20.0, PyTorch 2.9.0, JAX 0.4.33
- **Classical ML**: Scikit-learn 1.7.2, XGBoost 2.0.3, LightGBM 4.4.0
- **AutoML**: H2O AutoML 3.44.0, AutoGluon 1.0.0, TPOT 0.12.2
- **MLOps**: MLflow 2.9.0, Kubeflow 1.8.0, DVC 3.48.0
- **Deployment**: ONNX 1.16.0, TensorFlow Serving, TorchServe, Seldon Core

### Quick Setup Examples

```python
# TensorFlow 2.20.0 with modern Keras API
import tensorflow as tf
from tensorflow import keras

# Create a simple neural network
model = keras.Sequential([
    keras.layers.Dense(128, activation='relu', input_shape=(784,)),
    keras.layers.Dropout(0.2),
    keras.layers.Dense(64, activation='relu'),
    keras.layers.Dropout(0.2),
    keras.layers.Dense(10, activation='softmax')
])

model.compile(
    optimizer='adam',
    loss='sparse_categorical_crossentropy',
    metrics=['accuracy']
)

# Train with modern callbacks
callbacks = [
    keras.callbacks.EarlyStopping(patience=5, restore_best_weights=True),
    keras.callbacks.ReduceLROnPlateau(factor=0.5, patience=3),
    keras.callbacks.ModelCheckpoint('best_model.h5', save_best_only=True)
]

# model.fit(X_train, y_train, validation_data=(X_val, y_val), 
#           epochs=100, batch_size=32, callbacks=callbacks)
```

```python
# PyTorch 2.9.0 with modern features
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset

class NeuralNetwork(nn.Module):
    def __init__(self, input_size, hidden_size, num_classes):
        super().__init__()
        self.layers = nn.Sequential(
            nn.Linear(input_size, hidden_size),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(hidden_size, hidden_size // 2),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(hidden_size // 2, num_classes)
        )
    
    def forward(self, x):
        return self.layers(x)

# Initialize with device management
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
model = NeuralNetwork(784, 128, 10).to(device)
optimizer = optim.Adam(model.parameters(), lr=0.001)
criterion = nn.CrossEntropyLoss()

# Training loop with modern practices
# for epoch in range(epochs):
#     model.train()
#     for batch_x, batch_y in train_loader:
#         batch_x, batch_y = batch_x.to(device), batch_y.to(device)
#         optimizer.zero_grad()
#         outputs = model(batch_x)
#         loss = criterion(outputs, batch_y)
#         loss.backward()
#         optimizer.step()
```

## Level 2: Practical Implementation

### ML Pipeline Architecture

#### 1. Data Processing Pipeline

```python
# Modern data processing with scikit-learn 1.7.2
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import StandardScaler, OneHotEncoder
from sklearn.pipeline import Pipeline
from sklearn.impute import SimpleImputer
import pandas as pd
import numpy as np

class MlDataProcessor:
    def __init__(self):
        self.numeric_features = []
        self.categorical_features = []
        self.preprocessor = None
        self.feature_names = []
    
    def fit(self, X: pd.DataFrame):
        """Fit the data processor"""
        # Identify feature types
        self.numeric_features = X.select_dtypes(include=['int64', 'float64']).columns.tolist()
        self.categorical_features = X.select_dtypes(include=['object', 'category']).columns.tolist()
        
        # Create preprocessing pipeline
        numeric_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='median')),
            ('scaler', StandardScaler())
        ])
        
        categorical_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='most_frequent')),
            ('onehot', OneHotEncoder(handle_unknown='ignore', sparse_output=False))
        ])
        
        self.preprocessor = ColumnTransformer(
            transformers=[
                ('num', numeric_transformer, self.numeric_features),
                ('cat', categorical_transformer, self.categorical_features)
            ])
        
        # Fit and store feature names
        self.preprocessor.fit(X)
        self._generate_feature_names()
        
        return self
    
    def transform(self, X: pd.DataFrame) -> np.ndarray:
        """Transform the data"""
        return self.preprocessor.transform(X)
    
    def fit_transform(self, X: pd.DataFrame) -> np.ndarray:
        """Fit and transform in one step"""
        return self.fit(X).transform(X)
    
    def _generate_feature_names(self):
        """Generate feature names after transformation"""
        feature_names = []
        
        # Numeric features keep their names
        feature_names.extend(self.numeric_features)
        
        # Categorical features get prefixed names
        cat_transformer = self.preprocessor.named_transformers_['cat']
        cat_encoder = cat_transformer.named_steps['onehot']
        
        for i, feature in enumerate(self.categorical_features):
            categories = cat_encoder.categories_[i]
            feature_names.extend([f"{feature}_{cat}" for cat in categories])
        
        self.feature_names = feature_names
    
    def get_feature_names(self) -> list:
        """Get the transformed feature names"""
        return self.feature_names

# Usage example
# processor = MlDataProcessor()
# X_processed = processor.fit_transform(X_train)
# feature_names = processor.get_feature_names()
```

#### 2. Model Training with Experiment Tracking

```python
# MLflow experiment tracking integration
import mlflow
import mlflow.pytorch
import mlflow.tensorflow
from mlflow.tracking import MlflowClient
import json
from datetime import datetime

class ExperimentTracker:
    def __init__(self, experiment_name: str, tracking_uri: str = None):
        self.experiment_name = experiment_name
        self.tracking_uri = tracking_uri
        
        if tracking_uri:
            mlflow.set_tracking_uri(tracking_uri)
        
        mlflow.set_experiment(experiment_name)
        self.client = MlflowClient()
    
    def log_experiment(self, model, X_train, X_val, y_train, y_val, 
                      model_params: dict, metrics: dict, tags: dict = None):
        """Log a complete ML experiment"""
        
        with mlflow.start_run() as run:
            # Log parameters
            for param, value in model_params.items():
                mlflow.log_param(param, value)
            
            # Log tags
            if tags:
                mlflow.set_tags(tags)
            
            # Train and evaluate model
            model.fit(X_train, y_train)
            train_pred = model.predict(X_train)
            val_pred = model.predict(X_val)
            
            # Calculate and log metrics
            from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
            
            train_metrics = {
                'train_accuracy': accuracy_score(y_train, train_pred),
                'train_precision': precision_score(y_train, train_pred, average='weighted'),
                'train_recall': recall_score(y_train, train_pred, average='weighted'),
                'train_f1': f1_score(y_train, train_pred, average='weighted')
            }
            
            val_metrics = {
                'val_accuracy': accuracy_score(y_val, val_pred),
                'val_precision': precision_score(y_val, val_pred, average='weighted'),
                'val_recall': recall_score(y_val, val_pred, average='weighted'),
                'val_f1': f1_score(y_val, val_pred, average='weighted')
            }
            
            # Log all metrics
            all_metrics = {**train_metrics, **val_metrics, **metrics}
            for metric, value in all_metrics.items():
                mlflow.log_metric(metric, value)
            
            # Log model artifacts
            if hasattr(model, 'feature_importances_'):
                # Log feature importance for tree-based models
                feature_importance = dict(zip(range(len(model.feature_importances_)), 
                                           model.feature_importances_))
                mlflow.log_dict(feature_importance, 'feature_importance.json')
            
            # Log the model
            try:
                mlflow.sklearn.log_model(model, 'model')
            except:
                # Fallback for other model types
                mlflow.log_dict({'model_type': type(model).__name__}, 'model_info.json')
            
            return run.info.run_id
    
    def compare_experiments(self, metric: str = 'val_accuracy', top_n: int = 5):
        """Compare experiments and return top performers"""
        
        # Get experiment info
        experiment = self.client.get_experiment_by_name(self.experiment_name)
        runs = self.client.search_runs(
            experiment_ids=[experiment.experiment_id],
            order_by=[f"metrics.{metric} DESC"]
        )
        
        # Extract relevant information
        results = []
        for run in runs[:top_n]:
            results.append({
                'run_id': run.info.run_id,
                'status': run.info.status,
                'start_time': datetime.fromtimestamp(run.info.start_time / 1000),
                'end_time': datetime.fromtimestamp(run.info.end_time / 1000) if run.info.end_time else None,
                'metrics': run.data.metrics,
                'params': run.data.params,
                'tags': run.data.tags
            })
        
        return results

# Usage example
# tracker = ExperimentTracker("text_classification_experiments")
# run_id = tracker.log_experiment(
#     model=rf_classifier,
#     X_train=X_train, X_val=X_val, y_train=y_train, y_val=y_val,
#     model_params={'n_estimators': 100, 'max_depth': 10},
#     metrics={'training_time': 45.2},
#     tags={'model_type': 'random_forest', 'dataset_version': 'v1.2'}
# )
```

#### 3. Model Evaluation and Validation

```python
# Comprehensive model evaluation framework
from sklearn.model_selection import cross_val_score, learning_curve, validation_curve
from sklearn.metrics import classification_report, confusion_matrix, roc_auc_score
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np

class ModelEvaluator:
    def __init__(self):
        self.results = {}
    
    def evaluate_classification(self, model, X_test, y_test, class_names=None):
        """Comprehensive classification evaluation"""
        
        # Make predictions
        y_pred = model.predict(X_test)
        y_pred_proba = None
        
        if hasattr(model, 'predict_proba'):
            y_pred_proba = model.predict_proba(X_test)
        
        # Calculate metrics
        from sklearn.metrics import (
            accuracy_score, precision_score, recall_score, f1_score,
            classification_report, confusion_matrix, roc_auc_score
        )
        
        metrics = {
            'accuracy': accuracy_score(y_test, y_pred),
            'precision_macro': precision_score(y_test, y_pred, average='macro'),
            'recall_macro': recall_score(y_test, y_pred, average='macro'),
            'f1_macro': f1_score(y_test, y_pred, average='macro'),
            'precision_weighted': precision_score(y_test, y_pred, average='weighted'),
            'recall_weighted': recall_score(y_test, y_pred, average='weighted'),
            'f1_weighted': f1_score(y_test, y_pred, average='weighted')
        }
        
        # ROC AUC for binary classification
        if len(np.unique(y_test)) == 2 and y_pred_proba is not None:
            metrics['roc_auc'] = roc_auc_score(y_test, y_pred_proba[:, 1])
        
        # Generate classification report
        report = classification_report(y_test, y_pred, target_names=class_names, 
                                     output_dict=True)
        
        # Confusion matrix
        cm = confusion_matrix(y_test, y_pred)
        
        self.results = {
            'metrics': metrics,
            'classification_report': report,
            'confusion_matrix': cm.tolist(),
            'predictions': y_pred.tolist(),
            'probabilities': y_pred_proba.tolist() if y_pred_proba is not None else None
        }
        
        return self.results
    
    def cross_validate_model(self, model, X, y, cv=5, scoring=['accuracy', 'f1_weighted']):
        """Perform cross-validation with multiple metrics"""
        
        cv_results = {}
        
        for metric in scoring:
            scores = cross_val_score(model, X, y, cv=cv, scoring=metric)
            cv_results[metric] = {
                'scores': scores.tolist(),
                'mean': scores.mean(),
                'std': scores.std(),
                'cv': cv
            }
        
        self.results['cross_validation'] = cv_results
        return cv_results
    
    def learning_curve_analysis(self, model, X, y, cv=5, train_sizes=None):
        """Generate learning curve analysis"""
        
        if train_sizes is None:
            train_sizes = np.linspace(0.1, 1.0, 10)
        
        train_sizes_abs, train_scores, val_scores = learning_curve(
            model, X, y, cv=cv, train_sizes=train_sizes, n_jobs=-1
        )
        
        learning_curve_data = {
            'train_sizes': train_sizes_abs.tolist(),
            'train_scores_mean': train_scores.mean(axis=1).tolist(),
            'train_scores_std': train_scores.std(axis=1).tolist(),
            'val_scores_mean': val_scores.mean(axis=1).tolist(),
            'val_scores_std': val_scores.std(axis=1).tolist()
        }
        
        self.results['learning_curve'] = learning_curve_data
        return learning_curve_data
    
    def generate_evaluation_report(self, model_name: str = "Model"):
        """Generate a comprehensive evaluation report"""
        
        report = f"# {model_name} Evaluation Report\n\n"
        
        if 'metrics' in self.results:
            report += "## Performance Metrics\n\n"
            for metric, value in self.results['metrics'].items():
                report += f"- **{metric}**: {value:.4f}\n"
            report += "\n"
        
        if 'cross_validation' in self.results:
            report += "## Cross-Validation Results\n\n"
            for metric, cv_data in self.results['cross_validation'].items():
                report += f"- **{metric}**: {cv_data['mean']:.4f} (±{cv_data['std']:.4f})\n"
            report += "\n"
        
        if 'classification_report' in self.results:
            report += "## Classification Report\n\n"
            report += "```\n"
            for class_name, metrics in self.results['classification_report'].items():
                if isinstance(metrics, dict):
                    report += f"Class: {class_name}\n"
                    for metric, value in metrics.items():
                        if isinstance(value, float):
                            report += f"  {metric}: {value:.4f}\n"
                        else:
                            report += f"  {metric}: {value}\n"
                    report += "\n"
            report += "```\n\n"
        
        return report

# Usage example
# evaluator = ModelEvaluator()
# results = evaluator.evaluate_classification(model, X_test, y_test, class_names=['A', 'B', 'C'])
# cv_results = evaluator.cross_validate_model(model, X, y)
# report = evaluator.generate_evaluation_report("Random Forest Classifier")
```

### AutoML Implementation

#### 4. Automated Machine Learning Pipeline

```python
# AutoML with hyperparameter optimization
from sklearn.model_selection import RandomizedSearchCV, GridSearchCV
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.svm import SVC
import numpy as np
import time

class AutoMLPipeline:
    def __init__(self, task_type='classification', random_state=42):
        self.task_type = task_type
        self.random_state = random_state
        self.best_model = None
        self.best_score = None
        self.results = {}
        
        # Define model search space
        self.models = {
            'random_forest': {
                'model': RandomForestClassifier(random_state=random_state),
                'params': {
                    'n_estimators': [50, 100, 200],
                    'max_depth': [5, 10, 15, None],
                    'min_samples_split': [2, 5, 10],
                    'min_samples_leaf': [1, 2, 4]
                }
            },
            'gradient_boosting': {
                'model': GradientBoostingClassifier(random_state=random_state),
                'params': {
                    'n_estimators': [50, 100, 200],
                    'learning_rate': [0.01, 0.1, 0.2],
                    'max_depth': [3, 5, 7],
                    'subsample': [0.8, 0.9, 1.0]
                }
            },
            'logistic_regression': {
                'model': LogisticRegression(random_state=random_state, max_iter=1000),
                'params': {
                    'C': [0.1, 1.0, 10.0],
                    'penalty': ['l1', 'l2'],
                    'solver': ['liblinear', 'saga']
                }
            }
        }
        
        if task_type == 'classification':
            self.models['svm'] = {
                'model': SVC(random_state=random_state, probability=True),
                'params': {
                    'C': [0.1, 1.0, 10.0],
                    'kernel': ['rbf', 'linear'],
                    'gamma': ['scale', 'auto']
                }
            }
    
    def search_best_model(self, X_train, X_val, y_train, y_val, 
                         search_method='random', n_iter=50, cv=5):
        """Search for the best model using hyperparameter optimization"""
        
        print(f"Starting AutoML search with {search_method} search...")
        
        best_score = -np.inf
        best_model = None
        best_params = None
        search_results = []
        
        for model_name, model_config in self.models.items():
            print(f"Testing {model_name}...")
            
            start_time = time.time()
            
            # Choose search method
            if search_method == 'random':
                search = RandomizedSearchCV(
                    model_config['model'],
                    model_config['params'],
                    n_iter=n_iter,
                    cv=cv,
                    scoring='accuracy',
                    random_state=self.random_state,
                    n_jobs=-1
                )
            else:  # grid search
                search = GridSearchCV(
                    model_config['model'],
                    model_config['params'],
                    cv=cv,
                    scoring='accuracy',
                    n_jobs=-1
                )
            
            # Fit the search
            search.fit(X_train, y_train)
            
            # Evaluate on validation set
            val_score = search.score(X_val, y_val)
            
            search_time = time.time() - start_time
            
            result = {
                'model_name': model_name,
                'best_params': search.best_params_,
                'cv_score': search.best_score_,
                'val_score': val_score,
                'search_time': search_time,
                'best_estimator': search.best_estimator_
            }
            
            search_results.append(result)
            
            print(f"  - CV Score: {search.best_score_:.4f}")
            print(f"  - Val Score: {val_score:.4f}")
            print(f"  - Time: {search_time:.2f}s")
            
            # Update best model
            if val_score > best_score:
                best_score = val_score
                best_model = search.best_estimator_
                best_params = search.best_params_
        
        self.best_model = best_model
        self.best_score = best_score
        self.results = {
            'search_results': search_results,
            'best_model_name': best_model.__class__.__name__,
            'best_params': best_params,
            'best_score': best_score
        }
        
        print(f"\nBest model: {best_model.__class__.__name__}")
        print(f"Best validation score: {best_score:.4f}")
        
        return best_model
    
    def ensemble_models(self, X_train, y_train, top_k=3):
        """Create an ensemble of the top k models"""
        
        if not self.results:
            raise ValueError("Run search_best_model first")
        
        # Sort models by validation score
        sorted_results = sorted(
            self.results['search_results'],
            key=lambda x: x['val_score'],
            reverse=True
        )
        
        # Select top k models
        top_models = [result['best_estimator'] for result in sorted_results[:top_k]]
        
        from sklearn.ensemble import VotingClassifier
        
        # Create voting ensemble
        ensemble = VotingClassifier(
            estimators=[(f"model_{i}", model) for i, model in enumerate(top_models)],
            voting='soft'
        )
        
        # Train ensemble
        ensemble.fit(X_train, y_train)
        
        return ensemble
    
    def generate_search_report(self):
        """Generate a comprehensive AutoML search report"""
        
        if not self.results:
            return "No search results available. Run search_best_model first."
        
        report = "# AutoML Search Report\n\n"
        report += f"**Best Model**: {self.results['best_model_name']}\n"
        report += f"**Best Score**: {self.results['best_score']:.4f}\n\n"
        
        report += "## Search Results\n\n"
        report += "| Model | CV Score | Val Score | Time (s) |\n"
        report += "|-------|----------|-----------|----------|\n"
        
        for result in sorted(self.results['search_results'], 
                           key=lambda x: x['val_score'], reverse=True):
            report += f"| {result['model_name']} | {result['cv_score']:.4f} | "
            report += f"{result['val_score']:.4f} | {result['search_time']:.2f} |\n"
        
        report += "\n## Best Hyperparameters\n\n"
        report += f"**Model**: {self.results['best_model_name']}\n\n"
        
        for param, value in self.results['best_params'].items():
            report += f"- **{param}**: {value}\n"
        
        return report

# Usage example
# automl = AutoMLPipeline(task_type='classification')
# best_model = automl.search_best_model(X_train, X_val, y_train, y_val)
# ensemble_model = automl.ensemble_models(X_train, y_train)
# report = automl.generate_search_report()
```

## Level 3: Advanced Integration

### MLOps and Production Deployment

#### 1. Model Serving with REST API

```python
# FastAPI model serving with MLflow integration
from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel
import mlflow.pyfunc
import numpy as np
import pandas as pd
import json
from typing import List, Optional, Dict, Any
import asyncio
from concurrent.futures import ThreadPoolExecutor
import logging

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class PredictionRequest(BaseModel):
    features: List[Dict[str, Any]]
    model_version: Optional[str] = "latest"
    return_probabilities: Optional[bool] = False

class PredictionResponse(BaseModel):
    predictions: List[Any]
    probabilities: Optional[List[List[float]]] = None
    model_version: str
    prediction_time_ms: float
    request_id: str

class ModelServer:
    def __init__(self, model_uri: str, app_name: str = "ml-model-server"):
        self.app = FastAPI(title=app_name)
        self.model_uri = model_uri
        self.model = None
        self.model_version = "unknown"
        self.executor = ThreadPoolExecutor(max_workers=4)
        
        # Load model
        self._load_model()
        
        # Setup routes
        self._setup_routes()
    
    def _load_model(self):
        """Load model from MLflow"""
        try:
            self.model = mlflow.pyfunc.load_model(self.model_uri)
            
            # Extract model version from URI if possible
            if "runs:/" in self.model_uri:
                self.model_version = self.model_uri.split("/")[-1]
            
            logger.info(f"Model loaded successfully from {self.model_uri}")
            logger.info(f"Model version: {self.model_version}")
            
        except Exception as e:
            logger.error(f"Failed to load model: {str(e)}")
            raise
    
    def _setup_routes(self):
        """Setup FastAPI routes"""
        
        @self.app.get("/health")
        async def health_check():
            return {"status": "healthy", "model_version": self.model_version}
        
        @self.app.get("/model/info")
        async def model_info():
            if not self.model:
                raise HTTPException(status_code=503, detail="Model not loaded")
            
            return {
                "model_uri": self.model_uri,
                "model_version": self.model_version,
                "model_type": type(self.model).__name__
            }
        
        @self.app.post("/predict", response_model=PredictionResponse)
        async def predict(request: PredictionRequest, background_tasks: BackgroundTasks):
            """Make predictions with the loaded model"""
            
            if not self.model:
                raise HTTPException(status_code=503, detail="Model not loaded")
            
            start_time = time.time()
            request_id = f"req_{int(time.time() * 1000)}"
            
            try:
                # Convert features to DataFrame
                df = pd.DataFrame(request.features)
                
                # Make prediction in background thread
                loop = asyncio.get_event_loop()
                predictions = await loop.run_in_executor(
                    self.executor,
                    self._predict_sync,
                    df,
                    request.return_probabilities
                )
                
                prediction_time = (time.time() - start_time) * 1000
                
                # Log prediction request
                background_tasks.add_task(
                    self._log_prediction,
                    request_id,
                    len(request.features),
                    prediction_time
                )
                
                response = PredictionResponse(
                    predictions=predictions["predictions"],
                    probabilities=predictions.get("probabilities"),
                    model_version=self.model_version,
                    prediction_time_ms=prediction_time,
                    request_id=request_id
                )
                
                return response
                
            except Exception as e:
                logger.error(f"Prediction failed: {str(e)}")
                raise HTTPException(status_code=500, detail=str(e))
        
        @self.app.post("/predict/batch")
        async def predict_batch(requests: List[PredictionRequest]):
            """Batch prediction endpoint"""
            
            if not self.model:
                raise HTTPException(status_code=503, detail="Model not loaded")
            
            start_time = time.time()
            
            try:
                results = []
                
                for request in requests:
                    df = pd.DataFrame(request.features)
                    
                    loop = asyncio.get_event_loop()
                    predictions = await loop.run_in_executor(
                        self.executor,
                        self._predict_sync,
                        df,
                        request.return_probabilities
                    )
                    
                    results.append({
                        "predictions": predictions["predictions"],
                        "probabilities": predictions.get("probabilities"),
                        "model_version": self.model_version
                    })
                
                total_time = (time.time() - start_time) * 1000
                
                return {
                    "results": results,
                    "total_requests": len(requests),
                    "total_time_ms": total_time,
                    "avg_time_per_request_ms": total_time / len(requests)
                }
                
            except Exception as e:
                logger.error(f"Batch prediction failed: {str(e)}")
                raise HTTPException(status_code=500, detail=str(e))
    
    def _predict_sync(self, df: pd.DataFrame, return_probabilities: bool = False):
        """Synchronous prediction method"""
        
        if return_probabilities and hasattr(self.model, 'predict_proba'):
            predictions = self.model.predict(df)
            probabilities = self.model.predict_proba(df).tolist()
            return {
                "predictions": predictions.tolist(),
                "probabilities": probabilities
            }
        else:
            predictions = self.model.predict(df)
            return {
                "predictions": predictions.tolist()
            }
    
    def _log_prediction(self, request_id: str, num_samples: int, prediction_time: float):
        """Log prediction request for monitoring"""
        
        log_entry = {
            "request_id": request_id,
            "timestamp": datetime.now().isoformat(),
            "num_samples": num_samples,
            "prediction_time_ms": prediction_time,
            "model_version": self.model_version
        }
        
        logger.info(f"Prediction logged: {json.dumps(log_entry)}")

# Usage example
# server = ModelServer("runs:/12345abcdef/model")
# app = server.app
```

#### 2. Model Monitoring and Drift Detection

```python
# Production model monitoring and drift detection
import numpy as np
import pandas as pd
from scipy import stats
from sklearn.metrics import accuracy_score, precision_score, recall_score
import plotly.graph_objects as go
import plotly.express as px
from datetime import datetime, timedelta
import json
import asyncio
from typing import Dict, List, Tuple

class ModelMonitor:
    def __init__(self, model_performance_threshold: float = 0.05):
        self.performance_threshold = performance_threshold
        self.reference_data = None
        self.performance_history = []
        self.drift_alerts = []
    
    def set_reference_data(self, X_ref: pd.DataFrame, y_ref: pd.Series = None):
        """Set reference data for drift detection"""
        self.reference_data = {
            'X': X_ref,
            'y': y_ref,
            'feature_stats': self._calculate_feature_statistics(X_ref)
        }
        
        if y_ref is not None:
            self.reference_data['target_distribution'] = y_ref.value_counts(normalize=True).to_dict()
    
    def _calculate_feature_statistics(self, X: pd.DataFrame) -> Dict:
        """Calculate feature statistics for reference"""
        stats = {}
        
        for column in X.columns:
            if X[column].dtype in ['int64', 'float64']:
                stats[column] = {
                    'mean': X[column].mean(),
                    'std': X[column].std(),
                    'min': X[column].min(),
                    'max': X[column].max(),
                    'q25': X[column].quantile(0.25),
                    'q75': X[column].quantile(0.75)
                }
            else:
                stats[column] = {
                    'value_counts': X[column].value_counts(normalize=True).to_dict()
                }
        
        return stats
    
    def detect_data_drift(self, X_current: pd.DataFrame, 
                          method: str = 'ks_test') -> Dict:
        """Detect data drift compared to reference data"""
        
        if self.reference_data is None:
            raise ValueError("Reference data not set")
        
        X_ref = self.reference_data['X']
        ref_stats = self.reference_data['feature_stats']
        
        drift_results = {}
        
        for column in X_current.columns:
            if column not in X_ref.columns:
                continue
            
            if X_current[column].dtype in ['int64', 'float64']:
                # Numerical features
                current_data = X_current[column].dropna()
                ref_data = X_ref[column].dropna()
                
                if method == 'ks_test':
                    # Kolmogorov-Smirnov test
                    statistic, p_value = stats.ks_2samp(current_data, ref_data)
                    drift_detected = p_value < 0.05
                    
                    drift_results[column] = {
                        'drift_detected': drift_detected,
                        'ks_statistic': statistic,
                        'p_value': p_value,
                        'ref_mean': ref_stats[column]['mean'],
                        'current_mean': current_data.mean(),
                        'ref_std': ref_stats[column]['std'],
                        'current_std': current_data.std()
                    }
                
                elif method == 'wasserstein':
                    # Wasserstein distance
                    from scipy.stats import wasserstein_distance
                    distance = wasserstein_distance(current_data, ref_data)
                    
                    drift_results[column] = {
                        'wasserstein_distance': distance,
                        'ref_mean': ref_stats[column]['mean'],
                        'current_mean': current_data.mean(),
                        'drift_detected': distance > 0.1  # Threshold can be adjusted
                    }
            
            else:
                # Categorical features
                current_dist = X_current[column].value_counts(normalize=True).to_dict()
                ref_dist = ref_stats[column]['value_counts']
                
                # Calculate KL divergence
                kl_divergence = self._calculate_kl_divergence(current_dist, ref_dist)
                
                drift_results[column] = {
                    'kl_divergence': kl_divergence,
                    'drift_detected': kl_divergence > 0.1,
                    'ref_distribution': ref_dist,
                    'current_distribution': current_dist
                }
        
        # Calculate overall drift score
        drift_columns = [col for col, result in drift_results.items() 
                        if result.get('drift_detected', False)]
        overall_drift_score = len(drift_columns) / len(drift_results)
        
        return {
            'feature_drift': drift_results,
            'overall_drift_score': overall_drift_score,
            'drifted_features': drift_columns,
            'timestamp': datetime.now().isoformat()
        }
    
    def _calculate_kl_divergence(self, p: Dict, q: Dict) -> float:
        """Calculate KL divergence between two distributions"""
        
        # Ensure both distributions have the same keys
        all_keys = set(p.keys()) | set(q.keys())
        
        for key in all_keys:
            if key not in p:
                p[key] = 1e-10  # Small value to avoid division by zero
            if key not in q:
                q[key] = 1e-10
        
        # Normalize
        p_total = sum(p.values())
        q_total = sum(q.values())
        
        p_norm = {k: v / p_total for k, v in p.items()}
        q_norm = {k: v / q_total for k, v in q.items()}
        
        # Calculate KL divergence
        kl_div = 0
        for key in all_keys:
            kl_div += p_norm[key] * np.log(p_norm[key] / q_norm[key])
        
        return kl_div
    
    def monitor_model_performance(self, y_true: np.ndarray, y_pred: np.ndarray,
                                 model_name: str = "model"):
        """Monitor model performance and detect degradation"""
        
        # Calculate performance metrics
        accuracy = accuracy_score(y_true, y_pred)
        precision = precision_score(y_true, y_pred, average='weighted')
        recall = recall_score(y_true, y_pred, average='weighted')
        
        performance_entry = {
            'timestamp': datetime.now().isoformat(),
            'model_name': model_name,
            'accuracy': accuracy,
            'precision': precision,
            'recall': recall
        }
        
        self.performance_history.append(performance_entry)
        
        # Check for performance degradation
        if len(self.performance_history) > 1:
            prev_performance = self.performance_history[-2]
            accuracy_change = abs(accuracy - prev_performance['accuracy'])
            
            if accuracy_change > self.performance_threshold:
                alert = {
                    'timestamp': datetime.now().isoformat(),
                    'alert_type': 'performance_degradation',
                    'model_name': model_name,
                    'metric': 'accuracy',
                    'previous_value': prev_performance['accuracy'],
                    'current_value': accuracy,
                    'change': accuracy_change,
                    'threshold': self.performance_threshold
                }
                
                self.drift_alerts.append(alert)
        
        return performance_entry
    
    def generate_monitoring_report(self) -> str:
        """Generate comprehensive monitoring report"""
        
        report = "# Model Monitoring Report\n\n"
        report += f"**Generated**: {datetime.now().isoformat()}\n\n"
        
        # Performance history
        if self.performance_history:
            report += "## Performance History\n\n"
            report += "| Timestamp | Model | Accuracy | Precision | Recall |\n"
            report += "|-----------|-------|----------|-----------|--------|\n"
            
            for entry in self.performance_history[-10:]:  # Last 10 entries
                report += f"| {entry['timestamp']} | {entry['model_name']} | "
                report += f"{entry['accuracy']:.4f} | {entry['precision']:.4f} | "
                report += f"{entry['recall']:.4f} |\n"
            report += "\n"
        
        # Drift alerts
        if self.drift_alerts:
            report += "## Drift Alerts\n\n"
            
            for alert in self.drift_alerts[-5:]:  # Last 5 alerts
                report += f"### {alert['alert_type'].replace('_', ' ').title()}\n"
                report += f"- **Time**: {alert['timestamp']}\n"
                report += f"- **Model**: {alert['model_name']}\n"
                report += f"- **Metric**: {alert['metric']}\n"
                report += f"- **Change**: {alert['previous_value']:.4f} → {alert['current_value']:.4f}\n"
                report += f"- **Threshold**: {alert['threshold']}\n\n"
        
        return report

# Usage example
# monitor = ModelMonitor(performance_threshold=0.05)
# monitor.set_reference_data(X_train, y_train)
# drift_results = monitor.detect_data_drift(X_current)
# performance = monitor.monitor_model_performance(y_true, y_pred)
# report = monitor.generate_monitoring_report()
```

## Related Skills

- **moai-domain-data-science**: Data science workflows and analysis
- **moai-domain-testing**: ML model testing and validation
- **moai-domain-devops**: MLOps infrastructure and deployment
- **moai-essentials-refactor**: Model optimization and refactoring

## Quick Start Checklist

- [ ] Select appropriate ML framework (TensorFlow 2.20/PyTorch 2.9/Scikit-learn 1.7)
- [ ] Set up experiment tracking with MLflow
- [ ] Implement data preprocessing pipeline
- [ ] Configure AutoML for hyperparameter optimization
- [ ] Setup model monitoring and drift detection
- [ ] Deploy model serving with FastAPI
- [ ] Implement performance monitoring
- [ ] Create model retraining pipeline

## Performance Optimization Tips

1. **Data Preprocessing**: Use ColumnTransformer for consistent preprocessing
2. **Hyperparameter Tuning**: Use RandomizedSearchCV for efficient optimization
3. **Model Selection**: Consider ensemble methods for better performance
4. **Experiment Tracking**: Use MLflow for reproducible experiments
5. **Monitoring**: Implement drift detection and performance monitoring
6. **Deployment**: Use FastAPI for scalable model serving
7. **AutoML**: Leverage AutoML for automated model selection
8. **MLOps**: Implement CI/CD pipelines for model deployment

---

**Enterprise Machine Learning** - Build production-ready ML systems with comprehensive monitoring, automated experimentation, and scalable deployment infrastructure.
