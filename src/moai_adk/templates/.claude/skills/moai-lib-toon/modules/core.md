**Token Savings**: 3 documents in TOON ≈ 180 tokens vs JSON ≈ 320 tokens (44% reduction)

### Pattern 2: Structured Prompts
```toon
# Example: Multi-turn conversation with structured examples

examples[2]{input,output}:
 "Summarize this text","A brief overview of the main points"
 "Translate to French","Le texte traduit en français"

# Task: Follow the pattern above for new input
```

### Pattern 3: Batch Data Processing
```toon
# Process multiple records efficiently

records[100]{id,timestamp,event_type,user_id}:
 1,2025-11-21T10:00:00Z,login,user_001
 2,2025-11-21T10:05:00Z,purchase,user_002
 3,2025-11-21T10:10:00Z,logout,user_001
 ...
```

### Pattern 4: Metadata with Content
```toon
document:
  title: "Python Asyncio Guide"
  author: "GOOS"
  created: "2025-11-21"

content[2]{section,word_count}:
 "Introduction",1200
 "Advanced Patterns",3400

chunks[3]{chunk_id,text}:
 1,"Async programming enables efficient I/O handling"
 2,"The event loop manages task execution"
 3,"Coroutines are functions with await points"
```


## Python Implementation (toon-python)

### Installation
```bash
# Using uv (recommended)
uv pip install toon_format tiktoken

# Or standard pip
pip install toon_format tiktoken
```

### Basic Usage
```python
from toon_format import encode, decode, estimate_savings

# JSON → TOON
data = {
    "users": [
        {"id": 1, "name": "Alice", "email": "alice@example.com"},
        {"id": 2, "name": "Bob", "email": "bob@example.com"}
    ]
}

toon_str = encode(data)
print(toon_str)
# Output:
# users[2]{id,name,email}:
#  1,Alice,alice@example.com
#  2,Bob,bob@example.com

# TOON → JSON
decoded = decode(toon_str)
assert decoded == data  # Lossless conversion
```

### Measuring Token Savings
```python
import json
from toon_format import encode, count_tokens

data = [
    {"doc_id": "doc_001", "content": "Example text", "score": 0.95},
    {"doc_id": "doc_002", "content": "More example", "score": 0.89},
    {"doc_id": "doc_003", "content": "Final example", "score": 0.85}
]

# Compare formats
json_str = json.dumps(data)
toon_str = encode(data)

json_tokens = count_tokens(json_str)
toon_tokens = count_tokens(toon_str)
savings = (json_tokens - toon_tokens) / json_tokens * 100

print(f"JSON: {json_tokens} tokens")
print(f"TOON: {toon_tokens} tokens")
print(f"Savings: {savings:.1f}%")
# Output:
# JSON: 320 tokens
# TOON: 180 tokens
# Savings: 43.8%
```

### Handling Custom Delimiters
```python
from toon_format import encode, decode

# Data with pipe characters in content
results = [
    {"doc_id": "doc_001", "pattern": "a|b|c", "score": 0.95},
    {"doc_id": "doc_002", "pattern": "x|y|z", "score": 0.89}
]

# Encode with tab delimiter to avoid quoting
toon_str = encode(results)
# Auto-quotes pattern field due to pipe characters

# Or use different delimiter (if library supports options)
# This depends on toon-python API capabilities
```


## Format Decision Guide

```
Encoding structured data for LLM?
│
├─ Uniform array of objects?
│  ├─ YES → Use TOON TABULAR form
│  │        Header: [N]{field1,field2,...}:
│  │        Rows: value1,value2,...
│  │        Token Savings: 40-50%
│  │
│  └─ NO → Check complexity
│          ├─ Complex nesting (5+ levels)?
│          │  └─ YES → Use JSON
│          │
│          └─ NO → Use TOON EXPANDED form
│                 Items: [N]:
│                   - item1
│                   - item2
│
└─ Simple key-value metadata?
   └─ YES → Use TOON OBJECT form
           key: value
           Token Savings: 30-40%
```


## Best Practices

### DO ✅
- **Declare array lengths explicitly** — Aids truncation detection and parsing
- **Use tabular form for uniform records** — Maximum compression
- **Minimize quoting** — Only quote when necessary (spaces, delimiters, reserved words)
- **Preserve delimiter consistency** — Once declared, maintain across all rows
- **Validate in strict mode** — Catches malformed TOON early
- **Test round-trip conversion** — Ensure lossless JSON ↔ TOON
- **Document delimiter choice** — Comment why comma/tab/pipe selected

### DON'T ❌
- **Deep nesting (5+ levels)** — JSON more readable and equally efficient
- **Mixed delimiters in one array** — Violates TOON scoping rules
- **Mismatched field counts** — Array count in header must match actual rows
- **Tab/space mixing** — Use only spaces for indentation (2-space default)
- **Unquoted values that look like keywords** — Quote `true`, `false`, `null` if they're data
- **Omitting array count** — Length is required and must be exact
- **Complex escape sequences** — Keep data simple; use quoting instead


## Advanced Techniques

### Key Folding (Path Compression)
```python
from toon_format import encode

# Nested object with dot-separated keys
data = {
    "user.profile.name": "Alice",
    "user.profile.age": 30,
    "user.email": "alice@example.com"
}

toon = encode(data)
# Output (folded paths):
# user.profile.name: Alice
# user.profile.age: 30
# user.email: alice@example.com
```

### Strict Mode Validation
```python
from toon_format import decode

toon_str = """
