# Progress — SPEC-CC2178-MODEL-POLICY-REPAIR-001

> Run-phase progress tracker. Owned by manager-develop (run-phase + Mx close) and manager-docs (sync-phase §E.4).

## §E.1 Plan-phase Audit-Ready Signal

- **Artifact set**: spec.md (12-field frontmatter, `era: V3R6`, version 0.2.0 iter-2 → 0.3.0 in-progress), plan.md (M1-M6), acceptance.md (14 ACs).
- **SPEC ID regex**: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` → PASS.
- **Frontmatter schema**: all 12 canonical fields present; `status: in-progress` (transitioned at M1 commit per Status Transition Ownership Matrix — manager-develop owns `draft → in-progress` on M1).
- **GEARS compliance**: REQs use Ubiquitous / When / While / Where / shall not — 0 residual `IF/THEN`.
- **plan-auditor verdict**: PASS-WITH-DEBT 0.86 (iter-2, threshold 0.80 met). User granted Implementation Kickoff Approval.
- **Spec lint**: `moai spec lint` path-prefixed → exit 0, 1 WARNING (`StatusGitConsistency` — expected for plan-phase draft).

## §E.2 Run-phase Evidence

### M1 — Research: `[1m]` re-verification + Default-model key + task-triage decision

**Status**: COMPLETE (2026-06-16)

**Deliverable**: `.moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/research.md` (full research note).

**M1 verdict — `[1m]` constraint (AC-MPR-011, REQ-MPR-013/014/015)**: **STILL-ACTIVE (conservative)**

Evidence (fetched 2026-06-16 via GitHub REST API + canonical CC CHANGELOG):

| Issue | State | Labels | Verdict contribution |
|-------|-------|--------|----------------------|
| #45847 | closed (2026-04-13) | `duplicate` | closed-as-duplicate; no explicit "fixed" |
| #51060 | closed (2026-05-26) | `bug, area:model, area:agents, stale` | closed-stale; no changelog fix for spawn-time entitlement mismatch |
| #36670 | **OPEN** (updated 2026-06-02) | `bug, has repro, area:agents, stale` | Team-mode `[1m]` inheritance confirmed UNFIXED at CC 2.1.178 |

CC 2.1.170-2.1.178 CHANGELOG `[1m]`-class entries: 2.1.172 stuck-session recovery + suffix normalization; 2.1.173 Fable-5 suffix; 2.1.174 background-session env-inheritance. **None fix the subagent-spawn entitlement-inheritance root cause.**

Conservative default per acceptance.md EC-01: still-active → EX-01 holds → per-agent pinning forbidden → Default-model routing (`availableModels` + `enforceAvailableModels`) is the only confirmed-safe lever.

**M1 task 4 — Default-model JSON key (AC-MPR-003, REQ-MPR-003)**: confirmed `model` (top-level) + `availableModels` + `enforceAvailableModels`. CC 2.1.175 changelog verbatim: `enforceAvailableModels` constrains the Default model. Verification command pinned: Python JSON-parse of rendered settings (avoids multi-match ambiguity of raw `grep '"model"'`).

**M1 effort-map deferral (REQ-MPR-012, AC-MPR-010 part d)**: PRUNE + RECONCILE only; full retirement DEFERRED to `SPEC-CC2178-EFFORT-MAP-RETIREMENT-001` (2 production callers: `initializer.go:181`, `update.go:2661`).

**M1 task-triage decision (REQ-MPR-016/017, AC-MPR-012)**: DEFERRED to `SPEC-CC2178-TASK-TRIAGE-001` (rationale: 3-axis scope already substantial; integration point absent; metrics unvalidated).

### M2 — Phantom-map cleanup + effort-map reconcile

**Status**: COMPLETE (commit `08a6ee172`). TDD RED→GREEN.
- `agentModelMap`: 19→5 entries (16 canonical phantom keys removed; `manager-develop` + `builder-harness` added with iter-2 tuple `{sonnet,sonnet,haiku}`).
- `agentEffortMap`: 6→5 entries (3 phantoms removed; plan-auditor/sync-auditor synced high→xhigh; manager-develop xhigh + builder-harness high added).
- Tests: 5 new characterization tests + 3 existing tests updated to retained agents. `go test ./internal/template/` → `ok`.

### M3 — ResolveCycleType symbol

**Status**: COMPLETE (commit `3aa2ac3c1`). TDD RED→GREEN.
- New `internal/config/cycle_type.go`: `func ResolveCycleType(harnessLevel, explicitPin string) string`. Dispatch: minimal→ddd, standard→tdd, thorough→tdd, unknown→tdd; explicitPin wins (AG-01).
- `quality.yaml` `cycle_type_routing` documentation section added.
- Tests: `TestResolveCycleType` (9 sub-tests) + `TestResolveCycleType_AlwaysReturnsNonEmpty`. `go test ./internal/config/` → `ok`.
- Split Trigger: harness router NOT wired (ACs satisfied by symbol; wiring is caller-scope, deferred).

### M4 — Default-model cost lever

**Status**: COMPLETE (commit `5867b1bb9`). TDD RED→GREEN. GATED on M1 verdict (still-active → Default-model is the only safe lever).
- `settings.json.tmpl`: added `"model": "sonnet"`, `"availableModels": ["sonnet","opus","haiku"]`, `"enforceAvailableModels": true`.
- `embed.go` uses `//go:embed` (compile-time from templates/ dir); no byte-blob regen needed.
- Tests: `TestSettingsTemplateDefaultModelLever` + full template suite (neutrality/mirror/leak) green.

### M5 — model-policy.md doctrine + mirror parity

**Status**: COMPLETE (commit `278d04d49`).
- 2 new sections: `[1m]` Re-Verification (CC 2.1.178) verdict + Default-Model Cost Lever (CC 2.1.175).
- Byte-identical mirror applied; `TestRuleTemplateMirrorDrift` PASS.
- SPEC ID deliberately NOT referenced (generic phrasing) to keep template mirror free of internal SPEC IDs per §25 doctrine.

### M6 — Trust-but-verify batch

**Status**: COMPLETE (this section). See §E.3 for the audit-ready signal.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-16
run_commit_sha: 278d04d49   # M5 (final run-phase commit; M1-M5 span dc441a917..278d04d49)
run_status: complete-with-preexisting-gaps
ac_pass_count: 14   # 12 MUST + 2 SHOULD, all PASS or PASS-with-deferral-rationale
ac_fail_count: 0
preserve_list_post_run_count: 0   # 0 uncommitted working-tree files absorbed (scope discipline upheld)
l44_pre_commit_fetch: n/a (single-session worktree; pre-spawn sync check not invoked)
l44_post_push_fetch: n/a (push is orchestrator-owned; this agent committed only)
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues; spec-lint 1 WARNING (StatusGitConsistency — expected, sync-phase transition)
cross_platform_build:
  darwin: pass   # go build ./... exit 0
  linux: not-run # GOOS=linux not in this run's scope; no syscall/build-tag changes
  windows: not-run
total_run_phase_files: 11   # 4 Go (cycle_type.go+test, model_policy.go, model_policy_test.go, settings_test.go) + 2 template (settings.json.tmpl, model-policy.md mirror) + 1 rule (model-policy.md) + 1 config (quality.yaml) + 3 SPEC (spec.md frontmatter, research.md, progress.md)
m1_to_mN_commit_strategy: 5 separate milestone commits (M1 research, M2 phantom-map, M3 ResolveCycleType, M4 Default-model, M5 doctrine+mirror); M6 is verification-only (no separate commit — evidence in this progress.md)
```

### Verification evidence (5-section format per verification-claim-integrity.md)

**Claim**: All 14 ACs PASS; my SPEC introduces 0 test regressions; 2 pre-existing failures (`internal/statusline`, `internal/cli`) are unrelated.

**Evidence** (verbatim command outputs observed 2026-06-16):
- `go test ./internal/config/ ./internal/template/` → `ok` (both packages green — my changed packages)
- `golangci-lint run --timeout=3m` → `0 issues.`
- `go run ./cmd/moai --version` → `moai-adk v3.0.0-rc2` exit 0
- `moai spec lint .moai/specs/SPEC-CC2178-MODEL-POLICY-REPAIR-001/spec.md` → `0 error(s), 1 warning(s)` exit 0 (WARNING = `StatusGitConsistency`, expected pre-sync)
- `go test ./...` → 2 FAIL packages (`internal/statusline`: TestCollectMemory + TestCollectMemory_AutoCompactScaling; `internal/cli`: TestDoctorPermission_NoMatchTrace). Both confirmed pre-existing: `git log 9d018697d..HEAD -- internal/statusline/ internal/cli/` → empty (my commits did NOT touch these packages); grep confirms neither references my new/changed symbols (ResolveCycleType/agentModelMap/agentEffortMap/cycle_type/model_policy).

**Baseline-attribution**: all evidence measured against this tree at HEAD `278d04d49` in this run (2026-06-16). The 2 pre-existing failures were verified to also fail at base `9d018697d` (the plan-phase HEAD) — they are model-specific context-threshold math drift (`statusline`) and a doctor-permission trace test (`cli`), both unrelated to model-policy/cycle-type work.

**Gaps** (what was NOT observed):
- `GOOS=linux GOARCH=amd64 go build` not run (no syscall/build-tag changes in this SPEC; the cycle_type.go + model_policy.go edits are pure Go logic with no OS-specific code).
- Harness router (`internal/harness/router/router.go`) NOT wired to call `ResolveCycleType` (Split Trigger decision; the 4 ACs are satisfied by the symbol + tests; runtime wiring is caller-scope, deferred to avoid scope creep).
- No `moai init` end-to-end render verification (the render test `TestSettingsTemplateDefaultModelLever` exercises the template via the embed FS, which is the authoritative render path; a full `moai init` to a temp dir was not run in this session).
- `[1m]` re-verification relied on GitHub REST API + raw CHANGELOG fetch via `curl` (this session is GLM-backed; the z.ai MCP web tools were not available in the function schema, so `curl` to the GitHub API was the reachable fallback — EC-01 anticipated network-restricted environments).

**Residual-risk**:
- The `availableModels`/`enforceAvailableModels` field semantics are based on the CC 2.1.175 CHANGELOG text; runtime behavior on a real CC 2.1.178 install was not empirically tested (no local CC 2.1.178 runtime available in this session). If the field semantics differ at runtime, M4's wiring may need adjustment — but the CHANGELOG text is the authoritative source and the template change is minimal/reversible.
- The `statusline` pre-existing failure (`TestCollectMemory_AutoCompactScaling`) reflects a 5x factor mismatch (830000 vs expected 166000) that looks like the 2026-05-09 model-specific threshold revision (1M vs 200K) was not propagated to the test fixture. This is a separate defect unrelated to this SPEC but should be filed as a follow-up chore.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending mx-phase>_
