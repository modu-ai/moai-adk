# t341 — plan-phase discovery: test-selector census

Card: t341 · worktree `.claude/worktrees/t341` · branch `WT-selector-census`
Tree measured: `a6bbbf82b` (= `origin/develop` at fetch time, 2026-08-29)

---

## 1. Claim

The card's premise — "a verification that uses a test selector without counting what it
actually selected has no mechanical-layer warning" — is **true, and understated**. On this
tree the mechanical layer does not merely stay silent: it **records a zero-execution
`go test` run as an observed test PASS**, which is the signal the Stop evidence gate reads.

Second claim: the defect is already normative in a landed artifact — a release-blocking
acceptance criterion asserts the opposite of the measured behaviour, and passed audit.

## 2. Evidence

### E1 — the classifier reads a zero-match run as a pass (RED-now)

Seam: `internal/hook/evidence_writer.go` — `deriveFromOutputText` (line 194) treats
`"ok  \t"` as a *precise pass marker* (line 213-215) and returns before any count is
consulted. `[no tests to run]` appears nowhere in the repository
(`grep -rn 'no tests to run'` → no hits outside doc comments; independently confirmed by a
read-only sweep of `internal/hook/**`).

Probe (temporary `internal/hook/zzz_t341_probe_test.go`, removed after the run):

    $ go test ./internal/hook/ -run '^TestT341ProbeZeroMatchSelector$' -count=1 -v
    === RUN   TestT341ProbeZeroMatchSelector
        zero-match specimen -> pass=true fail=false ok=true
        classifyTestCommand  -> isTest=true isPass=true isFail=false
        no-test-files        -> pass=false fail=false ok=false
        pytest-no-tests-ran  -> pass=false fail=false ok=false
    --- PASS: TestT341ProbeZeroMatchSelector (0.00s)
    ok  	github.com/modu-ai/moai-adk/internal/hook	0.590s

Input line fed to the classifier is the t350 specimen verbatim:

    ok  \tgithub.com/modu-ai/moai-adk/internal/config\t0.424s [no tests to run]

Reading: a run that executed **zero** tests is written to the evidence ledger as
`IsTestPass=true`, `Outcome=success` (`buildBashRecord`, evidence_writer.go:296-330). The
Stop evidence gate then sees an observed pass and stays quiet.

Two adjacent shapes are correct today **by accident, not by assertion**: `[no test files]`
and pytest `no tests ran` return no signal only because they lack the `ok  \t` token. No
test pins that behaviour, so it is one marker-list edit away from regressing.

### E2 — a landed release-blocking AC asserts the opposite

`origin/develop:.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md:13`, AC-TOSQ-001,
RED-now cell, verbatim:

> Test name does not exist → suite failure ("no tests to run" surfaces red).

Falsified on this tree with the counterfactual the cell describes — a test name that does
not exist:

    $ go test ./internal/kanban -run TestMigrationParityDoesNotExistXYZ -count=1 ; echo "exit=$?"
    ok  	github.com/modu-ai/moai-adk/internal/kanban	0.429s [no tests to run]
    exit=0

Not red. The cell's premise is false, and six sibling ACs in the same matrix
(AC-TOSQ-002..005, 007, 008) rest on the same phrase, "Red via missing test".

> Correction to the material as relayed: the reproduction quoted to this lane ran
> `-run TestMigrationParity`, which **does** exist on develop
> (`internal/kanban/backlog_migrate_test.go:70`), so its `ok (cached) exit=0` shows a
> passing test, not a zero-match. The falsification above re-runs the right
> counterfactual. The conclusion stands; the cited command did not establish it.

### E3 — sixth live specimen, produced by this lane

Composing the probe, the heredoc that created the probe file was refused by the
worktree guard, so the file never existed. The run that followed printed:

    testing: warning: no tests to run
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/hook	1.029s [no tests to run]

`PASS` on a file that was never written. The shape reproduces itself inside the card that
exists to catch it.

## 3. Baseline-attribution

Every command above was run in `.claude/worktrees/t341` at `a6bbbf82b`, in this session,
on 2026-08-29. Working tree clean at measurement time except the probe file, since removed
(`git status --short` → empty).

## 4. Gaps

- **Not observed**: whether the live Claude Code `PostToolUse` payload actually delivers
  Bash stdout for a `go test` call. `evidence_writer.go` is written for a wrapped
  `tool_response` object and its tests use synthesized fixtures; no captured live payload
  was read. Run-phase must observe one before relying on it.
- **Not observed**: `jest` / `vitest` / `cargo` zero-execution output tokens. Only the go
  and pytest forms were exercised.
- **Not measured**: how many other landed SPEC acceptance matrices carry the same
  "missing test → red" premise. Only SPEC-TODO-SQLITE-001 was read.
- **Not run**: the full suite (policy — CLAUDE.local.md §4).

## 5. Residual risk

- E1's fix touches a marker list that four runner families share. Widening it can turn a
  genuine pass into "no signal" and quiet the gate in the other direction; the run-phase
  criteria must pin both directions.
- E2's cell is cited by t343 (lane-7) from the opposite axis. **Whoever edits that cell
  first silently removes the other card's evidence** — any change to
  `SPEC-TODO-SQLITE-001/acceptance.md:13` must update both SPECs in the same change.

---

## Scope proposal (lane decision, open to the lead)

The card leaves the axis open: `-run` only, or the cousins too. Proposed split:

**IN — one seam, one mechanism.** Zero-execution detection in the existing
`classifyTestCommand` / `deriveFromOutputText` pure functions: a zero-swept run is never
recorded as an observed pass, for every runner already in `testCommandSignatures`
(go `[no tests to run]` / `[no test files]`, pytest `no tests ran`, jest-vitest `0 passed`,
cargo `0 passed`). Plus the surfacing path, so the run is visible rather than merely
unrecorded — silence is what the card is about.

**OUT — named non-goals.** grep-based AC with a zero-hit pass condition, `t.Skip` before an
assertion, `sg test` over an empty ruleset. They share the reasoning error but no execution
seam: the first two are authoring-layer, and `sg test`'s `0 passed` falls out of the same
count taxonomy for free if its command signature is ever added. Covering them mechanically
would need a second, different instrument — a separate card.

**Doc layer.** `verification-completeness.md` §1.1 already carries this exact form as an
*"Evidence footnote (not a rule)"*. This card promotes it to a clause, so the mechanical
layer has a rule to cite. That is the whole doc scope — no rewrite.

**Detection keys on structure, not vocabulary** (lane-8's t342 measurement: lexical
discriminators are unsound in both directions, 7 false positives). The key here is an
output token emitted by the runner itself, plus the absence of an executed-test count —
not the wording of a claim.

**Self-application (t241).** The guard's own criteria must observe both directions: the
planted specimen fires, and a genuine `ok  \t` pass does **not**. A guard verified in one
direction only is indistinguishable from a guard that was switched off.

Proposed Tier: **M** — one Go package plus a rule clause plus the template mirror.
