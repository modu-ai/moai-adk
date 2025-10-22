# Machine Learning & MLOps Reference

> Official documentation for ML model training, evaluation, deployment, and MLOps workflows

---

## Official Documentation Links

### ML Frameworks

| Framework | Version | Documentation | Status |
|-----------|---------|--------------|--------|
| **PyTorch** | 2.5.1 | https://pytorch.org/docs/ | ✅ Current (2025) |
| **TensorFlow** | 2.18.0 | https://www.tensorflow.org/api_docs | ✅ Current (2025) |
| **scikit-learn** | 1.6.0 | https://scikit-learn.org/stable/ | ✅ Current (2025) |
| **Hugging Face** | 4.47.0 | https://huggingface.co/docs | ✅ Current (2025) |

### MLOps Platforms

| Platform | Documentation | Status |
|----------|--------------|--------|
| **MLflow** | https://mlflow.org/docs/latest/ | ✅ Current (2025) |
| **Kubeflow** | https://www.kubeflow.org/docs/ | ✅ Current (2025) |
| **Weights & Biases** | https://docs.wandb.ai/ | ✅ Current (2025) |
| **Azure ML** | https://learn.microsoft.com/en-us/azure/machine-learning/ | ✅ Current (2025) |

---

## Model Training Best Practices

### Data Pipeline
```python
import torch
from torch.utils.data import Dataset, DataLoader

class CustomDataset(Dataset):
    def __init__(self, data, labels):
        self.data = data
        self.labels = labels
    
    def __len__(self):
        return len(self.data)
    
    def __getitem__(self, idx):
        return self.data[idx], self.labels[idx]

# Create DataLoader with optimization
train_loader = DataLoader(
    dataset=train_dataset,
    batch_size=32,
    shuffle=True,
    num_workers=4,
    pin_memory=True  # Faster data transfer to GPU
)
```

### Model Training Loop
```python
import torch.nn as nn
import torch.optim as optim

def train_epoch(model, loader, criterion, optimizer, device):
    model.train()
    total_loss = 0
    
    for batch_idx, (data, target) in enumerate(loader):
        data, target = data.to(device), target.to(device)
        
        optimizer.zero_grad()
        output = model(data)
        loss = criterion(output, target)
        
        loss.backward()
        optimizer.step()
        
        total_loss += loss.item()
    
    return total_loss / len(loader)
```

---

## Model Evaluation

### Metrics
```python
from sklearn.metrics import accuracy_score, precision_recall_fscore_support

def evaluate_model(model, test_loader, device):
    model.eval()
    predictions, true_labels = [], []
    
    with torch.no_grad():
        for data, target in test_loader:
            data = data.to(device)
            output = model(data)
            pred = output.argmax(dim=1).cpu().numpy()
            
            predictions.extend(pred)
            true_labels.extend(target.numpy())
    
    accuracy = accuracy_score(true_labels, predictions)
    precision, recall, f1, _ = precision_recall_fscore_support(
        true_labels, predictions, average='weighted'
    )
    
    return {
        'accuracy': accuracy,
        'precision': precision,
        'recall': recall,
        'f1': f1
    }
```

---

## MLOps Deployment

### MLflow Tracking
```python
import mlflow

mlflow.set_experiment("my-experiment")

with mlflow.start_run():
    # Log parameters
    mlflow.log_param("learning_rate", 0.001)
    mlflow.log_param("batch_size", 32)
    
    # Train model
    model = train_model()
    
    # Log metrics
    mlflow.log_metric("accuracy", 0.95)
    mlflow.log_metric("loss", 0.1)
    
    # Log model
    mlflow.pytorch.log_model(model, "model")
```

### Model Serving (FastAPI)
```python
from fastapi import FastAPI
import torch

app = FastAPI()
model = torch.load("model.pth")
model.eval()

@app.post("/predict")
async def predict(data: dict):
    tensor = torch.tensor(data['features'])
    with torch.no_grad():
        output = model(tensor)
    return {"prediction": output.tolist()}
```

---

**Last Updated**: 2025-10-22
**Frameworks**: PyTorch 2.5.1, TensorFlow 2.18.0, scikit-learn 1.6.0
