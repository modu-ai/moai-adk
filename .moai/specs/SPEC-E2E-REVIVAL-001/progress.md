# SPEC-E2E-REVIVAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-07-13
- artifacts: spec.md, plan.md, acceptance.md, research.md, design.md, progress.md (6, Tier L)
- status: draft v0.1.2 — iter-1 FAIL 0.80 → D1-D12 fixed (v0.1.1); iter-2 PASS-WITH-DEBT 0.88 → D13-D15 residual-fixed (v0.1.2: repo-wide count-literal rings + REQ-E2E-305 Go display pin + CMD-019-INV widened baseline 38 / INV-B 7 + lint gate "0 errors" recalibration); run-ready
- clarifications: ALL RESOLVED 2026-07-13 via orchestrator AskUserQuestion (D-7 desktop-native → deferred to follow-up; D-8 mobile default → Maestro confirmed; H-3 docs-site → deferred to follow-up) — clarification markers removed from plan.md/research.md

## §E.2 Run-phase Evidence

Run-phase executed 2026-07-13 by manager-develop (cycle_type=tdd, Mode 5 sequential milestones M1→M5). Verbatim command outputs persisted under `.moai/state/verify/e2e-revival/` (gitignored runtime state; paths cited per row).

### Milestone commits

| Milestone | Commit | Scope |
|-----------|--------|-------|
| M1 | `8190c3340` | e2e-specialist agent (both trees, byte-identical) + spec.md draft→in-progress |
| M2 | `a500fafe8` | e2e workflow skill (both trees, 294 lines) |
| M3 | `cc0156396` | thin command pair + SKILL.md router + CLAUDE.md §3/§4 + count-literal rings 1-3 (28 files) |
| M4 | `3235a7233` | catalog.yaml entry + expectedAgentCount 10 / expectedTotal 38 / wantTotal 38 + model_policy tier-profile pin (66 cells) + make build |
| M5 | (this commit — see §E.3 run_commit_sha) | evidence + [HARD] CLI-first single-line hardening + docs-libraries metadata-key convergence + hash regen |

### AC matrix (28/28 PASS)

| AC | Status | Verification command | Actual output / evidence path |
|----|--------|---------------------|-------------------------------|
| AC-E2E-001 | PASS | Read workflow §Detection Matrix | 9-row matrix (web / mobile×3 / desktop×2 / desktop-native / mixed / none), ≥2 markers per class, marker-driven prose; `m5-cmd027.log` → `0` |
| AC-E2E-002 | PASS | Read workflow §Tool Matrix + §Toolchain Probe + Installation | Defaults: web→Playwright CLI, mobile→Maestro, desktop→Playwright `_electron` / WDIO+tauri-service; per-default probe + install commands in table |
| AC-E2E-003 | PASS | `grep -n 'AskUserQuestion' <workflow>` | 7 hits (L40/100/110/158/258/272/283) — every hit orchestrator-addressed; agent grep → 0; `m5-ac003-askuser.log` |
| AC-E2E-004 | PASS | grep `--tool` flags row + Phase 0.5 bypass | Flag row (workflow L40) + "If `--tool` is provided: bypass the selection question" branch (Phase 0.5) |
| AC-E2E-005 | PASS | Read workflow §Toolchain Probe + Installation | Probe → Surface (orchestrator AskUserQuestion approval) → Install → Re-probe 4-step sequence; per-toolchain probe/install table |
| AC-E2E-006 | PASS | Read workflow §No-Target Graceful Exit | "no e2e target detected" branch with marker evidence + `desktop-native` deferral notice; "There is no opt-in automation path for desktop-native" |
| AC-E2E-007 | PASS | `grep -nE '\[HARD\]'` both files | Workflow L52 `- [HARD] **CLI-first**`; agent L92 `[HARD] CLI-first`; Tool Matrix carries token-cost column; `m5-final-parity.log` |
| AC-E2E-008 | PASS | CMD-008 | `1` per file (both trees); redirect example in agent rung 1; `m5-cmd008.log` |
| AC-E2E-009 | PASS | Read agent §ladder rung 3 | "snapshot/aggregate reads … over per-element round-trips; never per-element polling loops" |
| AC-E2E-010 | PASS | Read workflow Phase 4/5 | Native-facility recording table (`e2e/traces/`, `e2e/recordings/`), "cited by path … never inlined"; no MCP-screenshot-loop path |
| AC-E2E-011 | PASS | Read Tool Matrix tier column | All 4 platform DEFAULT rows CLI-class; MCP rows marked "MCP (conditional)"; "No MCP hard dependency" [HARD] |
| AC-E2E-012 | PASS | `test -f` both + `go test -run TestCommandsThinPattern` | BOTH-COMMAND-FILES-EXIST; exit 0; `m5-ac012-014-tests.log` |
| AC-E2E-013 | PASS | `diff` workflow both trees + `grep -c 'user-invocable: false'` | WORKFLOW-IDENTICAL; count 1; `m5-final-parity.log` |
| AC-E2E-014 | PASS | `diff` agent both trees + manual field checklist + `go test -run TestAgentFrontmatterAudit` | AGENT-IDENTICAL; frontmatter complete (name/description w/ PROACTIVELY+NOT-for/tools CSV no leading `-`/model inherit/effort high/color cyan/permissionMode default/memory project/skills 1 entry); audit exit 0 |
| AC-E2E-015 | PASS | CMD-015 + `grep -c 'Missing Inputs'` | `0` violations; Missing Inputs = 1; `m5-cmd015.log`, `m5-final-parity.log` |
| AC-E2E-016 | PASS | `grep -c e2e-specialist <workflow>` + `test -f` + CMD-016 | 22 references (≥3); both agent paths exist; CMD-016 → 2 `=== RUN` + exit 0; `m5-cmd016.log` |
| AC-E2E-017 | PASS | `grep -cE '^- \*\*e2e\*\*'` both SKILL.md | 1 each (baseline 0); `m5-ac017-020-router.log` |
| AC-E2E-018 | PASS | frontmatter grep + CMD-018 | description enumeration carries e2e (1); CMD-018 → 2; `m5-cmd018.log` |
| AC-E2E-019 | PASS | `grep -c '11 retained agents'` + `grep -c e2e-specialist` CLAUDE.md + CMD-019-INV + CMD-019-INV-B | 2 + 2 per tree; INV → 0 (baseline 38); INV-B → 0 (baseline 7); `m5-cmd019-inv.log`, `m5-cmd019-invb.log` |
| AC-E2E-020 | PASS | `grep -n 'routes to \*\*e2e\*\*'` P3 section | L92 both trees, phrased as semantic exemplar ("any conversation_language … routes identically") |
| AC-E2E-021 | PASS | `grep -A4 'name: e2e-specialist' catalog.yaml` + hash regex + `TestAllAgentsInCatalog` | tier core / path resolves / version 1.0.0 / hash `20c16b1ec435…` matches `^[0-9a-f]{64}$` (regenerated after M5 agent edit); test exit 0; `m5-ac021-catalog.log` (pre-M5 hash), `m5-make-build-2.log` (final hash) |
| AC-E2E-022 | PASS | `git show --stat` M1-M5 | Every commit carries template + local siblings together (no local-only additions); parity diffs green; `m5-ac022-provenance.log` |
| AC-E2E-023 | PASS | constants greps | `expectedAgentCount = 10` (L234), `expectedTotal = 38` (L57), `expectedSkillCount = 28` STILL present (L158), `wantTotal = 38` (embed_catalog_test.go L48, same pin class); ledger comments adjacent; `m5-ac023-028-constants.log` |
| AC-E2E-024 | PASS | `make build` + `go test ./internal/template/...` | build exit 0 (×2 runs); template tests exit 0; `m5-make-build-2.log`, `m5-final-template-tests.log` |
| AC-E2E-025 | PASS | CMD-025 | `0` matches; neutrality CI tests green within suite; `m5-cmd025.log`, re-verified post-M5-edit in `m5-final-parity.log` → 0 |
| AC-E2E-026 | PASS | CMD-026 + expectedSkillCount + AC-015 | `0` design-pack dirs; `expectedSkillCount = 28` passing; boundary greps 0; `m5-cmd026.log` |
| AC-E2E-027 | PASS | manual matrix review + CMD-027 | ≥2 markers per platform class; web class documented marker-driven ("framework list is exemplary"); CMD-027 → 0; `m5-cmd027.log` |
| AC-E2E-028 | PASS | `grep -c e2e-specialist model_policy.go` + CMD-028 | 4 references (order list + 2 profile rows + comment); CMD-028 → 12 `=== RUN` + exit 0 with pins (len 11, 66 cells); `m5-cmd028.log` |

### Quality gates (G1-G6)

| Gate | Result | Evidence |
|------|--------|----------|
| G1 | PASS — `go test ./...` exit 0 (100 packages ok, 0 FAIL, full suite at final tree state) | `m5-go-test-full-final.log` |
| G2 | PASS — `golangci-lint run` exit 0, "0 issues." | `m5-lint.log` |
| G3 | PASS — `moai spec lint --strict spec.md` → 0 errors; 1 EXPECTED `StatusGitConsistency` WARNING (in-progress vs git-implied — expected until sync close, not suppressed) | `m5-spec-lint.log` |
| G4 | PASS — CMD-025 → 0 + neutrality CI green | `m5-cmd025.log` |
| G5 | PASS — CMD-015 → 0 | `m5-cmd015.log` |
| G6 | PASS — agent + workflow diffs identical both trees | `m5-final-parity.log` |

### PRESERVE-list / invariants

- No modifications outside the declared surface (DoD item 5): M1-M5 pathspec-scoped commits only; parallel-session in-flight edits excluded (local agent-authoring.md foreign `disallowedTools` hunk filtered out of M3 via index-only patch staging).
- Mirror-registration decision (pre-flight #9): sibling top-level workflow files are NOT in `workflowOptMirroredPaths` → no-register (audit-verified precedent followed).
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exit 0 + native build exit 0 (`m5-build-windows.log`, `m5-build-native.log`).
- Coverage snapshot: internal/template 85.3%, internal/web 70.8% (`m5-coverage.log`; C-7 — no new Go runtime logic, one data-row addition covered by TierProfile tests; web coverage is pre-existing baseline, untouched by this SPEC's comment-only edits).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-13
run_commit_sha: 5f22d9022ff69c13d272be2bdb105450d970f1f2
run_status: audit-ready
ac_pass_count: 28
ac_fail_count: 0
preserve_list_post_run_count: 0   # no out-of-scope modifications
new_warnings_or_lints_introduced: 0   # golangci-lint "0 issues."; spec-lint warning pre-existing/expected (D14)
cross_platform_build:
  native: exit 0
  windows_amd64: exit 0
total_run_phase_files: 39   # 3 new artifacts ×2 trees (agent/workflow ×2 + command pair) + SKILL.md ×2 + CLAUDE.md ×2 + ring1 rule/agent files ×2 trees (10) + ring2 skill files ×2 trees (8) + README ×4 + catalog.yaml + 4 Go test files + model_policy.go + spec.md + progress.md
m1_to_mN_commit_strategy: per-milestone pathspec-scoped commits on main (Route A Hybrid Trunk), M1=8190c3340 M2=a500fafe8 M3=cc0156396 M4=3235a7233 M5=5f22d9022
evidence_dir: .moai/state/verify/e2e-revival/
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
