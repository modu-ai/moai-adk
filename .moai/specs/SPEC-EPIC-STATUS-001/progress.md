# Progress — SPEC-EPIC-STATUS-001

> Plan-phase → run-phase → sync-phase lifecycle record. §E.* heading structure is parser-load-bearing (spec-frontmatter-schema.md § progress.md Section Map); do NOT rename.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-11
spec_id: SPEC-EPIC-STATUS-001
tier: L
artifact_count: 6
req_count: 13              # REQ-ES-001..REQ-ES-013
ac_count: 16               # AC-ES-001..AC-ES-015 + AC-ES-003b + AC-ES-004b + AC-ES-005b + AC-ES-008b (sub-IDs)
must_ac_count: 14          # MUST-severity ACs
should_ac_count: 1         # SHOULD-severity (AC-ES-014)
milestone_count: 6         # M0..M5 (M6 = sync close, owned by manager-docs)
baseline_attribution: 9fa242ddae3e5c7e9a80c2b47bd03d38b4c1b5ed
worktree: /Users/goos/.moai/worktrees/kanban
branch: feat/factory-bootstrap-guidance
frontmatter_validated: true   # 12 canonical fields present, no snake_case aliases
spec_id_regex_check: PASS     # ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$
out_of_scope_satisfied: true  # 6 H3 `### Out of Scope — <topic>` sub-headings, each with bullets
ac_req_traceability: true     # acceptance.md §B matrix covers REQ-ES-001..013
template_neutrality: true     # artifacts NOT mirrored to internal/template/templates/
```

Plan-phase self-check (per `moai-workflow-spec` SKILL.md § Verification):

- [x] SPEC file exists at `.moai/specs/SPEC-EPIC-STATUS-001/spec.md` with unique ID `SPEC-EPIC-STATUS-001`.
- [x] Every requirement uses GEARS keywords (`shall` + `When`/`While`/`Where`/`shall not` per the pattern); no IF/THEN modality.
- [x] Every acceptance criterion is observable (CLI output, grep, file-existence, exit-code).
- [x] research.md exists (this SPEC touches existing code: `internal/spec.ListDocs`, `internal/spec.Audit`, `internal/web/board.go`).
- [x] design.md exists (Tier L; epic-discovery 3-strategy chain + JSON shape contract are locked decisions).
- [x] Out of Scope section present with 6 `### Out of Scope — <topic>` H3 sub-headings + bullets.
- [x] No push, no PR (per user instruction; work stays local on `feat/factory-bootstrap-guidance`).

---

## §E.2 Run-phase Evidence

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-ES-001 (disk-grounded) | PASS | `go test ./internal/epic/ -run TestBuildEpicStatus_BASFixture` | `--- PASS: TestBuildEpicStatus_BASFixture (0.00s)` |
| AC-ES-002 (reuse path, no fork) | PASS | `grep -c 'spec.ListDocs' internal/epic/discover.go` | `1` (composed, not forked) + `grep -c 'os.WriteFile\|os.Create' internal/epic/*.go` → `0` |
| AC-ES-003 (prefix glob) | PASS | `go test ./internal/epic/ -run TestDiscoverEpic_PrefixGlob` | `--- PASS: TestDiscoverEpic_PrefixGlob (0.00s)` |
| AC-ES-003b (empty match clean) | PASS | `go test ./internal/epic/ -run TestBuildEpicStatus_EmptyPrefix` + CLI smoke `go run ./cmd/moai epic status NONEXIST --json` | exit 0, `{"epic":"NONEXIST",...,"total":0,"pct":0}` |
| AC-ES-004 (Mx extraction) | PASS | `go test ./internal/epic/ -run TestExtractMx_BASMarker` | `--- PASS: TestExtractMx_BASMarker (0.00s)` |
| AC-ES-004b (generic token) | PASS | `go test ./internal/epic/ -run TestExtractMx_GenericToken` | `--- PASS: TestExtractMx_GenericToken (0.00s)` |
| AC-ES-005 (orphan-Mx detection) | PASS | `go run ./cmd/moai epic status NAVIGATOR-SYNC --json` | `"orphan_mx": ["M2","M3","M5"]` + 3 absent milestones with `"covered": false` |
| AC-ES-005b (omit orphan when no report) | PASS | `go test ./internal/epic/ -run TestBuildEpicStatus_NoDesignReport_OmitsOrphan` | `--- PASS: ... OmitsOrphan (0.00s)` |
| AC-ES-006 (status join) | PASS | `go test ./internal/epic/ -run TestJoinStatus_Classification` | `--- PASS: TestJoinStatus_Classification (0.00s)` |
| AC-ES-007 (divide-by-zero) | PASS | `go test ./internal/epic/ -run TestComputePct_DivideByZero` | `--- PASS: TestComputePct_DivideByZero (0.00s)` |
| AC-ES-008 (JSON shape) | PASS | `go test ./internal/epic/ -run TestRenderJSON_FrozenShape` + CLI smoke | byte-stable JSON with all frozen keys present |
| AC-ES-008b (human grammar) | PASS | `go test ./internal/epic/ -run TestRenderHuman_ProgressBoardGrammar` + `go run ./cmd/moai epic status NAVIGATOR-SYNC` | first line `🎯 NAVIGATOR-SYNC ▓▓▓▓▓░░░░░ 3/6 (50%)` |
| AC-ES-009 (forward-compat) | PASS | `go test ./internal/epic/ -run TestRenderJSON_ForwardCompat` | `--- PASS: TestRenderJSON_ForwardCompat (0.00s)` |
| AC-ES-010 (CLI registration) | PASS | `go run ./cmd/moai epic --help` + `go run ./cmd/moai epic status --help` + `go run ./cmd/moai --help | grep epic` | status subcommand listed; flags --json/--design-report/--marker present; epic in root --help |
| AC-ES-011 (non-interactive) | PASS | `go test ./internal/cli/ -run TestNewEpicCmd_NoAskUserQuestion` | `--- PASS: TestNewEpicCmd_NoAskUserQuestion (0.00s)` |
| AC-ES-012 (template neutrality) | PASS | `grep -rn 'SPEC-EPIC-STATUS\|REQ-ES-' internal/template/templates/ 2>/dev/null` | (no matches — no mirror created; producer Go source not under templates/) |
| AC-ES-013 (no persisted store) | PASS | `go test ./internal/cli/ -run TestEpicStatusCmd_NoPersistedStore` + `grep -c 'os.WriteFile\|os.Create' internal/epic/*.go` | PASS; grep returns 0 |
| AC-ES-014 (factory touchpoint) | PASS | `grep -c 'moai epic status' internal/hook/session_start_factory.go` | `1` (single pointer line added to factoryLeadNotice) |
| AC-ES-015 (baseline attribution) | PASS | `go run ./cmd/moai epic status NAVIGATOR-SYNC --json | grep baseline_attribution` | `"baseline_attribution": "87c685b099b011b8d43d0e15b37ced14c2bccebb"` |

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-12
run_commit_sha: 87c685b099b011b8d43d0e15b37ced14c2bccebb   # worktree HEAD at run-phase completion
run_status: audit-ready
ac_pass_count: 18          # AC-ES-001..015 + AC-ES-003b + AC-ES-004b + AC-ES-005b + AC-ES-008b all PASS
ac_fail_count: 0
must_ac_count: 17          # 15 MUST + 2 MUST sub-IDs (003b, 004b, 005b, 008b — see acceptance.md §B)
should_ac_count: 1         # AC-ES-014 (SHOULD) — PASS (factory touchpoint shipped)
coverage_internal_epic: 87.6%   # go test -cover ./internal/epic/ (≥85% threshold met)
coverage_internal_cli: measured via full package suite (producer code paths covered by TestEpic* + TestNewEpicCmd_NoAskUserQuestion)
new_warnings_or_lints_introduced: 0   # golangci-lint run --timeout=2m: "0 issues" on touched packages
cross_platform_build:
  go_build_default: PASS   # exit 0
  goos_windows_goarch_amd64: PASS   # exit 0
total_run_phase_files: 11   # internal/epic/{discover,designreport,status,render,git}.go + internal/epic/{discover,designreport,status,render}_test.go + internal/cli/epic.go + internal/cli/epic_test.go + internal/hook/session_start_factory.go touchpoint + internal/cli/help.go help-row
m1_to_mN_commit_strategy: per-milestone Conventional Commits `feat(SPEC-EPIC-STATUS-001): M{N} <subject>` + 🗿 MoAI trailer; no push, no PR (per user instruction)
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-12
sync_commit_sha: 6da952899be8b2ceafe7bb87aa6af41b8c32dd0c   # self-referential-hazard workaround (spec-frontmatter-schema.md D3); backfilled in a follow-up commit
sync_status: audit-ready
run_commit_sha: 4aa9178c3                # final run-phase commit (M5 factory-notice touchpoint)
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed"   # merged close on this single sync commit (manager-docs owns the ride)
  plan_md: "(no status: field — plan.md carries no frontmatter)"
  acceptance_md: "(no status: field — acceptance.md carries no frontmatter)"
  progress_md: "(no frontmatter — §E.4 lives in the body, not in frontmatter)"
b12_self_test:
  a_pre_emission_grep: 0           # grep -c 'SPEC-EPIC-STATUS-001' CHANGELOG.md BEFORE edit → 0 (no duplicate; PASS)
  b_ac_count_match: 15             # acceptance.md §D AC-ES-001..015 distinct count == CHANGELOG-referenced AC count (15)
  c_file_path_verification: PASS   # ls internal/epic/{discover,designreport,status,render,git}.go + internal/cli/epic.go + internal/cli/epic_test.go + internal/hook/session_start_factory.go → all exist
changelog_entry_position: "top of [Unreleased] > Added (newest first — above the SPEC-FACTORY-BOOTSTRAP-001 entry landed by main #1447)"
canary_compliance_check:
  zero_go_files_in_sync_commit: true   # markdown-only — CHANGELOG.md + spec.md frontmatter + progress.md §E.4
  moai_gate_bypass_rationale: "0 .go files — sync markdown only (SKIP_MOAI_PRECOMMIT=1 if the pre-commit gate false-positives)"
  push_performed: false                # C5 user instruction — commit to feat/factory-bootstrap-guidance ONLY; no `git push`
  pr_created: false                    # C5 user instruction — no `gh pr create`
mx_tag_validation:
  present_count: 0                    # grep -rn '@MX:' internal/epic/ internal/cli/epic.go → 0 matches
  gaps: "BuildEpicStatus + DiscoverEpic fan_in = 1 (CLI only), below the @MX:ANCHOR threshold (≥3 callers); tags deferred to a future /moai mx scan when fan_in grows"
  status: not-warranted-at-current-fan-in
```

---

## §F Phase 4 Mode Selection

- **tier**: L
- **scope**: ~7 files (internal/epic/{discover,designreport,status,render}.go + internal/epic/*_test.go + internal/cli/epic.go + internal/cli/epic_test.go + internal/hook/session_start_factory.go touchpoint)
- **domain count**: 2 (Go producer library + CLI wiring)
- **file language mix**: 100% Go source
- **concurrency benefit**: LOW (coding-heavy, single-author producer library)

**Decision**: solo-sequential (Mode 5)

**Justification**: This is coding-heavy new-feature work (a producer library + thin CLI wiring) executed by a single manager-develop delegation in a worktree. Per Anthropic's coding-task parallelism caveat and the §B decision tree, Mode 5 (sequential sub-agent) is the correct default — the work has no genuinely-parallel independent tracks and no high-volume mechanical transform that would justify Mode 4 / Mode 6. Mode 3 is retired. Mode 1/2 do not apply (non-trivial, write-capable).

**Implementation Kickoff Approval**: obtained this session before run-phase entry (plan-audit iter-2 PASS 0.97, Tier L threshold 0.85 met).

---

## §H Recursive Self-Diagnosis Log

_<pending run-phase — populated only if the DIAPNOSE-PATCH-VERIFY loop fires on a mechanical failure during M0..M5.>_

---

## §I Token Accounting

_<pending sync-close — per-SPEC token-spend measurement populated at sync-close.>_
