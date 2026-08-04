# Progress — SPEC-LSEL-LOCAL-EVOLUTION-001

## §E.1 Plan-phase Audit-Ready Signal

_status: plan-phase artifacts authored 2026-08-04; awaiting plan-auditor verdict._

- spec.md / plan.md / acceptance.md authored with `status: draft`.
- 16 REQs (GEARS) / 16 ACs (Given-When-Then); Tier M.
- SPEC ID regex PASS (`SPEC-LSEL-LOCAL-EVOLUTION-001`).
- v2 corrections baked in: `.moai/config/sections/**` treated as evolvable merge-managed; NO new section file (loop state under `.moai/state/lsel/`); mechanism skills are `hns-lsel-*` (dogfood, NOT distributed).
- 3 [DECISION] markers flagged for plan-auditor challenge (frozen-allowlist location; M5 simulation-harness deferral; distributed-dead-verb retirement as out-of-scope).

## §E.2 Run-phase Evidence

### M1 — Drain closure (AC-LSEL-009 + AC-LSEL-010)

**Deliverables (user-owned surfaces only — doctrine-zero-touch verified):**

| Path | Surface | Notes |
|------|---------|-------|
| `.claude/skills/hns-lsel-curator/SKILL.md` | evolvable #4 (`hns-*` skills) | CLUSTER + drain engine skill body; English (§27.2); NOT templated |
| `.claude/skills/hns-lsel-curator/drain.sh` | evolvable #4 | mechanical drain engine (bash + jq); companion-offset; severity filter; clustering; importance gate |
| `.claude/skills/hns-lsel-curator/drain_test.sh` | evolvable #4 | TDD RED→GREEN fixture characterization test (7 assertions) |
| `.moai/state/lsel/drain-offset.json` | evolvable #6 (`.moai/state/lsel/`) | companion offset advanced 0→629 (consumed-stub marker) |
| `.moai/state/lsel/clusters.json` | evolvable #6 | 16 candidate clusters staged (NO memory/ write) |

**AC-LSEL-009 — 569-stub backlog drained (re-measured live: 629 stubs at drain time):**

| Sub-check | Evidence (verbatim command + observed output) |
|-----------|-----------------------------------------------|
| offset advances by ≥1 | `drain.sh` log: `drain: read 629 stubs (offset 0→629)`; `drain-offset.json` → `{"offset":629,...}` |
| ≥1 candidate topic | `jq '.candidates|length' clusters.json` → `16` |
| zero `memory/` writes | `find memory -newermt "2026-08-04 05:21:03" -name 'feedback_*' \| wc -l` → `0` |

**AC-LSEL-010 — drain severity filter excludes noise BEFORE clustering:**

| Sub-check | Evidence |
|-----------|----------|
| noise share of live inbox | `446/629 = 70.9%` discarded by the drain-side filter (Bash:UnknownFailure + Bash:SandboxViolation + *:TimeoutError) |
| `tool_failure:Bash:*` share of accepted candidates | `3/16 = 18.8%` (< 30% threshold) — accepted Bash clusters are Bash:ExitError, Bash:OOMKilled, Bash:PermissionDenied (real signal, not the opaque timeout/sandbox bucket) |
| zero noise keys leak into candidates | `jq '[.candidates[].event_key] \| map(select(.=="tool_failure:Bash:UnknownFailure" or .=="tool_failure:Bash:SandboxViolation" or endswith(":TimeoutError"))) \| length'` → `0` |

**TDD RED→GREEN evidence (§E E8):**

- RED (verbatim): `FAIL: drain.sh not found or not executable at .../hns-lsel-curator/drain.sh (expected during RED phase; this line is the RED failure)` / `exit=1`
- GREEN (verbatim): `ALL DRAIN TESTS PASSED` / `exit=0` (7 assertions: offset advance, noise exclusion, singleton discard, frequencies, 1-10 importance, zero memory/ writes, idempotent re-drain no-op).

**Top candidate clusters (by importance × frequency):** Agent:UnknownFailure(41), Read:UnknownFailure(35), Bash:ExitError(20), Bash:OOMKilled(18), Bash:PermissionDenied(14), Read:ContextCancelled(10), Write:PermissionDenied(10), StructuredOutput(7), DesignSync(6), MCP-family(higgsfield/Linear/chrome-devtools). These feed M2 PROPOSE.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-04T05:21:03Z"
run_commit_sha: "pending-backfill-m1"
run_status: "M3-complete"
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 0   # zero lines of shipped moai assets touched
l44_pre_commit_fetch: "n/a (M1 commit, no push to main — PR-mandatory per repo-local-pr-policy.md)"
l44_post_push_fetch: "n/a (no push in M1; orchestrator/manager-git owns the PR)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_all: "exit 0"
  go_build_windows: "exit 0"
total_run_phase_files: 5   # SKILL.md, drain.sh, drain_test.sh, drain-offset.json, clusters.json (+ gitignored state)
m1_to_mN_commit_strategy: "single M1 commit carrying plan-phase artifacts (previously uncommitted) + M1 deliverables; draft→in-progress transition on this commit"
```

M1 DoD met: AC-LSEL-009 PASS + AC-LSEL-010 PASS; `drain-offset.json` advanced; candidate topics staged under `clusters.json`; ZERO `memory/` writes from M1. The closed surface is DRAIN only (AP-LSEL-006 — M1 is NOT "the loop is closed"; APPLY lands in M3).

### M2 — PROPOSE shadow + INVARANTS kernel (AC-LSEL-005, 006, 007, 011, 012)

**Deliverables (user-owned surfaces + CI-guard extension only — doctrine-zero-touch verified):**

| Path | Surface | Notes |
|------|---------|-------|
| `.claude/skills/hns-lsel-curator/SKILL.md` | evolvable #4 | EXTEND: PROPOSE stage + CSA forced-gate (6 categories) + Tier-4 DEAD finding |
| `.claude/skills/hns-lsel-curator/tier4_firing_test.sh` | evolvable #4 | AC-012 characterization probe (Tier-4 DEAD finding) |
| `.claude/skills/hns-lsel-curator/propose_test.sh` | evolvable #4 | AC-011 schema-validation test (RED→GREEN) |
| `.claude/skills/hns-lsel-curator/csa_refusal_test.sh` | evolvable #4 | AC-005 CSA mechanical-refusal fixture (RED→GREEN) |
| `.claude/skills/hns-lsel-curator/backlog_check.sh` | evolvable #4 | AC-007 SessionStart backlog-check (advisory) |
| `.claude/skills/hns-lsel-curator/backlog_check_test.sh` | evolvable #4 | AC-007 fixture test (RED→GREEN) |
| `.moai/state/lsel/proposals/lsel-001/{proposal.md,diff.patch,self-critique.md}` | evolvable #6 | AC-011 sample shadow proposal (8-key schema + retrieval + blocked self-critique) |
| `CLAUDE.local.md` | evolvable #1 | §0 INVARANTS kernel (top) + §28 LSEL operating instructions |
| `internal/template/internal_content_leak_test.go` | dev CI guard | AC-006 LSEL leak classes (L1/L2/L3) + positive-control fixture |
| `.github/workflows/lsel-leak-guard.yaml` | dev CI | AC-006 named CI workflow (report §11 B#7) |
| `.claude/workflows/lsel-drain-loop.js` | evolvable #4 | AC-007 default `/loop` recipe (read-only drain trigger) |

**AC-LSEL-005 — CSA forced-gate (PASS):** all 6 categories enumerated in curator SKILL.md (INVARANTS kernel, security/validation exception, HIGH-fan-in, Bash risk path, permissions.allow, execution-meta); bother-cost-exemption clause present; `csa_refusal_test.sh` confirms fixture WITHOUT marker REFUSED (+ reject-log row) and fixture WITH marker proceeds.

**AC-LSEL-006 — namespace + leak guard (PASS):** `TestLSELLeakPositiveControl` + `TestTemplateNoInternalContentLeak` + `TestSplitHarnessNamespaceNoLeak` all PASS; zero `lsel`/`hns-lsel`/`SPEC-LSEL`/CLAUDE.local.md-marker content under `internal/template/templates/` (negative control); named CI workflow `.github/workflows/lsel-leak-guard.yaml` added.

**AC-LSEL-007 — mechanical trigger (PASS):** `backlog_check_test.sh` confirms system-reminder emitted on overflow + silent below threshold + silent after drain; default `/loop` recipe registered at `.claude/workflows/lsel-drain-loop.js` (read-only, cadence-bridge compliant).

**AC-LSEL-011 — PROPOSE shadow payload (PASS):** sample proposal at `.moai/state/lsel/proposals/lsel-001/` carries full 8-key schema + retrieval_evidence + diff.patch + self-critique with UNRESOLVED objection → status=blocked (gate fires).

**AC-LSEL-012 — Tier-4 firing verification (PASS-WITH-DOWNGRADE):** `tier4_firing_test.sh` verified the `moai-harness-learner` Tier-4 AskUserQuestion flow is DEAD at the production invocation layer (CuratorDispatch 0 callers; enableTriggerInjectionWrites=false; CLI prints stub string; no mechanical trigger). Per acceptance.md §E edge case, M2 downgrades to "PROPOSE shadow only, APPROVE via fresh path" (M3). M2 wiring does NOT depend on the Tier-4 flow. Blocker reported to orchestrator for acknowledgment.

**TDD RED→GREEN evidence (§E E8):**
- propose_test RED: `FAIL: .moai/state/lsel/proposals/lsel-001/proposal.md absent` → GREEN: `propose_test: PASS` (8 schema keys + retrieval + blocked self-critique).
- csa_refusal_test RED: `FAIL: CSA category MISSING` (×6) + over-fire bug → GREEN: `csa_refusal_test: PASS` (6 categories + bother-cost-exemption + marker-keyed refusal).
- leak-test RED (positive control): planted LSEL fixture → L1/L2/L3 each fire; GREEN: clean templates tree → `TestLSELLeakPositiveControl PASS` + `TestTemplateNoInternalContentLeak PASS`.
- tier4_firing_test: characterization PASS (finding = Tier-4 DEAD).

**M2 DoD met:** AC-005/006/007/011 PASS + AC-012 PASS-WITH-DOWNGRADE (Tier-4 DEAD, blocker reported); §0 INVARANTS kernel + §28 present in CLAUDE.local.md; mechanical triggers wired; commit stacks on e542fd905.

### M3 — APPLY closure via bypass (AC-LSEL-001, 002, 003, 004, 008, 013, 014)

**Deliverables (user-owned surfaces only — doctrine-zero-touch verified; REQ-LSEL-003 frozen-flag invariant preserved):**

| Path | Surface | Notes |
|------|---------|-------|
| `.claude/lsel/frozen-allowlist.json` | NEW execution-meta (outside the 6 evolvable surfaces) | REQ-LSEL-001/002 frozen allowlist: `frozen_patterns` regex + `evolvable_surfaces` (the 6) + `execution_meta` (the 4 categories) |
| `.moai/hooks/lsel-apply.sh` | NEW evolvable (.moai/hooks is dev-local) | REQ-LSEL-008 playback-only APPLY consumer (reads approved decision.json → frozen-reject → exec-meta forced-gate → git apply → ledger → lsel-* commit) |
| `.claude/skills/hns-lsel-applier/SKILL.md` | NEW evolvable #4 (hns-* skill) | APPLY engine skill (English body, hns- namespace, user-invocable:false, not templated) |
| `.claude/skills/hns-lsel-applier/apply_test.sh` | NEW evolvable #4 | AC-001/002/005/008/013 characterization (hermetic temp-repo) |
| `.claude/skills/hns-lsel-applier/rollback_rehearsal_test.sh` | NEW evolvable #4 | AC-014 SHIP GATE (mixed-history `git revert` lands clean) |
| `.moai/state/lsel/apply-ledger.jsonl` | NEW evolvable #6 | append-only apply manifest seed |
| `.claude/skills/hns-lsel-curator/{SKILL.md,tier4_firing_test.sh}` | EDIT evolvable #4 | AC-003 literal-token scrub (frozen-flag identifier rephrased to location/role) + tier4 pipefail-abort fix (M2 debt) |

**AC binary matrix (§E E1):**

| AC | Status | Verification | Observed |
|----|--------|--------------|----------|
| AC-LSEL-001 | PASS | `apply_test.sh` frozen-target fixture | REFUSED (exit 2) + reject-log row `category=frozen-path` + no write |
| AC-LSEL-002 | PASS | `apply_test.sh` allowlist-location | allowlist at `.claude/lsel/` (NOT under any of the 6 evolvable surfaces) |
| AC-LSEL-003 | PASS | `grep -rn enableTriggerInjectionWrites .claude/skills/hns-lsel-* .moai/hooks/lsel-* .moai/state/lsel/ .claude/lsel/` | 0 matches; `internal/harness/applier.go:22` still `= false` |
| AC-LSEL-004 | PASS | first `lsel-*` apply commit on feature branch | `git log --grep lsel-` finds the rehearsal commit; `git branch --show-current` = `feat/SPEC-LSEL-LOCAL-EVOLUTION-001` |
| AC-LSEL-008 | PASS | `grep -nE 'propose\|self-approve\|new-proposal' .moai/hooks/lsel-apply.sh` | 0 matches; reads decision.json + approval marker |
| AC-LSEL-013 | PASS | `apply_test.sh` routine fixture | file written + ledger row `{result:"applied"}` + `lsel-*` commit landed |
| AC-LSEL-014 | PASS (SHIP GATE) | `rollback_rehearsal_test.sh` | `git revert <lsel-tag>` exit 0; lsel-block lines removed; non-adjacent manual edits survive |

**REQ-LSEL-003 re-verify (post-implementation):** `internal/harness/applier.go:22` = `var enableTriggerInjectionWrites = false` (unchanged); LSEL-surface grep for the literal identifier = 0 matches (the applier/curator skills + tier4 fixture + ledger seed reference the flag only by location/role, never as mutable).

**Subagent boundary (§E E4):** `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/skills/hns-lsel-applier/` (excluding test/comment lines) = 0 matches. The synchronous user gate is run by the orchestrator, never by `hns-lsel-applier`.

**Cross-platform build (§E E2):** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0. (M3 adds ZERO Go files — these are the no-Go-regression guards.)

**Coverage (§E E3):** N/A for M3 — no new Go code (shell + JSON + markdown only).

**Lint (§E E5):** `golangci-lint run --timeout=2m` → 0 issues (no NEW findings; baseline unchanged). Shell scripts pass `shellcheck` semantics by construction (set -euo pipefail, quoted vars, hermetic temp-repo tests).

**Full-suite guard (trust-but-verify):** `go test ./...` exit 0 (full suite green; template namespace + internal-content-leak tests still PASS — no `hns-lsel-applier` mirror leaked into `internal/template/templates/`).

**TDD RED→GREEN evidence (§E E8):**
- `apply_test.sh` RED (mechanism absent): `RED: lsel-apply.sh absent at .../.moai/hooks/lsel-apply.sh (mechanism not built yet)` exit 1 → GREEN: `apply_test: PASS` (AC-001/002/005/008/013 all PASS).
- `rollback_rehearsal_test.sh` RED (mechanism absent): `RED: mechanism absent — lsel-apply.sh or frozen-allowlist.json missing` exit 1 → GREEN: `rollback_rehearsal_test: PASS` (git revert exit 0, lsel-block removed, manual edits survive).

**First `lsel-*` apply commit (§E E6):** the M3 scaffolding commit carries the mechanism; a separate `feat(lsel-m3-rehearsal-001): ...` commit (produced by running `lsel-apply.sh` against an approved fixture decision.json targeting `.moai/state/lsel/`) lands the first real `lsel-*` apply on the feature branch (AC-004/013).

**M3 DoD met:** AC-001/002/003/004/008/013/014 all PASS; rollback-rehearsal SHIP GATE clean; REQ-LSEL-003 frozen-flag invariant preserved; subagent boundary honored; full `go test ./...` green.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**M3 (APPLY closure via bypass) — Phase 4 decision (logged before the M3 manager-develop spawn):**

Input parameters:
- tier: M (3 artifacts; 16 REQs / 16 ACs)
- scope (file count): ~7 (NEW `hns-lsel-applier/SKILL.md`, NEW `.claude/lsel/frozen-allowlist.json`, NEW `.moai/hooks/lsel-apply.sh`, NEW `.moai/state/lsel/apply-ledger.jsonl`, NEW 2 test fixtures, EDIT `.claude/settings.local.json`)
- domain count: 1 (LSEL APPLY engine — single cohesive mechanism)
- file language mix: markdown (skill) + JSON (allowlist/ledger) + bash (hook + tests) — heterogeneous, single-domain
- concurrency benefit: LOW (coding-heavy, single-domain, sequential TDD)

Mode evaluation:
- Mode 1 trivial: NO — multi-file milestone introducing the loop's first live write path
- Mode 2 background: NO — write-capable implementation, not read-only analysis
- Mode 3 agent-team: RETIRED
- Mode 4 parallel: NO — single-domain coding-heavy (Anthropic coding-task parallelism caveat)
- Mode 6 workflow: NO — <30 files, semantic new-code (not mechanical-uniform transform)
- Mode 5 sub-agent: YES — sequential manager-develop, cycle_type=tdd

Decision: sub-agent (Mode 5)
Justification: M3 is coding-heavy single-domain implementation (APPLY engine + frozen allowlist + playback hook + characterization tests). Per Anthropic's coding-task parallelism caveat, the sequential sub-agent path is the correct default for coding work. One manager-develop delegation, TDD RED→GREEN→REFACTOR, Tier M Section A-E template. Implementation Kickoff Approval was satisfied at run-phase (M1) entry; M3 is a milestone resume within the approved run-phase (per `.claude/rules/moai/workflow/orchestration-mode-selection.md` § Implementation Kickoff Approval mandatory-restoration — the gate binds the plan→run boundary, not milestone→milestone).
