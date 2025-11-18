---
name: moai-ml-llm-fine-tuning
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: LoRA, QLoRA, PEFT techniques for LLM customization and domain adaptation with Llama 3.1, Mistral, Falcon models
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# LLM Fine-Tuning and Parameter-Efficient Adaptation — Enterprise v4.0

## Quick Summary

**Primary Focus**: Parameter-efficient fine-tuning methods (LoRA, QLoRA, PEFT) for large language models with production deployment strategies

**Best For**: Domain-specific LLM customization, cost-effective training, memory-constrained environments, multi-task learning

**Key Libraries**: PEFT 0.13+, Transformers 4.45+, bitsandbytes 0.45+, accelerate 0.34+, TRL 0.11+

**Key Models**: Llama 3.1 (8B/70B), Mistral 7B, Mixtral 8x7B, Falcon 7B/40B, OpenLLaMA 7B/13B

**Auto-triggers**: fine-tuning, LoRA, QLoRA, peft, llama, mistral, parameter-efficient, adapter, quantization

| Framework | Version | Release | Support |
|-----------|---------|---------|---------|
| PEFT | 0.13.0+ | 2025-10 | Active |
| Transformers | 4.45.0+ | 2025-11 | Active |
| bitsandbytes | 0.45.0+ | 2025-09 | Active |
| Llama 3.1 | 8B/70B | 2024-07 | Active |
| Mistral | 7B/Large | 2024-12 | Active |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core fine-tuning concepts with practical examples:

- **Full Fine-tuning vs Parameter-Efficient**: Comparison of memory, cost, and performance tradeoffs
- **LoRA Basics**: Low-rank decomposition, rank selection, intuition
- **QLoRA Basics**: Quantization fundamentals, 4-bit storage, scaling laws
- **Dataset Preparation**: Tokenization, batching, data validation
- **Training Setup**: Optimizer selection, learning rate, checkpointing
- **Examples**: See `examples.md` for complete code samples

### Level 2: Advanced Patterns (See reference.md)

Production-ready enterprise patterns:

- **LoRA Implementation**: PEFT configuration, hyperparameter tuning, multi-model adaptation
- **QLoRA with Large Models**: Llama 3.1 70B fine-tuning on consumer GPUs
- **PEFT Framework**: Multi-adapter composition, domain-specific task learning
- **Distributed Training**: Multi-GPU fine-tuning, gradient accumulation, deepspeed
- **Model Evaluation**: Perplexity, BLEU, ROUGE metrics, domain-specific assessment
- **Pattern Reference**: See `reference.md` for API details and enterprise patterns

### Level 3: Production Deployment (Consult security/performance skills)

Enterprise deployment and optimization:

- **Model Export**: HuggingFace → GGUF format, llama.cpp compatibility
- **Inference Optimization**: Ollama, vLLM, TGI servers
- **Quantized Inference**: CPU execution with 4-bit models, latency optimization
- **Cost Analysis**: Training cost vs inference cost, ROI calculation
- **Monitoring**: Loss tracking, inference latency, memory usage
- **Details**: Skill("moai-essentials-perf"), Skill("moai-security-backend")

---

## Technology Stack (November 2025 Stable)

### Core Libraries

- **PEFT 0.13.0** (Parameter-Efficient Fine-Tuning)
  - LoRA (Low-Rank Adaptation)
  - QLoRA (Quantized LoRA)
  - Prefix Tuning, Prompt Tuning, Adapters
  - Multi-task learning with adapters

- **Transformers 4.45.0** (HuggingFace)
  - AutoModelForCausalLM with quantization support
  - BitsAndBytesConfig for 4-bit/8-bit quantization
  - Trainer API with distributed support
  - Latest model implementations

- **bitsandbytes 0.45.0** (Quantization)
  - 8-bit optimizer (Linear8bitLt)
  - 4-bit quantization (nf4, fp4, int4)
  - CPU offloading, gradient checkpointing
  - Zero3 integration

- **accelerate 0.34.0** (Distributed Training)
  - Multi-GPU synchronization
  - DeepSpeed integration
  - Mixed precision training (FP16, BF16)
  - Checkpoint management

- **torch 2.2.0+** (PyTorch)
  - Flash Attention 2 integration
  - Automatic mixed precision
  - Gradient checkpointing
  - Tensor parallelism support

- **TRL 0.11.0** (Transformer Reinforcement Learning)
  - Supervised Fine-Tuning (SFT) trainer
  - DPO (Direct Preference Optimization)
  - ORPO, KTO training
  - Chat template support

### Models (Production-Ready)

- **Llama 3.1** (Meta)
  - 8B (11.5GB VRAM, 4-bit: 3GB)
  - 70B (140GB VRAM, 4-bit: 40GB)
  - Superior reasoning, coding, multilingual
  - Chat-optimized tokenizer

- **Mistral 7B** (Mistral AI)
  - Efficient, fast inference
  - Sliding window attention
  - 4K context window
  - Strong performance-efficiency ratio

- **Mixtral 8x7B** (Mistral AI)
  - Mixture of Experts architecture
  - 47B parameters (12B active per token)
  - Superior reasoning vs Llama equivalents
  - Advanced problem-solving

- **Falcon 7B/40B** (Technology Innovation Institute)
  - Multi-query attention (efficient inference)
  - Extended context windows
  - Strong code understanding
  - Commercially permissive license

- **OpenLLaMA 7B/13B** (Open source, Llama-compatible)
  - Free commercial use
  - PEFT-friendly architecture
  - Strong baseline for domain adaptation
  - Diverse training data

---

## 1. Fine-Tuning Fundamentals

### Full Fine-tuning vs Parameter-Efficient Methods

Fine-tuning updates model weights to adapt to specific domains or tasks. Two primary strategies exist:

**Full Fine-tuning**: Updates all model parameters (100% of weights)
- **Cost**: Extremely high (requires high-end GPUs: A100/H100)
- **Memory**: 4-6x model size (Llama 70B → 420-840GB VRAM)
- **Training Time**: Long (days/weeks for large models)
- **Performance**: Marginal improvement over parameter-efficient methods
- **Use Case**: Only when maximum performance is required (rare)

**Parameter-Efficient Fine-tuning**: Updates 0.1-1% of parameters via adapters
- **Cost**: Low (consumer GPUs: RTX 4090, A10)
- **Memory**: Minimal (Llama 70B 4-bit QLoRA → 24GB VRAM)
- **Training Time**: Hours (2-8 hours for reasonable-sized datasets)
- **Performance**: 95-99% of full fine-tuning performance
- **Use Case**: Standard production approach (recommended)

#### Example 1: Memory Consumption Comparison

```python
# memory_comparison.py
import torch
from transformers import AutoModelForCausalLM, BitsAndBytesConfig
from peft import get_peft_model, LoraConfig

def calculate_model_memory(model_name, method):
    """Calculate memory requirements for different fine-tuning approaches"""
    
    print(f"\n{method.upper()} - {model_name}")
    print("=" * 60)
    
    # Method 1: Full Fine-tuning (FP16)
    if method == "full_fp16":
        # 70B parameters × 2 bytes (FP16) = 140GB base
        # + optimizer state (Adam: 2x) = 420GB total
        base_model = 140  # GB
        optimizer_state = 280  # GB (2x for Adam)
        gradient = 140  # GB
        activation = 50  # GB (approximate)
        total = base_model + optimizer_state + gradient + activation
        
        print(f"Model weights (FP16):    {base_model} GB")
        print(f"Optimizer state (Adam):  {optimizer_state} GB")
        print(f"Gradients:               {gradient} GB")
        print(f"Activations (approx):    {activation} GB")
        print(f"TOTAL REQUIRED:          {total} GB")
        print(f"GPU needed:              8x H100 (80GB each)")
        
    # Method 2: QLoRA (4-bit quantization + LoRA)
    elif method == "qlora":
        # 70B parameters × 0.125 bytes (int4) = 8.75GB base
        # + LoRA params (0.1% additional) = ~0.5GB
        # + gradients for LoRA = 0.5GB
        # + optimizer state = 1GB
        quantized_model = 8.75
        lora_params = 0.5
        lora_gradients = 0.5
        lora_optimizer = 1.0
        kv_cache = 8.0  # KV cache during training
        total = quantized_model + lora_params + lora_gradients + lora_optimizer + kv_cache
        
        print(f"Model (int4 quantized):  {quantized_model} GB")
        print(f"LoRA weights (rank=8):   {lora_params} GB")
        print(f"LoRA gradients:          {lora_gradients} GB")
        print(f"LoRA optimizer state:    {lora_optimizer} GB")
        print(f"KV cache:                {kv_cache} GB")
        print(f"TOTAL REQUIRED:          {total} GB")
        print(f"GPU needed:              1x RTX 4090 (24GB)")
        
    # Method 3: LoRA (FP16 model + LoRA adapters)
    elif method == "lora":
        # 70B × 2 bytes = 140GB base
        # + LoRA params (0.5% additional) = 2-3GB
        # + gradients + optimizer = 20GB
        base_model = 140
        lora_params = 2.5
        gradients = 15
        optimizer = 20
        total = base_model + lora_params + gradients + optimizer
        
        print(f"Model weights (FP16):    {base_model} GB")
        print(f"LoRA weights (rank=16):  {lora_params} GB")
        print(f"Gradients & Optimizer:   {gradients + optimizer} GB")
        print(f"TOTAL REQUIRED:          {total} GB")
        print(f"GPU needed:              2x H100 (80GB each)")
        
    print()

# Main comparison
models = ["Llama 3.1 70B"]
methods = ["full_fp16", "lora", "qlora"]

for model in models:
    for method in methods:
        calculate_model_memory(model, method)

# Summary table
print("\nSUMMARY - Memory Requirements for Llama 3.1 70B Fine-tuning")
print("=" * 70)
print(f"{'Method':<20} {'Memory':<15} {'GPU':<30} {'Cost':<10}")
print("-" * 70)
print(f"{'Full Fine-tuning':<20} {'420 GB':<15} {'8x H100':<30} {'~$400/hr':<10}")
print(f"{'LoRA':<20} {'180 GB':<15} {'2-3x H100':<30} {'~$100/hr':<10}")
print(f"{'QLoRA':<20} {'20 GB':<15} {'1x RTX 4090':<30} {'~$2/hr':<10}")
print("-" * 70)
print("\nRecommendation: Use QLoRA for 99% of fine-tuning tasks")
```

### Dataset Preparation

High-quality datasets are critical for effective fine-tuning. Preparation involves:

1. **Data Collection**: Gather domain-specific examples (1K-50K examples typically)
2. **Formatting**: Convert to instruction-response pairs
3. **Tokenization**: Convert text to model-compatible token sequences
4. **Validation**: Check quality, remove corrupted examples
5. **Batching**: Group examples by length for efficiency

#### Example 2: HuggingFace Dataset Preparation

```python
# dataset_preparation.py
from datasets import Dataset, load_dataset
from transformers import AutoTokenizer
import numpy as np

def prepare_instruction_dataset(raw_data):
    """
    Prepare instruction-following dataset for fine-tuning
    
    Expected raw_data format:
    [
        {"instruction": "What is Python?", "response": "Python is a..."},
        {"instruction": "How to loop?", "response": "Use for..."},
    ]
    """
    
    # Create HuggingFace Dataset
    dataset = Dataset.from_list(raw_data)
    
    return dataset

def format_instruction_pairs(example, template="llama3"):
    """Format instruction-response pairs with model-specific templates"""
    
    instruction = example.get("instruction", "")
    response = example.get("response", "")
    
    if template == "llama3":
        # Llama 3 chat template
        text = f"""<|begin_of_text|><|start_header_id|>user<|end_header_id|>

{instruction}<|eot_id|><|start_header_id|>assistant<|end_header_id|>

{response}<|eot_id|>"""
        
    elif template == "mistral":
        # Mistral chat template
        text = f"""[INST] {instruction} [/INST]

{response} </s>"""
        
    elif template == "generic":
        # Generic template for custom models
        text = f"Instruction: {instruction}\nResponse: {response}"
    
    return {"text": text}

def tokenize_dataset(dataset, tokenizer, max_length=2048):
    """
    Tokenize dataset with padding/truncation
    
    Args:
        dataset: HuggingFace Dataset
        tokenizer: AutoTokenizer instance
        max_length: Maximum sequence length
    
    Returns:
        Tokenized dataset with "input_ids" and "attention_mask"
    """
    
    def tokenize_function(examples):
        # Tokenize with truncation and padding
        tokenized = tokenizer(
            examples["text"],
            truncation=True,
            max_length=max_length,
            padding="max_length",
            return_tensors=None
        )
        
        # Add labels (same as input_ids for causal language modeling)
        tokenized["labels"] = tokenized["input_ids"].copy()
        
        # Mask padding tokens in labels (-100 is ignored in loss calculation)
        for i, attention_mask in enumerate(tokenized["attention_mask"]):
            tokenized["labels"][i] = [
                token if mask == 1 else -100
                for token, mask in zip(tokenized["labels"][i], attention_mask)
            ]
        
        return tokenized
    
    # Apply tokenization with batching
    tokenized_dataset = dataset.map(
        tokenize_function,
        batched=True,
        batch_size=100,
        remove_columns=dataset.column_names,
        desc="Tokenizing dataset"
    )
    
    return tokenized_dataset

def validate_dataset(dataset, tokenizer, sample_size=5):
    """Validate tokenized dataset quality"""
    
    print(f"\nDataset Validation ({sample_size} samples)")
    print("=" * 70)
    
    # Check basic statistics
    print(f"Total examples: {len(dataset)}")
    print(f"Columns: {dataset.column_names}")
    
    # Sample and decode to verify
    sample_indices = np.random.choice(len(dataset), sample_size, replace=False)
    
    for idx in sample_indices:
        example = dataset[idx]
        
        # Get token count
        input_ids = example["input_ids"]
        token_count = sum(1 for x in input_ids if x != tokenizer.pad_token_id)
        
        # Decode first 100 tokens for verification
        decoded = tokenizer.decode(
            [x for x in input_ids if x != tokenizer.pad_token_id],
            skip_special_tokens=False
        )
        
        print(f"\nExample {idx}:")
        print(f"  Token count: {token_count}")
        print(f"  Preview: {decoded[:200]}...")
        print(f"  Labels: {example['labels'][:20]}...")
    
    print("\n" + "=" * 70)

# Usage example
if __name__ == "__main__":
    from transformers import AutoTokenizer
    
    # Load tokenizer
    tokenizer = AutoTokenizer.from_pretrained(
        "meta-llama/Llama-2-7b-hf",
        trust_remote_code=True,
        padding_side="right"
    )
    
    # Add padding token if missing
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token
    
    # Sample data
    raw_data = [
        {
            "instruction": "What is machine learning?",
            "response": "Machine learning is a subset of AI..."
        },
        {
            "instruction": "Explain neural networks",
            "response": "Neural networks are computing systems..."
        },
    ]
    
    # Prepare dataset
    dataset = prepare_instruction_dataset(raw_data)
    dataset = dataset.map(format_instruction_pairs)
    tokenized_dataset = tokenize_dataset(dataset, tokenizer)
    
    # Validate
    validate_dataset(tokenized_dataset, tokenizer)
```

### Training Setup

Configuring optimizer, learning rate, and checkpointing:

```python
# training_setup.py
from transformers import TrainingArguments
from torch.optim import AdamW
import torch

# Learning rate schedule
learning_rate = 2e-4  # Start with conservative LR

# Warmup strategy
# Llama 3.1: warmup_ratio = 0.03 (3% of total steps)
# Mistral: warmup_ratio = 0.1 (10% of total steps)
# Conservative: warmup_ratio = 0.1

# AdamW optimizer configuration
# - betas (0.9, 0.999): momentum and RMSprop beta (standard)
# - eps (1e-8): numerical stability
# - weight_decay (0.01): L2 regularization

training_args = TrainingArguments(
    output_dir="./outputs",
    
    # Learning rate
    learning_rate=2e-4,
    lr_scheduler_type="cosine",  # Linear, cosine, constant
    warmup_ratio=0.03,  # 3% warmup for faster convergence
    
    # Batch size and gradient accumulation
    per_device_train_batch_size=4,
    per_device_eval_batch_size=8,
    gradient_accumulation_steps=4,  # Effective batch: 4*4=16
    
    # Training schedule
    num_train_epochs=3,
    max_steps=-1,  # Set to positive for early stopping
    
    # Evaluation and checkpointing
    eval_strategy="steps",
    eval_steps=100,
    save_steps=100,
    save_total_limit=3,  # Keep only 3 recent checkpoints
    
    # Logging
    logging_steps=10,
    logging_dir="./logs",
    
    # Optimization
    optim="paged_adamw_8bit",  # 8-bit AdamW for VRAM efficiency
    max_grad_norm=1.0,
    gradient_checkpointing=True,  # Reduce memory at cost of speed
    
    # Mixed precision
    bf16=True,  # BF16 on newer GPUs (A100+)
    tf32=False,  # Disable for reproducibility
    
    # Distributed training
    dataloader_pin_memory=True,
    dataloader_num_workers=4,
    
    # Reporting
    report_to=["tensorboard"],
    push_to_hub=False,
)
```

### Evaluation Metrics

Metrics for assessing fine-tuning quality:

```python
# evaluation_metrics.py
import numpy as np
from nltk.translate.bleu_score import corpus_bleu
from rouge_score import rouge_scorer
from datasets import load_metric

def calculate_perplexity(predictions, references):
    """
    Perplexity measures model confidence on held-out data
    
    Lower is better. Typical ranges:
    - Base model: 15-30 (on test set)
    - After fine-tuning: 8-15 (domain-specific)
    - Excellent fine-tuning: 5-10
    """
    
    # Cross-entropy loss
    loss = np.mean([
        -np.log(pred) for pred in predictions
    ])
    
    # Perplexity = exp(loss)
    perplexity = np.exp(loss)
    
    return perplexity

def calculate_bleu(predictions, references):
    """
    BLEU measures n-gram overlap between prediction and reference
    
    Typical ranges:
    - Poor (0-20): Significant differences
    - Fair (20-40): Some overlap
    - Good (40-60): Strong similarity
    - Excellent (60+): Nearly identical
    """
    
    # Convert to format expected by corpus_bleu
    refs = [[ref.split()] for ref in references]
    preds = [pred.split() for pred in predictions]
    
    bleu = corpus_bleu(refs, preds)
    return bleu

def calculate_rouge(predictions, references):
    """
    ROUGE measures overlap and quality
    
    ROUGE-1: Unigram overlap
    ROUGE-2: Bigram overlap
    ROUGE-L: Longest common subsequence
    """
    
    scorer = rouge_scorer.RougeScorer(['rouge1', 'rouge2', 'rougeL'])
    
    scores = {"rouge1": [], "rouge2": [], "rougeL": []}
    
    for pred, ref in zip(predictions, references):
        score = scorer.score(ref, pred)
        
        for key in scores:
            scores[key].append(score[key].fmeasure)
    
    # Average across corpus
    return {k: np.mean(v) for k, v in scores.items()}

def compute_metrics(predictions, references):
    """Compute all metrics for fine-tuning evaluation"""
    
    print("\nFine-Tuning Evaluation Metrics")
    print("=" * 70)
    
    perplexity = calculate_perplexity(predictions, references)
    bleu = calculate_bleu(predictions, references)
    rouge = calculate_rouge(predictions, references)
    
    print(f"Perplexity:        {perplexity:.2f} (lower is better)")
    print(f"BLEU Score:        {bleu:.4f} (higher is better)")
    print(f"ROUGE-1 (F1):      {rouge['rouge1']:.4f}")
    print(f"ROUGE-2 (F1):      {rouge['rouge2']:.4f}")
    print(f"ROUGE-L (F1):      {rouge['rougeL']:.4f}")
    print("=" * 70)
    
    return {
        "perplexity": perplexity,
        "bleu": bleu,
        "rouge": rouge
    }
```

---

## 2. LoRA (Low-Rank Adaptation)

### LoRA Principles and Intuition

LoRA enables efficient fine-tuning by adding small trainable matrices alongside frozen model weights:

**Core Idea**: Instead of updating all weights W → W + ΔW, introduce trainable low-rank matrices:

```
ΔW = AB^T  (where A and B are small matrices)

Original: h = Wx
With LoRA: h = Wx + ABx^T

Benefits:
- A and B have much fewer parameters (rank r << dim)
- Original W stays frozen (no backward passes)
- Can save/load only A, B (small files)
```

**Rank Selection**:
- **rank=8**: Default, works for most tasks, minimal overhead
- **rank=16**: Better for complex domain shifts
- **rank=32+**: Only for highly specialized domains

**Alpha (Scaling)**: Controls contribution strength
- **alpha=16**: Standard (2x rank), scales LoRA output
- **alpha=32**: Stronger adaptation (4x rank)
- Conservative: alpha=rank

#### Example 3: PEFT LoRA Configuration

```python
# lora_config.py
from peft import LoraConfig, TaskType

# Llama 3.1 LoRA configuration
llama3_lora_config = LoraConfig(
    # Task type
    task_type=TaskType.CAUSAL_LM,
    
    # LoRA rank and alpha
    r=8,                    # Rank of low-rank matrices
    lora_alpha=16,          # Scaling factor (typically 2*r)
    
    # Which layers to fine-tune
    target_modules=[
        "q_proj",           # Query projection
        "v_proj",           # Value projection
        "k_proj",           # Key projection (optional)
        "o_proj",           # Output projection (optional)
        "up_proj",          # Feed-forward up projection
        "down_proj"         # Feed-forward down projection
    ],
    
    # LoRA dropout
    lora_dropout=0.05,      # Dropout in LoRA layers
    
    # Bias configuration
    bias="none",            # "none", "all", "lora_only"
    
    # Module names to exclude
    modules_to_save=["embed_tokens", "lm_head"],  # Save full weights
    
    # Inference mode
    inference_mode=False,   # False during training
)

# Mistral 7B LoRA configuration (similar but different targets)
mistral_lora_config = LoraConfig(
    task_type=TaskType.CAUSAL_LM,
    r=8,
    lora_alpha=16,
    target_modules=["q_proj", "v_proj"],  # Mistral uses fewer proj layers
    lora_dropout=0.05,
    bias="none",
)

# Multi-task learning: Different adapters for different domains
domain_specific_configs = {
    "code": LoraConfig(
        task_type=TaskType.CAUSAL_LM,
        r=16,  # Higher rank for code (more complex patterns)
        lora_alpha=32,
        target_modules=["q_proj", "v_proj", "up_proj", "down_proj"],
        lora_dropout=0.1,
    ),
    "medical": LoraConfig(
        task_type=TaskType.CAUSAL_LM,
        r=8,  # Standard rank
        lora_alpha=16,
        target_modules=["q_proj", "v_proj"],
        lora_dropout=0.05,
    ),
}

print("LoRA configurations prepared for different domains")
```

#### Example 4: LoRA Fine-Tuning Pipeline for Llama 2

```python
# lora_finetuning_llama.py
import torch
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    TrainingArguments,
)
from trl import SFTTrainer  # Supervised Fine-Tuning Trainer
from peft import LoraConfig, get_peft_model, TaskType

def fine_tune_llama_lora(
    model_name="meta-llama/Llama-2-7b-hf",
    dataset_path="./data/training_data.json",
    output_dir="./llama-lora-finetuned"
):
    """
    Complete LoRA fine-tuning pipeline for Llama 2
    
    Args:
        model_name: HuggingFace model identifier
        dataset_path: Path to training data (JSONL format)
        output_dir: Directory to save fine-tuned model
    """
    
    # 1. Load tokenizer
    tokenizer = AutoTokenizer.from_pretrained(
        model_name,
        trust_remote_code=True,
        padding_side="right"
    )
    
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token
    
    print(f"Loaded tokenizer: {model_name}")
    
    # 2. Load model
    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        torch_dtype=torch.float16,
        device_map="auto",
        trust_remote_code=True
    )
    
    print(f"Loaded model: {model_name}")
    
    # 3. Configure LoRA
    lora_config = LoraConfig(
        r=8,
        lora_alpha=16,
        target_modules=["q_proj", "v_proj", "k_proj", "o_proj"],
        lora_dropout=0.05,
        bias="none",
        task_type=TaskType.CAUSAL_LM,
    )
    
    # 4. Wrap model with LoRA
    model = get_peft_model(model, lora_config)
    
    trainable_params = sum(p.numel() for p in model.parameters() if p.requires_grad)
    total_params = sum(p.numel() for p in model.parameters())
    
    print(f"\nLoRA Configuration:")
    print(f"  Trainable params: {trainable_params:,}")
    print(f"  Total params: {total_params:,}")
    print(f"  Trainable %: {100.0 * trainable_params / total_params:.2f}%")
    
    # 5. Configure training arguments
    training_args = TrainingArguments(
        output_dir=output_dir,
        num_train_epochs=3,
        per_device_train_batch_size=4,
        per_device_eval_batch_size=4,
        gradient_accumulation_steps=4,
        learning_rate=2e-4,
        lr_scheduler_type="cosine",
        warmup_ratio=0.03,
        optim="paged_adamw_8bit",
        max_grad_norm=1.0,
        bf16=True,
        logging_steps=10,
        save_steps=100,
        eval_strategy="steps",
        eval_steps=100,
        save_total_limit=3,
    )
    
    # 6. Create SFT Trainer
    trainer = SFTTrainer(
        model=model,
        tokenizer=tokenizer,
        args=training_args,
        train_dataset=None,  # Load from dataset_path
        dataset_text_field="text",
        max_seq_length=2048,
        packing=True,  # Pack multiple sequences to maximize GPU usage
    )
    
    # 7. Fine-tune
    print("\nStarting fine-tuning...")
    trainer.train()
    
    # 8. Save model and adapter
    print("\nSaving fine-tuned model...")
    
    # Save full model
    trainer.save_model(f"{output_dir}/full_model")
    
    # Save only LoRA adapter (lightweight, ~10MB)
    model.save_pretrained(f"{output_dir}/adapter")
    
    print(f"\nFine-tuning complete!")
    print(f"  Full model: {output_dir}/full_model")
    print(f"  Adapter only: {output_dir}/adapter (~10MB)")
    
    return model, tokenizer

# Usage
if __name__ == "__main__":
    model, tokenizer = fine_tune_llama_lora(
        model_name="meta-llama/Llama-2-7b-hf",
        dataset_path="./training_data.jsonl",
        output_dir="./llama-lora-finetuned"
    )
```

---

## 3. QLoRA (Quantized LoRA)

### Quantization Fundamentals

Quantization reduces model size by storing weights in lower precision formats:

**Standard Approaches**:
- **FP32 (32-bit)**: Full precision, high memory, slow
- **FP16 (16-bit)**: Half precision, standard for training
- **BF16 (Brain Float 16)**: Truncated FP32, good for numerical stability
- **8-bit**: 8x smaller, maintains quality, enables consumer GPU fine-tuning
- **4-bit (NormalFloat)**: 16x smaller, minimal quality loss, enables 70B models on RTX 4090

**QLoRA Innovation**: Combine quantization with LoRA for extreme efficiency

```
Normal LoRA: 70B model (140GB) + LoRA (1GB) = 141GB
QLoRA: 70B model (8.75GB) + LoRA (0.5GB) = 9.25GB → 93% memory reduction!
```

#### Example 5: QLoRA 4-bit Quantization Setup

```python
# qlora_config.py
from transformers import BitsAndBytesConfig, AutoModelForCausalLM
from peft import LoraConfig, get_peft_model, TaskType

# 4-bit Quantization Configuration
bnb_config = BitsAndBytesConfig(
    load_in_4bit=True,
    bnb_4bit_quant_type="nf4",              # NormalFloat4 (optimal for LLMs)
    bnb_4bit_compute_dtype=torch.bfloat16,  # Compute in BF16 for stability
    bnb_4bit_use_double_quant=True,         # Double quantization for extra savings
)

# Load 70B model with QLoRA in just 24GB VRAM!
model = AutoModelForCausalLM.from_pretrained(
    "meta-llama/Llama-2-70b-hf",
    quantization_config=bnb_config,
    device_map="auto",
    trust_remote_code=True
)

# Add LoRA on top
lora_config = LoraConfig(
    r=8,
    lora_alpha=16,
    target_modules=["q_proj", "v_proj"],
    lora_dropout=0.05,
    bias="none",
    task_type=TaskType.CAUSAL_LM,
)

model = get_peft_model(model, lora_config)

print("QLoRA Configuration Applied:")
print(f"  Quantization: 4-bit NormalFloat (nf4)")
print(f"  Double Quantization: Enabled")
print(f"  Compute dtype: bfloat16")
print(f"  Memory usage: ~24GB for Llama 70B")

# 8-bit Alternative (for less aggressive quantization)
bnb_config_8bit = BitsAndBytesConfig(
    load_in_8bit=True,
    # No compute_dtype needed for 8-bit
)

# Use case: When 4-bit causes slight accuracy loss
model_8bit = AutoModelForCausalLM.from_pretrained(
    "meta-llama/Llama-2-13b-hf",
    quantization_config=bnb_config_8bit,
    device_map="auto"
)
```

#### Example 6: Llama 3.1 QLoRA Fine-Tuning

```python
# llama31_qlora_finetuning.py
import torch
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    BitsAndBytesConfig,
    TrainingArguments,
)
from trl import SFTTrainer
from peft import LoraConfig, get_peft_model, TaskType
from datasets import load_dataset

def fine_tune_llama31_qlora(
    model_size="8b",  # "8b" or "70b"
    dataset_name="openassistant-guanaco",  # HuggingFace dataset
    learning_rate=2e-4,
    num_epochs=3,
):
    """
    QLoRA fine-tuning for Llama 3.1
    
    Memory requirements:
    - Llama 3.1 8B: 10-12GB VRAM (RTX 3090/4090)
    - Llama 3.1 70B: 24-32GB VRAM (RTX 4090/A100)
    """
    
    # Model selection
    model_mapping = {
        "8b": "meta-llama/Llama-2-8b-hf",
        "70b": "meta-llama/Llama-2-70b-hf",
    }
    model_name = model_mapping[model_size]
    
    print(f"\n{'='*70}")
    print(f"QLoRA Fine-Tuning: {model_name}")
    print(f"{'='*70}")
    
    # 1. Quantization configuration (4-bit)
    bnb_config = BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_quant_type="nf4",
        bnb_4bit_compute_dtype=torch.bfloat16,
        bnb_4bit_use_double_quant=True,
    )
    
    # 2. Load tokenizer
    tokenizer = AutoTokenizer.from_pretrained(
        model_name,
        trust_remote_code=True,
        padding_side="right",
        add_eos_token=True,
    )
    
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token
    
    # 3. Load quantized model
    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        quantization_config=bnb_config,
        torch_dtype=torch.bfloat16,
        device_map="auto",
        trust_remote_code=True,
    )
    
    # 4. Prepare model for training
    model.config.use_cache = False  # Disable cache during training
    model.config.pretraining_tp = 1  # Tensor parallelism size
    
    # 5. LoRA configuration
    lora_config = LoraConfig(
        r=8,
        lora_alpha=16,
        target_modules=["q_proj", "v_proj", "k_proj", "o_proj"],
        lora_dropout=0.05,
        bias="none",
        task_type=TaskType.CAUSAL_LM,
    )
    
    # 6. Wrap model with LoRA
    model = get_peft_model(model, lora_config)
    
    # Print trainable parameters
    trainable = sum(p.numel() for p in model.parameters() if p.requires_grad)
    total = sum(p.numel() for p in model.parameters())
    
    print(f"\nModel Information:")
    print(f"  Total parameters: {total:,}")
    print(f"  Trainable parameters: {trainable:,}")
    print(f"  Trainable %: {100.0 * trainable / total:.2f}%")
    
    # 7. Load dataset
    dataset = load_dataset(dataset_name, split="train")
    
    # 8. Training arguments
    training_args = TrainingArguments(
        output_dir=f"./llama{model_size}_qlora",
        num_train_epochs=num_epochs,
        per_device_train_batch_size=4,
        gradient_accumulation_steps=4,
        learning_rate=learning_rate,
        lr_scheduler_type="cosine",
        warmup_ratio=0.03,
        optim="paged_adamw_8bit",
        max_grad_norm=1.0,
        max_steps=-1,
        logging_steps=25,
        save_steps=500,
        save_total_limit=3,
        bf16=True,
        logging_dir="./logs",
        report_to=["tensorboard"],
    )
    
    # 9. Create trainer
    trainer = SFTTrainer(
        model=model,
        tokenizer=tokenizer,
        args=training_args,
        train_dataset=dataset,
        dataset_text_field="text",
        max_seq_length=2048,
        packing=True,
    )
    
    # 10. Fine-tune
    print(f"\nStarting QLoRA fine-tuning...")
    trainer.train()
    
    # 11. Save
    print(f"\nSaving QLoRA adapter...")
    model.save_pretrained(f"./llama{model_size}_qlora/adapter")
    tokenizer.save_pretrained(f"./llama{model_size}_qlora/tokenizer")
    
    return model, tokenizer

# Performance Comparison
print("""
QLoRA vs Full Fine-tuning (Llama 3.1 70B):

Metric              QLoRA       Full FT    Ratio
─────────────────────────────────────────────
Memory              24GB        420GB      17.5x
Training Time       8 hours     3 days     9x
GPU Cost/hour       $2          $400       200x
Final Quality       99%         100%       -1%

Verdict: QLoRA provides 99% of full FT quality with 200x cost savings!
""")
```

---

## 4. PEFT (Parameter-Efficient Fine-Tuning) Framework

### PEFT Overview

PEFT (Parameter-Efficient Fine-Tuning) is the HuggingFace library providing multiple parameter-efficient methods:

**Available Methods**:

1. **LoRA**: Low-rank matrices alongside weights
2. **QLoRA**: LoRA + quantization
3. **Prefix Tuning**: Learnable prefix tokens for each layer
4. **Prompt Tuning**: Prepend learnable vectors to input embeddings
5. **Adapters**: Small feed-forward networks inserted in layers
6. **IA3**: Infuse Adapter by Inhibiting and Amplifying Inner Activations
7. **LLaMA-Adapter**: Adapter for Llama models with residual connections

**Comparison Table**:

| Method | Trainable % | Memory | Speed | Use Case |
|--------|------------|--------|-------|----------|
| LoRA | 0.1-1% | 95% reduction | Fast | Standard (recommended) |
| QLoRA | 0.05-0.5% | 98% reduction | Medium | Extreme constraints |
| Prefix | 0.5-2% | 90% reduction | Slower | Sequence prepending |
| Prompt | 0.01-0.1% | 99% reduction | Fast | Prompt optimization |
| Adapters | 1-3% | 90% reduction | Fast | Modular learning |

#### Example 7: PEFT Multi-Adapter Configuration

```python
# peft_multi_adapter.py
from peft import LoraConfig, TaskType, get_peft_model
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch

def setup_multi_task_adapters(model_name="meta-llama/Llama-2-7b-hf"):
    """
    Set up multiple task-specific adapters for multi-task learning
    
    Allows a single model to handle different domains by switching adapters
    """
    
    print(f"\nMulti-Task Adapter Setup for {model_name}")
    print("=" * 70)
    
    # Load base model
    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        torch_dtype=torch.float16,
        device_map="auto",
    )
    
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    
    # Define task-specific adapters
    adapters = {
        "code": {
            "config": LoraConfig(
                r=16,
                lora_alpha=32,
                target_modules=["q_proj", "v_proj", "up_proj", "down_proj"],
                lora_dropout=0.1,
                task_type=TaskType.CAUSAL_LM,
            ),
            "description": "Code generation and completion"
        },
        "medical": {
            "config": LoraConfig(
                r=8,
                lora_alpha=16,
                target_modules=["q_proj", "v_proj"],
                lora_dropout=0.05,
                task_type=TaskType.CAUSAL_LM,
            ),
            "description": "Medical question answering"
        },
        "legal": {
            "config": LoraConfig(
                r=8,
                lora_alpha=16,
                target_modules=["q_proj", "v_proj"],
                lora_dropout=0.05,
                task_type=TaskType.CAUSAL_LM,
            ),
            "description": "Legal document analysis"
        },
    }
    
    # Add adapters to model
    for task_name, adapter_info in adapters.items():
        # Add first adapter directly
        if task_name == "code":
            model = get_peft_model(model, adapter_info["config"])
            model.set_adapter("code")
        else:
            # Add subsequent adapters
            model.add_adapter(task_name, adapter_info["config"])
            print(f"Added adapter: {task_name}")
            print(f"  Description: {adapter_info['description']}")
    
    # List all adapters
    print(f"\nAvailable adapters: {model.active_adapters()}")
    
    # Switch between adapters
    print(f"\nSwitching to 'medical' adapter...")
    model.set_adapter("medical")
    
    print(f"Current active adapter: {model.active_adapters()}")
    
    # Combine adapters (sequential composition)
    print(f"\nCombining adapters (medical + legal)...")
    # Requires special handling - see PEFT documentation
    
    return model, tokenizer, adapters

def demonstrate_adapter_switching():
    """Demonstrate switching between task-specific adapters"""
    
    model, tokenizer, adapters = setup_multi_task_adapters()
    
    prompts = {
        "code": "def fibonacci(n):",
        "medical": "What are the symptoms of diabetes?",
        "legal": "What is the definition of negligence?",
    }
    
    for task, prompt in prompts.items():
        print(f"\n{'='*70}")
        print(f"Task: {task}")
        print(f"Prompt: {prompt}")
        print(f"{'='*70}")
        
        # Switch adapter
        model.set_adapter(task)
        
        # Generate with active adapter
        inputs = tokenizer(prompt, return_tensors="pt")
        
        with torch.no_grad():
            outputs = model.generate(
                **inputs,
                max_length=100,
                temperature=0.7,
                top_p=0.9,
            )
        
        response = tokenizer.decode(outputs[0], skip_special_tokens=True)
        print(f"Response: {response}")

if __name__ == "__main__":
    demonstrate_adapter_switching()
```

### Adapter Merging and Model Export

```python
# adapter_export.py
from peft import AutoPeftModelForCausalLM
from transformers import AutoModelForCausalLM

def merge_adapter_into_base(
    adapter_dir="./code_adapter",
    base_model_dir="meta-llama/Llama-2-7b",
    output_dir="./llama-code-merged"
):
    """Merge LoRA adapter into base model weights"""
    
    # Load model with adapter
    model = AutoPeftModelForCausalLM.from_pretrained(
        adapter_dir,
        device_map="auto"
    )
    
    # Merge
    merged_model = model.merge_and_unload()
    
    # Save merged model
    merged_model.save_pretrained(output_dir)
    
    print(f"Merged model saved to {output_dir}")
    print(f"File size: {merged_model.get_memory_footprint() / 1e9:.2f} GB")

def save_only_adapter(
    model_with_adapter,
    adapter_name="code",
    save_dir="./code_adapter"
):
    """Save only the adapter weights (lightweight, ~10MB)"""
    
    # Save adapter
    model_with_adapter.save_pretrained(save_dir)
    
    print(f"Adapter saved to {save_dir}")
    import os
    size = sum(os.path.getsize(os.path.join(save_dir, f)) 
               for f in os.listdir(save_dir)) / 1e6
    print(f"Adapter size: {size:.2f} MB")

def load_adapter_for_inference(
    base_model_name="meta-llama/Llama-2-7b",
    adapter_dir="./code_adapter"
):
    """Load base model with adapter for inference"""
    
    # Load base model
    model = AutoModelForCausalLM.from_pretrained(
        base_model_name,
        device_map="auto"
    )
    
    # Load adapter on top
    from peft import PeftModel
    model = PeftModel.from_pretrained(
        model,
        adapter_dir
    )
    
    return model
```

---

## 5. Training Optimization

### Gradient Accumulation

Allows training with larger effective batch sizes on limited VRAM:

#### Example 8: Gradient Accumulation Configuration

```python
# gradient_accumulation.py
from transformers import TrainingArguments

# Scenario: Want batch_size=64 but GPU only fits batch_size=16
# Solution: gradient_accumulation_steps=4

training_args = TrainingArguments(
    # Actual batch size per GPU
    per_device_train_batch_size=16,
    
    # Accumulate gradients across 4 steps
    gradient_accumulation_steps=4,
    
    # Effective batch size: 16 * 4 = 64
    # Equivalent to single-GPU batch_size=64
    
    # For distributed training (2 GPUs):
    # Effective batch: 16 * 4 * 2 = 128
)

# Memory implications:
# - Activations stored: ~batch_size (dynamic based on seq_len)
# - Gradients stored: ~model_size (fixed)
# - With grad checkpointing: activations can be recomputed (save 80%)

# Trade-off:
# - Benefit: Larger effective batch → better convergence
# - Cost: Slower training (more optimizer steps)

# Best practice: Use highest batch_size that fits in VRAM
# Then accumulate if you want larger effective batch
```

### Mixed Precision Training

Using different precisions for different operations:

```python
# mixed_precision_training.py
from transformers import TrainingArguments

training_args = TrainingArguments(
    # BF16 (recommended for modern GPUs: A100, H100, RTX 4090)
    bf16=True,
    
    # FP16 (for older GPUs: V100, RTX 2080)
    # fp16=True,  # Don't set both to True!
    
    # Both FP16 and BF16 are disabled by default (full FP32)
)

# Mixed Precision Breakdown:
# Forward pass: BF16 (half memory)
# Loss computation: FP32 (numerical stability)
# Backward pass: BF16 (half memory)
# Optimizer: BF16 (half memory)

# Benefits:
# - 50% memory reduction
# - 2-3x faster on modern GPUs with tensor cores
# - Maintained accuracy (BF16 better than FP16 for stability)

# When to use:
# BF16: GPU with native support (A100, H100, RTX 4090, RTX 3090)
# FP16: Older GPUs, ensure proper scaling/loss clipping
# FP32: When accuracy critical, older GPUs without FP16 hardware
```

### Distributed Training

Training across multiple GPUs with DeepSpeed:

```python
# distributed_training.py
from transformers import TrainingArguments

training_args = TrainingArguments(
    # For 2-4 GPU setup
    
    # Distributed strategy
    ddp_backend="nccl",  # Multi-GPU sync
    ddp_find_unused_parameters=False,
    
    # DeepSpeed integration (optional)
    deepspeed=None,  # Can provide JSON config file
    
    # Gradient checkpointing
    gradient_checkpointing=True,
    
    # Mixed precision
    bf16=True,
)

# Expected speedup:
# 2 GPUs: 1.8x (slight overhead)
# 4 GPUs: 3.5-3.8x
# 8+ GPUs: Diminishing returns (communication overhead)

# For multi-GPU training on single machine:
# python -m torch.distributed.launch \
#   --nproc_per_node=2 train.py

# For multi-machine setup:
# torchrun --nproc_per_node=2 --nnodes=4 train.py
```

---

## 6. Model Selection and Preparation

### Llama 3.1 Fine-Tuning

Llama 3.1 is the flagship open-source model with excellent fine-tuning characteristics:

**Strengths**:
- Superior reasoning (vs older Llama 2)
- Excellent coding capabilities
- 128K context window (8x Llama 2)
- Multilingual support
- Strong instruction-following

**Characteristics**:
- Uses byte-level BPE tokenization
- Chat-specific special tokens: `<|start_header_id|>`, `<|end_header_id|>`, `<|eot_id|>`
- Different from Llama 2 (not backward compatible)
- Requires specific chat template formatting

#### Example 9: Llama 3.1 Chat Template

```python
# llama31_chat_template.py
from transformers import AutoTokenizer

def format_llama31_instruction(instruction, response=""):
    """Format instruction-response for Llama 3.1 chat"""
    
    # Llama 3.1 official chat template
    text = f"""<|begin_of_text|><|start_header_id|>user<|end_header_id|>

{instruction}<|eot_id|><|start_header_id|>assistant<|end_header_id|>

{response}<|eot_id|>"""
    
    return text

def format_llama31_system(system_prompt, instruction, response=""):
    """Format with system prompt (for advanced scenarios)"""
    
    text = f"""<|begin_of_text|><|start_header_id|>system<|end_header_id|>

{system_prompt}<|eot_id|><|start_header_id|>user<|end_header_id|>

{instruction}<|eot_id|><|start_header_id|>assistant<|end_header_id|>

{response}<|eot_id|>"""
    
    return text

def format_llama31_multiturn(conversations):
    """Format multi-turn conversation
    
    Args:
        conversations: List of {"role": "user"/"assistant", "content": "..."}
    """
    
    messages = []
    text = "<|begin_of_text|>"
    
    for conv in conversations:
        role = conv["role"]
        content = conv["content"]
        
        text += f"<|start_header_id|>{role}<|end_header_id|>\n\n"
        text += content
        text += "<|eot_id|>"
    
    return text

# Tokenizer configuration for Llama 3.1
def setup_llama31_tokenizer(model_name="meta-llama/Llama-3.1-8B"):
    """Proper tokenizer setup for Llama 3.1"""
    
    tokenizer = AutoTokenizer.from_pretrained(
        model_name,
        trust_remote_code=True,
        padding_side="right",  # Right-pad for training
        add_eos_token=True,    # Add EOS to prompts
    )
    
    # Llama 3.1 doesn't have a pad token by default
    if tokenizer.pad_token is None:
        # Option 1: Use EOS as pad (common)
        tokenizer.pad_token = tokenizer.eos_token
        
        # Option 2: Add special pad token
        # tokenizer.add_special_tokens({"pad_token": "[PAD]"})
    
    return tokenizer

# Verify format
if __name__ == "__main__":
    tokenizer = setup_llama31_tokenizer()
    
    # Test formatting
    prompt = "What is Python?"
    response = "Python is a programming language..."
    
    formatted = format_llama31_instruction(prompt, response)
    print("Formatted prompt:")
    print(formatted)
    print("\n" + "="*70)
    
    # Tokenize and check
    tokens = tokenizer.encode(formatted)
    print(f"Token count: {len(tokens)}")
    print(f"Special tokens: {tokenizer.special_tokens_map}")
```

### Mistral and Mixtral Models

**Mistral 7B**:
- More efficient than Llama equivalent
- Sliding window attention (4K context)
- Fast inference
- Strong for code

**Mixtral 8x7B**:
- Mixture of Experts (MoE) architecture
- 12B dense model performance with 47B total size
- Best-in-class reasoning for size
- More expensive to fine-tune (need to activate all experts)

```python
# model_selection.py
MODEL_PROFILES = {
    "llama3-8b": {
        "model": "meta-llama/Llama-3.1-8B",
        "memory_fp16": 20,  # GB
        "memory_int4": 4,
        "inference_speed": "fast",
        "best_for": "code, general purpose",
        "recommendation": "First choice for most tasks"
    },
    "llama3-70b": {
        "model": "meta-llama/Llama-3.1-70B",
        "memory_fp16": 140,
        "memory_int4": 40,  # With QLoRA
        "inference_speed": "medium",
        "best_for": "complex reasoning, code, large domains",
        "recommendation": "When accuracy critical, with QLoRA"
    },
    "mistral-7b": {
        "model": "mistralai/Mistral-7B-v0.1",
        "memory_fp16": 16,
        "memory_int4": 3,
        "inference_speed": "very_fast",
        "best_for": "speed-critical, code",
        "recommendation": "When latency matters"
    },
    "mixtral-8x7b": {
        "model": "mistralai/Mixtral-8x7B-v0.1",
        "memory_fp16": 100,
        "memory_int4": 25,  # With QLoRA
        "inference_speed": "medium",
        "best_for": "complex tasks, reasoning",
        "recommendation": "Best reasoning per token"
    },
}

def select_model(task_requirements):
    """Select model based on requirements"""
    
    print("\nModel Selection Helper")
    print("="*70)
    
    # Check GPU memory
    gpu_memory = task_requirements.get("gpu_memory_gb")
    
    if gpu_memory < 8:
        return MODEL_PROFILES["llama3-8b"]
    elif gpu_memory < 24:
        if task_requirements.get("speed_critical"):
            return MODEL_PROFILES["mistral-7b"]
        return MODEL_PROFILES["llama3-8b"]
    elif gpu_memory < 40:
        return MODEL_PROFILES["llama3-70b"]  # Use QLoRA
    else:
        return MODEL_PROFILES["mixtral-8x7b"]  # Use QLoRA
```

---

## 7. Evaluation and Validation

### Evaluation Metrics

Comprehensive evaluation for fine-tuned models:

#### Example 10: Evaluation Metrics Implementation

```python
# evaluation_complete.py
import numpy as np
import torch
from datasets import load_dataset
from transformers import AutoModelForCausalLM, AutoTokenizer
from tqdm import tqdm

def compute_perplexity_batch(model, tokenizer, texts, device="cuda"):
    """
    Compute perplexity on batch of texts
    
    Perplexity measures model confidence:
    - Lower is better
    - Typical: Base model 15-30, Fine-tuned 8-15
    """
    
    total_loss = 0
    total_tokens = 0
    
    for text in tqdm(texts, desc="Computing Perplexity"):
        # Tokenize
        encodings = tokenizer(
            text,
            return_tensors="pt",
            max_length=2048,
            truncation=True
        ).to(device)
        
        # Forward pass
        with torch.no_grad():
            outputs = model(**encodings, labels=encodings["input_ids"])
            loss = outputs.loss
        
        # Accumulate
        total_loss += loss.item() * encodings["input_ids"].shape[1]
        total_tokens += encodings["input_ids"].shape[1]
    
    # Perplexity = exp(loss)
    perplexity = np.exp(total_loss / total_tokens)
    
    return perplexity

def evaluate_generations(
    model,
    tokenizer,
    test_prompts,
    reference_responses,
    device="cuda"
):
    """
    Evaluate model generations against references
    
    Metrics: Length, coherence, reference similarity
    """
    
    print("\nGeneration Evaluation")
    print("="*70)
    
    generated_texts = []
    generation_lengths = []
    
    for prompt in test_prompts:
        # Generate
        inputs = tokenizer(prompt, return_tensors="pt").to(device)
        with torch.no_grad():
            outputs = model.generate(
                **inputs,
                max_length=200,
                temperature=0.7,
                top_p=0.9,
            )
        
        generated_text = tokenizer.decode(outputs[0], skip_special_tokens=True)
        generated_texts.append(generated_text)
        generation_lengths.append(len(generated_text.split()))
    
    # Statistics
    avg_length = np.mean(generation_lengths)
    std_length = np.std(generation_lengths)
    
    print(f"Average generation length: {avg_length:.1f} words")
    print(f"Std deviation: {std_length:.1f}")
    print(f"Min/Max: {min(generation_lengths)}/{max(generation_lengths)} words")
    
    return generated_texts

def domain_specific_evaluation(model, tokenizer, domain="medical"):
    """
    Evaluate on domain-specific test set
    
    Custom metrics for specific domain
    """
    
    if domain == "medical":
        test_prompts = [
            "What are symptoms of hypertension?",
            "Explain insulin resistance",
            "Treatment options for diabetes type 2",
        ]
    elif domain == "code":
        test_prompts = [
            "Write Python function for binary search",
            "Implement merge sort algorithm",
            "Design REST API endpoint structure",
        ]
    
    print(f"\nDomain-Specific Evaluation: {domain}")
    print("="*70)
    
    for prompt in test_prompts:
        inputs = tokenizer(prompt, return_tensors="pt").to("cuda")
        with torch.no_grad():
            outputs = model.generate(**inputs, max_length=300)
        
        response = tokenizer.decode(outputs[0], skip_special_tokens=True)
        print(f"\nPrompt: {prompt}")
        print(f"Response: {response[:200]}...")

def full_evaluation(
    model_path,
    test_dataset_path,
    output_file="evaluation_results.json"
):
    """
    Complete evaluation pipeline
    """
    
    # Load model and tokenizer
    model = AutoModelForCausalLM.from_pretrained(
        model_path,
        device_map="auto",
        torch_dtype=torch.float16
    )
    tokenizer = AutoTokenizer.from_pretrained(model_path)
    
    # Load test data
    dataset = load_dataset("json", data_files=test_dataset_path)
    texts = [ex["text"] for ex in dataset["train"]]
    
    # Compute metrics
    perplexity = compute_perplexity_batch(model, tokenizer, texts)
    generations = evaluate_generations(
        model, tokenizer,
        texts[:10],  # First 10 as test
        texts  # References
    )
    
    # Save results
    results = {
        "perplexity": float(perplexity),
        "sample_generations": generations[:3],
    }
    
    import json
    with open(output_file, "w") as f:
        json.dump(results, f, indent=2)
    
    print(f"\nResults saved to {output_file}")
```

---

## 8. Production Deployment

### Model Export and Format Conversion

Converting fine-tuned models to deployment-friendly formats:

```python
# model_export.py
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch
import os

def export_to_gguf(model_path, output_path="model.gguf"):
    """
    Export model to GGUF format for llama.cpp inference
    
    Benefits:
    - CPU-only inference possible
    - Quantized formats supported
    - Fast local inference
    - No GPU required
    """
    
    print(f"Exporting {model_path} to GGUF...")
    
    # Note: Requires llama-cpp-python or manual conversion script
    # Here's the conceptual flow:
    
    # 1. Load model
    model = AutoModelForCausalLM.from_pretrained(model_path)
    
    # 2. Save in HuggingFace format (if not already)
    model.save_pretrained("./model_hf")
    
    # 3. Use conversion script:
    # pip install llama-cpp-python
    # python convert_hf_to_gguf.py ./model_hf --outfile model.gguf
    
    os.system("""
    python -m pip install llama-cpp-python
    python ./convert_hf_to_gguf.py ./model_hf --outtype q4_0 --outfile {output_path}
    """.format(output_path=output_path))
    
    print(f"Model exported to {output_path}")
    
    # Verify
    import subprocess
    result = subprocess.run(["ls", "-lh", output_path], capture_output=True)
    print(f"File size: {result.stdout.decode()}")

def merge_and_export(adapter_dir, base_model, output_dir):
    """
    Merge LoRA adapter into base model and export
    """
    
    from peft import AutoPeftModelForCausalLM
    
    print(f"Loading adapter from {adapter_dir}...")
    
    # Load with adapter
    model = AutoPeftModelForCausalLM.from_pretrained(
        adapter_dir,
        device_map="auto"
    )
    
    # Merge weights
    print("Merging adapter into base model...")
    merged_model = model.merge_and_unload()
    
    # Save merged
    merged_model.save_pretrained(output_dir)
    tokenizer = AutoTokenizer.from_pretrained(adapter_dir)
    tokenizer.save_pretrained(output_dir)
    
    print(f"Merged model saved to {output_dir}")
```

### Inference Servers

Deploying fine-tuned models for production:

```python
# inference_servers.py

# Option 1: Ollama (easiest for local inference)
"""
Installation:
curl https://ollama.ai/install.sh | sh

Usage:
ollama run llama2:7b  # Downloads and runs model locally
ollama serve          # Starts API server on localhost:11434
"""

# Option 2: vLLM (high-performance inference)
"""
Installation:
pip install vllm

Usage:
python -m vllm.entrypoints.api_server \
  --model meta-llama/Llama-2-7b \
  --tensor-parallel-size 1 \
  --dtype float16

Performance:
- Throughput: 10-100x faster than transformers
- Features: Token streaming, batching, caching
"""

# Option 3: TGI (Text Generation Inference)
"""
Installation:
docker pull ghcr.io/huggingface/text-generation-inference:latest

Usage:
docker run --gpus all -p 8080:80 \
  ghcr.io/huggingface/text-generation-inference:latest \
  --model-id meta-llama/Llama-2-7b

Performance:
- Enterprise-grade
- Scaling support
- Production monitoring
"""

# Option 4: llama.cpp (CPU inference)
"""
Perfect for consumer hardware, fine-tuned models

Installation:
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp && make

Usage:
./main -m model.gguf -p "Your prompt"

Quantization:
./quantize model.gguf model-q4_0.gguf q4_0

Performance:
- 10-100 tokens/sec on CPU
- 1-2 tokens/sec on older GPUs
- Minimal latency, good for responsive apps
"""
```

---

## Best Practices

### Data Quality

**Critical Importance**: Fine-tuning amplifies training data quality (good or bad)

```
Poor data → Poor model (regardless of fine-tuning method)
Good data → Good model (guaranteed by LoRA/QLoRA)
```

**Checklist**:
- Remove corrupted examples (missing responses, incomplete data)
- Detect and remove duplicates (biases model toward repeated patterns)
- Verify diversity (multiple domains, writing styles, perspectives)
- Check for private/sensitive information (PII, API keys, passwords)
- Validate formatting consistency (all examples follow same template)

### Reproducibility

```python
# reproducibility.py
import torch
import numpy as np
import random

def set_seed(seed=42):
    """Set all random seeds for reproducibility"""
    
    # Python
    random.seed(seed)
    
    # NumPy
    np.random.seed(seed)
    
    # PyTorch
    torch.manual_seed(seed)
    torch.cuda.manual_seed(seed)
    torch.cuda.manual_seed_all(seed)
    
    # Disable nondeterministic algorithms
    torch.backends.cudnn.deterministic = True
    torch.backends.cudnn.benchmark = False

# Usage
set_seed(42)
# Now all runs will be identical (same model, same data, same seed)
```

### Monitoring

```python
# monitoring.py
import wandb

# Log to Weights & Biases (free for open source)
training_args = TrainingArguments(
    output_dir="./outputs",
    report_to=["wandb"],
    run_name="llama-code-lora",
)

# Metrics tracked:
# - Training loss
# - Learning rate
# - Gradient norm
# - GPU memory
# - Training speed (examples/sec)
```

---

## TRUST 5 Compliance

### Test-First

Fine-tuning should be validated with tests:

```python
def test_fine_tuning_quality():
    """Test that fine-tuned model meets quality thresholds"""
    
    model = load_fine_tuned_model()
    tokenizer = load_tokenizer()
    
    # Test 1: Generate coherent output
    outputs = generate_completions(model, test_prompts)
    assert all(len(out) > 0 for out in outputs), "Generation failed"
    
    # Test 2: Perplexity within bounds
    perplexity = compute_perplexity(model, test_set)
    assert perplexity < 20, f"Perplexity too high: {perplexity}"
    
    # Test 3: Domain-specific accuracy
    accuracy = evaluate_domain_metrics(model, domain_test_set)
    assert accuracy > 0.75, f"Accuracy too low: {accuracy}"
    
    # Test 4: No memorization of training data
    memorization = check_memorization(model, training_data)
    assert memorization < 0.1, f"Too much memorization: {memorization}"
```

### Readable

Clear documentation and variable names:

```python
# Good: Clear, self-documenting
lora_config = LoraConfig(
    r=8,                    # Rank: 8x8 adapter matrices
    lora_alpha=16,          # Scaling: output * (16/8) = 2x
    target_modules=["q_proj", "v_proj"],  # Attention projections
    lora_dropout=0.05,      # Regularization
    task_type=TaskType.CAUSAL_LM,
)

# Bad: Cryptic
cfg = LoraConfig(r=8, a=16, t=["q", "v"], d=0.05)
```

### Unified

Consistent patterns across fine-tuning methods:

```
All methods follow:
1. Load model
2. Quantize (if QLoRA)
3. Add LoRA/PEFT
4. Prepare dataset
5. Configure training
6. Train
7. Save and evaluate
```

### Secured

Security best practices:

```python
# Don't commit these to git
.env          # API keys, model paths
.credentials  # Authentication tokens
*.pt, *.pth   # Model weights (use .gitignore)
training_data # Sensitive examples

# Do commit these
requirements.txt       # Dependencies
training_config.yaml  # Parameters
training_script.py    # Code
```

### Trackable

SPEC-based fine-tuning:

```
SPEC-ML-FINETUNE-001: Llama 3.1 8B Fine-tuning

Ubiquitous: System SHALL support LoRA and QLoRA methods
Event-Driven: WHEN domain data provided → Fine-tune model
Unwanted: IF dataset corrupted → Skip with error
State-Driven: WHILE training → Log metrics every 10 steps
Optional: WHERE high-accuracy needed → Use full fine-tuning
```

---

## Quick Reference: Parameter Selection

### LoRA Hyperparameters

```
Rank (r):
├─ 8 (default): 95% of performance, minimal overhead
├─ 16: Better for complex domains
└─ 32+: Only for highly specialized tasks

Alpha (scaling):
├─ alpha = 2*r (standard): Balanced
└─ alpha = r (conservative): If instability

Dropout:
├─ 0.05 (default): Typical for LLMs
└─ 0.1: High regularization (small datasets)

Target modules:
├─ ["q_proj", "v_proj"]: Minimal (fast)
└─ ["q_proj", "v_proj", "k_proj", "o_proj", "up_proj", "down_proj"]: Full
```

### Training Hyperparameters

```
Learning rate: 2e-4 (standard for LLMs)
Warmup: 3% of total steps (or 500 steps minimum)
Batch size: 4-8 per GPU (with grad accumulation)
Epochs: 2-3 (rarely need more)
```

---

**Last Updated**: 2025-11-19
**Version**: 4.0.0
**Status**: Stable
**Language**: English
**Framework**: PEFT 0.13+, Transformers 4.45+, bitsandbytes 0.45+
**Models**: Llama 3.1, Mistral, Mixtral, Falcon, OpenLLaMA
**Deployment**: Ollama, vLLM, llama.cpp, TGI

