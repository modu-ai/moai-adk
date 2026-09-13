# acceptance.md — SPEC-GITSTRAT-WORKFLOW-READER-001

All ACs are binary-testable: each names the command whose exit code / observed output decides PASS.

## §D AC Matrix

| AC | Requirement | Given | When | Then (machine-verifiable) |
|---|---|---|---|---|
| AC-GWS-001 | REQ-GWS-001 | The characterization baseline is green on the unmodified tree | The allowed-set validation is implemented | `go test ./internal/config/ -run TestWorkflowDisposition -count=1` exits 0 with all 4 allowed values classified and ≥1 invalid value classified |
| AC-GWS-002 | REQ-GWS-002 | git-strategy.yaml has `workflow: git-flwo` (typo fixture) | `LoadGitFlowIntegrationConfig` runs | Test asserts disposition == invalid, carried raw value == "git-flwo", `IsGitFlow()` == false; `go test` exit 0 |
| AC-GWS-003 | REQ-GWS-002 | git-strategy.yaml has `workflow: trunk-based` | reader runs | disposition == invalid (exclusion enforced); test exit 0 |
| AC-GWS-004 | REQ-GWS-003 | invalid-disposition fixture project | `moai integration acquire` runs | stderr contains the offending value AND one of the 4 allowed entries; exit code and lock record identical to the pre-change non-git-flow path (fixture test asserts both) |
| AC-GWS-005 | REQ-GWS-004 | flow = github-flow | target resolution runs | resolved target == "main" without reading any key; test exit 0 |
| AC-GWS-006 | REQ-GWS-004/005 | flow = git-flow, develop_branch = develop | target resolution runs | resolved target == develop; `develop_branch` still gated on Manual && GitFlowWorkflow (M1 characterization tests pass unmodified) |
| AC-GWS-007 | REQ-GWS-004/006 | flow = gitlab-flow, environment key empty | target resolution runs | target == "" (caller-fallback neutral); test exit 0 |
| AC-GWS-008 | REQ-GWS-004/006 | flow = release-flow, release_branch_prefix = release/ | target resolution runs | target == "release/"; test exit 0 |
| AC-GWS-009 | REQ-GWS-007 | — | run phase starts | `git log --oneline` in the SPEC branch shows the M1 characterization commit BEFORE the first production (M2) commit; ordering clause of verification-claim-integrity §2.3 satisfied |
| AC-GWS-010 | REQ-GWS-007 | M1 suite committed | after M2+M3 land | `go test ./internal/config/ -run TestLoadGitFlowIntegrationConfig -count=1` passes with the M1 test file byte-unmodified (`git diff M1..HEAD -- internal/config/loader_integration_branch_test.go` shows no M1-test edits) |
| AC-GWS-011 | REQ-GWS-008 | template + defaults | grep + test | `grep -c 'workflow: github-flow' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` == 3 AND a github-flow fixture yields identical acquire/auto-merge observable behavior pre/post change (characterization assertion) |
| AC-GWS-012 | NFR | — | final gate | `go test -cover ./internal/config/` output shows ≥ 85.0% coverage AND `golangci-lint run internal/config/... internal/cli/...` exits 0 |

## §D.1 Severity

- Must-pass (reject on FAIL): AC-GWS-001..006, 009, 010, 012.
- Should-pass: AC-GWS-007, 008, 011.

## §D.2 Traceability

REQ-GWS-001→AC-001; 002→AC-002/003; 003→AC-004; 004→AC-005/006/007/008; 005→AC-006; 006→AC-007/008; 007→AC-009/010; 008→AC-011. NFR→AC-012.

## §D.3 Edge cases

- Whitespace-padded develop_branch / environment values → trimmed (t449 trim contract generalized to new target keys).
- Unparseable YAML → zero-value struct (invalid-unknown, not "invalid value present") — doctor distinguishes "file unreadable" from "value invalid".
- mode != manual with a git-flow workflow → not git-flow (existing t449 comment contract preserved and characterized).

## §D.4 Indirect verification

- Doctor item verified through its check function's table test rather than a live `moai doctor` run in CI; one live run captured as §E.2 evidence.

## §D.5 Definition of Done

All must-pass ACs PASS with verbatim outputs persisted under §E.2; characterization ordering evidenced by git history; no regression in t637 coverage baseline; commit subjects follow Conventional Commits with `🗿 MoAI` trailer.
