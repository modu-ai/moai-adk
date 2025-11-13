# Agent Selection Examples

## Example 1: New Feature Request

**User**: "Add two-factor authentication"

**Decision Tree**:
```
Task = Feature Implementation
  ↓
Create SPEC first? → YES
Invoke: spec-builder
  ↓
Write Tests? → YES
Invoke: tdd-implementer
  ↓
Need Code Review? → YES
Invoke: quality-gate + Skill("moai-foundation-trust")
  ↓
Update Documentation? → YES
Invoke: doc-syncer
```

**Result**: Proper agent sequence avoids rework

---

## Example 2: Bug Fix Triage

**Error**: "TypeError: undefined is not a function"

**Agent Sequence**:
1. Invoke: debug-helper
   - Analyze stack trace
   - Identify root cause
2. Invoke: tdd-implementer
   - Write regression test
   - Fix implementation
3. Invoke: git-manager

---

## Example 3: Large Codebase Exploration

**Task**: "How is authentication implemented?"

✅ Correct:
```
Invoke: Explore(prompt="Find auth flow, entry points, models")
Result: Clear architecture map
```

❌ Wrong:
```
Manual grep for 20+ files
Result: Bloated context, still confused
```

---

Learn more in `reference.md` for complete decision tree and multi-agent patterns.
