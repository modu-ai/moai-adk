auditor-model: claude-sonnet-5-5

# Sync audit: card t1800 (Class A, Template-First, no SPEC)

verdict: PASS-WITH-DEBT
audited_sha: d1481b3215796fa62cace8dcebdbb0d8a9959d55

Overall: PASS-WITH-DEBT. The change is correct and complete for its stated scope. The debt is a pre-existing parity divergence left by ec193e477 that d1481b321 neither introduced nor claims to fix, plus environment-dependent Codex-wiring test failures that I did not baseline.

Scores (flat mode): Functionality 95, Security 100 (docs only), Craft 85, Consistency 90. Must-pass dimensions (Functionality, Security) pass.

## Claim / Evidence / Baseline-attribution

| # | Claim | Evidence (command, verbatim output) | Baseline |
|---|---|---|---|
| C1 | Commit scope is the two files only | `git diff --name-only ec5a4f392 d1481b321` printed `.claude/rules/moai/core/hooks-system.md` and `internal/template/templates/.claude/rules/moai/core/hooks-system.md`. `git show --stat`: `2 files changed, 4 insertions(+), 2 deletions(-)` | this tree, run now |
| C2 | Restored text equals what ec193e477 deleted, modulo neutralisation | ec193e477 removed line A `- Hook stdout over 50K characters ... (v2.1.89+) — this 50K cap is on a hook's TOTAL stdout.` and line B `- A second, distinct cap applies per \`additionalContext\` string: ... 2,000-character preview (decision-index Q4; observed live by a factory lane session's session-start overflow notice — a 23,166-character assembled context)`. d1481b321 re-adds line A byte-identical and line B with the parenthetical replaced by `(measured live: a 23,166-character assembled session-start context was replaced by a file path and preview)`. Everything before the parenthetical is identical. | `git show ec193e477 -- <both files>` and `git show d1481b321` |
| C3 | Twins byte-identical, `10,000` present | `cmp` exit 0 (`cmp=0`); `grep -c '10,000'` printed `1` for each of the two files. Both blobs in d1481b321 share index `e3255abc5` | this tree |
| C4 | Other ec193e477 deletions in rules / template / `.moai/policies` | `git show ec193e477 --stat -- .claude/rules internal/template .moai/policies`: the only deletions are in the two hooks-system.md files (`5 +-` each, 6 deletions total). The `.moai/policies` files are pure additions. The remaining deletion in those two files is the UserPromptSubmit row `5s | 30s | ... reduced max`, replaced by the intentional `15s` row. That is a modification, not a lost hunk | this tree |
| C5 | Restored rows are consistent with the file | The current file has line 131 (50K tail, TOTAL stdout), line 132 (10,000 per-string cap, 2,000 preview), line 318 (`additionalContext` retained from every hook and passed together), and line 369 (`UserPromptSubmit | 15s | 30s`). No contradiction. The 50K cap (total stdout) and the 10,000 cap (per string) are explicitly distinguished. The restored text carries no SPEC, card or date citation, and the old `decision-index Q4` citation is gone | grep of this tree |
| C6 | `go build ./...` | exit=0 (`$S/build.out` empty) | this tree |
| C7 | `go test ./internal/template/... -count=1` | exit=0. `ok  internal/template 265.617s`, `ok agentemit 0.559s`, `ok commandemit 0.330s`, `ok embedemit 0.351s`, `scripts [no test files]` | this tree, run now |
| C8 | Cross-model gate (codex required) | `moai verify audit-plan`: `config_status: ok`, `cross_model_required: true`, `enforced_required: ["codex"]`. `audit_multi` (project_root = this worktree, no gates arg): `overall_verdict: pass`, `gate_unmet` absent, `plan_source: config`, backend `codex` gate `required` verdict `pass`, `audit_receipt: rcpt-062951309100c8ee65d64e08` | this tree |

## Findings

- F1 [low] [optional] `internal/cli` run `go test ./internal/cli -run 'Mirror|Parity|Doc' -count=1` exited 1 (`FAIL github.com/modu-ai/moai-adk/internal/cli 330.208s`). Failing tests:
  - `TestHookEntryParity_NoDivergenceEitherDirection`, with `hook_entry_parity_test.go:172: template-only ... UserPromptSubmit/handle-user-prompt-submit.sh timeout=5 async=false` and `:175: project-only ... timeout=15 async=false`.
  - `TestCheckCodexWiring_Mirror*` (six tests; for example `absent mirror on a wired project status = ok, want Warn`). The `~/.codex/config.toml: 1 stale skill entry` text suggests the tests read the operator's real Codex home.
  
  Attribution: d1481b321 changes only two markdown files, so it cannot affect either. For the parity test I confirmed the cause independently: `git grep` at ec5a4f392 shows the template settings at `"timeout": 5`, while ec193e477 changed project `.claude/settings.json` from 5 to 15 without the template twin. This divergence was introduced by ec193e477 and pre-exists d1481b321. Required fix (separate card): set the template `settings.json.tmpl` UserPromptSubmit timeout to 15, or revert the project value. It was not run against a clean baseline (see Gaps), so the Codex-wiring failures are an inference, not a measurement.
  
  This is not blocking for t1800: the failures have no path through a markdown-only rule edit.
- F2 [info] [optional] ec193e477 raised the rule's UserPromptSubmit row to 15s while the shipped template settings still say 5s. This is the same issue as F1, noted for rule-vs-template consistency. Not part of this card's scope.

No blocking findings.

## Gaps

- I did not run the failing `internal/cli` tests against base ec5a4f392. The claim that the Codex-wiring failures are environment-related is inferred from the diff scope and the failure messages. It is a hypothesis, not a measurement.
- `go test ./internal/cli` was run only with the `Mirror|Parity|Doc` filter, as instructed. No full-suite cli run.
- `moai` CLI binary lag: the installed binary was built from ec5a4f392, an ancestor of HEAD d1481b321 (per `audit_multi` `build_lag`). `moai verify audit-plan` therefore describes ec5a4f392's code. The tree is docs-only relative to that commit, so the lag does not change this audit's verdict.
- The `audit_multi` result does not carry `gate_unmet`. The orchestrator's `--result-file` convergence check was not run by me; my verdict is not final until that check passes.
- The Verification-batch output from `moai verify run` reuse was not used; every cited command was run directly.

## Residual-risk

- Rule/template drift is mechanically guarded for these two files only by the Template-First convention plus the template Go tests, which passed. The cli parity test currently fails for an unrelated pre-existing reason, so that guard is red until F1 is fixed.
- The restored row's "(measured live: ...)" claim rests on the authors' observation; I did not re-measure the 10,000-character cap against a live runtime.

## Iteration history

Round 1 (this audit). No prior audit.

AUDIT-VERDICT: PASS-WITH-DEBT spec=t1800 receipts=rcpt-062951309100c8ee65d64e08
