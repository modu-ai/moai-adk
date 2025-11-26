# TOON Format Examples: LLM Communication Patterns

This document demonstrates practical patterns for using TOON format to optimize token usage in LLM interactions.

---

## Pattern 1: Search Results as Context

### Use Case: Injecting RAG/search results into an LLM prompt

**Original JSON Format** (320 tokens):
```json
{
  "query": "Python async programming",
  "results": [
    {
      "doc_id": "doc_001",
      "title": "Async/Await Fundamentals",
      "content": "Async/await enables concurrent I/O operations without blocking",
      "score": 0.95,
      "source": "research/python-async.md"
    },
    {
      "doc_id": "doc_002",
      "title": "Event Loop Architecture",
      "content": "The event loop manages task scheduling and execution",
      "score": 0.89,
      "source": "research/event-loop.md"
    },
    {
      "doc_id": "doc_003",
      "title": "Coroutine Context",
      "content": "Coroutines are functions that can suspend and resume execution",
      "score": 0.85,
      "source": "context7/python-docs.md"
    }
  ]
}
```

**TOON Format** (180 tokens, 44% reduction):
```toon
query: Python async programming
results[3]{doc_id,title,content,score,source}:
 doc_001,"Async/Await Fundamentals","Async/await enables concurrent I/O operations without blocking",0.95,research/python-async.md
 doc_002,"Event Loop Architecture","The event loop manages task scheduling and execution",0.89,research/event-loop.md
 doc_003,"Coroutine Context","Coroutines are functions that can suspend and resume execution",0.85,context7/python-docs.md
```

**In LLM Prompt**:
```python
from toon_format import encode

# Generate search results
search_results = rag.search("Python async programming", top_k=3)

# Convert to TOON for prompt injection
toon_context = encode(search_results)

prompt = f"""
You are a Python expert. Analyze the following reference materials and answer the user's question.

Reference Materials:
{toon_context}

User Question: Explain the relationship between async/await and the event loop

Provide a comprehensive answer referencing the materials above.
"""

response = llm.generate(prompt)
```

**Token Savings**: ~140 tokens per RAG query (average 3-5 documents)

---

## Pattern 2: Structured Prompt Examples

### Use Case: Few-shot learning with formatted examples

**Use case**: Training an LLM to follow a specific output format

**JSON Examples** (450 tokens):
```json
{
  "task": "Extract key information from text",
  "examples": [
    {
      "input": "The quick brown fox jumps over the lazy dog",
      "output": {
        "entities": ["fox", "dog"],
        "actions": ["jumps"],
        "descriptors": ["quick", "brown", "lazy"]
      }
    },
    {
      "input": "Alice runs to the store",
      "output": {
        "entities": ["Alice", "store"],
        "actions": ["runs"],
        "descriptors": []
      }
    }
  ]
}
```

**TOON Examples** (280 tokens, 38% reduction):
```toon
task: Extract key information from text
examples[2]{input,output}:
 "The quick brown fox jumps over the lazy dog","{""entities"":[""fox"",""dog""],""actions"":[""jumps""],""descriptors"":[""quick"",""brown"",""lazy""]}"
 "Alice runs to the store","{""entities"":[""Alice"",""store""],""actions"":[""runs""],""descriptors"":[]}"
```

**Better Alternative - Semantic TOON**:
```toon
task: Extract key information from text

example1:
  input: "The quick brown fox jumps over the lazy dog"
  output:
    entities[2]: fox,dog
    actions[1]: jumps
    descriptors[3]: quick,brown,lazy

example2:
  input: "Alice runs to the store"
  output:
    entities[2]: Alice,store
    actions[1]: runs
    descriptors[0]:
```

**In LLM Prompt**:
```python
from toon_format import encode

examples = [
    {
        "input": "The quick brown fox jumps over the lazy dog",
        "entities": ["fox", "dog"],
        "actions": ["jumps"],
        "descriptors": ["quick", "brown", "lazy"]
    },
    {
        "input": "Alice runs to the store",
        "entities": ["Alice", "store"],
        "actions": ["runs"],
        "descriptors": []
    }
]

toon_examples = encode(examples)

prompt = f"""
Extract entities, actions, and descriptors from text. Follow these examples:

{toon_examples}

Now extract from: "The red car accelerates quickly"
"""
```

---

## Pattern 3: Batch Data Processing

### Use Case: Processing large datasets with LLM

**Dataset** (100 records):
```toon
users[100]{id,name,email,country,signup_date}:
 1,Alice Smith,alice@example.com,US,2025-01-15
 2,Bob Johnson,bob@example.com,UK,2025-02-20
 3,Carol Lee,carol@example.com,JP,2025-03-10
 ...
 100,Zoe Wang,zoe@example.com,CN,2025-11-15
```

**Processing Pattern**:
```python
from toon_format import encode, decode
import json

def process_batch_with_llm(users_file: str, batch_size: int = 10):
    """Process user data in batches through LLM"""

    with open(users_file) as f:
        all_users = json.load(f)

    results = []

    for i in range(0, len(all_users), batch_size):
        batch = all_users[i:i+batch_size]

        # Convert to TOON for efficient prompting
        toon_batch = encode(batch)

        prompt = f"""
        Analyze these user profiles and categorize by engagement level (high/medium/low):

        {toon_batch}

        Output: JSON array with added 'engagement_level' field
        """

        response = llm.generate(prompt)
        batch_results = decode(response)  # Parse LLM output
        results.extend(batch_results)

    return results
```

**Token Savings**: 40-45% per batch (vs JSON) = ~200 tokens saved per 10 users

---

## Pattern 4: API Response Compression

### Use Case: Optimizing costs for API-dependent LLM systems

**API Response** (JSON, 500 tokens):
```json
{
  "status": "success",
  "data": {
    "page": 1,
    "per_page": 50,
    "total": 5000,
    "items": [
      {
        "id": "item_001",
        "name": "Product A",
        "sku": "SKU-001",
        "price": 29.99,
        "quantity": 150,
        "status": "active"
      },
      // ... 49 more items
    ]
  }
}
```

**TOON Response** (280 tokens, 44% reduction):
```toon
status: success
data:
  page: 1
  per_page: 50
  total: 5000
  items[50]{id,name,sku,price,quantity,status}:
   item_001,Product A,SKU-001,29.99,150,active
   item_002,Product B,SKU-002,39.99,120,active
   ...
   item_050,Product AJ,SKU-050,49.99,200,active
```

**Implementation**:
```python
from toon_format import encode
import requests

def fetch_and_optimize_api(api_url: str):
    """Fetch API response and convert to TOON for LLM"""

    response = requests.get(api_url).json()

    # Convert to TOON
    toon_response = encode(response)

    # Use in LLM prompt
    prompt = f"""
    Analyze this product inventory data and identify stock levels requiring attention:

    {toon_response}

    Highlight items with quantity < 100
    """

    analysis = llm.generate(prompt)
    return analysis
```

**Cost Benefit**: For 1000 API calls/month:
- JSON: ~500K tokens
- TOON: ~280K tokens
- **Savings**: 220K tokens = ~$0.66/month (at $3/M tokens)

---

## Pattern 5: Metadata with Mixed Content

### Use Case: Combining structured data with content for document analysis

```toon
document:
  title: "Python Concurrency Guide"
  author: "GOOS"
  created: "2025-11-21"
  status: draft
  version: "1.0.0"

sections[3]{section_id,section_title,word_count,key_concepts}:
 1,"Introduction",1200,"concurrency,performance,overview"
 2,"Threads vs Processes",2400,"threading,multiprocessing,GIL"
 3,"Async/Await",3600,"asyncio,coroutines,event_loop"

metadata:
  total_words: 7200
  reading_time_minutes: 25
  difficulty: intermediate
```

**In LLM Prompt**:
```python
from toon_format import encode

document_structure = {
    "document": {
        "title": "Python Concurrency Guide",
        "author": "GOOS",
        "status": "draft"
    },
    "sections": [
        {"section_id": 1, "title": "Introduction", "word_count": 1200},
        {"section_id": 2, "title": "Threads vs Processes", "word_count": 2400},
        {"section_id": 3, "title": "Async/Await", "word_count": 3600}
    ]
}

toon_structure = encode(document_structure)

prompt = f"""
Review this document structure and provide feedback on organization:

{toon_structure}

Is the progression logical? Should any sections be reordered?
"""
```

---

## Pattern 6: Query-Result Pairs for Context

### Use Case: Chain-of-thought prompting with historical queries

```toon
queries[5]{query_id,question,answer,confidence}:
 1,"What is async programming?","Concurrent I/O without blocking",0.95
 2,"When to use threading?","I/O-bound tasks with low concurrency",0.89
 3,"Explain the GIL","Global lock preventing true parallelism",0.92
 4,"Asyncio vs threads?","Async lighter, threads more mature",0.88
 5,"Best for CPU tasks?","Multiprocessing or ProcessPoolExecutor",0.91
```

**Usage**:
```python
from toon_format import encode

qa_history = [
    {"query": "What is async?", "answer": "...", "confidence": 0.95},
    {"query": "When threading?", "answer": "...", "confidence": 0.89},
    # ... more Q&A
]

toon_qa = encode(qa_history)

prompt = f"""
Based on these Q&A examples showing high confidence:

{toon_qa}

Apply the same reasoning style to answer: "How does asyncio handle errors?"
"""
```

---

## Pattern 7: Time Series Data

### Use Case: Analyzing metrics or events efficiently

```toon
metrics:
  service: api-server
  period: 24h
  interval: 1h
  start: 2025-11-20T00:00:00Z

hourly[24]{hour,cpu_percent,memory_mb,requests,errors,latency_ms}:
 0,12.5,450,1450,2,45
 1,11.8,431,1120,1,42
 2,10.2,415,890,0,38
 3,9.8,402,780,0,35
 4,10.5,418,820,1,40
 5,14.2,520,1500,3,85
 6,45.8,890,6200,15,320
 7,52.3,950,8100,22,450
 8,48.6,920,7800,18,425
 # ... more hours
```

**Processing**:
```python
from toon_format import encode

metrics = {
    "service": "api-server",
    "period": "24h",
    "data": [
        {"hour": 0, "cpu": 12.5, "memory": 450, "requests": 1450, "errors": 2},
        {"hour": 1, "cpu": 11.8, "memory": 431, "requests": 1120, "errors": 1},
        # ... 22 more hours
    ]
}

toon_metrics = encode(metrics)

prompt = f"""
Analyze these 24-hour server metrics:

{toon_metrics}

Identify anomalies and suggest optimization opportunities.
"""

analysis = llm.generate(prompt)
```

---

## Pattern 8: Structured Output Enforcement

### Use Case: Getting LLM to produce TOON-formatted responses

```python
def get_structured_response(prompt: str, output_schema: dict):
    """
    Request LLM output in TOON format matching a schema
    """

    schema_toon = encode(output_schema)

    system_prompt = f"""
    You are a data extraction assistant.

    Output format (TOON):
    {schema_toon}

    Rules:
    - Use TOON format exactly as specified
    - Maintain field order and types
    - Quote values with special characters
    - Ensure array counts match actual items
    """

    response = llm.generate(
        system_prompt=system_prompt,
        user_prompt=prompt
    )

    # Parse TOON response back to Python
    from toon_format import decode
    return decode(response)

# Usage
result = get_structured_response(
    prompt="Extract entities from: 'Alice and Bob work at Google'",
    output_schema={
        "extraction": {
            "entities": [],
            "relationships": []
        }
    }
)
```

---

## Pattern 9: Error Handling with Format

### Use Case: Transmitting error information efficiently

**Error Report** (JSON, 180 tokens):
```json
{
  "status": "error",
  "error": {
    "code": "PARSE_ERROR",
    "message": "Invalid TOON syntax",
    "details": {
      "line": 15,
      "column": 23,
      "context": "users[3]{name,age}:",
      "expected": "field separator or array start",
      "got": "unexpected character"
    },
    "suggestions": [
      "Check field count matches [N]",
      "Verify no mixed delimiters",
      "Ensure proper indentation"
    ]
  }
}
```

**TOON** (120 tokens, 33% reduction):
```toon
status: error
error:
  code: PARSE_ERROR
  message: "Invalid TOON syntax"
  details:
    line: 15
    column: 23
    context: "users[3]{name,age}:"
    expected: "field separator or array start"
    got: "unexpected character"
  suggestions[3]:
   "Check field count matches [N]"
   "Verify no mixed delimiters"
   "Ensure proper indentation"
```

---

## Pattern 10: Model Comparison Context

### Use Case: Comparing LLM outputs side-by-side

```toon
comparison:
  task: "Summarize Python async/await"
  models[3]{model_id,response_length,quality_score,speed_rank}:
   gpt4,450,"Comprehensive, clear examples",1
   claude,380,"Concise, well-structured",2
   llama,520,"Detailed, some repetition",3
  best_for[3]:
   "Production systems",gpt4
   "Quick overviews",claude
   "Deep learning contexts",llama
```

---

## Pattern 11: Token Budget Planning

### Use Case: Planning token allocation in multi-step LLM workflows

```toon
task: "Multi-turn document analysis"
token_budget: 10000

steps[4]{step_id,description,format,estimated_tokens,actual_tokens}:
 1,"Initial document",json,2000,2050
 2,"Extract entities",toon,1000,650
 3,"Generate summary",json,1500,1480
 4,"Analysis feedback",toon,800,520

summary:
  total_planned: 5300
  total_actual: 4700
  saved: 600
  efficiency: 88.7%
```

---

## Common Pitfalls and Solutions

### Pitfall 1: Forgetting Array Count

**Wrong**:
```toon
users{name,age}:
 Alice,30
 Bob,25
```

**Correct**:
```toon
users[2]{name,age}:
 Alice,30
 Bob,25
```

### Pitfall 2: Inconsistent Field Count

**Wrong**:
```toon
items[2]{name,price}:
 Widget,9.99,InStock
 Gadget,19.99
```

**Correct**:
```toon
items[2]{name,price,status}:
 Widget,9.99,InStock
 Gadget,19.99,OutOfStock
```

### Pitfall 3: Mixing Delimiters

**Wrong**:
```toon
data[2]{a,b}:
 1,2
 3,4|5  # Mixed delimiters
```

**Correct**:
```toon
data[2]{a,b}:
 1,2
 3,4
```

### Pitfall 4: Unquoted Special Characters

**Wrong**:
```toon
title: This has, and: special characters
```

**Correct**:
```toon
title: "This has, and: special characters"
```

---

## Performance Benchmarks (2025)

### Token Savings by Data Type

| Data Type | JSON Tokens | TOON Tokens | Savings | Use Case |
|-----------|------------|-----------|---------|----------|
| Search Results | 320 | 180 | 43.8% | RAG context |
| Examples | 450 | 280 | 37.8% | Few-shot learning |
| Batch Data | 2400 | 1350 | 43.8% | Bulk processing |
| API Response | 500 | 280 | 44.0% | API optimization |
| Metrics | 800 | 420 | 47.5% | Time series |
| Error Report | 180 | 120 | 33.3% | Error handling |

### Aggregate Savings for Multi-Turn Conversations

```
5-turn conversation with structured data:
- JSON total: 5000 tokens
- TOON total: 2800 tokens
- Savings: 2200 tokens (44%)

Cost benefit (Sonnet 4.5 @ $3/1M tokens):
- JSON: $0.015
- TOON: $0.0084
- Saved: $0.0066 per conversation
- Annual (1M conversations): $6,600
```

---

## Implementation Checklist

- [ ] Identify data type (tabular vs nested)
- [ ] Choose encoding format (tabular for arrays, object for metadata)
- [ ] Declare array count explicitly `[N]`
- [ ] Maintain consistent field order
- [ ] Quote values with special characters
- [ ] Test round-trip conversion (encode/decode)
- [ ] Measure token savings with benchmarks
- [ ] Document delimiter choice if non-standard
- [ ] Validate strict mode before production
- [ ] Monitor LLM parsing accuracy

---

**End of TOON Examples**
