# Examples

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
            report += "```

```

### AutoML Implementation

#### 4. Automated Machine Learning Pipeline

```

```

## Level 3: Advanced Integration

### MLOps and Production Deployment

#### 1. Model Serving with REST API

```

```

#### 2. Model Monitoring and Drift Detection

```