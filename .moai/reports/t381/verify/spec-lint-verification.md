# Plan-phase lint verification — SPEC-IGNORED-EVIDENCE-CITATION-001

Tree `3f03d9c36`, worktree `.claude/worktrees/t381`, measured 2026-08-31.

## 1. Why the first attempt was NOT evidence

```
$ moai spec lint 2>&1 | grep -i 'IGNORED-EVIDENCE' ; echo "rc=$?"
rc=1
```

`rc=1` is a zero-hit grep, and the exit code `0` reported by the runner came from `echo`, not from
the linter. The whole-catalog output went into the pipe and was never seen, so nothing established
that the linter visited this SPEC at all. This is a **vacuous green**, recorded rather than
discarded.

## 2. Targeted lint — attributable, because the file is named on argv

```
$ moai spec lint .moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md
✓ No findings — all SPEC documents are valid
rc=0
```

## 3. Mutation — establishes the rule is LIVE

A clean result only counts once the check is shown to be capable of failing. A scratch copy of the
SPEC directory had its five `### Out of Scope — ` headings renamed:

```
$ sed -i '' 's/^### Out of Scope — /### MUTANT-HEADING /' <scratch>/spec.md
$ grep -c '^### Out of Scope —' <scratch>/spec.md
0

$ moai spec lint <scratch>/spec.md
SEVERITY  CODE               LINE  MESSAGE
WARNING   MissingExclusions  1     'Out of Scope' section has no items — minimum one item required
                                   [grandfathered era — downgraded to warning]

0 error(s), 1 warning(s)
```

The rule fired. The clean result in §2 is therefore non-vacuous.

## 4. Finding surfaced by the mutation — era misclassification of every plan-phase SPEC

The mutant's message carries `[grandfathered era — downgraded to warning]`. Confirmed against the
audit engine:

```
$ mcp__moai__spec_audit --filter-spec SPEC-IGNORED-EVIDENCE-CITATION-001
{"era":"V3R5","finding_type":"EraAutoDetected","severity":"INFO",
 "details":{"heuristic_matched":"H-3 (§E.2 present, sync_commit_sha missing)"}}
```

`internal/spec/era.go` evaluates its heuristics in order and returns on the first match:

| Heuristic | Predicate | Result |
|---|---|---|
| H-2 | progress.md without §E.* markers | V3R2-R4 |
| **H-3** | **§E.2 present AND sync_commit_sha missing** | **V3R5 (grandfathered)** |
| H-4 | §E.2 + §E.4 + sync_commit_sha | V3R6 |
| H-5 | modern phase or `created >= 2026-04-01` | V3R6 |

A plan-phase SPEC has §E.2 **by construction** (the progress skeleton is mandated at plan-phase) and
has no `sync_commit_sha` (sync has not run). So H-3 matches and returns **before** H-5's
modern-date tie-breaker is ever reached — even though this SPEC was created today.

**Consequence**: lint errors are demoted to warnings for the SPEC's entire plan and run phase —
precisely when the gate is meant to be strongest. The skeleton is mandated in order to prevent an
H-2 misclassification; it prevents H-2 and lands in H-3 instead.

**Scale (measured, whole catalog, this tree)**:

```
$ mcp__moai__spec_audit  →  total=714 grandfathered=286 modern_clean=421
198  H-4 (§E.2 + §E.4 + sync_commit_sha)
144  H-1 (progress.md absent)
118  H-2 (progress.md without §E.* markers)
 24  H-3 (§E.2 present, sync_commit_sha missing)   ← the trap
  5  H-5 (modern phase or created date)
  1  H-4-legacy
```

24 SPECs currently sit in H-3. This is systemic, not a property of this SPEC.

**Not acted on here.** It is a lint/era-engine concern, outside t381's scope (which is 5 citation
lines), and fixing it would mean editing `internal/spec/era.go` — unrelated to this card and owned by
nobody in this lane. Reported to the lead as a candidate card.

## 5. Gaps

- The whole-catalog `moai spec lint` run was started and its full output captured to
  `spec-lint-full.txt`, but the verdict cited above rests on the **targeted** run (§2), not on it.
- No claim is made about the other four artifacts (`plan.md`, `acceptance.md`, `progress.md`); the
  linter's schema rules bind `spec.md`, and only `spec.md` was linted.
