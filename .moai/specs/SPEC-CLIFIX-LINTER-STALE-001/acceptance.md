# SPEC-CLIFIX-LINTER-STALE-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given an agent file declaring a hook whose script path does not exist, When the maintainer runs agentlint, Then LR-04 reports the dead hook (previously: silent pass forever).
2. Given the shipped 10-agent catalog and its template mirrors, When agentlint runs over the repository, Then the report contains no live/mirror duplicate noise and every retained agent is effort-checked.
3. Given a teammate claiming task IDs, When it claims a completed or unknown ID, Then the claim fails and the orchestrator sees an actionable error instead of a phantom success.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-LINT-001-001 | REQ-LINT-001-001 | `go test ./internal/cli/agentlint/ -run 'LR04\|DeadHook' -count=1 -v` && `grep -n 'yaml.Unmarshal\|yaml.v3' internal/cli/agentlint/agent_lint.go` | PASS — dead-hook fixture yields an LR-04 finding; parser uses yaml.v3; Hooks/Skills/Sandbox populated in parse test |
| AC-LINT-001-002 | REQ-LINT-001-002 | `go test ./internal/cli/agentlint/ -run 'EffortMatrixRoster' -count=1 -v` | PASS — all 10 retained agents covered; test fails if a retired name (e.g. manager-quality, expert-backend) remains in the matrix |
| AC-LINT-001-003 | REQ-LINT-001-003 | `go test ./internal/cli/agentlint/ -run 'LR07Dedupe' -count=1 -v` | PASS — live+mirror pair → 0 duplicate findings; genuine same-name duplicate fixture → 1 finding |
| AC-LINT-001-004 | REQ-LINT-001-004 | `go test ./internal/cli/ -run 'DoctorSkillsCatalog' -count=1 -v` && `grep -c 'moai-' internal/cli/doctor_skills.go` | PASS — allowlist built from catalog at runtime; static skill-name literals reduced to zero (or a documented bootstrap minimum); retired-skill fixture WARNs, catalog-live skill PASSes |
| AC-LINT-001-005 | REQ-LINT-001-005 | `! go run ./cmd/moai --help 2>&1 \| grep -q 'brain'` (negated grep — exit 0 when zero matches, safe for `&&` chaining) | Command succeeds (no `brain` entry); help-entry gate test asserts every advertised command is registered |
| AC-LINT-001-006 | REQ-LINT-001-006 | `go test ./internal/cli/ -run 'ClaimTaskValidate' -count=1 -v` | PASS — nonexistent ID → error, completed task → error, pending task → success; ledger line count unchanged on failures |
| AC-LINT-001-007 | REQ-LINT-001-007 | `go test ./internal/cli/agentlint/... ./internal/cli/ -run 'LR04\|EffortMatrixRoster\|LR07Dedupe\|DoctorSkillsCatalog' -count=1` | PASS — each gate has both a firing (positive) and non-firing (negative) case |

## §C Edge Cases

- Agent frontmatter with CSV `tools:` string vs YAML list — yaml.v3 parse must handle the project's documented CSV-string convention without false findings.
- Template mirror that is a sanitized (non-byte-identical) variant of the live agent — dedupe must still pair them (path-mapping key).
- Catalog file unreadable/corrupt — doctor_skills emits a loud diagnostic and non-zero doctor status, never a silent all-pass.
- Help output width/wrapping — the brain-absence grep must not be defeated by line wrapping (assert on full buffer).
- ClaimTask racing a concurrent claim of the same task — combined with the P0 O_APPEND fix, second claim must fail pending-validation or be detectably duplicate.

## §D.5 Quality Gate / Definition of Done

- All 7 AC rows PASS with verbatim command output cited in progress.md §E.2.
- Before/after agentlint run over the real repository recorded: LR-07 false-positive count drops to 0; any NEW true findings surfaced by the fixed parser are triaged in progress.md (fix-or-file follow-up), not suppressed.
- `go test ./internal/cli/... -count=1` green; `golangci-lint run` no new findings.
