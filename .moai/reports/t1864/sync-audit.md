auditor-model: claude-sonnet-5-5

verdict: PASS
audited_sha: 026340b5e21cb60c7de55dc6e3c7560e4a1d7856
required_backend: codex pass
convergence_overall: pass

# Sync audit: card t1864 (Class A, main-red express, no SPEC)

SPEC: none (Class A card). Tree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1864, branch WT-run-adopt-errcheck, merge base bc5817d75.

## Scores (0.00-1.00)
- Functionality: 1.00
- Security: 1.00
- Craft: 0.95
- Consistency: 1.00

Must-pass dimensions (Functionality, Security) meet their thresholds. Overall: PASS.

## Audit plan (verbatim extract)
`moai verify audit-plan --project-root <tree>`: config_status ok; model codex; backends claude off, codex required, glm off; cross_model_active true; enforced_required ["codex"].

## 1. Claim
The three `fmt.Fprintf(warn, ...)` call sites in `adoptResumedLeaderRun` now discard their results with `_, _ =`. Nothing else in Go source changed. The card's errcheck command exits 0, with no lint relaxation and no `//nolint`.

## 2. Evidence (verbatim, this run)

A1, errcheck command and default lint:
```
$ golangci-lint run --enable-only errcheck ./internal/cli/ 2>&1 | tail -20
0 issues.
$ golangci-lint run ./internal/cli/ 2>&1 | tail -20
0 issues.
```
The `exit` echo printed an empty value because `PIPESTATUS` is not set the way I used it in this shell. golangci-lint prints `0 issues.` only on success, and the lint output was fed through `tail`, so the exit code itself was not captured. This is recorded as a Gap, not a Claim.

A2, diff scope:
```
$ git diff --stat bc5817d75..HEAD
 .moai/reports/t1864/card-review.md | 11 +++++++++++
 .moai/reports/t1864/progress.md    | 35 +++++++++++++++++++++++++++++++++++
 internal/cli/factory_run_adopt.go  |  6 +++---
 3 files changed, 49 insertions(+), 3 deletions(-)
$ git diff --word-diff=porcelain bc5817d75..HEAD -- internal/cli/factory_run_adopt.go | grep '^[+-]'
--- a/internal/cli/factory_run_adopt.go
+++ b/internal/cli/factory_run_adopt.go
+_, _ =
+_, _ =
+_, _ =
```
The word-level diff shows only the `_, _ =` token added at lines 79, 91 and 103. Format strings and arguments are byte-identical. No `.golangci.yml` or other lint config appears in the stat.

```
$ git diff bc5817d75..HEAD | grep -n nolint
44:+Only the three call sites change: `fmt.Fprintf(warn, ...)` -> `_, _ = fmt.Fprintf(warn, ...)`. No lint configuration change, no `//nolint`.
```
The only `nolint` hit is prose in progress.md (diff line 44) saying that none was added. No directive was added to source.

A3, convention and callers:
```
internal/cli/factory.go:916:		_, _ = fmt.Fprintf(notes, "factory: %s is held by a live session; launching as %s\n", canonical, final)
internal/cli/factory_adopt.go:95:			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "resuming card %s (lane %s)\n", card, lane)
internal/cli/cc.go:214:		adopted, adoptErr := leaderEntryRunAdoption(launchProjectRoot(), filteredArgs, leaderLabel, priorRunID, entry.FactoryRun, cmd.ErrOrStderr())
internal/cli/glm.go:274:		adopted, adoptErr := leaderEntryRunAdoption(launchProjectRoot(), filteredArgs, leaderLabel, priorRunID, entry.FactoryRun, cmd.ErrOrStderr())
```
Both callers (cc.go, glm.go) pass `cmd.ErrOrStderr()` as `warn`. It is a best-effort stderr notice at launch time. The function's own comment says only a factory database failure propagates. All three writes sit on paths that return `"", nil` (or on the preceding branch), so `adopted`/`err` are unaffected by the discarded write results. The `adopted string, err error` contract is unchanged. The form `_, _ = fmt.Fprintf(` matches factory.go:916 and factory_adopt.go:95-104.

A4, tests, build, format:
```
--- PASS: TestAdoptResumedLeaderRunAdoptsSingleCandidate (2.46s)
--- PASS: TestAdoptResumedLeaderRunSkipsWhenNotResumeShaped (0.75s)
--- PASS: TestAdoptResumedLeaderRunWarnsOnMultipleCandidates (1.21s)
--- PASS: TestAdoptResumedLeaderRunMintsWhenNoCandidates (0.89s)
--- PASS: TestAdoptResumedLeaderRunWarnsCodexLedCandidate (1.40s)
--- PASS: TestStrictResumeCarrierSeen (0.00s)
--- PASS: TestLeaderRunIDWasMinted (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	8.072s
build=0
gofmt: (empty output)
vet: (no output; exit code not captured, see Gaps)
```

A5, regression sanity: the errcheck command over the whole `./internal/cli/` package reports `0 issues.`, so no other unchecked results remain in the file or the package. The progress.md refusal note is consistent with its description. It records a guard refusal of `git branch -m` in the primary checkout, and says the rename was repeated from the worktree rather than routed around. The branch `WT-run-adopt-errcheck` matches that. I did not reproduce the refusal.

Cross-model audit (`mcp__moai__audit_multi`, card_id t1864, target baseBranch, focus "errcheck fix: three call sites, behavior unchanged, no lint relaxation"):
```
{"per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[]}],"overall_verdict":"pass","disagreement_flag":null,"participant_count":1,"fail_open_backends":[],"audit_receipt":"rcpt-542faae67b11d116c0e195ad","plan_source":"config","build_commit":"ee981a932", ... "tree_root":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1864"}
```
Digest members for the orchestrator's result check: overall_verdict pass; gate_unmet absent; plan_source config; per-backend: codex / required / pass. My verdict is not final until the orchestrator's `moai verify audit-plan --result-file` check passes.

## 3. Baseline-attribution
- Tree: HEAD 026340b5e21cb60c7de55dc6e3c7560e4a1d7856, base bc5817d75, measured in this run.
- Judging builds: golangci-lint and the Go toolchain are third-party and installed (no tree lag applies). The moai MCP/CLI build was ee981a932, an ancestor of the tree HEAD (reported `build_lag`: the binary describes older code). It was used only to read the audit plan and run the backend audit, not to measure the change. The change itself was measured with golangci-lint, go test, go build and gofmt.
- I ran the card's exact command myself; I did not rely on the progress.md numbers.

## 4. Gaps
- The exit code of the two golangci-lint runs and of `go vet` was not captured, because my `PIPESTATUS` usage returned an empty value. The `0 issues.` output corroborates exit 0 but is not the code itself.
- The whole internal/cli suite was not run (scoped by instruction).
- The branch-guard refusal described in progress.md was not independently reproduced.
- The codex backend's reasoning is not visible beyond `Verdict: pass` with empty findings.

## 5. Residual-risk
- A write failure to stderr is now silently ignored, which matches the package convention. The only effect is a lost warning line, and no caller can act on it. There is no propagation path to lose, because `warn` is stderr.
- CI Lint runs only on origin/main, per progress.md. A different golangci-lint version could flag additional issues. This is theoretical.

## Defect list
- No BLOCKING findings.
- F1 [MINOR] [optional] Gap only: exit codes of the lint and vet runs were not captured. The corroborating output is `0 issues.`.

Required-backend verdicts: codex = pass (required gate, receipt rcpt-542faae67b11d116c0e195ad). claude and glm are off per config.

AUDIT-VERDICT: PASS spec=t1864 receipts=rcpt-542faae67b11d116c0e195ad
