# progress.md — SPEC-CONTEXT-ENGINE-RIGHTSIZE-001

> Tier M progress tracker. Skeleton emitted at plan-phase; §E.2-§E.4 are placeholder headings only (per canonical §E skeleton generation policy — era.go parses literal `§E.2`/`§E.3`/`§E.4` tokens). manager-spec populates only §E.1; run/sync evidence is owned by manager-develop / manager-docs respectively.

---

## §A. Status

- **Phase**: run (manager-develop)
- **Status**: in-progress
- **Tier**: M (3 artifacts: spec.md + plan.md + acceptance.md)
- **Plan-phase commit**: 1fb78d44f (plan v0.1.1, plan-auditor PASS 0.86 review-2)

---

## §B. Milestone Tracker

| Milestone | Title | Status | Commit SHA |
|---|---|---|---|
| M1 | Expressive transition: code_comments line | ✅ done | (pending push) |
| M2 | Tool Selection consolidation + informational reframing | ✅ done | (pending push) |
| M3 | Template mirror synchronization + §25 neutralization | ✅ done | (pending push) |
| M4 | Regression verification (A-group + C-group + lint parity) | ✅ done | (pending push) |

Legend: ⬜ pending / 🟡 in-progress / ✅ done / ⚠️ blocked

---

## §C. Decision Log

| Date | Decision | Rationale | Authority |
|---|---|---|---|
| 2026-07-28 | Conservative "B-group only" scope | GOOS decision; preserve all A-group Frozen + C-group mechanical guardrails | GOOS |
| 2026-07-28 | Tier M classification | 3 files, no Go changes, no architectural decisions deferred | manager-spec |
| 2026-07-28 | M1.3 = verification-only (no defect) | Direct grep confirmed SSOT reference already exists at `plan-auditor.md:~144` | manager-spec (per `feedback_defect_claim_verification`) |
| 2026-07-28 | M1.4 (CLAUDE.md Tool Selection) out of scope | Direct grep confirmed CLAUDE.md has no such section | manager-spec |

---

## §D. Baseline Captures (pre-edit, locked at plan-phase)

| Baseline | Value | Source |
|---|---|---|
| `[HARD]` in CLAUDE.md | 15 | `grep -c '\[HARD\]' CLAUDE.md` |
| `[ZONE:Frozen]` in `.claude/rules/moai/` | 66 across 13 files | `grep -rc '\[ZONE:Frozen\]'` |
| `[ZONE:Evolvable]` in `.claude/rules/moai/` | 98 | `grep -rc '\[ZONE:Evolvable\]'` |
| `MUST` in `.claude/rules/moai/` | 269 | `grep -rc '\bMUST\b'` |
| `NEVER` in `.claude/rules/moai/` | 14 | `grep -rc '\bNEVER\b'` |
| `moai-constitution.md` "Use X instead of Y" bullets | 5 | `grep -c '^- Use .* instead of'` |
| `moai-constitution.md` "English comments" line | 1 (line ~77) | `grep -n 'English comments'` |
| `plan-auditor.md` SSOT cross-reference | observable (line ~144) | `grep -n 'agent-common-protocol.md.*Tool Selection by Task'` |

---

## §E.1 Plan-phase Audit-Ready Signal

- **plan_status**: audit-ready (plan-auditor PASS 0.84; D1/D2/AC-CER-006b defects corrected + directly verified 2026-07-29)
- **plan_complete_at**: 2026-07-29
- **artifact_set**: spec.md + plan.md + acceptance.md + progress.md (Tier M)
- **frontmatter_validated**: 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + optional (tier, related_specs)
- **SPEC ID regex**: PASS (`SPEC-CONTEXT-ENGINE-RIGHTSIZE-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`)
- **out_of_scope_section**: present (8 `### Out of Scope — <topic>` H3 sub-headings)
- **applied_lessons_linked**: feedback_claimed_correction_never_applied, feedback_defect_claim_verification, feedback_local_template_sync_neutralize_first, feedback_guard_observation_must_be_falsifiable, feedback_guard_signal_proves_call_not_effect, feedback_plan_commit_subject_feat_prefix, feedback_shared_checkout_concurrent_commit_race, feedback_index_level_commit_shared_checkout

---

## §E.2 Run-phase Evidence

Run-phase completed 2026-07-30 by manager-develop. All 4 milestones applied; 10/10 ACs PASS with verbatim grep evidence cited below.

### AC PASS/FAIL matrix (acceptance.md §B)

| AC | Status | Verification command | Observed output |
|---|---|---|---|
| AC-CER-001 (M1 code_comments) | PASS | (a) `grep -c 'Clear naming, English comments'` / (b) `grep -c 'match the surrounding code.*language and density'` / (c) `grep -c 'code_comments.*language.yaml'` | (a)=0 / (b)=1 / (c)=1 |
| AC-CER-002 (M2 Tool Selection consolidation) | PASS | (a) `grep -c '^- Use .* instead of'` / (b) `grep -c 'agent-common-protocol.md.*Tool Selection by Task'` (in constitution) / (c) `grep -Ec 'prefer the dedicated tool\|fit for purpose'` | (a)=0 / (b)=1 / (c)=1 |
| AC-CER-003 (canonical SSOT retained) | PASS | `grep -c '^### Tool Selection by Task'` agent-common-protocol.md | 1 |
| AC-CER-004 (plan-auditor SSOT regression guard) | PASS | `grep -c 'agent-common-protocol.md.*Tool Selection by Task'` plan-auditor.md | 1 |
| AC-CER-005 (A-group Frozen count) | PASS | (a) total `grep -rc '\[ZONE:Frozen\]' \| awk` / (b) per-file diff vs baseline-frozen-distribution.txt | (a)=66 (>=66) / (b)=empty diff |
| AC-CER-006 (A-group AskUserQuestion doctrines) | PASS | (a) AskUserQuestion-Only / (b) `[HARD] Subagents MUST NOT prompt the user` / (c) `ToolSearch(query: "select:AskUserQuestion")` | (a)=1 / (b)=1 / (c)=8 |
| AC-CER-007 (A-group safety invariants) | PASS | (a) BRANCH_GUARD_VIOLATION / (b) `no unobserved-claim\|1.1 Binding scope` / (c) Native `/goal` Prohibition | (a)=1 / (b)=2 / (c)=1 |
| AC-CER-008 (C-group mechanical guardrails) | PASS | (a) Multi-File Decomposition / (b) Reproduction-First Bug Fix / (c) `[HARD]` markers | (a)=3 (>=2) / (b)=3 (>=2) / (c)=2 (>=2) |
| AC-CER-009 (template mirror + §25) | PASS | (a) M1 mirrored / (b) M2 mirrored / (c) no SPEC ID / (d) no REQ-CER / (e) no audit citation | (a)=1 / (b)=1 / (c)=0 / (d)=0 / (e)=0 (every file :0) |
| AC-CER-010 (lint + test parity) | PASS | (a) `go run ./cmd/moai spec lint` / (b) `go test ./...` | (a)=0 errors, 64 warnings (all pre-existing grandfathered-era, none reference moai-constitution.md) / (b)=ok (internal/web flake re-ran clean 0.857s; no Go touched) |

### Files edited (run-phase scope)

- `.claude/rules/moai/core/moai-constitution.md` — M1 (line 78) + M2 (§ Tool Selection Priority)
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — M3 mirror (same M1+M2 edits)
- `.moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/spec.md` — frontmatter only (status draft→in-progress, updated 2026-07-30)
- `.moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/progress.md` — §B/§E.2/§E.3 population
- `.moai/specs/SPEC-CONTEXT-ENGINE-RIGHTSIZE-001/baseline-frozen-distribution.txt` — captured at M1 start (13 files / sum 66)

### Invariants preserved (all PASS)

- A-group Frozen: 66 (no downgrade, per-file distribution byte-identical to baseline)
- A-group anchor doctrines (AskUserQuestion-Only / Subagent Prohibitions / Deferred Tool Preload / Branch Guard / Verification-Claim Integrity / Native /goal Prohibition) — all observable
- C-group mechanical guardrails (Multi-File Decomposition / Reproduction-First Bug Fix) — unchanged, [HARD] markers intact
- §25 template neutrality — zero SPEC ID / REQ token / audit citation leaked into template mirror

---

## §E.3 Run-phase Audit-Ready Signal

- **run_status**: audit-ready
- **run_complete_at**: 2026-07-30
- **run_commit_sha**: a75d95e76 (M4 regression-verification commit; per the SHA-placeholder backfill exemption D3, a commit cannot reference its own hash, so this is backfilled in a follow-up commit on the same branch)
- **ac_pass_count**: 10
- **ac_fail_count**: 0
- **preserve_list_post_run_count**: 66 ([ZONE:Frozen] preserved; per-file distribution diff-empty vs baseline)
- **l44_pre_commit_fetch**: not applicable (Route B PR — single worktree, no main-checkout mutation)
- **l44_post_push_fetch**: not applicable (Route B PR)
- **new_warnings_or_lints_introduced**: 0 (moai spec lint: 0 errors / 64 warnings, all pre-existing grandfathered-era SPEC findings unrelated to moai-constitution.md)
- **cross_platform_build**: not applicable (rules-edit SPEC, no Go code touched; `go build ./...` not required per spec.md §E constraint 3)
- **total_run_phase_files**: 5 (2 source rules + 1 template mirror + 2 SPEC artifacts [spec.md frontmatter + progress.md])
- **m1_to_mN_commit_strategy**: 3 pathspec commits — M1+M2 bundled (both edits co-locate in moai-constitution.md, cannot split one file across two commits without interactive staging); M3 template mirror; M4 regression verification record. draft→in-progress frontmatter transition rides the M1+M2 commit. Bundled into a single PR (Route B).

### D7 audit-finding correction (Evolvable count baseline-attribution drift)

Audit finding **D7** flagged that the plan/spec narrative states `[ZONE:Evolvable]` x 98, but the actual measured count is **102**. Re-measured 2026-07-30 at run-phase close:

```
$ grep -rc '\[ZONE:Evolvable\]' .claude/rules/moai/ | awk -F: '{s+=$2} END{print s}'
102
```

**Authoritative count: 102** (the 98 in spec.md §HISTORY v0.1.0 / plan.md §A line 14 + §C Baseline 2 / progress.md §D table is baseline-attribution drift — likely measured at a different point in time before 4 additional `[ZONE:Evolvable]` markers landed in sibling rule edits).

**Why no AC depends on the Evolvable count**: per acceptance.md, every A-group preservation AC (AC-CER-005) binds the **Frozen** count (>=66), not Evolvable. The SPEC's conservative scope decision preserves all Frozen markers and does NOT transition any marker between zones — so the Evolvable count is informational baseline context, not a falsifiable gate.

**Where the narrative cannot be corrected**: spec.md §HISTORY + plan.md §A/§C are plan-phase body content — per the Status Transition Ownership Matrix, manager-develop MUST NOT modify spec.md/plan.md body content (frontmatter status+updated only on the draft→in-progress transition). This §E.3 record is the authoritative correction site; the plan-phase narrative drift is documented for a future manager-spec amendment if exact-count parity is desired.

### Residual risk

- The `internal/web` test package produced `signal: terminated` (11.550s) on the first `go test ./...` run but passed in 0.857s on isolated re-run — an environmental flake (test parallelism resource contention), not a regression. No Go code was touched by this SPEC.
- Template byte-parity was verified for the M1 line region (lines 75-82); the M2 region (Tool Selection Priority) is a full-block replacement verified via grep content-match, not a line-range diff.

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates>_

---
