auditor-model: claude-sonnet-5-5

# Sync audit, card t1858, iteration 3 (final delta)

verdict: PASS-WITH-DEBT
audited_sha: 55cac554f56b02bd4ace7e6c29952fa952bb1844

Delta audited: `git diff 16ba65ac9..f85666c0b` (correction commit f85666c0b; HEAD 55cac554f is docs-only on top). Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858.

## Iteration history
- iter1: FAIL @f9458ae18, findings F1-F4 (closed, not re-judged).
- iter2: FAIL @16ba65ac9, findings F5 (codex P1: SSoT-only PR skipped the pin tests) and F6 (codex P2: empty release_gate entries silently dropped).
- iter3 (this file): delta f85666c0b, scoped to F5 and F6 plus regressions the commit could introduce.

## Scores (0.00-1.00)
| Dimension | Score | Note |
|---|---|---|
| Functionality | 0.92 | F5 and F6 closed on observed evidence; hosted CI not observable |
| Security | 0.93 | Fail-closed awk; no realistic path to a short list on the real SSoT |
| Craft | 0.90 | Tests added and shown sensitive; lint/gofmt clean |
| Consistency | 0.92 | New filter entry follows neighbouring literal-path entries and comment style |

## Required backend verdicts (verbatim from audit_multi)
`{"per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[]}],"overall_verdict":"pass","disagreement_flag":null,"participant_count":1,"residual_risk_note":"","fail_open_backends":[],"audit_receipt":"rcpt-2d74fd9b7e1845c52910b486","plan_source":"config"}`

audit-plan: codex required (config.gates); claude off, glm off; gate_unmet absent. Codex returned pass with no findings, so no backend finding is left to quote. Per the audit contract this verdict is not final until the orchestrator's convergence check passes.

## 1. Claim
1. F5 is closed: `.github/required-checks.yml` is now a literal `go_code` entry, and the parity guard requires it.
2. F6 is closed: the release.sh awk refuses empty, whitespace-only and space-padded entries and any non-entry line inside the block, and still accepts tab indent, CRLF, comments and blank lines.
3. No regression: no check literal in the script, internal/config untouched, all named test and lint commands green.

## 2. Evidence (verbatim)
Filter diff: adds `- '.github/required-checks.yml'` after `- '.goreleaser.yml'` plus a comment block; nothing else in the file changes. Neighbouring entries are quoted literal paths (`'scripts/release.sh'`, `'.goreleaser.yml'`).
Consumers: ci.yml:79, codeql.yml:43, release-pr-multi-os.yml:68 all pass `filters: .github/test-input-filters.yml`, so the one new entry is seen identically by all three. Only a path change on `.github/required-checks.yml` newly yields go_code true.

F5 guard mutation (scratch copy at /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/3485ddff-4832-470a-9a17-ee4dac7ad178/scratchpad/cp, entry removed):
```
--- FAIL: TestTestInputFiltersCoversGoTestInputs (0.00s)
    test_input_filter_parity_test.go:143: .github/test-input-filters.yml: go_code list lost required root ".github/required-checks.yml" — TEST_INPUT_FILTER_DRIFT: ...
FAIL	github.com/modu-ai/moai-adk/internal/template	0.466s
```

F6 awk probe (program extracted from scripts/release.sh, run on crafted files; header `release_gate:` then lines; output shown as names, `|` = newline):
```
real SSoT          rc=0 Lint|Test (ubuntu-latest)|Build (linux/amd64)|
dash alone         rc=1 empty entry
- ""               rc=1 empty entry
- "  "             rc=1 empty entry
- " Lint"          rc=1 space-padded name: " Lint"
- "Lint "          rc=1 space-padded name
- Lint<space>      rc=0 Lint|B   (trailing space stripped: right name)
tab indent         rc=0 Lint|B
CRLF               rc=0 Lint|B
comment+blank      rc=0 Lint|B
--- after block    rc=0 Lint|B
column-0 key after rc=0 Lint|B
column-0 item      rc=0 (empty) -> caught by the non-empty check after the awk
"Lint #1"          rc=1 trailing comment
single quote       rc=1
-Lint              rc=1 unrecognized line
duplicate          rc=1
indented 2nd key   rc=1 unrecognized line
nested sub-item    rc=0 Lint|sub|B   (extra name demanded: fails closed at check stage)
lone quote         rc=0 `"` demanded: fails closed at check stage
release_gate: [Lint, B] + "  - Lint"  rc=0 Lint   (see M1)
release_gate: []   rc=0 (empty)       -> caught by the non-empty check
```
Sensitivity of the new rows (scratch copy with release.sh reverted to 16ba65ac9):
```
--- FAIL: TestReleaseScriptCIGateSSoTFailsClosed/bare_dash_entry_is_malformed
--- FAIL: .../dash_space_entry_is_malformed
--- FAIL: .../empty_quoted_entry_is_malformed
--- FAIL: .../blank_quoted_entry_is_malformed
--- FAIL: .../space_padded_quoted_name_is_malformed
--- FAIL: .../unrecognized_line_is_malformed
```
(`tab_indented_list_passes` passes before and after: a regression guard, not a defect guard.)

Regression commands in the tree:
```
go test ./internal/config -count=1                      ok  github.com/modu-ai/moai-adk/internal/config 11.830s
go test ./internal/template -run 'TestReleaseScript|TestTestInputFilters|BranchProtection' -count=1   ok  .../internal/template 85.473s
gofmt -l internal/config internal/template              (no output)
golangci-lint run ./internal/config/ ./internal/template/   0 issues.
scripts/ci-mirror/validate-required-checks.sh           All validations passed
grep -c 'Test (ubuntu-latest)' scripts/release.sh       0
git diff --stat 16ba65ac9..f85666c0b -- internal/config 0 lines (untouched)
```
The delta touches no `Release Verify (...)` leg. The tree stayed clean (`git status --short` empty).

## 3. Baseline-attribution
All commands ran in this run, in /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1858 at HEAD 55cac554f56b02bd4ace7e6c29952fa952bb1844 (delta f85666c0b). Mutation tests ran on a `git archive HEAD` copy under the scratchpad. The moai MCP server reported build lag (installed build ee981a932 is an ancestor of HEAD); the cross-model verdict was produced by that server's codex backend, and the lag note is recorded here as required. All other measurements used the tree's own scripts and `go test`, not the moai binary.

## 4. Gaps
- Hosted CI was not run. That `dorny/paths-filter` classifies an edit to `.github/required-checks.yml` as go_code was inferred from the filter file and consumer wiring plus the guard test, not observed on a real PR.
- No live GitHub check-run call was made.
- Not run: the full `go test ./...`; only the scoped packages and commands listed.
- audit_multi returned no findings text, so no backend finding could be quoted.

## 5. Residual-risk
- M1 (MINOR): `release_gate: [Lint, B]` (inline value) followed by an indented `- Lint` yields rc=0 and the one-name list `Lint`. This cannot occur on the real SSoT: internal/config `required_checks_release_gate_test.go` parses the YAML and pins exactly the three names, and `TestReleaseScriptRealSSoTDemandsEveryPinnedCheck` runs the script against it; both are green. Optional hardening: refuse any non-empty text after `release_gate:` in the awk.
- M2 (MINOR, cosmetic): on a refusal the die message carries the names printed before the failing line (e.g. `Lint|empty entry`). The refusal is still correct.
- M3 (MINOR): a nested sub-item or a lone `"` is read as an extra name and demanded; the release fails closed (blocked, not passed). The real SSoT has neither.
- Hosted CI behaviour of the new filter entry stays unobserved until a PR that edits only the SSoT runs; the guard test and literal-path neighbours make a different result unlikely.

## Defect list
| ID | Severity | Item |
|---|---|---|
| F5 | closed | Filter entry present and guarded; guard fails without it |
| F6 | closed | Empty/padded/unrecognized entries refused; valid forms accepted; real SSoT yields exactly three names |
| M1 | MINOR | Inline value after `release_gate:` not refused (not reachable on the real SSoT) |
| M2 | MINOR | Refusal message carries preceding names |
| M3 | MINOR | Nested item or lone quote becomes a demanded name (fail-closed) |

No BLOCKING or SHOULD-FIX defect. Overall: PASS-WITH-DEBT (MINOR items M1-M3, plus the unobserved hosted-CI gap above).

AUDIT-VERDICT: PASS-WITH-DEBT spec=t1858 receipts=rcpt-2d74fd9b7e1845c52910b486
