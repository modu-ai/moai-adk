# Progress — SPEC-TEAMMATE-REVIVAL-GUARD-001

> Card t267 · plan-phase artifacts emitted 2026-08-26 on tree `c9eed8ac6` (branch `WT-taskstop-name-reclaim`).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md` (frontmatter 12 canonical fields + tier; status: draft; v0.1.1), `plan.md` (§A–§H; design decision §C with per-layer feasibility evidence E1–E8), `acceptance.md` (AC-TRG-001..011, two-cell adoption, RED pinned to `c9eed8ac6`), this `progress.md` (§E skeleton).
- SPEC ID pre-write check (executed): `ID="SPEC-TEAMMATE-REVIVAL-GUARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`.
- Catalog uniqueness check (executed): `ls .moai/specs/ | grep -c "SPEC-TEAMMATE-REVIVAL-GUARD"` → `0` (pre-emission).
- Out of Scope lint shape: 5 `### Out of Scope — <topic>` H3 sub-headings, each with `-` bullets.
- Enforcement default RESOLVED (plan-audit iteration 2, D1): shipped default **false** — operator decision 2026-08-26, orchestrator AskUserQuestion round, recommended option taken (plan.md §C.3). M3 default-flip trigger remains evidence-gated.
- `plan_status: audit-ready`
- `plan_complete_at: 2026-08-26` — plan-audit iter-2 PASS-WITH-DEBT 0.88 (all MUST-PASS clear, D1–D5 confirmed resolved); report: `.moai/reports/plan-audit/SPEC-TEAMMATE-REVIVAL-GUARD-001-review-2.md`

## §E.2 Run-phase Evidence

Commits (branch `WT-taskstop-name-reclaim`, card-local — push deferred to sync per B9):
`7759b8130` (M1 observe+record) → `f7bd5bdc7` (M2 enforcement + config-cache schema bump) →
`70541af5c` (M2 coverage closure) → M3 deliverables commit (this file set; SHA backfilled in §E.3).

### AC binary matrix (all attribution triples measured on this branch's tree)

| AC | Status | Evidence (command → observed output) |
|----|--------|--------------------------------------|
| AC-TRG-001 stop records registry+audit | PASS | `go test ./internal/hook/ -run TestRecordAgentStopRecordsRegistryAndAudit -count=1` → `ok … internal/hook`; live: registry `.moai/state/agent-stops/612bd790-….json` + `stop_recorded` row 2026-08-26T05:43:06Z (orchestrator recording probe) |
| AC-TRG-002 deny + sentinel + audit | PASS | `TestAgentStopGuardDenySentinelPath` → deny, reason prefix `STOPPED_TEAMMATE_VIOLATION:`, names teammate + orchestrator route, `send_denied` row; live: rows 2026-08-26T17:12:53Z & 17:13:10Z (orchestrator + teammate-issued, session `703df7e1-…`) |
| AC-TRG-003 uncertain never denies | PASS | `TestAgentStopGuardUncertainNeverDenies` (malformed JSON / missing `to` / corrupt registry → allow + observe-only row) |
| AC-TRG-004 live teammates unaffected (mirror mutant) | PASS | `TestAgentStopGuardLiveTeammatesUnaffected` (live Y / `main` / `notify_when_idle`→Y all allow); live: `send_observed name:"main"` allowed 17:13:33Z |
| AC-TRG-005 same-name spawn clears | PASS | `TestSpawnNameClearsStoppedEntry` (+`TestSpawnOtherNameKeepsEntry` complement); live: `respawn_cleared` 17:13:21Z, post-clear send delivered and acknowledged end-to-end (orchestrator-verified step-6) |
| AC-TRG-006 session end clears | PASS | `TestSessionEndClearsStops` (sess-1 file removed + `session_cleared` row; sess-2 survives); live: session `703df7e1`'s registry file absent after its end. Honest gap (orchestrator-noted): live removal is inferred from file absence — unit-test direct assertion + audit row carry the contract |
| AC-TRG-007 gate off ⇒ advise never deny | PASS | `TestAgentStopGuardGateOffNeverDenies` (advisory surfaces; row records `advisory:true` would-have-denied state) + `TestWorkflowAgentStopGuardDefaultFalse` (defaults.go ships false) |
| AC-TRG-008 audit row shape exactly-once | PASS | `TestAgentStopAuditRowShape` — all rows carry `{timestamp UTC RFC3339, session_id, kind, name, agent_id, decision}`; M2 kinds (`send_denied`, `respawn_cleared`, `session_cleared`) same shape + `advisory` flag; live rows in `.moai/logs/agent-stop-audit.jsonl` conform |
| AC-TRG-009 wiring + mirror + neutrality | PASS | `grep -c '"matcher": "SendMessage\|TaskStop"' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` → `1`/`1` (pre-`|` alternation escaped in shell; verbatim matcher `SendMessage|TaskStop` ×1 each); `make build` clean; `go test ./internal/template/...` → `ok 34.408s` (neutrality + leak audits); config default exists in `internal/config/defaults.go`; twin neutrality greps 0 hits (no SPEC IDs / SHAs / dates) |
| AC-TRG-010 live-session firing gate | PASS | Orchestrator E-P1 (PATH-pinned fresh headless sessions, worktree binary): probe A 2026-08-26T05:43Z — spawn→TaskStop→registry entry present; probe B 2026-08-26T17:12–17:13Z (2026-08-27 KST) — full recipe: deny (lead-issued), deny (teammate-issued, E4 parity), same-name respawn → cleared, post-respawn send delivered + acknowledged. Live-shape ground truth recorded: TaskStop payload carries the spawn name in the field mapped to `agent_id` with display-name empty — pinned as `TestExtractTaskStopTargetLiveShape` + `TestAgentStopGuardLiveShapeLifecycle` (matcher compares BOTH registry fields) |
| AC-TRG-011 `name [ref]` both directions | PASS | `TestAgentStopGuardNameRefBothDirections` (stopped `X [ref]` DENIED / live `Y [ref]` ALLOWED) + live-shape `[ref]` subtest |

### Cross-cutting findings recorded for sync

- **Pre-existing defect fixed by this card**: config disk-cache schema poisoning — a cache
  written by an older binary (struct lacking newly added fields) is served as fingerprint-valid
  to a newer binary, which unmarshals zero-values over live file values (observed:
  `enabled:false` served over `enabled:true`). Fix: `configCacheSchemaVersion` 1→2
  (`internal/config/cache.go`, field-addition-is-a-compat-break rationale in the constant's
  comment). **CHANGELOG must carry this** — it affects every config-field addition, not just
  this SPEC. Isolation evidence: `MOAI_CONFIG_CACHE_DISABLED=1` load read the true value.
- Live tests wrote two audit rows under the worktree `.moai/logs/` during binary smokes; all
  unit tests write only under `t.TempDir()`.
- PRESERVE surfaces untouched across the run (`.claude/rules/**`, other SPECs, research,
  reports, runtime state/logs beyond the guard's own runtime artifacts); PRETOOL-PERF
  fixture rewrites from test runs were restored before each commit.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: "aac630f25"   # M3 deliverables commit (backfilled one commit later — self-referential-hazard workaround)
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "origin/main...HEAD = 3 7 (origin advanced 3 since base c9eed8ac6 via parallel landings; card branch 7 ahead; reconciliation at sync)"
l44_post_push_fetch: "no push (B9: card-local branch; push + PR at sync via manager-git)"
new_warnings_or_lints_introduced: 0      # golangci-lint 0 issues, baseline 0
cross_platform_build:
  darwin: exit-0
  windows_amd64: exit-0
total_run_phase_files: 17                # 14 code/wiring/config (M1 7 unique + M2 10, overlap 3) + proposal-rule-amendment.md + progress.md §E.2/§E.3
m1_to_mN_commit_strategy: "per-milestone commits (M1 / M2 feat+test pair / M3 deliverables docs), no amend, no push; run_commit_sha backfilled one commit later (self-referential-hazard workaround)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier M (3 artifacts); scope ≈6–9 files (`internal/hook/agent_stop_guard.go` + tests, `internal/config` defaults+loader, `.claude/settings.json` + template twin `settings.json.tmpl`, catalog parity if templates change); domain count 3 (hook Go code / config / template twins); language mix Go + JSON-settings + markdown; concurrency benefit LOW (coding-heavy, interdependent milestone chain); Agent Teams prereqs not requested (no `--team`).

Mode evaluation:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file Go implementation with tests — not a typo/single-line fix |
| serial | **SELECTED** | coding-heavy Tier M; single implementation agent keeps the write surface single-owner; milestones interdependent (M2 deny builds on M1 registry; M3 dogfood builds on M2) |
| fanout | no | coding-heavy per Anthropic's coding-task parallelism caveat; no multi-domain research fan-out needed |
| sweep | no | ≈9 files (< ~30); semantic new-code work, not a uniform mechanical transform |
| agent-team | no | experimental + explicit-request-only; not requested |

Decision: serial

Justification: the run-phase scope is Go hook-handler code + config + template twins with a strictly sequential dependency chain, so sequential single-writer delegation is the safe default (Anthropic coding-task parallelism caveat: most coding tasks involve fewer truly parallelizable tasks than research). Progression mode: **autonomous** — operator choice at Implementation Kickoff Approval (2026-08-26); the `ac_converge` goal is armed alongside run-phase start (arm-only; it substitutes for no gate). Boundary Case: none — no threshold boundary was hit (scope and domain count sit clearly below fanout/sweep thresholds).
