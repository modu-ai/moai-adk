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

_<pending>_

### M3 — negative constraint + verification batch (AC-LCP-006 + E1-E6)

_<pending>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
