# progress.md — SPEC-LLMCFG-PRESERVE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (plan-auditor PASS 1.00/1.00, 2026-09-02)
plan_complete_at: 2026-09-02
card: t239
branch: WT-llm-yaml-preserve
plan_authored_at_head: b7462203a

## §E.2 Run-phase Evidence

Baseline attribution for every item below: worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t239`, branch
`WT-llm-yaml-preserve`, pre-M1-commit HEAD `ffdd024eb` (measurements taken in
this run against this tree; evidence logs committed under `.moai/reports/t239/`).

### M1 — template-sync path contracts (AC-LCP-001..004)

New test file: `internal/cli/update_llm_preserve_test.go` (package cli).
Fixtures derive from the REAL embedded template llm.yaml
(`template.EmbeddedTemplates`); every fixture edit uses must-replace semantics
(a template drift fails the fixture builder instead of producing a vacuous
green). Update cycle driven via the production seam
`runTemplateSyncWithReporter` (chdir fixture), the same driver the existing
home-seam and mirror-notice tests use.

**E2/M1 scoped GREEN** — command: `go test -run 'TestUpdateLLMYAML' -count=1
./internal/cli/`; observed output (verbatim): `ok
github.com/modu-ai/moai-adk/internal/cli	1.746s`; exit 0.
Evidence: `.moai/reports/t239/green-m1-all-contracts.log`.

**E3/M1 mutant teeth (RED observed, scratch only — not committed)**:

- AC-LCP-001 — mutation: restore step bypassed in
  `internal/cli/update_template_sync.go` (`if false && configBackupPath != ""`
  scratch edit). Command: `go test -run
  'TestUpdateLLMYAMLPreserveTemplateSync' -count=1 ./internal/cli/`; exit 1;
  verbatim RED lines: `llm.glm.models.high = "glm-5.3-flash"; want user pin
  "glm-user-pinned-model" (template default reset the user's value?)` and
  `llm.yaml: key "manager-develop" missing (path [llm agent_overrides
  manager-develop model])`. Evidence:
  `.moai/reports/t239/mutant-red-ac-lcp-001-restore-bypass.log`. Reverted;
  GREEN re-observed: `.moai/reports/t239/green-after-restore-ac-lcp-001.log`
  (exit 0, `ok ... 1.094s`).
- AC-LCP-002 — merge-function-level probe (scratch): `DeepMerge3Way(new
  missing performance_tier, old missing performance_tier, base)` → merged
  contains `performance_tier:` → **false** (the e2e delivery assertion is
  sensitive to the delivery path; a fixture whose key was never removed would
  not deliver and the assertion would fail). Evidence:
  `.moai/reports/t239/mutant-red-ac-lcp-002-004-merge-level.log`.
- AC-LCP-003 — mutation: deploy walk skips
  `.moai/config/sections/llm.yaml` (scratch edit in
  `internal/template/deployer.go`). Command: `go test -run
  'TestUpdateLLMYAMLFirstDeployCalm' -count=1 ./internal/cli/`; exit 1;
  verbatim RED line: `llm.yaml absent after update on a project that had
  none: open /var/folders/.../TestUpdateLLMYAMLFirstDeployCalm.../llm.yaml:
  no such file or directory`. Evidence:
  `.moai/reports/t239/mutant-red-ac-lcp-003-deploy-skip.log`. Reverted.
- AC-LCP-004 — merge-function-level probe (scratch): real template bytes
  through a `map[string]any` round trip (pre-#1243 defect class) →
  `Profile matrix` sentinel kept → **false**, `GLM reasoning-effort mapping`
  sentinel kept → **false** (the comment assertions detect comment loss).
  Evidence: `.moai/reports/t239/mutant-red-ac-lcp-002-004-merge-level.log`.

**Design observation recorded (fixture placement)**: the node merge walks the
NEW template's key nodes, so a user comment attached to a key the template
carries is replaced by the template's own comment (measured via scratch
probe, 2026-09-02). The AC-LCP-001 marker therefore rides a USER-ADDED key
(`t239_user_marker_key`), which the merge retains through its old-only pass —
this is the documented retention semantics (REQ-UYP-006/007), not a defect;
no repair named.

### M2 — clean-reinstall parity (AC-LCP-005)

Extended the `update_clean_install_config_preserve_test.go` fixture family:
`makeConfigPreserveFixture` now seeds llm.yaml with the SAME user-edit pattern
(value re-pin + agent_overrides entry + marker comment above a user-added
key), and `overwritingDeployer.Deploy` clobbers llm.yaml with the REAL
embedded template bytes (read via `readEmbeddedLLMYAML`, error-propagating) —
so survival proves the restore's 3-way merge, not a clobber that never
happened. The five existing family tests run unchanged (verified below with
the widened fixture + clobber in place).

**E2/M2 scoped GREEN** — commands and observed outputs (verbatim):

- `go test -run 'TestCleanReinstallLLMYAMLPreserved' -count=1 ./internal/cli/`
  → `ok github.com/modu-ai/moai-adk/internal/cli	1.002s`, exit 0
  (`.moai/reports/t239/green-m2-clean-reinstall.log`).
- Family re-run with the widened fixture: `go test -run 'TestCleanReinstall'
  -count=1 ./internal/cli/` → `ok
  github.com/modu-ai/moai-adk/internal/cli	1.106s`, exit 0
  (`.moai/reports/t239/green-m2-clean-reinstall-family.log`).

**E3/M2 mutant teeth** — mutation: clean-reinstall restore step bypassed in
`internal/cli/update_clean_install.go` (`if false && configBackupPath != ""`
scratch edit). Command: `go test -run 'TestCleanReinstallLLMYAMLPreserved'
-count=1 ./internal/cli/`; exit 1; verbatim RED line: `clean-reinstall lost
user llm.glm.models.high = "glm-5.3-flash"; want "glm-user-pinned-model"
(clobbered to template default?)`. Evidence:
`.moai/reports/t239/mutant-red-ac-lcp-005-clean-restore-bypass.log`.
Reverted; GREEN re-observed across the whole family.

### M3 — negative constraint + verification batch (AC-LCP-006 + E1-E6)

**AC-LCP-006 (negative guard, grep form per acceptance.md)** — commands and
observed outputs (verbatim):

- `grep -rnE "\.moai-[a-z-]+\.sha256" internal/cli/update/ --include="*.go" |
  grep -v _test | wc -l` → `0` (`.moai/reports/t239/ac-lcp-006-update-scope-zero.log`).
- `grep -rlnE "\.moai-[a-z-]+\.sha256" internal/cli/ --include="*.go" | grep
  -v _test` → `internal/cli/hook_install.go`,
  `internal/cli/hook_install_precommit.go` — hook tier files only
  (`.moai/reports/t239/ac-lcp-006-cli-scope-files.log`).
- Teeth (pattern non-vacuity): the same scan that returns 0 in the update
  scope matches the hook tier's own constant
  `moaiPreCommitProvenanceName = ".moai-pre-commit.sha256"`
  (`internal/cli/hook_install_precommit.go:19`).

**E1 final re-measurement** — `go test -run
'TestUpdateLLMYAML|TestCleanReinstallLLMYAMLPreserved' -count=1
./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli	2.121s`,
exit 0 (`.moai/reports/t239/e1-final-all-contracts.log`).

**E2 scoped packages** — `go test -count=1 ./internal/cli/update/backup/
./internal/cli/update/merge/` → `ok ... update/backup 0.624s` + `ok ...
update/merge 0.774s`, exit 0 (`.moai/reports/t239/e2-scoped-packages.log`).

**E4 vet + lint** — `go vet ./internal/cli/...` exit 0, no output
(`.moai/reports/t239/e4-vet.log`); `golangci-lint run --timeout=2m
./internal/cli/...` → `0 issues.`, exit 0 (`.moai/reports/t239/e4-lint-cli.log`)
— 0 NEW findings vs the plan §C lint baseline.

**E5 cross-platform** — `GOOS=windows GOARCH=amd64 go build ./...` exit 0
(`.moai/reports/t239/e5-windows-build.log`); `GOOS=windows GOARCH=amd64 go vet
./internal/cli/` exit 0 (test files compile cross-platform,
`.moai/reports/t239/e5-windows-vet-cli.log`).

**E6 PRESERVE-list adherence** — `git diff --stat ffdd024eb..HEAD` over the
§A.5 paths (`internal/cli/update/backup/`, `internal/cli/update/merge/`,
`internal/cli/update/deploy/deploy.go`,
`internal/template/templates/.moai/config/sections/llm.yaml`,
`internal/cli/hook_install_precommit.go`, `internal/settings/`,
`.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/`) is EMPTY — no §A.5 file
modified (`.moai/reports/t239/e6-preserve-adherence-head-state.txt`). The
change set is exactly: `internal/cli/update_llm_preserve_test.go` (new),
`internal/cli/update_clean_install_config_preserve_test.go` (fixture family
extension, M2-scope per plan §D), SPEC artifacts, and evidence logs.

**Known minor defect D1 handoff (manager-spec ownership, noted not fixed)**:
spec.md §A.2 cites `update_template_sync.go` without its package path — the
file lives at `internal/cli/update_template_sync.go`. Same for §A.2's bare
`update_clean_install.go` (`internal/cli/update_clean_install.go`). Left
untouched per the ownership boundary; sync-phase should relay to manager-spec
if a body correction round opens.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-09-02
run_commit_sha: pending-backfill-final-commit (this §E.3 block ships in the
final M3 commit; the commit SHA cannot reference itself)
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 7
l44_pre_commit_fetch: not-applicable (lane does not push; develop integration
is the lead's window)
l44_post_push_fetch: not-applicable (no push from this lane)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin_arm64: exit 0 (native run of all scoped tests)
cross_platform_build.windows_amd64: exit 0 (go build ./... + go vet
./internal/cli/)
total_run_phase_files: 3 (2 test files + 5 evidence log groups under
.moai/reports/t239/, plus SPEC artifacts)
m1_to_mN_commit_strategy: one commit per milestone (M1 8bbd10a67, M2
5b5561254, M3 = this commit) — contract-value-first ordering per plan §F
with teeth evidence captured before each GREEN counted.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
