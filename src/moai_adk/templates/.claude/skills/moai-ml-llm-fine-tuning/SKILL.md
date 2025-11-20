---
name: moai-ml-llm-fine-tuning
version: 4.0.0
updated: 2025-11-20
status: stable
description: Enterprise LLM Fine-Tuning with LoRA, QLoRA, and PEFT techniques
category: Machine Learning
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# LLM Fine-Tuning Expert

**Parameter-Efficient Fine-Tuning (PEFT) for Enterprise LLMs**

> **Focus**: LoRA, QLoRA, Domain Adaptation  
> **Models**: Llama 3.1, Mistral, Mixtral, Falcon  
> **Stack**: PyTorch, Transformers, PEFT, bitsandbytes

---

## Overview

Enterprise-grade fine-tuning strategies for customizing Large Language Models (LLMs) with minimal resource requirements.

### Core Capabilities

- **Parameter-Efficient Fine-Tuning (PEFT)**: LoRA, QLoRA, Prefix Tuning
- **Quantization**: 4-bit/8-bit training with bitsandbytes
- **Distributed Training**: Multi-GPU, DeepSpeed, FSDP
- **Optimization**: Flash Attention 2, Gradient Checkpointing
- **Evaluation**: Perplexity, BLEU, ROUGE, Domain benchmarks

### Technology Stack

- **PEFT 0.13+**: Adapter management
- **Transformers 4.45+**: Model architecture
- **TRL 0.11+**: Supervised Fine-Tuning (SFT), DPO
- **Accelerate 0.34+**: Training loop orchestration
- **bitsandbytes 0.45+**: Low-precision optimization

---

## Fine-Tuning Strategies

| Method               | Params Updated | VRAM (70B) | Use Case                               |
| -------------------- | -------------- | ---------- | -------------------------------------- |
| **Full Fine-Tuning** | 100%           | ~420GB     | Foundation model creation              |
| **LoRA**             | 0.1-1%         | ~180GB     | Domain adaptation, Style transfer      |
| **QLoRA**            | 0.1-1%         | ~24GB      | Consumer GPU training, Cost efficiency |

---

## Implementation Patterns

### 1. QLoRA Configuration (Recommended)

Efficient 4-bit training for large models.

```python
import torch
from transformers import AutoTokenizer, AutoModelForCausalLM, BitsAndBytesConfig
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training

def setup_qlora_model(model_id="meta-llama/Llama-3.1-8B"):
    # 1. Quantization Config
    bnb_config = BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_quant_type="nf4",
        bnb_4bit_compute_dtype=torch.bfloat16,
        bnb_4bit_use_double_quant=True,
    )

    # 2. Load Model
    model = AutoModelForCausalLM.from_pretrained(
        model_id,
        quantization_config=bnb_config,
        device_map="auto",
        trust_remote_code=True
    )

    # 3. Enable Gradient Checkpointing & K-bit Training
    model.gradient_checkpointing_enable()
    model = prepare_model_for_kbit_training(model)

    # 4. LoRA Config
    peft_config = LoraConfig(
        r=16,                    # Rank
        lora_alpha=32,           # Alpha (scaling)
        target_modules=[
            "q_proj", "k_proj", "v_proj", "o_proj",
            "gate_proj", "up_proj", "down_proj"
        ],
        lora_dropout=0.05,
        bias="none",
        task_type="CAUSAL_LM"
    )

    # 5. Apply Adapter
    model = get_peft_model(model, peft_config)
    model.print_trainable_parameters()

    return model
```

### 2. Training Loop with TRL

Supervised Fine-Tuning (SFT) using HuggingFace TRL.

```python
from trl import SFTTrainer
from transformers import TrainingArguments

def train_model(model, tokenizer, dataset, output_dir="./results"):
    training_args = TrainingArguments(
        output_dir=output_dir,
        num_train_epochs=3,
        per_device_train_batch_size=4,
        gradient_accumulation_steps=4,
        learning_rate=2e-4,
        lr_scheduler_type="cosine",
        warmup_ratio=0.03,
        fp16=True,             # or bf16=True for Ampere+
        logging_steps=10,
        save_strategy="epoch",
        optim="paged_adamw_8bit",
        report_to="tensorboard"
    )

    trainer = SFTTrainer(
        model=model,
        train_dataset=dataset,
        dataset_text_field="text", # Column containing formatted prompt
        max_seq_length=2048,
        tokenizer=tokenizer,
        args=training_args,
        packing=False,
    )

    trainer.train()
    trainer.save_model()
```

### 3. Data Preparation

Formatting data for instruction tuning.

```python
from datasets import load_dataset

def format_instruction(sample):
    return f"""### Instruction:
{sample['instruction']}

### Input:
{sample['input']}

### Response:
{sample['output']}
"""

def prepare_dataset(path="databricks/databricks-dolly-15k"):
    dataset = load_dataset(path, split="train")

    # Format for SFTTrainer
    dataset = dataset.map(lambda x: {"text": format_instruction(x)})

    return dataset
```

---

## Advanced Techniques

### Multi-GPU Distributed Training

Using Accelerate and DeepSpeed for scaling.

```bash
# config.yaml for accelerate
compute_environment: LOCAL_MACHINE
distributed_type: MULTI_GPU
deepspeed_config:
  gradient_accumulation_steps: 4
  gradient_clipping: 1.0
  offload_optimizer_device: cpu
  offload_param_device: cpu
  zero_optimization:
    stage: 2
```

Run command:

```bash
accelerate launch --config_file config.yaml train.py
```

### Model Merging

Merging LoRA adapters back into the base model for deployment.

```python
from peft import PeftModel

def merge_adapter(base_model_id, adapter_path, output_path):
    # Load base model in FP16 (not 4-bit)
    base_model = AutoModelForCausalLM.from_pretrained(
        base_model_id,
        torch_dtype=torch.float16,
        device_map="auto"
    )

    # Load adapter
    model = PeftModel.from_pretrained(base_model, adapter_path)

    # Merge
    model = model.merge_and_unload()

    # Save
    model.save_pretrained(output_path)
    base_model.tokenizer.save_pretrained(output_path)
```

---

## Validation Checklist

**Setup**:

- [ ] GPU environment verified (CUDA available)
- [ ] Dependencies installed (peft, trl, bitsandbytes)
- [ ] HuggingFace token configured

**Data**:

- [ ] Dataset formatted correctly (Instruction/Input/Output)
- [ ] Tokenization length checked (< context window)
- [ ] Train/Val split created

**Training**:

- [ ] QLoRA config applied (4-bit, nf4)
- [ ] Gradient checkpointing enabled
- [ ] Learning rate scheduled (warmup + decay)
- [ ] Loss monitoring active

**Evaluation**:

- [ ] Perplexity calculated
- [ ] Generation quality manually verified
- [ ] Adapter merged successfully

---

## Related Skills

- `moai-domain-ml`: General ML workflows
- `moai-domain-data-science`: Data preparation
- `moai-essentials-perf`: Inference optimization

---

**Last Updated**: 2025-11-20
