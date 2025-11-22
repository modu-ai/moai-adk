users[2]{id,name}:
 1,Alice
 2,Bob
"""

# Strict mode (default): Validates array count exactly
try:
    data = decode(toon_str, {"strict": True})
    print("Valid TOON")
except Exception as e:
    print(f"Validation error: {e}")
```

### Cost Analysis Across Formats
```python
from toon_format import encode, estimate_savings
import json

datasets = {
    "search_results": [
        {"doc_id": f"doc_{i:03d}", "score": 0.95 - i*0.01, "source": f"src_{i}"}
        for i in range(100)
    ],
    "metadata": {"title": "Data", "author": "GOOS", "version": "1.0.0"},
    "nested": {"level1": {"level2": {"level3": {"level4": {"level5": "data"}}}}}
}

for name, data in datasets.items():
    savings = estimate_savings(data)
    print(f"\n{name}:")
    print(f"  Savings: {savings.get('savings_percent', 0):.1f}%")
    print(f"  Original: {savings.get('original_tokens', 0)} tokens")
    print(f"  Optimized: {savings.get('optimized_tokens', 0)} tokens")
```


## Escaping and Special Cases

### Valid Escape Sequences (Quoted Strings Only)
```
\\  → Backslash
\"  → Quote
\n  → Newline
\r  → Carriage return
\t  → Tab
```

### Automatic Quoting Rules
A value is automatically quoted if:
- Empty string: `""`
- Starts with whitespace: `"  leading"`
- Contains declared delimiter: `"a,b,c"` (if comma-delimited)
- Matches reserved keyword: `"true"`, `"false"`, `"null"`
- Looks like a number: `"123"`, `"1.5"`
- Looks like array header: `"[3]:"` (if quoted)
- Contains newline, CR, backslash, or quote

Example:
```toon
# Automatic quoting
fields[3]:
 "value with spaces"
 "123"  # looks like number, must quote
 ""     # empty string, must quote
 normal # no quoting needed
```


## Integration with Yoda Project

### Universal Usage Pattern
```python
# In any yoda module or agent

from toon_format import encode, decode

def format_llm_context(data: dict | list) -> str:
    """Convert Python data to TOON for LLM prompts"""
    return encode(data)

def parse_llm_output(toon_str: str):
    """Parse TOON from LLM back to Python"""
    return decode(toon_str)

# Example: RAG integration
def format_search_results_for_prompt(results: list[dict]) -> str:
    """Results = [{'doc_id': str, 'content': str, 'score': float}, ...]"""
    return encode(results)  # Automatic tabular format

# Example: Batch processing
def convert_data_batch(json_file: str) -> str:
    import json
    with open(json_file) as f:
        data = json.load(f)
    return encode(data)
```


## Performance Characteristics (2025)

| Metric | Value | Notes |
|--------|-------|-------|
| Token Reduction | 40-50% | Average across typical datasets |
| Array Overhead | 12-15 tokens | Per array declaration and count |
| Table Efficiency | 45% best case | Uniform objects, minimal quoting |
| Nesting Penalty | +5% per level | YAML-like indentation cost |
| Escape Cost | Variable | Only quoted strings escape |

### Benchmarks
- **100-row dataset**: 3200 JSON tokens → 1680 TOON tokens (47.5% savings)
- **Nested metadata**: 450 JSON tokens → 280 TOON tokens (37.8% savings)
- **Mixed structure**: 1200 JSON tokens → 720 TOON tokens (40% savings)


## Troubleshooting

### Common Issues

**Issue**: "Array count mismatch"
```
Solution: Ensure [N] matches actual row count in strict mode
users[2]{id,name}:  # declares 2 rows
 1,Alice
 2,Bob
 # 3,Carol  ← ERROR: declared [2] but 3 rows provided
```

**Issue**: "Unterminated string"
```
Solution: Close all quotes properly
description: "This is unclosed    ← ERROR: missing closing "
description: "This is closed"     ← OK
```

**Issue**: "Invalid escape sequence"
```
Solution: Use only valid escapes: \\ \" \n \r \t
invalid: "path\windows\file"      ← ERROR: \w \i \l not valid
valid: "path\\windows\\file"       ← OK: backslashes escaped
```

**Issue**: "Tab/space mixing"
```
Solution: Use consistent indentation (spaces only)
users:
→name: Alice        ← ERROR: tab character used
  age: 30           ← OK: spaces only
```


## TOON Spec v2.0 Compliance

This skill implements TOON v2.0 working draft (2025-11-10) with the following features:

- ✅ Core syntax (primitives, objects, arrays, tabular form)
- ✅ Strict mode validation with exact array count checking
- ✅ All delimiter types (comma, tab, pipe)
- ✅ Valid escape sequences in quoted strings
- ✅ Key folding with path expansion (optional)
- ✅ Lossless JSON conversion
- ✅ Human-readable indentation

**Spec**: [github.com/toon-format/spec](https://github.com/toon-format/spec/blob/main/SPEC.md)


## References

- **Official TOON Format**: https://toonformat.dev
- **GitHub Repository**: https://github.com/toon-format/toon
- **Specification**: https://github.com/toon-format/spec/blob/main/SPEC.md
- **Python Library**: https://github.com/toon-format/toon-python
- **MIME Type**: `application/toon+text`
- **File Extension**: `.toon`

## Skill Documentation

- [examples.md](examples.md) — Practical use cases and patterns
- [reference.md](reference.md) — Complete API reference
- [patterns.md](patterns.md) — Anti-patterns and common mistakes


## Integration Matrix

Works best with:
- `moai-lang-python` — Native toon-python library integration
- `moai-context7-integration` — Latest TOON spec and best practices
- `moai-essentials-perf` — Performance optimization
- `moai-core-code-reviewer` — Code quality for TOON handlers


**Skill Version**: 2.0.0
**TOON Spec Version**: 2.0 (working draft, 2025-11-10)
**Status**: Production Ready
**License**: MIT
**Last Updated**: 2025-11-21


**End of TOON Format Specialist Skill**
