# Plan: SPEC-CONTEXT-CLEANUP-001

## Task Decomposition

### Phase 1: Deletion (Tasks 1-2)
Independent, parallelizable.

| Task | Files | Complexity |
|------|-------|------------|
| T1: Delete context memory files | 2 files x2 (local+template) = 4 | Low |
| T2: Clean SKILL.md routing | 1 file x2 = 2 | Low |

### Phase 2: Core Cleanup (Tasks 3-5)
Independent, parallelizable.

| Task | Files | Complexity |
|------|-------|------------|
| T3: Clean CLAUDE.md | 1 file x2 = 2 | Medium (renumbering) |
| T4: Clean sync.md | 1 file x2 = 2 | Medium (preserve tags) |
| T5: Clean manager-git.md | 1 file x2 = 2 | Low |

### Phase 3: Cross-Reference Cleanup (Task 6)
Depends on Phase 1-2 understanding.

| Task | Files | Complexity |
|------|-------|------------|
| T6: Clean cross-references | ~9 files x2 = ~18 | Medium (many files) |

### Phase 4: Enhancement (Tasks 7-9)
Independent of deletion, but logically follows.

| Task | Files | Complexity |
|------|-------|------------|
| T7: Enhance SPEC template | ~2 files x2 = 4 | Low |
| T8: Update run.md | 1 file x2 = 2 | Low |
| T9: Update moai-memory.md | 1 file x2 = 2 | Low |

### Phase 5: Documentation & Verification (Tasks 10-11)

| Task | Files | Complexity |
|------|-------|------------|
| T10: Update READMEs | 2-3 files | Low |
| T11: Build & verify | make build + go test | Medium |

## Execution Order

```
Phase 1 (T1, T2) ──parallel──> Phase 2 (T3, T4, T5) ──parallel──>
Phase 3 (T6) ──> Phase 4 (T7, T8, T9) ──parallel──>
Phase 5 (T10, T11)
```

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Missed reference to context memory | Medium | Low | Grep verification in T11 |
| Section renumbering error in CLAUDE.md | Low | Medium | Careful review |
| Template/local out of sync | Medium | Medium | Parallel edits per file |
| embedded.go regeneration failure | Low | High | make build + go test |
