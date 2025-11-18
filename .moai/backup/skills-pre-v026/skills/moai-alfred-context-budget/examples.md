# Context Budget Optimization Examples

## Memory File Structure

```
.moai/memory/
├── session-summary.md      # Current task, key files, assumptions
├── architecture.md         # System design reference
├── gotchas.md             # Common pitfalls
└── archive/               # Completed sessions
    ├── feature-auth-*
    └── feature-api-*
```

## JIT Retrieval Example

**Task**: Add email verification

**JIT Loading**:
- Read: User model (src/domain/user.ts)
- Read: Signup endpoint (src/api/auth.ts)
- Cache: Signup flow summary in memory

**Result**: Context efficiency, clear understanding

---

Learn more in `reference.md`.
