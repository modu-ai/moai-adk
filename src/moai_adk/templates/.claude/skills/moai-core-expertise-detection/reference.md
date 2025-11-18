# Alfred Expertise Detection - Quick Reference

> **Main Skill**: [SKILL.md](SKILL.md)  
> **Examples**: [examples.md](examples.md)

---

## Signal Detection Cheat Sheet

| Signal Type | Beginner | Intermediate | Expert | Weight |
|-------------|----------|--------------|--------|--------|
| **Command Syntax** | Trial-and-error | Mostly correct | Always correct | +2/+1/0 |
| **Questions** | "How?", "Why?" | "Can I?" | Direct commands | +2/+1/0 |
| **Error Recovery** | Needs help | Self-corrects | Debugs alone | +2/+1/0 |
| **Documentation** | Frequent refs | Occasional refs | Rare refs | +2/+1/0 |
| **Git Workflow** | `/alfred:*` only | Mixed usage | Direct git | +2/+1/0 |

**Score calculation**: Sum all weights → 0-3=Expert, 4-7=Intermediate, 8-10=Beginner

---

## Behavioral Adaptation Matrix

### Response Verbosity

| Level | Length | Examples | Context |
|-------|--------|----------|---------|
| Beginner | 200-400 words | 2-3 | Deep background |
| Intermediate | 100-200 words | 1-2 | Key points |
| Expert | 50-100 words | 0-1 | Action-focused |

### Confirmation Gates

| Risk Level | Beginner | Intermediate | Expert |
|------------|----------|--------------|--------|
| **Low** | Confirm | Skip | Skip |
| **Medium** | Confirm | Confirm | Skip |
| **High** | Confirm + explain | Confirm | Confirm |

### Proactive Suggestions

| Level | Suggestions/Session | Threshold |
|-------|---------------------|-----------|
| Beginner | 3-5 | Low (common patterns) |
| Intermediate | 2-3 | Medium (optimizations) |
| Expert | 1-2 | High (advanced techniques) |

---

## Quick Decision Tree

```
User Request
    ↓
Analyze 5 signals (commands, questions, errors, docs, git)
    ↓
Calculate weighted score (0-10)
    ↓
├─ 0-3: Expert → Efficiency Coach bias
├─ 4-7: Intermediate → Project Manager bias
└─ 8-10: Beginner → Technical Mentor bias
    ↓
Apply adaptations (verbosity, confirmations, suggestions)
```

---

## Override Mechanisms

**User keywords override detection**:
- "quick", "fast" → Force Expert mode
- "explain", "how", "why" → Force Beginner mode
- `/alfred:*` commands → Force Project Manager (Intermediate)

**Example**:
- Detected: Expert (score 2)
- Override: Beginner mode (verbose explanation)

---

## Signal Examples

### Beginner Signal Pattern

```
User: "How do I create a SPEC?"
Alfred detects:
  - Question keyword: "How" (+2)
  - Documentation unfamiliarity (+2)
  - No direct command usage (+2)
  
Score: 6 → Beginner

Response: Verbose explanation with Skill("moai-foundation-specs")
```

### Expert Signal Pattern

```
User: "/alfred:2-run SPEC-AUTH-001"
Alfred detects:
  - Direct command with correct syntax (0)
  - No questions (0)
  - Knows SPEC ID format (0)
  
Score: 0 → Expert

Response: Execute immediately, minimal output
```

---

## Integration Hooks

**Called by**:
- Every Alfred request (implicit)
- `moai-core-persona-roles` (role selection)
- `moai-core-proactive-suggestions` (suggestion threshold)

**Calls**:
- None (standalone signal analysis)

---

**End of Reference** | 2025-11-02
