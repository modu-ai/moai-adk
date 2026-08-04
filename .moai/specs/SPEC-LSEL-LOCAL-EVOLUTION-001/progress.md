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
run_status: "M1-complete"
ac_pass_count: 2
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

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
