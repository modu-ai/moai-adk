# decision-index.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

Questions the interview did not settle, one row each. Labels use the fixed vocabulary `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`. All five rows surfaced during plan assembly for card t1568; none is settled by a committed operator setting, a constitution clause, or a completed SPEC's HISTORY/Amendments row, so all route `FOUNDER`. All five are implementation-level (none changes a shipped command's default user-visible behavior beyond the card's own defect fix, removes a user-facing feature, or changes a template default), each carries a `Default:` selected by the published default rule, and each is stamped `DEFAULT-APPLIED` at plan close — none blocks the Kickoff.

### Q1: Which cross-process mechanism — advisory flock, O_EXCL claim-stamp, or a dependency? (plan.md OD-1)

Label: FOUNDER
Class: implementation-level
Authority anchor: none — no committed setting or completed SPEC decides the mechanism for fold writes.
Why unresolved: three workable mechanisms exist in or near this codebase and the trade-offs (kernel-released lifetime vs stale-claim sweep vs dependency cost) are a judgment call.
Default: per-store advisory flock via the repository's own sessionmsg pattern (rule: step 3 of the published default rule — the smaller user-visible surface: no stale-claim cleanup path, no go.mod change; design.md carries the full rationale and the rejection table).
Alternate: O_EXCL claim-stamp with a staleness sweep; third-party flock package.
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec

### Q2: Lock span — the whole applyFold, or one acquisition per rename? (OD-2)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: both spans close the rename window; they differ in how much fold-vs-fold interleaving remains and in hold time.
Default: the whole `applyFold` span, one acquisition, deferred release (rule: step 3 — strictly stronger serialization with fewer acquisitions and a simpler contract; the hold-time cost is bounded and the close-path bound still expires the step).
Alternate: per-rename acquisition pairs (weaker: plan/re-check phases of two folds still interleave).
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec

### Q3: Acquisition semantics — blocking LOCK_EX, or non-blocking with a bounded retry? (OD-3)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: a blocking wait is friendlier under contention but introduces an unbounded user-facing wait and a goroutine the close-path bound cannot interrupt; the bounded refusal keeps every wait bounded.
Default: non-blocking acquire with a bounded retry (2s-class deadline, plan.md D-2), then the clean refusal of REQ-MRR-003 (rule: step 1 — preserves the current behavior class where no fold wait is unbounded, the MFB OD-11 boundedness precedent).
Alternate: blocking LOCK_EX with no retry loop.
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec

### Q4: Lock-file placement — inside the store directory, or outside keyed by the store path? (OD-4)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: a store-local dotfile is visible and simple but must stay invisible to the store's own tooling (REQ-MRR-005); a tmp-dir file keyed by the store's absolute path avoids touching the store but adds a hash-collision and stale-file surface.
Default: store-local `<store-dir>/.moai-fold.lock`, mode 0644, never removed on release (rule: step 3 — the smaller surface: no hashing scheme, no tmp cleanup; REQ-MRR-005's criterion polices the invisibility cost).
Alternate: `os.TempDir()` file keyed by the store path hash.
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec

### Q5: Should the other memory-writing verbs (drain, diet, archive) adopt the lock in this SPEC? (OD-5)

Label: FOUNDER
Class: implementation-level
Authority anchor: none.
Why unresolved: adopting now would broaden the protection to writers outside the observed defect class at the cost of touching three verbs this card does not name; deferring keeps this card's scope at the defect.
Default: defer — only fold applies acquire the lock in this SPEC; the convention is published for a follow-up (rule: step 1 — preserves the current behavior of the three verbs).
Alternate: adopt the lock in all memory-writing verbs now.
Operator verdict: DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec
