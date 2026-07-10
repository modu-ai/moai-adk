# SPEC-WEB-CONSOLE-012 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-10
artifacts: spec.md, plan.md, acceptance.md, progress.md (4)

## §E.2 Run-phase Evidence

Run-phase 2026-07-10, cycle_type=tdd. Base: `90f394685` (worktree fast-forwarded
to main HEAD). Milestone commits: M1 `cc1edd0ca`, M2 `1807197b1`,
M3 (NO-TOUCH gate — no diff by design), M4 `6e7f88cef`, M5 `4fba896db`.
Evidence logs (verbatim outputs, persisted to survive worktree disposal):
`/Users/goos/MoAI/moai-adk-go/.moai/state/verify/e07c0351/` (preflight-build,
preflight-lint-baseline, m1/m2/m4-fulltest, final-vet/lint/coverage/fulltest/makebuild).

### AC PASS/FAIL Matrix (SSOT: acceptance.md §D.1)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-WC12-001 | PASS | `sed -n '/^func llmFields/,/^}/p' internal/settings/schema_sections.go \| grep -c 'opus\|sonnet\|haiku'` | `0` |
| AC-WC12-002 | PASS | same window `grep -c 'fable'` | `1` (>= 1) |
| AC-WC12-003 | PASS | `sed -n '/^func applyLLMKey/,/^}/p' internal/settings/sectionapply.go \| grep -c 'glm.models.fable'` | `1` |
| AC-WC12-004 | PASS | same window `grep -c 'glm.models.opus\|glm.models.sonnet\|glm.models.haiku'` | `0` |
| AC-WC12-005 | PASS | fable.title count == high.title count in i18n.js | `PARITY` (4 == 4, sibling-derived) |
| AC-WC12-006 | PASS | `grep -c 'f\.llm\.glm\.models\.opus\|...sonnet\|...haiku' internal/web/assets/i18n.js` | `0` |
| AC-WC12-007 | PASS | `grep -c 'opus/sonnet/haiku' internal/web/schemaform.go` | `0` |
| AC-WC12-008 | PASS | `grep -c '"research"' sectionroute.go; sectionwrite.go` | `0` / `0` |
| AC-WC12-009 | PASS | `grep -c 'SectionResearch' schema.go; grep -rn ... \| wc -l` (`;` split per 0.2.0 D4) | `0`; `0` |
| AC-WC12-010 | PASS | `go test ./internal/settings/` (new `TestWriteSectionViaSeamRejectsResearchPreservesFile`: error + research.yaml byte-unchanged; RED-verified pre-M1) | `ok internal/settings 0.795s` |
| AC-WC12-011 | PASS | 3× FieldDef greps in schema.go + i18n grep (retention, 0.2.0 inverted) | `2` / `2` / `2` / `24` (all >= 1) |
| AC-WC12-012 | PASS | `grep -rc 'min_coverage_per_commit' internal/settings/*.go \| grep -v :0` AND `validation.enforce_on_push` 동형 | `schema.go:4` (+tests) / `schema.go:3` (+test) |
| AC-WC12-013 | PASS | struct + fallback preservation greps (REQ-WC12-006) | `1` / `1` / `1` (`AutoDetection AutoDetectionConfig`, `Opus   string`, `func resolveGLMModels`) |
| AC-WC12-014 | PASS | `grep -c '"auto_detection"' internal/settings/schema_sections.go` | `1` (harness field untouched) |
| AC-WC12-015 | PASS | Where-gate PASSED branch: `grep -rc 'errDictKey' internal/web/ \| grep -v ':0'` | (no output — sentinel + keep-alive removed) |
| AC-WC12-016 | PASS | `grep -rn 'WorkflowAgentPurposes' --include='*.go' internal/ cmd/ pkg/` | (no matches) |
| AC-WC12-017 | PASS | `grep -c '10 user-facing' server.go` → `0`; `grep -c '10개' projectconfig.go` → `0` | `0` / `0`; manual: research/db appear ONLY in excluded-set enumerations (server.go:24 "delisted sections db ... and research"; projectconfig.go:164 "폐선 섹션 db·research ... 쓰기 금지") |
| AC-WC12-018 | PASS | `go test ./internal/settings/... ./internal/web/... ./internal/cli/` + explicit `go test ./internal/cli/ -run TestI18nKeySetParity -v` (both surfaces) | all `ok`; `--- PASS: TestI18nKeySetParity` (cli 0.544s) + `--- PASS: TestI18nKeySetParity` (web 0.282s) |
| AC-WC12-019 | PASS | `git diff --name-only 90f394685..HEAD \| grep -c 'internal/statusline/\|internal/template/templates/'` | `0` (16 files, all in-scope) |
| AC-WC12-020 | PASS-WITH-DEBT | `go build ./... && GOOS=windows GOARCH=amd64 go build ./... && go vet ./... && go test ./...` | build/winbuild/vet all exit 0; `go test ./...` exit 0 observed on M2+M4 full runs (93 ok); final `-count=1` run hit 1 PRE-EXISTING flake outside scope: `internal/hook TestHookWrapper_TempFileCleanup` (passes in isolation 3/3; documented observed-flaky in base commit `90f394685` subject) |

### Invariant rows

| Invariant | Status | Evidence |
|-----------|--------|----------|
| REQ-WC12-006 (GLMModels legacy struct + resolveGLMModels fallback + backfill untouched) | PASS | AC-WC12-013 greps + `git diff 90f394685..HEAD` shows no `internal/config/` or `internal/cli/glm.go` change; EC-2 roundtrip `TestLLMTypedSavePreservesLegacyGhostKeys` PASS |
| REQ-WC12-020/021/023 (A5 5-field retention, harness field untouched) | PASS | M3 gate: §D.2 field-dot + bare-symbol pair re-run 2026-07-10 — classification unchanged (5/5 USED; consumers trust.go:788, hook_pre_push.go:251/258, hook_pre_push.go:146→:234 whole-struct bind→LoadConvention manager.go:46) |
| REQ-WC12-050 (both-surface schema regression guard) | PASS | web i18n parity + TUI bridge tests green in the SAME M2 commit (AC-WC12-018) |
| REQ-WC12-051 (statusline/** + template/templates/** untouched) | PASS | AC-WC12-019 = 0 |
| Existing test suite never broken | PASS | full `go test ./...` after every milestone (m1/m2/m4-fulltest.log; single failure each in m1/final = the pre-existing hook flake, in-isolation PASS) |

### M3 NO-TOUCH gate re-run (verification-claim-integrity §1.1 surface 3)

Re-executed 2026-07-10 (run-phase), identical classification to acceptance.md §D.2:

- field-dot `\.AutoDetection\.` (historical false-negative pattern): no output — as documented, cannot match the whole-struct bind
- bare-symbol `AutoDetection\|AutoDetectOptions` over hook_pre_push.go + internal/git/convention/: 14 lines incl. `hook_pre_push.go:146` `opts, maxLength := resolveAutoDetectOptions()`, `:234` `ad := cfg.GitConvention.AutoDetection`, `manager.go:46` `func (m *Manager) LoadConvention(name string, opts AutoDetectOptions) error`
- `MinCoveragePerCommit` behavioral reader: `internal/core/quality/trust.go:788`
- `EnforceOnPush` behavioral reader: `internal/cli/hook_pre_push.go:251` + `:258`

→ 5/5 fields remain USED; zero schema/i18n/bridge changes made in M3 (code-untouched by design).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-10
run_commit_sha: pending-backfill (last code commit: 4fba896db; M1 cc1edd0ca / M2 1807197b1 / M4 6e7f88cef / M5 4fba896db — this progress.md commit cannot self-reference)
run_status: audit-ready
ac_pass_count: 20   # 19 PASS + 1 PASS-WITH-DEBT (AC-WC12-020 — pre-existing internal/hook flake outside scope)
ac_fail_count: 0
preserve_list_post_run_count: 0 violations (plan §D PRESERVE list intact — internal/config/*, internal/cli/glm.go, internal/cli/hook_pre_push.go, internal/core/quality/trust.go, internal/statusline/**, internal/template/templates/**, live .moai/config/sections/*.yaml all untouched)
l44_pre_commit_fetch: worktree fast-forwarded to main HEAD 90f394685 before M1 (git merge --ff-only main; 0 own commits, clean baseline)
l44_post_push_fetch: n/a — NO push performed (orchestrator verifies and pushes per delegation contract)
new_warnings_or_lints_introduced: 0 (golangci-lint baseline "0 issues." -> final "0 issues.")
cross_platform_build:
  darwin: exit 0
  windows_amd64: exit 0
  make_build: exit 0 (catalog.yaml regen produced no working-tree drift)
total_run_phase_files: 16 (+ this progress.md)
m1_to_mN_commit_strategy: per-milestone specific-path commits on worktree branch worktree-agent-a35fa4a27516c8ee9 (M3 is a NO-TOUCH verification gate — no commit by design); TDD RED-first on M1/M2 behavior changes; frontmatter draft->in-progress rode M1
coverage:
  internal/settings: 88.6%
  internal/settings/agentfm: 85.3%
  internal/settings/yamlpatch: 86.3%
  internal/web: 69.8% (baseline-identical — measured 69.8% at base 90f394685 via detached worktree; net-negative dead-code removal, no regression)
evidence_dir: /Users/goos/MoAI/moai-adk-go/.moai/state/verify/e07c0351/
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-10
sync_commit_sha: 333203c32
sync_status: audit-ready
changelog_entry_position: CHANGELOG.md [Unreleased] § Changed, first entry (prepended above SPEC-MEMORY-DIET-001)
frontmatter_status_transitions:
  spec.md: "in-progress -> completed (updated: 2026-07-10, unchanged date — same-day plan/run/sync)"
  plan.md: "no frontmatter (Tier S convention — no status field present)"
  acceptance.md: "no frontmatter (Tier S convention — no status field present)"
readme_check: "no user-visible mention of moai web console research-seam / llm.glm tier config sections found in README.md — grep -n 'moai web console\\|research section\\|settings.*section' returned no matches; README left unchanged"
b12_self_test_a: "grep -c SPEC-WEB-CONSOLE-012 CHANGELOG.md (pre-emission) = 0 -> proceeded with emission"
b12_self_test_b: "acceptance.md AC row count = 20 (grep -cE '^\\| AC-WC12-[0-9]+ \\|') matches CHANGELOG claim 19 PASS + 1 PASS-WITH-DEBT = 20 total"
b12_self_test_c: "file paths cited in CHANGELOG entry verified via ls: internal/settings/{schema_sections,sectionapply,sectionroute,sectionwrite,schema}.go, internal/web/{assets/i18n.js,schemaform.go} — all exist"
```
