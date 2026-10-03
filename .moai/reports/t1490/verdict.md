# t1490 — red develop CI repair (7 tests)

Base: local develop 97c905858 (fast-forwarded), branch WT-develop-red-guards-repair.
CI reference: run 37122965651, job 111202629123 (origin/develop 19de344af).

## Claim

All seven CI-red tests pass on this tree after four scoped edits; none weakens a guard.

| Test | Package | Introduced by | Fix |
|---|---|---|---|
| TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration | internal/hook | t1404 (`9ef1cbedc` etc. add fixture `t1404@example.invalid`) | add `t1404@example.invalid` to `builtinCommitIdentityDenyEmails` |
| TestNumeralAxisFindsNoUndeclaredCountClaim, TestNumeralBreadthSetEqualsTheDeclaredUnion, TestNumeralResidualArithmeticCloses, TestSweepFindsNoUndeclaredRosterListing | internal/harness/rosterguard | t1435 (`815f04c83`, plugin payload `plugins/moai/`) | add `plugins/moai/` to `sweepSkipPrefixes` (shared by the numeral walk). The payload is a byte-identical machine emission of template files (`make plugin-emit`; read-only `plugin-emit-check` + pluginemit golden tests fail on divergence); every hit (8 paths) is the copy of a template site already registered in `Registry()` |
| TestCodex1718Fixtures_WidenedContent | internal/cli | t1404 `ae8c4bd46`/`872c6cb42` (codexFindingAnchorOf: several distinct path:line candidates leave the anchor unset, deliberately — N1/M1 repair) | fixture expectation aligned: S1 finding[0] names guide.md AND intro.md, so File/Line are now "" / 0. Code behavior is the intended one; test predated it |
| TestWSR006_ReviewGateRootMatrix | internal/cli | t1404 `697f18570` (primary_scope default `skip`, REQ-CGSC-002) | the matrix measures root resolution, so its fixture opts in with `primary_scope: review` (REQ-CGSC-004), the same pattern other t1404 tests use; under default skip the P-session rows measured the skip, not the root |

## Evidence

RED (scrubbed lane env, `go test -count=1 -run '<names>'`):
- internal/hook: `commit_identity_guard_list_test.go:156: fixture email literals found ... missing from builtinCommitIdentityDenyEmails ...: [t1404@example.invalid]`
- rosterguard: `numeral_rederivation_test.go:87: 8 breadth-set path(s) are neither registered nor exempted`; `numeral_test.go:352: undeclared count claim: plugins/moai/skills/moai-foundation-core/SKILL.md:239 ...` (+10 more, all under plugins/moai/); `rosterguard_test.go:148: undeclared roster listing: plugins/moai/skills/moai-foundation-core/modules/agents-reference.md ...`
- cli: `codex_1718_widening_test.go:81: S1.txt on turn/start: finding[0] got {sev="High" file="" line=0 ...} want {sev="High" file="docs/guide.md" line=28 ...}`; `wsr_state_root_test.go:403: row 3: observed other, required Pp` (rows 3,7,8,9,... — codex calls=[] with `primary_scope":"skip"` log rows)

GREEN (same command): `cli=0 rg=0 hook=0`
- `ok github.com/modu-ai/moai-adk/internal/cli 10.123s`
- `ok github.com/modu-ai/moai-adk/internal/harness/rosterguard 17.366s`
- `ok github.com/modu-ai/moai-adk/internal/hook 1.287s`

Package runs:
- `go test -count=1 -timeout 30m ./internal/hook/` (lane env incl. MOAI_KANBAN_LEAD_NAME scrubbed): `ok github.com/modu-ai/moai-adk/internal/hook 412.945s`
- `./internal/hook/... ./internal/harness/rosterguard/...`: subpackages all ok (rosterguard, handoff, memo, mx, perf, quality, security, testutil, trace)
- `go vet` on the three packages: exit 0
- `golangci-lint v2.1.6 run` on the three packages: `0 issues.`
- `make fmt-check`: exit 0

## Baseline-attribution

All measured in this run on worktree agent-ad5fe1785d42bd12e at base 97c905858 plus the edits above.

## Gaps

- `go test -timeout 30m ./internal/cli/` full package: `panic: test timed out after 30m0s` with 0 `--- FAIL` lines before the alarm; machine load average 16-17 at the time (other lanes). Full-package verdict for internal/cli is left to CI.
- The first hook package run failed 7 session-title tests with `"leader"` because `MOAI_KANBAN_LEAD_NAME=leader` was not in the scrub list; re-run with it unset passed. Lane-env artifact, not a defect on this tree.

## Residual-risk

- Excluding `plugins/moai/` relies on the emission staying byte-identical; if pluginemit ever transforms skill content, roster drift introduced only in the payload would escape this guard (the plugin-emit golden would still catch a hand edit).
