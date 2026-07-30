# SPEC-CLIFIX-LINTER-STALE-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then — test scenarios, NOT GEARS requirements)

> Per plan-audit D7: the §A scenarios and the §D AC matrix below are Given/When/Then test scenarios that map to GEARS REQs in spec.md §B, NOT GEARS requirements themselves. The GEARS requirement surface lives in spec.md §B; this file operationalizes each REQ as an observable, machine-verifiable test scenario.

1. Given an agent fixture declaring a hook whose script path does not exist, When the maintainer runs agentlint, Then LR-04 reports the dead hook (previously: silent pass forever because the hand-rolled parser never populated Hooks).
2. Given the shipped 10-agent catalog and its template mirrors, When agentlint runs over the repository, Then the report contains no live/mirror duplicate LR-07 noise and the `writeHeavyAgents` check fires on a retained write-heavy agent missing `isolation: worktree` but never on an archived name.
3. Given a teammate claiming task IDs, When it claims a completed or unknown ID, Then the claim fails and the orchestrator sees an actionable error instead of a phantom success.

## §D AC Matrix (machine-verifiable)

> Command-shape note: the agentlint CLI was extracted to a subcommand of `agent` during SPEC-CLI-SUBPKG-SPLIT-001 — the invocation is `moai agent lint` (parent `agent` + sub `lint`), NOT the legacy `moai agentlint`. Each AC below uses the correct shape.

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-LINT-001-001 | REQ-LINT-001-001 | `go test ./internal/cli/agentlint/ -run 'LR04\|DeadHook\|ParseFrontmatter' -count=1 -v` && `grep -n 'yaml.Unmarshal\|yaml\.v3\|"gopkg.in/yaml' internal/cli/agentlint/agent_lint.go` | PASS — dead-hook fixture yields an LR-04 finding; the negative fixture (no dead hooks) yields zero LR-04; parser source contains a yaml.v3 unmarshal call replacing the line-by-line `setField` stubs at `agent_lint.go:1229-1234`; Hooks/Skills/Sandbox populated in the parse test |
| AC-LINT-001-002 | REQ-LINT-001-002 | `go test ./internal/cli/agentlint/ -run 'WriteHeavyAgents\|LR05' -count=1 -v` && `! grep -n 'expert-backend\|expert-frontend\|expert-refactoring\|researcher' internal/cli/agentlint/agent_lint.go` | PASS — a fixture agent named `manager-develop` without `isolation: worktree` fires LR-05; an archived-name fixture (e.g. `expert-backend`) does NOT trigger the writeHeavyAgents branch; the 4 archived names are absent from the slice at `agent_lint.go:774-777` (negated grep — exit 0 when zero matches) |
| AC-LINT-001-003 | REQ-LINT-001-003 | `go test ./internal/cli/agentlint/ -run 'LR07\|DuplicateMandate\|LiveMirror' -count=1 -v` | PASS — live+mirror fixture pair → 0 LR-07 findings; genuine same-name duplicate fixture (two distinct live paths, same Skeptical-Evaluator Mandate block) → 1 finding |
| AC-LINT-001-005 | REQ-LINT-001-005 | `! go run ./cmd/moai --help 2>&1 \| grep -q 'brain'` (negated grep — exit 0 when zero matches, safe for `&&` chaining) && `go test ./internal/cli/ -run 'HelpRegisteredCommands\|NoPhantomBrain' -count=1 -v` | Command succeeds (no `brain` entry); the help-entry gate test asserts every advertised command resolves to a registered cobra command |
| AC-LINT-001-006 | REQ-LINT-001-006 | `go test ./internal/cli/taskledger/ -run 'ClaimTaskValidate\|ClaimTaskPending' -count=1 -v` | PASS — nonexistent ID → error, completed task → error, pending task → success; ledger line count unchanged on the two failure cases (verified by comparing `tasklist.md` byte count pre/post via `wc -l`) |
| AC-LINT-001-007 | REQ-LINT-001-007 | `go test ./internal/cli/agentlint/... ./internal/cli/taskledger/ -run 'LR04\|WriteHeavyAgents\|LR07\|ClaimTaskValidate' -count=1 -v` | PASS — each of the four previously-dead checks (LR-04, LR-05 writeHeavyAgents, LR-07 dedupe, ClaimTask pending-validation) has both a firing (positive) and non-firing (negative) case in the test output |

## §C Edge Cases

- Agent frontmatter with CSV `tools:` string vs YAML list — yaml.v3 parse must handle the project's documented CSV-string convention without false findings.
- Template mirror that is a sanitized (non-byte-identical) variant of the live agent — dedupe must still pair them via the path-mapping key (fallback content hash).
- An archived-name fixture (e.g. `expert-backend`) used to verify the writeHeavyAgents slice no longer matches — must NOT trigger LR-05, ensuring dead names cannot satisfy the rule by accident.
- Help output width/wrapping — the brain-absence grep must not be defeated by line wrapping (assert on the full `--help` buffer via `go run`, not a paged/pipe-truncated view).
- ClaimTask racing a concurrent claim of the same task — combined with the P0 O_APPEND fix, the second claim must fail pending-validation (the first claim flips the row to CLAIMED, so the second's pending search finds nothing).

## §D.5 Quality Gate / Definition of Done

- All 6 AC rows PASS with verbatim command output cited in progress.md §E.2.
- **LR-07 baseline falsifiability (D8)**: before/after `moai agent lint --format=json` output recorded verbatim in progress.md §E.2 — the §C pre-flight baseline (captured before M3) is cited alongside the post-M3 run, demonstrating the LR-07 false-positive count drops from the baseline value to 0. The baseline number is NOT predicted here; it is observed at run-phase entry and recorded in progress.md §E.2 as the denominator of the delta.
- Any NEW true findings surfaced by the fixed yaml.v3 parser (stricter than the hand parser) are triaged in progress.md (fix-or-file follow-up), not suppressed.
- `go test ./internal/cli/... -count=1` green; `golangci-lint run` no new findings.
