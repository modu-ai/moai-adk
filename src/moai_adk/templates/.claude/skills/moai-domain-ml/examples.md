# ML Domain - Practical Implementation Examples

_Last updated: 2025-11-13_

## Level 1: Quick Start Examples

### Example 1: Basic ML Project Setup
```bash
# Initialize ML project structure
mkdir ml-project && cd ml-project
python -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate  # Windows

# Install core ML libraries
pip install torch==2.9.0 scikit-learn==1.7.2 pandas==2.3.3
pip install matplotlib seaborn jupyter

# Create project structure
mkdir -p data/raw data/processed notebooks src models
touch requirements.txt README.md
```

### Example 2: TDD ML Workflow
```python
# test_model.py (RED phase)
import pytest
import numpy as np
from src.model import SimpleModel

def test_model_prediction():
    model = SimpleModel()
    X_train = np.array([[1], [2], [3], [4]])
    y_train = np.array([2, 4, 6, 8])
    
    model.fit(X_train, y_train)
    prediction = model.predict(np.array([[5]]))
    
    assert abs(prediction[0] - 10) < 0.1  # Allow small tolerance

# src/model.py (GREEN phase)
import numpy as np

class SimpleModel:
    def __init__(self):
        self.coefficient = None
    
    def fit(self, X, y):
        # Simple linear regression: y = 2x
        self.coefficient = 2.0
    
    def predict(self, X):
        return X * self.coefficient
```

### Example 3: Data Processing Pipeline
```python
# src/data_processor.py
import pandas as pd
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split

class DataProcessor:
    def __init__(self):
        self.scaler = StandardScaler()
    
    def load_data(self, filepath):
        """Load and validate data."""
        df = pd.read_csv(filepath)
        
        # Basic validation
        if df.empty:
            raise ValueError("Dataset is empty")
        
        print(f"Loaded {len(df)} records with {len(df.columns)} features")
        return df
    
    def preprocess(self, df, target_column):
        """Preprocess data for ML."""
        # Handle missing values
        df = df.fillna(df.median())
        
        # Separate features and target
        X = df.drop(columns=[target_column])
        y = df[target_column]
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=0.2, random_state=42
        )
        
        # Scale features
        X_train_scaled = self.scaler.fit_transform(X_train)
        X_test_scaled = self.scaler.transform(X_test)
        
        return X_train_scaled, X_test_scaled, y_train, y_test
```

---

## Level 2: Core Implementation Examples

### Example 4: Neural Network with PyTorch
```python
# src/neural_network.py
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset

class SimpleNN(nn.Module):
    def __init__(self, input_size, hidden_size, output_size):
        super(SimpleNN, self).__init__()
        self.fc1 = nn.Linear(input_size, hidden_size)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(hidden_size, output_size)
    
    def forward(self, x):
        x = self.fc1(x)
        x = self.relu(x)
        x = self.fc2(x)
        return x

def train_model(model, dataloader, criterion, optimizer, epochs):
    """Train neural network with proper tracking."""
    model.train()
    losses = []
    
    for epoch in range(epochs):
        epoch_loss = 0
        for batch_X, batch_y in dataloader:
            optimizer.zero_grad()
            outputs = model(batch_X)
            loss = criterion(outputs, batch_y)
            loss.backward()
            optimizer.step()
            epoch_loss += loss.item()
        
        avg_loss = epoch_loss / len(dataloader)
        losses.append(avg_loss)
        
        if epoch % 10 == 0:
            print(f'Epoch [{epoch}/{epochs}], Loss: {avg_loss:.4f}')
    
    return losses

# Usage example
def main():
    # Sample data
    X = torch.randn(1000, 10)  # 1000 samples, 10 features
    y = torch.randn(1000, 1)   # 1000 targets
    
    # Create dataset and dataloader
    dataset = TensorDataset(X, y)
    dataloader = DataLoader(dataset, batch_size=32, shuffle=True)
    
    # Initialize model
    model = SimpleNN(input_size=10, hidden_size=64, output_size=1)
    criterion = nn.MSELoss()
    optimizer = optim.Adam(model.parameters(), lr=0.001)
    
    # Train model
    losses = train_model(model, dataloader, criterion, optimizer, epochs=100)
    
    print("Training completed!")
    return model, losses
```

### Example 5: Model Evaluation Pipeline
```python
# src/evaluation.py
import numpy as np
from sklearn.metrics import accuracy_score, classification_report
from sklearn.metrics import mean_squared_error, r2_score
import matplotlib.pyplot as plt

class ModelEvaluator:
    def __init__(self):
        self.metrics = {}
    
    def evaluate_classification(self, y_true, y_pred, model_name):
        """Evaluate classification model."""
        accuracy = accuracy_score(y_true, y_pred)
        report = classification_report(y_true, y_pred, output_dict=True)
        
        self.metrics[model_name] = {
            'accuracy': accuracy,
            'precision': report['macro avg']['precision'],
            'recall': report['macro avg']['recall'],
            'f1_score': report['macro avg']['f1-score']
        }
        
        print(f"{model_name} Classification Results:")
        print(f"Accuracy: {accuracy:.4f}")
        print(f"F1-Score: {self.metrics[model_name]['f1_score']:.4f}")
        
        return self.metrics[model_name]
    
    def evaluate_regression(self, y_true, y_pred, model_name):
        """Evaluate regression model."""
        mse = mean_squared_error(y_true, y_pred)
        r2 = r2_score(y_true, y_pred)
        rmse = np.sqrt(mse)
        
        self.metrics[model_name] = {
            'mse': mse,
            'rmse': rmse,
            'r2_score': r2
        }
        
        print(f"{model_name} Regression Results:")
        print(f"RMSE: {rmse:.4f}")
        print(f"R² Score: {r2:.4f}")
        
        return self.metrics[model_name]
    
    def plot_predictions(self, y_true, y_pred, title="Model Predictions"):
        """Plot true vs predicted values."""
        plt.figure(figsize=(10, 6))
        plt.scatter(y_true, y_pred, alpha=0.6)
        plt.plot([y_true.min(), y_true.max()], [y_true.min(), y_true.max()], 'r--')
        plt.xlabel('True Values')
        plt.ylabel('Predictions')
        plt.title(title)
        plt.grid(True, alpha=0.3)
        plt.show()
```

---

## Level 3: Advanced Examples

### Example 6: Hyperparameter Optimization with Optuna
```python
# src/hyperparameter_optimization.py
import optuna
import torch
import torch.nn as nn
from sklearn.model_selection import cross_val_score

class HyperparameterOptimizer:
    def __init__(self, model_type='neural_network'):
        self.model_type = model_type
        self.best_params = None
        self.study = None
    
    def objective(self, trial):
        """Objective function for Optuna optimization."""
        if self.model_type == 'neural_network':
            # Suggest hyperparameters
            hidden_size = trial.suggest_int('hidden_size', 32, 256)
            learning_rate = trial.suggest_loguniform('learning_rate', 1e-5, 1e-1)
            batch_size = trial.suggest_categorical('batch_size', [16, 32, 64])
            dropout_rate = trial.suggest_uniform('dropout_rate', 0.0, 0.5)
            
            # Create and train model with suggested parameters
            model = self.create_model(hidden_size, dropout_rate)
            score = self.train_and_evaluate(model, learning_rate, batch_size)
            
            return score
    
    def create_model(self, hidden_size, dropout_rate):
        """Create neural network with specified parameters."""
        model = nn.Sequential(
            nn.Linear(10, hidden_size),
            nn.ReLU(),
            nn.Dropout(dropout_rate),
            nn.Linear(hidden_size, hidden_size // 2),
            nn.ReLU(),
            nn.Dropout(dropout_rate),
            nn.Linear(hidden_size // 2, 1)
        )
        return model
    
    def optimize(self, n_trials=100):
        """Run hyperparameter optimization."""
        self.study = optuna.create_study(direction='maximize')
        self.study.optimize(self.objective, n_trials=n_trials)
        
        self.best_params = self.study.best_params
        print("Best trial:")
        print(f"  Value: {self.study.best_value}")
        print(f"  Params: {self.best_params}")
        
        return self.best_params

# Usage example
def main():
    optimizer = HyperparameterOptimizer()
    best_params = optimizer.optimize(n_trials=50)
    
    # Train final model with best parameters
    final_model = optimizer.create_model(
        hidden_size=best_params['hidden_size'],
        dropout_rate=best_params['dropout_rate']
    )
    
    print("Hyperparameter optimization completed!")
    return final_model, best_params
```

### Example 7: ML Pipeline with sklearn
```python
# src/ml_pipeline.py
from sklearn.ensemble import RandomForestClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.svm import SVC
from sklearn.model_selection import GridSearchCV, cross_val_score
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler
import pandas as pd

class MLPipeline:
    def __init__(self):
        self.models = {
            'random_forest': RandomForestClassifier(random_state=42),
            'logistic_regression': LogisticRegression(random_state=42),
            'svm': SVC(random_state=42)
        }
        self.best_model = None
        self.best_score = None
    
    def create_pipeline(self, model_name, model_params=None):
        """Create sklearn pipeline with preprocessing."""
        if model_name not in self.models:
            raise ValueError(f"Model {model_name} not supported")
        
        model = self.models[model_name]
        if model_params:
            model.set_params(**model_params)
        
        pipeline = Pipeline([
            ('scaler', StandardScaler()),
            ('classifier', model)
        ])
        
        return pipeline
    
    def hyperparameter_search(self, X, y, model_name):
        """Perform grid search for hyperparameters."""
        param_grids = {
            'random_forest': {
                'classifier__n_estimators': [50, 100, 200],
                'classifier__max_depth': [5, 10, None],
                'classifier__min_samples_split': [2, 5, 10]
            },
            'logistic_regression': {
                'classifier__C': [0.1, 1, 10],
                'classifier__penalty': ['l1', 'l2']
            },
            'svm': {
                'classifier__C': [0.1, 1, 10],
                'classifier__kernel': ['rbf', 'linear'],
                'classifier__gamma': ['scale', 'auto']
            }
        }
        
        pipeline = self.create_pipeline(model_name)
        grid_search = GridSearchCV(
            pipeline, 
            param_grids[model_name], 
            cv=5, 
            scoring='accuracy',
            n_jobs=-1
        )
        
        grid_search.fit(X, y)
        
        print(f"Best parameters for {model_name}:")
        print(grid_search.best_params_)
        print(f"Best cross-validation score: {grid_search.best_score_:.4f}")
        
        return grid_search.best_estimator_, grid_search.best_score_
    
    def compare_models(self, X, y):
        """Compare multiple models and select the best."""
        results = {}
        
        for model_name in self.models.keys():
            print(f"\nEvaluating {model_name}...")
            best_model, score = self.hyperparameter_search(X, y, model_name)
            results[model_name] = {
                'model': best_model,
                'score': score
            }
            
            if self.best_score is None or score > self.best_score:
                self.best_score = score
                self.best_model = best_model
                self.best_model_name = model_name
        
        print(f"\nBest model: {self.best_model_name} with score: {self.best_score:.4f}")
        return results
```

---

## Level 4: Production Examples

### Example 8: Model Deployment Template
```python
# src/model_deployment.py
import joblib
import json
from flask import Flask, request, jsonify
import numpy as np

class ModelDeployment:
    def __init__(self, model_path=None):
        self.model = None
        self.model_path = model_path
        self.app = Flask(__name__)
        self.setup_routes()
    
    def load_model(self, model_path):
        """Load trained model from disk."""
        try:
            self.model = joblib.load(model_path)
            print(f"Model loaded from {model_path}")
            return True
        except Exception as e:
            print(f"Error loading model: {e}")
            return False
    
    def predict(self, features):
        """Make prediction using loaded model."""
        if self.model is None:
            raise ValueError("Model not loaded")
        
        # Convert to numpy array if needed
        if isinstance(features, list):
            features = np.array(features)
        
        # Reshape for single prediction
        if features.ndim == 1:
            features = features.reshape(1, -1)
        
        prediction = self.model.predict(features)
        return prediction.tolist()
    
    def setup_routes(self):
        """Setup Flask API routes."""
        
        @self.app.route('/health', methods=['GET'])
        def health_check():
            return jsonify({'status': 'healthy', 'model_loaded': self.model is not None})
        
        @self.app.route('/predict', methods=['POST'])
        def make_prediction():
            try:
                data = request.get_json()
                
                if 'features' not in data:
                    return jsonify({'error': 'Missing features in request'}), 400
                
                features = data['features']
                prediction = self.predict(features)
                
                return jsonify({
                    'prediction': prediction,
                    'status': 'success'
                })
                
            except Exception as e:
                return jsonify({'error': str(e)}), 500
        
        @self.app.route('/model-info', methods=['GET'])
        def model_info():
            if self.model is None:
                return jsonify({'error': 'No model loaded'}), 404
            
            return jsonify({
                'model_type': type(self.model).__name__,
                'model_path': self.model_path,
                'status': 'loaded'
            })
    
    def run(self, host='0.0.0.0', port=5000, debug=False):
        """Run the Flask application."""
        if self.model is None and self.model_path:
            self.load_model(self.model_path)
        
        print(f"Starting model deployment server on {host}:{port}")
        self.app.run(host=host, port=port, debug=debug)

# Usage example for deployment
if __name__ == "__main__":
    # Initialize deployment
    deployment = ModelDeployment(model_path="models/best_model.pkl")
    
    # Run the API server
    deployment.run(host='0.0.0.0', port=5000)
```

---

## Quick Testing Commands

```bash
# Run all tests with coverage
pytest tests/ -v --cov=src --cov-report=html

# Quality checks
flake8 src/ tests/
mypy src/

# Verify model performance
python -m src.evaluation --model models/best_model.pkl --test-data data/test.csv

# Load test API endpoint
curl -X POST http://localhost:5000/predict \
  -H "Content-Type: application/json" \
  -d '{"features": [1, 2, 3, 4, 5]}'
```

---

**Related Skills**:
- `Skill("moai-domain-data-science")` for data analysis patterns
- `Skill("moai-essentials-perf")` for ML performance optimization
- `Skill("moai-security-backend")` for ML security considerations

---

_For comprehensive implementation details, see the main SKILL.md file_
