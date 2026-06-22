# Progress — SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (standard). Artifact set authored: spec.md, plan.md, acceptance.md,
  research.md, progress.md.
- Status: `draft`. SPEC ID self-check: `SPEC ✓ | STEERING ✓ | ALIGN ✓ |
  GUARDRAIL ✓ | HOOK ✓ | 001 ✓ → PASS` (canonical regex
  `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- Requirements: 12 GEARS requirements (REQ-GH-001..012); Out of Scope section
  present with 4 `### Out of Scope —` H3 sub-headings.
- Load-bearing decisions resolved in plan.md §D + research.md: removal mechanism
  (`**/glm-web-tooling.md` self-glob — verified against loader behavior + live-path
  matching domain), injection point (SessionStart — once-per-session), payload
  (concise reminder, not full rule).
- Grounding facts tool-verified this session: live always-load = 11, template = 10
  (post-removal 10 / 9); `glm-web-tooling.md` byte-identical 7545 B both trees, no
  frontmatter; PROCESS-env `z.ai` signal reused from `cg_detect.go` `hasGLMEnv`
  (line 202); injection precedent `detectWorkflowContext` live; hook package
  coverage baseline 82.7%.
- iter-2 audit revision (plan-auditor iter-1 PASS-WITH-DEBT 0.83, Tier M bar 0.80;
  iter-1 carried a stale Tier-L 0.85 bar because `tier:` was absent — D7 fixed):
  D1 (MAJOR, two-detector hazard — hook pinned to PROCESS-env `hasGLMEnv`, NOT
  `sessionEnvHasGLM`/`IsCGMode`; AC-GH-009b/009c added); D2 (REQ-SARS-010
  reconciliation clause in §A); D3 (`**/glm-web-tooling.md` self-glob, live-path
  only, no template-tree path); D4 (AC-GH-005b behavioral load-on-touch parity);
  D5 (z.ai-vs-api.z.ai substring superset note); D6 (coverage rescoped to NEW code
  path 90%+ / package baseline ≥ 82.7%, not package-wide 90%); D7 (`tier: M`
  frontmatter added); MP-2 (REQ-GH-012 modality (Event-detected)→(Event-driven)).
  All 8 re-verified against actual source per verification-claim-integrity.md.
- Pending: plan-auditor iter-2 re-audit, then 구현 착수 승인 (plan-to-implement
  human gate) before run-phase entry.

## §E.2 Run-phase Evidence

TDD cycle: M2 RED (8 tests, `glmGuardrailReminder` undefined → build fail confirmed) →
M3 GREEN (`hookProcessEnvHasGLM` + `glmGuardrailReminder` in
`internal/hook/session_start_glm_guardrail.go`, wired into `session_start.go`
`Handle()` AdditionalContext) → M4 Template-First always-load removal → M5
REFACTOR + coverage → M6 acceptance sweep.

| AC | Severity | Status | Verification command | Actual output |
|----|----------|--------|----------------------|---------------|
| AC-GH-001 | baseline | PASS | `find .claude/rules/moai -name "*.md" -type f` per-file `sed 1,8p \| grep '^paths:'` count-WITHOUT — BEFORE M4 | `live=11` |
| AC-GH-002 | MUST-FIX | PASS | same command AFTER M4 | `live=10` |
| AC-GH-003 | baseline | PASS | same find over `internal/template/templates/.claude/rules/moai` — BEFORE M4 | `template=10` |
| AC-GH-004 | MUST-FIX | PASS | same template command AFTER M4 | `template=9` |
| AC-GH-005 | MUST-FIX | PASS | `grep -c '^paths: "\*\*/glm-web-tooling.md"' <live>; <template>` | `1` then `1` |
| AC-GH-005b | MUST-FIX | PASS | basename-vs-extracted-glob case match | `LOADS-ON-TOUCH` |
| AC-GH-006 | MUST-FIX | PASS | `diff <live> <template> && echo PARITY-OK` | `PARITY-OK` (no diff) |
| AC-GH-007 | SHOULD-FIX | PASS | `grep -rn 'z\.ai' internal/hook/*.go \| grep -v '_test.go'` | reuses `z.ai` substring (const `glmProcessEnvSubstring`); no second divergent literal |
| AC-GH-008 | MUST-FIX | PASS | `go test ./internal/hook/ -run 'GLM.*Guardrail\|GuardrailReminder'` | `ok ... 0.272s` — asserts `web_search_prime`, `web_reader`, `zai-mcp-server`, `ToolSearch` present |
| AC-GH-009 | MUST-FIX | PASS | `go test ./internal/hook/ -run 'GLM.*Guardrail.*NonGLM\|GuardrailReminder.*Absent'` | `--- PASS: TestGLMGuardrailReminder_NonGLMSession` |
| AC-GH-009b | MUST-FIX (D1 carve-out) | PASS | `go test ./internal/hook/ -run 'GLM.*Guardrail.*Leader\|GuardrailReminder.*CgLeader'` | `--- PASS: TestGLMGuardrailReminder_CgLeader` (PROCESS env clean + tmux SESSION GLM marker → no injection) |
| AC-GH-009c | MUST-FIX | PASS | `grep -n 'sessionEnvHasGLM\|IsCGMode' internal/hook/*.go \| grep -v '_test.go'` | (no output) — hook uses PROCESS-env check only |
| AC-GH-010 | SHOULD-FIX | PASS | `go test ./internal/hook/ -run 'GLM.*Guardrail.*Error\|GuardrailReminder.*AllowOnError'` | `--- PASS: TestGLMGuardrailReminder_AllowOnError` |
| AC-GH-011 | MUST-FIX | PASS | `go tool cover -func` grep `glm.*[Gg]uardrail\|[Gg]uardrail.*[Rr]eminder` + package `-cover` | `hookProcessEnvHasGLM 100.0%`, `glmGuardrailReminder 100.0%` (≥90%); package `coverage: 82.7%` (not below 82.7% baseline) |
| AC-GH-012 | MUST-FIX | PASS | `go test ./... && go vet ./... && golangci-lint run --timeout=2m` | full suite exit 0 (TestHookWrapper_* baseline timeout-flake under whole-module CPU contention — PASS in isolation + clean package re-run); `go vet` exit 0; golangci-lint `0 issues.` |
| AC-GH-013 | MUST-FIX | PASS | `go test ./internal/template/... -run TestTemplateNeutralityAudit` | `ok ... 0.352s` |
| AC-GH-014 | SHOULD-FIX | PASS-WITH-NOTE | `make build` + embedded FS read | embedded `glm-web-tooling.md` carries `paths:` frontmatter (verified via `EmbeddedTemplates()` FS read). NOTE: this repo uses live `//go:embed all:templates` (no generated `internal/template/embedded.go`); the AC's literal `embedded.go` reference is a plan-phase placeholder — the template edit IS the embedded content, re-embedded at `make build` compile time. `catalog.yaml` ran through `make build` but is byte-unchanged (rule frontmatter does not affect the skill-catalog hash), so it is NOT staged. |

Cross-platform build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64
go build ./...` not required (no syscall; pure env-read + string). Subagent
boundary (C-HRA-008): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/
| grep -v _test.go` — no new matches (the hook is advisory, no user interaction).

PRESERVE: zero diff to the `glm-web-tooling.md` rule BODY (frontmatter-only edit);
zero modification to `cg_detect.go`, the `detectWorkflowContext` precedent, the UUID
attribution injection, or any unrelated `internal/hook` code.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-23
run_commit_sha: <backfill-after-commit>
run_status: implemented
ac_pass_count: 17
ac_fail_count: 0
preserve_list_post_run_count: 0  # zero PRESERVE-list violations (rule body untouched, cg_detect.go untouched)
l44_pre_commit_fetch: pending  # git fetch origin main executed immediately before commit (race mitigation)
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: 0  # golangci-lint "0 issues."
cross_platform_build:
  linux_amd64: pass  # go build ./... exit 0
  windows_amd64: n-a  # no syscall — pure env-read + string concat
total_run_phase_files: 5  # session_start_glm_guardrail.go (new) + _test.go (new) + session_start.go (edit) + glm-web-tooling.md (live, edit) + glm-web-tooling.md (template, edit). catalog.yaml byte-unchanged by make build (rule frontmatter does not affect catalog hash); embedded.go absent (live go:embed).
m1_to_mN_commit_strategy: single-commit  # Tier M, one coherent feature — single run-phase commit
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
