# design.md — SPEC-WORKFLOW-CACHE-OPT-001

> Tier L design artifact (added per plan-audit iter-1 D9). Scope: the M1 shared diagnostic snapshot contract — the only axis with Go code. Axes 2-5 are doctrine edits whose design is fully carried by plan.md §F + the AC matrix. All open design decisions are settled (plan-audit iter-1 D1, user decisions); this document records the resulting architecture.

## §A — Snapshot Contract Architecture (M1)

### A.1 Component layout

```
internal/verify/                 (new package — name mirrors .moai/state/verify/)
├── schema.go        Snapshot + CheckEntry + Conditions types (loop-verdict-compatible)
├── key.go           Key() — HEAD SHA + porcelain-v2 digest
├── freshness.go     Fresh() — key-equality AND TTL predicate
└── store.go         atomic Load/Save (temp+rename, internal/goal/state.go pattern)

internal/cli/…                   `moai verify record|check` verb group, root-tree registered
internal/goal/evaluate.go        snapshot-aware Tier-1 path (exact-match reuse, time-boxed)
.moai/state/verify/snapshots/    <key-prefix>.json artifacts (gitignored)
```

### A.2 Data model (schema.go)

One snapshot JSON per key. Field shape (contract-level; exact Go naming is run-phase):

```json
{
  "key": "<head-sha>:<sha256(porcelain-v2)[:16]>",
  "recorded_at": "<ISO-8601>",
  "checks": [
    {
      "check_id": "test | lint | format | type | coverage | <free-form>",
      "command": "<exact byte-string executed>",
      "exit_code": 0,
      "duration_ms": 41200,
      "conditions": {
        "zero_errors": true, "error_count": 0,
        "tests_pass": true,
        "coverage_threshold": 85, "coverage_actual": 87.0,
        "zero_warnings": false
      }
    }
  ]
}
```

The `conditions` block is read-compatible with the loop-verdict `conditions` shape (REQ-SNAP-001) — existing loop-verdict readers keep working; per-check `conditions` fields are populated only where applicable.

### A.3 Key computation (key.go) — settled decision D1 #1

`key = HEAD commit SHA + digest(git status --porcelain=v2)`.

- Porcelain v2 lists staged/unstaged deltas AND untracked non-ignored paths → any tracked-content change, staged change, or untracked add/remove/rename invalidates the key.
- Accepted limitation: an in-place content edit to an already-listed untracked file changes neither HEAD nor its porcelain line → outside the digest. Mitigated by the TTL (A.4) and named as Residual-risk in consumer reports.
- Cost profile: two git subprocesses, constant w.r.t. repo history size — required for the stop-goal path (A.6).

### A.4 Freshness predicate (freshness.go) — settled decision D1 #2

`Fresh(snapshot) ⇔ key(now) == snapshot.key AND now - recorded_at <= TTL`.

- TTL default **10 minutes**, configurable (config key naming is run-phase; default constant lives in the package).
- Both legs binary; either failing ⇒ stale ⇒ the caller re-executes. A stale snapshot is never citable evidence (REQ-SNAP-011).

### A.5 CLI surface

- `moai verify record` — record one check result (flags or stdin JSON) under the current-tree key; creates/updates the key's snapshot file atomically.
- `moai verify check --key-current [--check <id>]` — freshness query; exit 0 fresh / exit 1 stale; prints snapshot path + key + matched entry for citation.
- Registered in the root command tree (reachability AC-WCO-004).

### A.6 stop-goal integration (internal/goal) — settled decision D1 #3

- Before running a Tier-1 mechanical condition, the evaluator looks up the current-key snapshot and matches the condition's `cmd` against recorded `command` values by **exact byte-string equality**. Hit + fresh ⇒ reuse recorded exit code, record attribution (snapshot path + key) in the verdict payload. Miss / stale ⇒ existing `CmdRunner` execution path, unchanged.
- **Time-box (Advisory-Check Discipline)**: key computation on the Stop-hook path runs under a context deadline; deadline exceeded ⇒ skip the optimization, re-execute. Correctness is never traded for the optimization.
- Normalized check-id matching: Out of Scope (spec.md §D), M2+ candidate.

### A.7 Consumer wiring + force-fresh (D7)

| Consumer | Role | Mode |
|----------|------|------|
| `/moai gate` | consume + produce | default reuse; **`--fresh` disables ALL consumption** (still records) |
| run Phase 2.75 | consume + produce | default reuse |
| sync Phase 0 (`gate-sync-1`) | consume | default reuse, citation-as-evidence |
| loop Step 3 | produce | writes via `moai verify record` |
| loop Step 1 | read own Step-3 snapshot | mechanical predicate re-evaluation |
| loop Step 1.5 | **force-fresh** | invokes `/moai gate --fresh` — no same-run consumption, direct or gate-mediated; MAY produce for downstream |
| stop-goal evaluator | consume | exact-match + time-boxed |

The `--fresh` mechanism exists because Step 1.5 is defined as a gate re-run: without a gate-side knob, the same-run snapshot would flow back transitively through the consuming gate (plan-audit iter-1 D7).

### A.8 Failure modes

- `.moai/state/` unwritable → record skipped with explicit note; consumers re-execute (fail-open, never fabricate).
- Concurrent writers, one checkout → atomic temp+rename; identical-key last-writer-wins is benign (same tree ⇒ equivalent results).
- First run / absent snapshot → plain re-execution (strictly additive contract).

## §B — Parity Model (D2)

- Workflow/rule files: **byte-parity** (`cmp -s`) local ↔ template.
- Agent files (`.claude/agents/moai/*.md`): **sanitized-parity** — byte-equal after normalizing internal long-form SPEC-ID refs to sanitized short form. Known instance: sync-auditor.md `(SPEC-V3R2-HRN-003)` ↔ `(HRN-003)`. Internal SPEC IDs are never copied into the template (§25 neutrality CI guard); each sanitization line is enumerated in the AC-WCO-036 evidence.

## §C — Design Cross-References

- spec.md §C.1 (REQ-SNAP-001..011) — the normative contract this design realizes.
- plan.md § Settled decisions — decision provenance (user decisions, plan-audit iter-1 D1) + package/CLI naming.
- research.md §A/§F — duplication inventory + contradiction ledger the design resolves.
- `internal/goal/state.go` — atomic-write pattern reused by store.go.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — attribution semantics of a reused snapshot.
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline — the Stop-hook time-box constraint.
