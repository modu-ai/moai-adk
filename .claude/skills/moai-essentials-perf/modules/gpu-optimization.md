# GPU Acceleration Optimization

Complete guide to GPU optimization patterns for performance-critical workloads.

## GPU Profiling with Scalene

```python
# GPU profiling setup
class GPUProfiler:
    async def profile_gpu(self, gpu_code):
        """Profile GPU code for optimization opportunities."""

        # Scalene GPU profiling
        result = self.run_scalene_gpu(
            gpu_code,
            flags=['--gpu', '--memory', '--html']
        )

        # Analyze results
        return {
            'gpu_utilization': result['gpu_util'],
            'memory_usage': result['gpu_memory'],
            'kernel_execution': result['kernel_time'],
            'data_transfer': result['host_device_time']
        }
```

## GPU Optimization Patterns

### Pattern 1: Kernel Optimization

```python
import numpy as np

# ❌ WRONG - Inefficient GPU kernel
def slow_matrix_operation():
    a = np.random.randn(10000, 10000)
    b = np.random.randn(10000, 10000)

    # Single kernel call (good)
    c = np.dot(a, b)  # Optimal on GPU

    # But with Python loop (bad)
    for i in range(1000):
        c = np.dot(c, b)  # 1000 kernel launches!

# ✅ CORRECT - Batched operations
def fast_matrix_operation():
    a = np.random.randn(10000, 10000)
    b_list = [np.random.randn(10000, 10000) for _ in range(1000)]

    # Batch process to reduce kernel overhead
    results = []
    for b in b_list:
        result = np.dot(a, b)
        results.append(result)

    return results
```

### Pattern 2: Memory Transfer Optimization

```python
import cupy as cp

class GPUMemoryOptimizer:
    def minimize_transfers(self, data):
        """Minimize host-device memory transfers."""

        # Keep data on GPU to avoid transfers
        gpu_data = cp.asarray(data)  # Transfer once

        # Do all GPU work on device
        result1 = self.gpu_operation1(gpu_data)  # GPU
        result2 = self.gpu_operation2(result1)   # GPU
        result3 = self.gpu_operation3(result2)   # GPU

        # Transfer result back (single transfer)
        return cp.asnumpy(result3)
```

### Pattern 3: Parallel GPU Operations

```python
import torch

class ParallelGPUProcessor:
    def parallel_gpu_compute(self, batches):
        """Process batches in parallel on GPU."""

        device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')

        results = []
        for batch in batches:
            # Load batch to GPU
            tensor = torch.from_numpy(batch).to(device)

            # Parallel GPU computation
            result = self.gpu_kernel(tensor)

            # Store result
            results.append(result.cpu().numpy())

        return results
```

## GPU Memory Management

### Reducing GPU Memory Footprint

```python
class GPUMemoryManager:
    def optimize_memory_usage(self, model, batch_size):
        """Optimize GPU memory usage."""

        # Gradient checkpointing (trade compute for memory)
        model.gradient_checkpointing_enable()

        # Smaller batch sizes
        small_batch_size = batch_size // 2

        # Mixed precision (FP16 + FP32)
        with torch.cuda.amp.autocast():
            output = model(data)

        return output

    def clear_gpu_cache(self):
        """Free GPU memory."""
        torch.cuda.empty_cache()
```

## Performance Monitoring

### GPU Performance Metrics

```python
class GPUMonitor:
    def monitor_gpu_performance(self):
        """Monitor GPU performance in real-time."""

        import pynvml

        pynvml.nvmlInit()
        device = pynvml.nvmlDeviceGetHandleByIndex(0)

        # Get GPU metrics
        util = pynvml.nvmlDeviceGetUtilizationRates(device)
        memory = pynvml.nvmlDeviceGetMemoryInfo(device)
        temp = pynvml.nvmlDeviceGetTemperature(device, 0)

        return {
            'gpu_util': util.gpu,
            'memory_used': memory.used / 1e9,  # GB
            'memory_total': memory.total / 1e9,
            'temperature': temp
        }
```

## GPU vs CPU Decision Framework

| Task | Ideal Platform | Why |
|------|---|---|
| Matrix operations (large) | GPU | Massive parallelism |
| Sequence processing (short) | CPU | Lower latency |
| Tensor computations | GPU | Native optimization |
| I/O bound tasks | CPU | GPU underutilized |
| Deep learning | GPU | CUDA/cuDNN optimized |
| Control flow heavy | CPU | Better for branches |

---

**Related Tools**: CUDA, cuDNN, Torch, TensorFlow

## CUDA Optimization Patterns (Advanced)

### CUDA Stream Parallelism

```python
import torch

class CUDAStreamOptimizer:
    """Optimize GPU workloads with CUDA streams."""

    def __init__(self):
        self.streams = [torch.cuda.Stream() for _ in range(4)]

    def parallel_inference(self, models, data_batches):
        """Run multiple models in parallel with CUDA streams."""

        results = []
        for idx, (model, data) in enumerate(zip(models, data_batches)):
            stream_idx = idx % len(self.streams)

            with torch.cuda.stream(self.streams[stream_idx]):
                # Async operations on different streams
                output = model(data.cuda(non_blocking=True))
                results.append(output.cpu(non_blocking=True))

        # Synchronize all streams
        torch.cuda.synchronize()

        return results
```

### Memory Pinning for Faster Transfer

```python
class CUDAMemoryOptimizer:
    """Optimize GPU memory transfer with pinned memory."""

    def optimize_data_transfer(self, dataset):
        """Use pinned memory for faster host-to-device transfer."""

        # Allocate pinned memory
        pinned_data = torch.empty(
            dataset.shape,
            pin_memory=True  # Faster transfer to GPU
        )
        pinned_data.copy_(dataset)

        # Async transfer
        gpu_data = pinned_data.cuda(non_blocking=True)

        return gpu_data
```

## cuDNN Optimization Patterns

### Automatic cuDNN Tuning

```python
import torch.backends.cudnn as cudnn

class CuDNNOptimizer:
    """Optimize neural network performance with cuDNN."""

    def __init__(self):
        # Enable cuDNN auto-tuner
        cudnn.benchmark = True  # Find best algorithms for your hardware

        # Deterministic mode (if reproducibility needed)
        cudnn.deterministic = False  # Faster, but non-deterministic

    def optimize_model(self, model):
        """Apply cuDNN optimizations to model."""

        model = model.cuda()

        # Enable cuDNN optimizations
        for module in model.modules():
            if isinstance(module, torch.nn.Conv2d):
                # Use cuDNN convolution
                module.groups = module.groups or 1

        return model
```

### Mixed Precision Training (cuDNN Tensor Cores)

```python
from torch.cuda.amp import autocast, GradScaler

class MixedPrecisionTrainer:
    """Leverage Tensor Cores with mixed precision."""

    def __init__(self):
        self.scaler = GradScaler()

    def train_step(self, model, data, target, optimizer):
        """Training step with automatic mixed precision."""

        optimizer.zero_grad()

        # Forward pass with autocasting
        with autocast():
            output = model(data)
            loss = criterion(output, target)

        # Backward pass with gradient scaling
        self.scaler.scale(loss).backward()
        self.scaler.step(optimizer)
        self.scaler.update()

        return loss.item()
```

## GPU Memory Management

### Memory Pool for Frequent Allocations

```python
class GPUMemoryPool:
    """Efficient GPU memory management with pooling."""

    def __init__(self, pool_size_mb=1024):
        self.pool_size = pool_size_mb * 1024 * 1024
        self.memory_pool = []

    def allocate_tensor(self, shape, dtype=torch.float32):
        """Allocate tensor from memory pool."""

        required_size = np.prod(shape) * torch.finfo(dtype).bits // 8

        # Reuse existing memory if available
        for idx, (tensor, size) in enumerate(self.memory_pool):
            if size >= required_size:
                self.memory_pool.pop(idx)
                return tensor.view(shape)

        # Allocate new memory
        new_tensor = torch.empty(shape, dtype=dtype, device='cuda')
        return new_tensor

    def release_tensor(self, tensor):
        """Return tensor to pool for reuse."""

        tensor_size = tensor.numel() * tensor.element_size()
        self.memory_pool.append((tensor, tensor_size))
```

## Best Practices Summary

✅ **DO**:
- Use CUDA streams for parallel GPU operations
- Enable cuDNN auto-tuner for optimal performance
- Apply mixed precision training on Tensor Core GPUs
- Pin memory for faster CPU-GPU transfers
- Profile with Scalene GPU metrics

❌ **DON'T**:
- Launch GPU kernels in tight Python loops
- Transfer data to GPU inside training loops
- Ignore GPU memory fragmentation
- Skip gradient scaling in mixed precision
- Use deterministic cuDNN without reason (slower)

