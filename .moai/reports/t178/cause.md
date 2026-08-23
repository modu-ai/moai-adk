# t178 — two backends reporting verdicts about code they never read

> Class B (재현 우선). No SPEC. Branch `WT-audit-verdict`, cut from `origin/main`
> at `1519f2660`. Both defects were surfaced by the SPEC-MCP-WORKTREE-ROOT-001
> M5 post-repair run and filed there as follow-ups 1 and 2.

## Overlap with t171 — checked before starting

t171 (`WT-audit-worktree-blind`) adds a `project_root` parameter and threads it
into codex's `cwd`. Measured against its branch:

```
$ git diff main...WT-audit-worktree-blind --stat -- internal/cli/mcp_codex.go \
      internal/cli/mcp_convergence.go internal/cli/mcp_audit_multi.go
 internal/cli/mcp_audit_multi.go | 18 ++++++++++--
 internal/cli/mcp_codex.go       | 25 +++++++++++++++--
 internal/cli/mcp_convergence.go | 61 ++++++++++++++++++++++++++++++-----------
```

It touches `handleCodexAudit`, `performCodexAudit`, `defaultBackendCaller`, and
`runMultiAudit`. It does **not** touch `synthesizeReviewOutput`, and it does not
change what GLM is sent — its own report says so, calling GLM's blindness "a
finding, not a footnote" and filing it as a new card. So neither defect here is
fixed by t171, and neither fix here duplicates it.

The branches overlap textually in two places: one line of `defaultBackendCaller`,
and `performGLMAudit`'s signature (t171 keeps two parameters and drops the root
for GLM deliberately; this card adds `target` and `projectRoot` because the diff
has to be read from a tree). Whichever merges second resolves both.

## (a) A merge-blocking codex review recorded as a pass

**Cause — `internal/cli/mcp_codex.go`, `synthesizeReviewOutput`.** codex returns
no verdict enum, so the verdict is inferred from prose. One signal was read,
`codexFindingBullet`, a regex for `- [P1] …` severity bullets:

```go
verdict := "pass"
if codexFindingBullet.MatchString(reviewText) {
    verdict = "fail"
}
```

That regex was measured against **review mode** output — the comment above it
says so, naming codex-cli 0.146.1 runs in both directions. The **adversarial**
path is a different codex method (`turn/start` plus a free-text prompt, reached
from `performCodexAudit` and from `handleCodexAudit` with `mode=adversarial`) and
returns ordinary prose: a leading `Verdict: fail — merge blocked.` line, findings
as numbered paragraphs, and no bracketed severity bullets anywhere. The regex
misses, and the function falls through to its `"pass"` initial value.

Nothing downstream recovers it. `converge()` reads the synthesized string, so
`overall_verdict` came out `pass` for a review whose own first line said the
merge was blocked — the required codex gate inverted, with nothing in the result
saying so.

**Reproduction (offline — no backend, no server restart).**
`TestSynthesizeReviewOutput_AdversarialVerdictLine` feeds the prose shape from
the M5 run straight into the synthesizer. Observed on the pre-fix tree:

```
--- FAIL: TestSynthesizeReviewOutput_AdversarialVerdictLine
    adversarial prose opening with "Verdict: fail" synthesized Verdict = "pass", want "fail"
--- FAIL: TestSynthesizeReviewOutput_VerdictLineDirections
    stated fail, no bullets:    Verdict = "pass", want "fail"
    stated FAIL uppercase:      Verdict = "pass", want "fail"
    stated fail, markdown bold: Verdict = "pass", want "fail"
```

**Fix.** A second regex, `codexStatedVerdict`, reads a verdict codex states in
its own words. The two signals combine **fail-biased**: a stated `pass` does not
clear finding bullets. When they disagree the blocking reading wins, because the
other direction is the one that launders a review carrying findings into a clean
verdict. The match is anchored to a line opening with the label, so "I could not
reach a verdict on the caching layer" stays prose — pinned as its own row.

## (b) GLM returning a confident verdict on nothing

**Cause — `internal/cli/mcp_glm.go`.** `glmAuditUserPrompt` said "Review the
proposed change …" and nothing more; `callGLMAudit` posted that with the system
prompt and no other content. The `glm_audit` tool had no `target` parameter, and
no diff, file, or path was ever assembled. Structurally there could not be one:
GLM is an HTTPS call to z.ai with no working directory — which is exactly why
t171 dropped `project_root` for this backend rather than threading it.

So the model was asked to review a change it had no way to see. It did not report
having nothing. In the M5 run it returned `fail` citing "a critical vulnerability
… failing to resolve the project root path before validating against the
repository whitelist" — there is no repository whitelist in this codebase, and
`validateProjectRoot` calls `filepath.Abs` first thing.

The attribution is clean precisely because there is no tree: reading the wrong
one cannot be the cause when none is read at all.

**Reproduction.** `TestGLMAudit_RequestCarriesTheDiff` builds a throwaway git
tree holding an uncommitted change with a canary symbol, stubs the HTTP doer, and
asserts the outgoing z.ai body contains it. Pre-fix the body carried only the
generic instruction — the canary appeared nowhere, because nothing put it there.
(Pre-fix the test does not compile against the old two-parameter
`performGLMAudit`, which is the same finding stated by the type checker: there
was no parameter through which review material could arrive.)

**Fix.** `internal/cli/mcp_review_material.go` collects the change as a unified
diff from the named tree — `uncommittedChanges` → `git diff HEAD`, `baseBranch` →
`git diff <merge-base>...HEAD` — bounded at 200 KB with a visible truncation
marker. `performGLMAudit` and `handleGLMAudit` both collect before calling z.ai,
the prompt carries the diff and tells the model to ground every finding in it,
and `glm_audit` gains the same `target` enum `codex_audit` already had.

The second half matters as much as the first: **when no diff can be produced,
z.ai is not called at all** and the result is `inconclusive`.
`TestGLMAudit_NoMaterialIsInconclusiveNotAVerdict` pins both halves — the verdict
and the absent HTTP call.

## Verification

| Claim | Command | Observed |
|---|---|---|
| both defects reproduce pre-fix | `go test ./internal/cli/ -run 'TestSynthesizeReviewOutput_(AdversarialVerdictLine\|VerdictLineDirections)' -count=1` | FAIL, verdict `pass` where `fail` was stated (quoted above) |
| (b) has no parameter to carry material | `go vet ./internal/cli/` on the pre-fix tree | `too many arguments in call to performGLMAudit … want (context.Context, string)` |
| both fixed | `go test ./internal/cli/ -run 'TestSynthesizeReviewOutput\|TestGLMAudit_\|TestPerformGLMAudit\|TestPerformCodexAudit\|TestDefaultBackendCaller' -count=1` | `ok … 3.368s` |
| package suite | `go test ./internal/cli/ -count=1 -timeout 900s` | `ok … 571.929s`, exit 0 (`suite.txt` — `.log` is gitignored, so the evidence carries a suffix that survives) |
| build | `go build ./...` | clean |

## What this does not establish

- **No live backend was exercised.** Both fixes are pinned by offline tests. The
  MCP server is a long-lived subprocess and does not reload when the binary
  changes, so a live check needs the binary installed and the server restarted —
  not done here.
- **The prompt change does not guarantee GLM stops inventing findings.** It
  removes the cause observed in the M5 run (no material at all) and instructs the
  model to ground findings in the diff. Whether a grounded prompt actually cuts
  invented findings is a claim about model behaviour that no offline test makes.
- **`codexStatedVerdict` reads the shapes it was given.** It covers the line form
  observed in the M5 run plus case and markdown-emphasis variants. A codex
  release that states its verdict some other way would fall back to the bullet
  heuristic — the same place this started, though now with one more signal
  rather than one fewer.
- **M5 follow-ups 3 and 4 are untouched** — convergence state persisting to the
  tree `resolveProjectDir()` names rather than `cfg.ProjectRoot`, and any
  `.moai`-bearing directory being an acceptable root. Both stay filed where that
  report put them.
