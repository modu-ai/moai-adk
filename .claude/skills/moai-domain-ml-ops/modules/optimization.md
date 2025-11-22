# ML/MLOps Performance Optimization & Scaling

**Version**: 4.0.0 (2025-11-22)
**Status**: Production Ready

---

## Training Performance Optimization

### Mixed Precision Training with PyTorch

```python
from torch.cuda.amp import autocast, GradScaler
import torch

class MixedPrecisionTrainer:
    """Train models faster with mixed precision (FP16 + FP32)."""

    def __init__(self, model, optimizer, device='cuda'):
        self.model = model
        self.optimizer = optimizer
        self.device = device
        self.scaler = GradScaler()  # Gradient scaling for FP16

    def train_step(self, batch_X, batch_y):
        """Single training step with mixed precision."""
        with autocast(device_type='cuda'):  # FP16 forward pass
            predictions = self.model(batch_X)
            loss = self.compute_loss(predictions, batch_y)

        self.scaler.scale(loss).backward()  # Scaled backward pass
        self.scaler.step(self.optimizer)
        self.scaler.update()

        return loss.item()

    def estimate_speedup(self):
        """Estimate training speedup from mixed precision."""
        # Typical: 1.5-3x faster, 2x memory reduction
        return {
            'speedup': '1.5-3x',
            'memory_reduction': '2x',
            'accuracy_impact': 'negligible'
        }
```

### Batch Processing Optimization

```python
import torch
from torch.utils.data import DataLoader

class BatchOptimizer:
    """Optimize batch size and data loading performance."""

    @staticmethod
    def find_optimal_batch_size(model, device='cuda', start_batch_size=32):
        """Binary search for maximum batch size that fits in GPU memory."""
        batch_size = start_batch_size

        while True:
            try:
                dummy_input = torch.randn(batch_size, 3, 224, 224).to(device)
                _ = model(dummy_input)
                batch_size *= 2
            except RuntimeError:  # Out of memory
                return batch_size // 2

    @staticmethod
    def create_optimized_dataloader(dataset, optimal_batch_size, num_workers=4):
        """Create DataLoader with optimal settings."""
        return DataLoader(
            dataset,
            batch_size=optimal_batch_size,
            shuffle=True,
            num_workers=num_workers,  # Parallel data loading
            pin_memory=True,  # Faster CPU to GPU transfer
            persistent_workers=True,  # Reduce worker initialization overhead
            prefetch_factor=2  # Pre-load next batches
        )
```

### Gradient Accumulation for Large Models

```python
class GradientAccumulationTrainer:
    """Train with larger effective batch sizes using gradient accumulation."""

    def __init__(self, model, optimizer, accumulation_steps=4):
        self.model = model
        self.optimizer = optimizer
        self.accumulation_steps = accumulation_steps
        self.step_count = 0

    def train_epoch(self, dataloader):
        """Training epoch with gradient accumulation."""
        self.optimizer.zero_grad()

        for step, (batch_X, batch_y) in enumerate(dataloader):
            predictions = self.model(batch_X)
            loss = self.compute_loss(predictions, batch_y)

            # Scale loss by accumulation steps
            loss = loss / self.accumulation_steps
            loss.backward()

            self.step_count += 1

            # Update weights after accumulation
            if (self.step_count + 1) % self.accumulation_steps == 0:
                self.optimizer.step()
                self.optimizer.zero_grad()

    def effective_batch_size(self, dataloader_batch_size):
        """Calculate effective batch size with accumulation."""
        return dataloader_batch_size * self.accumulation_steps
```

---

## Inference Optimization

### Model Quantization for Faster Inference

```python
import torch.quantization as q

class ModelQuantizer:
    """Reduce model size and improve inference speed with quantization."""

    @staticmethod
    def quantize_model(model, qat=False):
        """Quantize model to INT8 for 4x speedup."""
        model.eval()

        if qat:
            # Quantization-Aware Training (better accuracy)
            model.qconfig = q.get_default_qat_qconfig('fbgemm')
            q.prepare_qat(model, inplace=True)
            # Train here
            q.convert(model, inplace=True)
        else:
            # Post-Training Quantization (simpler)
            model.qconfig = q.get_default_qconfig('fbgemm')
            q.prepare(model, inplace=True)
            q.convert(model, inplace=True)

        return model

    @staticmethod
    def compare_performance(original_model, quantized_model, sample_input):
        """Compare model size and inference speed."""
        # Model size
        original_size = sum(p.numel() * 4 for p in original_model.parameters()) / (1024**2)
        quantized_size = sum(p.numel() * 1 for p in quantized_model.parameters()) / (1024**2)

        # Inference speed
        import time
        with torch.no_grad():
            start = time.time()
            for _ in range(100):
                _ = original_model(sample_input)
            original_time = time.time() - start

            start = time.time()
            for _ in range(100):
                _ = quantized_model(sample_input)
            quantized_time = time.time() - start

        return {
            'original_size_mb': original_size,
            'quantized_size_mb': quantized_size,
            'size_reduction': f'{(1 - quantized_size/original_size)*100:.1f}%',
            'speedup': f'{original_time/quantized_time:.1f}x'
        }
```

### Model Pruning for Inference Speed

```python
import torch.nn.utils.prune as prune

class ModelPruner:
    """Remove unimportant weights to reduce model size."""

    @staticmethod
    def prune_model(model, amount=0.3):
        """Prune 30% of least important weights."""
        for module in model.modules():
            if isinstance(module, torch.nn.Conv2d) or isinstance(module, torch.nn.Linear):
                prune.l1_unstructured(module, name='weight', amount=amount)
                prune.remove(module, 'weight')

        return model

    @staticmethod
    def calculate_sparsity(model):
        """Calculate model sparsity (% of zero weights)."""
        total = 0
        zeros = 0

        for module in model.modules():
            if hasattr(module, 'weight'):
                total += module.weight.numel()
                zeros += (module.weight == 0).sum().item()

        return {'sparsity': f'{(zeros/total)*100:.1f}%'}
```

### ONNX Export for Cross-Platform Inference

```python
import torch
import onnx

class ONNXExporter:
    """Export PyTorch models to ONNX for optimal inference."""

    @staticmethod
    def export_to_onnx(model, sample_input, output_path='model.onnx'):
        """Export model to ONNX format."""
        torch.onnx.export(
            model,
            sample_input,
            output_path,
            export_params=True,
            opset_version=17,
            input_names=['input'],
            output_names=['output'],
            dynamic_axes={
                'input': {0: 'batch_size'},
                'output': {0: 'batch_size'}
            },
            do_constant_folding=True
        )

        # Verify ONNX model
        onnx_model = onnx.load(output_path)
        onnx.checker.check_model(onnx_model)

        return output_path
```

---

## Scaling Strategies

### Distributed Inference with Ray Serve

```python
from ray import serve
import ray

@serve.deployment(num_replicas=3)
class DistributedModelServer:
    """Scale inference across multiple GPUs with Ray Serve."""

    def __init__(self, model_path: str):
        self.model = self.load_model(model_path)

    async def __call__(self, request):
        """Handle inference request."""
        input_data = await request.json()
        prediction = self.model.predict(input_data)
        return {'prediction': float(prediction)}

# Deploy with auto-scaling
serve.start()
DistributedModelServer.options(
    num_replicas=3,
    max_concurrent_queries=100
).deploy()
```

### Feature Store for Efficient Data Pipelines

```python
from feast import FeatureStore, FeatureView, Entity
from feast.infra.offline_stores.file import FileOfflineStoreConfig

class MLOpsFeatureStore:
    """Centralized feature management for training and serving."""

    def __init__(self):
        self.fs = FeatureStore('feature_repo')

    def get_training_features(self, entities, event_timestamp):
        """Get features for training with point-in-time accuracy."""
        training_data = self.fs.get_historical_features(
            entity_df=entities,
            features=[
                'user_features:age',
                'user_features:monthly_spend',
                'transaction_features:avg_transaction_amount'
            ],
            event_timestamp_column='event_time'
        )

        return training_data.to_df()

    def get_online_features(self, entity_id):
        """Get features for real-time inference."""
        feature_vector = self.fs.get_online_features(
            features=[
                'user_features:age',
                'user_features:monthly_spend'
            ],
            entity_rows=[{'user_id': entity_id}]
        )

        return feature_vector.to_dict()
```

---

## Resource Management

### GPU Memory Management

```python
import torch
import gc

class GPUMemoryManager:
    """Optimize GPU memory usage during training."""

    @staticmethod
    def clear_gpu_cache():
        """Force garbage collection and clear GPU cache."""
        gc.collect()
        torch.cuda.empty_cache()
        torch.cuda.reset_peak_memory_stats()

    @staticmethod
    def monitor_gpu_memory():
        """Monitor GPU memory usage."""
        for i in range(torch.cuda.device_count()):
            props = torch.cuda.get_device_properties(i)
            allocated = torch.cuda.memory_allocated(i) / 1024**3
            total = props.total_memory / 1024**3

            print(f'GPU {i}: {allocated:.1f}GB / {total:.1f}GB')

    @staticmethod
    def enable_gradient_checkpointing(model):
        """Reduce memory by recomputing gradients instead of storing."""
        if hasattr(model, 'gradient_checkpointing_enable'):
            model.gradient_checkpointing_enable()
```

---

## Best Practices

**DO**:
- Use mixed precision training (1.5-3x speedup)
- Implement gradient accumulation for large models
- Profile inference before optimization
- Use quantization for mobile/edge deployment
- Monitor GPU memory during training
- Implement batch size auto-tuning
- Use feature stores for consistency

**DON'T**:
- Skip profiling before optimization
- Over-optimize prematurely
- Ignore numerical stability in quantization
- Train without monitoring resources
- Skip validation after optimization
- Use stale models in production
- Hardcode batch sizes without tuning

---

**Related Skills**: moai-domain-backend, moai-essentials-perf
