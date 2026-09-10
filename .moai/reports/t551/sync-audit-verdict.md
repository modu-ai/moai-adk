# Sync-Audit Verdict — SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 (card t551)

## Evaluation Report

- SPEC: SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001
- Card: t551 · GitHub issue #1632 (blank-output axis)
- Tree: `.claude/worktrees/t551`, branch `WT-audit-fail-open`
- **Overall Verdict: PASS** (harmonic mean 89.1; both must-pass dimensions clear independently)

### Audit-window integrity

| Moment | HEAD | `git status --porcelain` | `internal/cli/mcp_codex.go` sha256 |
|---|---|---|---|
| open | `a4ff25f74346bfe5771f5480eb6de4ac9c63d5ff` | empty | `28b436563500a998526358a16e55857840ec069cd6e4d720c9078fb8ffaea1b1` |
| close | `a4ff25f74346bfe5771f5480eb6de4ac9c63d5ff` | empty | `28b436563500a998526358a16e55857840ec069cd6e4d720c9078fb8ffaea1b1` |

HEAD did not move; no artifact changed underneath the audit. Seven mutation probes and one
temporary probe test (`internal/cli/zz_audit_probe_test.go`) were applied and reverted; the
byte-identical sha256 above is the restoration proof. No push, no PR, no merge.

---

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | `go test ./internal/cli/...` → `ok github.com/modu-ai/moai-adk/internal/cli 426.811s`; 9/9 AC PASS on my own run; RED reproduced behaviorally against the unrepaired implementation |
| Security (25%) | 85/100 | PASS | Audit-laundering path closed and measured; fail-open contract preserved (`TestCharacterize_UnavailableBackend` byte-identical pin); one measured residual bypass (U+200B, F4) |
| Craft (20%) | 86/100 | PASS | New functions 100% covered; 5 of 8 mutants caught; `golangci-lint run --timeout=5m ./internal/cli/...` → `1 issues: * errcheck` (pre-existing, `todo.go:97`), NEW 0 |
| Consistency (15%) | 94/100 | PASS | Scope held to one production file; pinned test byte-identical; CHANGELOG format matches sibling entries and respects all three [HARD] accuracy constraints |

Must-pass firewall (Functionality + Security): both clear independently. No blocking finding.

---

## 1. Claim

The repair is correct, the acceptance suite genuinely discriminates it for the criteria that
matter, the RED was behavioral rather than a build failure, the preserved behaviors are actually
preserved, and the shipped prose does not overclaim. Three mutants survive — all of them test-
strength gaps, none a shipped defect.

## 2. Evidence

### 2.1 Acceptance suite — re-run, not accepted

```
$ go test ./internal/cli/...
ok  	github.com/modu-ai/moai-adk/internal/cli	426.811s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	(cached)
[... 15 further subpackages, all ok ...]
```

```
$ go test ./internal/cli/ -run 'TestCodexBlankReview' -v
--- PASS: TestCodexBlankReview_BlankDiscriminator (0.00s)
--- PASS: TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass (0.00s)
    --- PASS: .../space-only          --- PASS: .../newline-only
    --- PASS: .../mixed-whitespace    --- PASS: .../exactly-empty
--- PASS: TestCodexBlankReview_AC002_RealCleanReviewStillPasses (0.00s)
--- PASS: TestCodexBlankReview_AC003_BlankStructuredReviewDoesNotShadowAgentMessage (0.00s)
--- PASS: TestCodexBlankReview_AC004_UnavailableBackendStillFailsOpen (0.00s)
--- PASS: TestCodexBlankReview_AC005_BlankAndUnavailableAreDistinguishable (0.00s)
--- PASS: TestCodexBlankReview_AC006_BlankBodyIsNeverFail (0.00s)  [4 subtests PASS]
--- PASS: TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput (0.00s)
--- PASS: TestCodexBlankReview_AC008_NonBlankUnrecognizedBodyUnchanged (0.00s)
--- PASS: TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne (0.00s)
    --- PASS: .../review:_real_then_blank
    --- PASS: .../review:_blank_then_real_(kills_first-wins)
    --- PASS: .../agent:_real_then_blank
    --- PASS: .../agent:_blank_then_real_(kills_first-wins)
--- PASS: TestCodexBlankReview_M5_InconclusiveDoesNotBlockTheConvergenceLayer (0.00s)
--- PASS: TestCodexBlankReview_ThreeStateControlMatrix (0.00s)  [A/B/C subtests PASS]
ok  	github.com/modu-ai/moai-adk/internal/cli	0.682s
```

9/9 AC PASS, plus the three-state control matrix and the M5 convergence measurement.

### 2.2 Mutation testing — 8 mutants, 5 CAUGHT, 3 SURVIVED

| # | Mutant | Result | Killed by |
|---|---|---|---|
| A | collection: first-**non-blank**-wins (`&& codexReviewTextIsBlank(reviewText)`) | **SURVIVED** | — (F3) |
| A2 | collection: first-wins as acceptance.md describes it (`&& reviewText == ""`, records the blank) | **CAUGHT** | `AC009/review:_blank_then_real_(kills_first-wins)`, `AC009/agent:_blank_then_real_(kills_first-wins)` |
| B | discriminator trims only ASCII space (`strings.Trim(s, " ")`) | **CAUGHT** | 7 tests: `BlankDiscriminator`, `AC001` (newline-only, mixed-whitespace), `AC003`, `AC005`, `AC007`, `AC009`, `ThreeStateControlMatrix`, `TestCharacterize_ReviewTextPath` |
| C1 | blank summary rewritten to reuse the `codex unavailable: ` prefix (const rewrite) | **SURVIVED** | — (F1) |
| C2 | blank path routed through `inconclusiveReview` (prefix added at the factory) | **CAUGHT** | `TestCharacterize_ReviewTextPath` (all 4 blank rows) |
| D | guard `:855` reverted to `reviewText == ""` | **SURVIVED** | — (F2) |
| E | selection `:1308` reverted to `review != ""` | **CAUGHT** | `AC003` |
| F | both collection sites reverted to `!= ""` | **CAUGHT** | `AC009/review:_real_then_blank`, `AC009/agent:_real_then_blank` |

Verbatim, mutant A2 (the shape AC-CBR-009's reversed-ordering clause exists to kill):

```
--- FAIL: TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne (0.00s)
    --- FAIL: .../review:_blank_then_real_(kills_first-wins) (0.00s)
    --- FAIL: .../agent:_blank_then_real_(kills_first-wins) (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.912s
```

The reversed-ordering clause earns its keep: it is the **only** thing that kills A2.

Verbatim, mutants C1 and D:

```
# C1 — const rewritten to "codex unavailable: blank review output, no verdict text was produced"
ok  	github.com/modu-ai/moai-adk/internal/cli	0.875s
# D — guard at :855 reverted to `if reviewText == "" {`
ok  	github.com/modu-ai/moai-adk/internal/cli	0.854s
```

### 2.3 Vacuous-green hunt

- **AC-CBR-003's verdict half** — the named risk. The test does **not** rest on the verdict:
  `codex_blank_review_test.go:154` asserts `bestCodexReviewText("   \n", realCleanReview) ==
  realCleanReview`, and `:172` asserts the end-to-end `Summary` is the real agent text. Mutant E
  (selection reverted) kills it, and the recorded RED shows it failing on both assertions with the
  verdict never mentioned. **Not vacuous.**
- **AC-CBR-002 / 004 / 006 / 008** are declared preservation pins, excluded from the RED list by
  design; my RED replay confirms all four stay green on the unrepaired tree — which is the
  required behavior, not a vacuous pass.
- **No zero-match grep or count assertion exists in either new test file.** The suite asserts on
  values, not on absences; the only absence-shaped assertion (`gate_unmet != ""` in AC-CBR-007) has
  its positive control in the RED evidence, where it failed with `gate_unmet is empty`.
- **`TestCodexBlankReview_ThreeStateControlMatrix`** builds a signal map and errors on duplicate
  signals — a genuine three-way distinguishability check, not three independent single-state
  assertions.

### 2.4 The RED was real and behavioral

`red-evidence-20260908.txt` records its tree as `a87a7d8a7`. That commit's tree does **not** contain
the seam:

```
$ git show a87a7d8a7:internal/cli/mcp_codex.go | grep -n 'codexReviewTextIsBlank\|codexBlankReviewSummary\|blankReviewInconclusive'
(no output)
```

The seam-landed-but-unwired tree is `a5f83fd62`:

```
$ git show a5f83fd62:internal/cli/mcp_codex.go | grep -n 'codexReviewTextIsBlank\|codexBlankReviewSummary\|if reviewText == ""'
345:const codexBlankReviewSummary = "codex review output was blank: no verdict text was produced"
352:func blankReviewInconclusive() ReviewOutput {
367:func codexReviewTextIsBlank(s string) bool {
855:	if reviewText == "" {
```

I replayed the RED by restoring that implementation into the current tree:

```
$ git show a5f83fd62:internal/cli/mcp_codex.go > internal/cli/mcp_codex.go
$ go test ./internal/cli/ -run 'TestCodexBlankReview|TestCharacterize_UnavailableBackend'
--- FAIL: TestCodexBlankReview_AC001_BlankBodyDoesNotSynthesizePass (0.00s)
    --- FAIL: .../space-only   --- FAIL: .../newline-only   --- FAIL: .../mixed-whitespace
--- FAIL: TestCodexBlankReview_AC003_BlankStructuredReviewDoesNotShadowAgentMessage (0.00s)
--- FAIL: TestCodexBlankReview_AC005_BlankAndUnavailableAreDistinguishable (0.00s)
--- FAIL: TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput (0.00s)
--- FAIL: TestCodexBlankReview_AC009_BlankItemDoesNotClobberRealOne (0.00s)
    --- FAIL: .../review:_real_then_blank   --- FAIL: .../agent:_real_then_blank
--- FAIL: TestCodexBlankReview_ThreeStateControlMatrix (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.948s
```

**It compiled and failed on behavior.** The failing set matches the recorded RED exactly
(AC-001, 003, 005, 007, 009, matrix), and the four preservation pins stayed green.
The RED is genuine; only the cited attribution SHA is imprecise (F5).

### 2.5 What was preserved

- **`bestCodexReviewText` preference ordering** — pinned at `codex_blank_review_test.go:159`:
  `bestCodexReviewText("structured wins", "agent loses") == "structured wins"`. Plan-phase
  residual risk F4 is now pinned, and mutant E confirms the site is live. **Verified.**
- **Backend-unavailable fail-open** — `TestCharacterize_UnavailableBackend` asserts the
  byte-identical summary `"codex unavailable: codex session start failed: codex binary could not be
  started"`, `Findings` length 0, and the single next-step
  `"fall back to the active auditor (claude)"`. PASS on my run. **Verified.**
- **The pinned synthesizer test** — `git diff --stat 3ac58b5a1..HEAD --
  internal/cli/codex_review_rpc_test.go` → empty. Line 119's `"": "pass"` row is untouched;
  production never reaches it because the guard intercepts first. **Verified byte-identical.**
- **Scope** — `git diff --stat 3ac58b5a1..HEAD` shows exactly one production file changed
  (`internal/cli/mcp_codex.go`, +69/−7); `synthesizeReviewOutput` and `codexUnrecognizedVerdict`
  appear nowhere in the diff hunks. **Verified.**

### 2.6 CHANGELOG and progress.md — no overclaim

All three [HARD] accuracy constraints hold in the shipped text:

| Constraint | Shipped wording |
|---|---|
| must NOT claim #1632 closed / gate enforced | "**[HARD] This does not close the gate hole, and the entry does not claim it does.**" … "Making a `required` codex gate actually **block** is #1632 axis 3, a sibling card" |
| must NOT claim codex-absent behavior changed | "**Backend-absent behavior is unchanged** — an unavailable codex still fails open to `inconclusive` by design, and only blankness was reclassified" |
| must NOT describe as "fail-closed" | "The verdict is `inconclusive`, not `fail`, and the SPEC's `FAILCLOSED` name refers to closing the loophole rather than to adopting a blocking verdict" |

I re-verified the convergence claim rather than inheriting it. `internal/cli/mcp_convergence.go`:

```go
	required := filterRequired(verdicts)
	requiredFails := filterVerdict(required, "fail")
	switch {
	case len(requiredFails) > 0:
		overall = overallVerdictFail
	case allPass(required):
		overall = overallVerdictPass
	default:
		overall = claudeVerdictOrDefault(verdicts, overallVerdictPass)
	}
```

A required codex `inconclusive` produces no `requiredFails` and makes `allPass(required)` false, so
control reaches `default:` → the claude anchor. With a required claude `pass`, `overall_verdict` is
`pass`. `TestCodexBlankReview_M5_...` measures exactly this and PASSes. **The non-closure claim is
true as written.**

`progress.md` §E.4 additionally states its own gap honestly: *"No test, build, lint, or coverage
command was re-run in this sync phase."* That is accurate and is the right disclosure.

### 2.7 Cross-platform and lint

```
$ go build ./...                              → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...    → exit 0
$ go vet ./internal/cli/...                   → exit 0, silent
$ golangci-lint run --timeout=5m ./internal/cli/...
internal/cli/todo.go:97:13: Error return value of `fmt.Fprintf` is not checked (errcheck)
1 issues:
* errcheck: 1
```

Both halves of the claimed baseline confirmed: **one pre-existing `errcheck` at `todo.go:97`
(inherited, card t536/t577 axis), NEW findings 0.**

### 2.8 Coverage of the changed code

```
$ go tool cover -func=… | grep <changed funcs>
mcp_codex.go:333	inconclusiveReviewWithSummary	100.0%
mcp_codex.go:352	blankReviewInconclusive		100.0%
mcp_codex.go:367	codexReviewTextIsBlank		100.0%
mcp_codex.go:825	runTurn				 71.4%
mcp_codex.go:1136	awaitCodexTurnReview		 65.5%
mcp_codex.go:1307	bestCodexReviewText		100.0%
```

Every function this card introduced is fully covered by this SPEC's own suite alone.

## 3. Baseline-attribution

Every figure above was measured in **this** run, in **this** tree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t551`) at HEAD `a4ff25f74`, on darwin/arm64.
Nothing was read out of `green-evidence-20260908.txt`, `full-suite-final-20260908.txt`, or
`lint-postrepair.txt`; those files were compared against my own output, not substituted for it.
The RED replay was measured against the implementation blob of `a5f83fd62` restored into this tree
and reverted immediately after. The plan base for all `git diff` ranges is `3ac58b5a1`.

## 4. Gaps — what this audit did NOT observe

- **No live `codex` binary was exercised.** Every fixture drives `runCodexReviewRPC` over a stubbed
  connection. Whether real codex-cli 0.146.1 ever emits a whitespace-only body is unmeasured here,
  exactly as the SPEC itself records.
- **linux and windows runtime cells are unmeasured.** `GOOS=windows` compiles but `go build` does
  not compile test files, so windows test compilability is unverified. Cross-platform runtime
  verdicts belong to CI, and this lane does not push — **no CI verdict exists for this work.**
- **`go test ./...` was not run**, per repository discipline; only `./internal/cli/...`. Packages
  outside that tree are unmeasured by me.
- **`-race` was not run** on the changed code. The change adds no concurrency.
- **The 4350-line `spec-lint-wholetree*.txt` evidence files were not read**; SPEC-lint conformance
  is outside this audit's scope and I make no claim about it.
- **Mutation coverage is not exhaustive** — 8 mutants over the four repaired sites plus the
  summary/discriminator, not a systematic mutation run.

## 5. Residual-risk

- **U+200B is a live, measured bypass of the blankness check.** My probe (temporary test, since
  removed) observed: `unicode.IsSpace(U+200B) = false`, `codexReviewTextIsBlank(ZWSP-only) = false`,
  and end-to-end `verdict="pass" summary="​​"`. That is the same laundering shape the card
  repairs, surviving on a different axis. It is honestly disclosed in both `progress.md` and the
  CHANGELOG, and no criterion pins it.
- **Three surviving mutants mean the suite is weaker than its own narrative claims** — specifically
  around the guard site and the summary wording. A future edit to either can pass CI while
  regressing the design intent the code comments state.
- **The convergence layer still does not block**, by design and by measurement. A required codex
  gate over a passing claude yields `pass`. The honest `inconclusive` this card produces is
  visible in `FailOpenBackends` and `GateUnmet` but is not enforced — issue #1632 stays open.

---

## Findings (structured defect-list)

- **F1** [Medium] [optional] `internal/cli/codex_blank_review_test.go:222-231` (AC-CBR-005) —
  **Mutant C1 SURVIVED.** Rewriting `codexBlankReviewSummary` to `"codex unavailable: blank review
  output, no verdict text was produced"` passes the entire suite: AC-CBR-005 checks only string
  inequality plus the substrings `"blank"` and `"codex unavailable"`, and the mutant satisfies all
  three. The design intent stated in the code's own comment at `mcp_codex.go:331` — *"prefixing that
  with 'codex unavailable' would make the two states indistinguishable"* — is therefore unpinned.
  *Required fix (follow-up):* add
  `if strings.HasPrefix(blankOut.Summary, "codex unavailable: ") { t.Error(...) }` to AC-CBR-005.
  Not blocking: the shipped summary is correct; this is test strength, not behavior.

- **F2** [Medium] [optional] `internal/cli/mcp_codex.go:855` (the guard) — **Mutant D SURVIVED.**
  Reverting the guard to `if reviewText == ""` passes every test. The reason is structural: with the
  collection sites blank-aware, a whitespace-only body is never stored, so `reviewText` reaches the
  guard as exactly `""` and the old test catches it. The guard's blank-awareness is unreachable-as-
  blank on the only producing path (`awaitCodexTurnReview` → `bestCodexReviewText`). Consequence:
  `acceptance.md` AC-CBR-003's assertion that the guard is one of the load-bearing repair sites is
  not measurable, and AC-CBR-001's discriminating power rests on the collection sites, not the
  guard. *Required fix (follow-up):* either add a unit-level criterion calling `runTurn`'s guard with
  a blank-but-non-empty `reviewText` directly, or record in `acceptance.md` that the guard is
  defense-in-depth rather than independently load-bearing. Not blocking: the shipped code is
  correct and the redundancy is safe.

- **F3** [Low] [optional] `internal/cli/mcp_codex.go:1184,1187` (collection) — **Mutant A
  SURVIVED.** A first-**non-blank**-wins variant passes the suite, because it differs from the
  shipped last-non-blank-wins only when **two real non-blank items of the same type** arrive, a case
  no criterion covers. The shipped semantics (last real item wins) are therefore unpinned.
  *Required fix (follow-up):* add an AC-CBR-009 row `real-A then real-B` asserting the summary is
  `real-B`, or state in `acceptance.md` that multi-real ordering is deliberately undefined.

- **F4** [Low] [optional] `internal/cli/mcp_codex.go:367` (`codexReviewTextIsBlank`) —
  **U+200B measured bypass.** Probe output: `codexReviewTextIsBlank(ZWSP-only) = false`, end-to-end
  `verdict="pass"`. Same failure class as the repaired defect (an invisible body reported as a
  clean review), different axis (`unicode.IsSpace` classification rather than exact equality).
  **Judgement: acceptable residual risk for this card.** The card's requirement set (REQ-CBR-001..009)
  is scoped to the exact-equality defect; the gap is disclosed in both `progress.md` and the
  CHANGELOG rather than silently dropped, which is the correct handling. *Required fix (follow-up
  card):* widen the discriminator to `strings.TrimFunc(s, func(r rune) bool { return
  unicode.IsSpace(r) || unicode.Is(unicode.Cf, r) }) == ""` and add the ZWSP fixture to
  `blankFixtures`.

- **F5** [Low] [optional] `.moai/reports/t551/red-evidence-20260908.txt` (trailing SHA) —
  **attribution imprecision.** The file cites `a87a7d8a7e033ac040b6b8e66d47112c425ad8cb`, but that
  commit's tree contains none of `codexReviewTextIsBlank` / `codexBlankReviewSummary` /
  `blankReviewInconclusive`, so replaying the acceptance suite at that SHA would fail to compile
  rather than reproduce the recorded RED. The cited value is the run-time HEAD with the new files
  uncommitted; the reproducible unrepaired tree is `a5f83fd62`. The recorded RED **output is
  truthful** — I reproduced it exactly against `a5f83fd62`'s implementation. *Required fix
  (follow-up):* record the reproducing tree (`a5f83fd62`) alongside the run-time HEAD, or note that
  the run carried uncommitted working-tree changes.

**No blocking findings.** All five are optional under the finding-consumption discipline: F1/F2/F3
are test-strength gaps over correct shipped behavior, F4 is a disclosed residual risk on a
different axis from the card's requirement set, and F5 is an evidence-annotation imprecision whose
underlying claim I independently reproduced.

---

## Recommendations

1. Open a follow-up card for **F4 (U+200B)** — it is the only finding that names a live behavioral
   bypass, and the fix is one line plus one fixture.
2. Fold **F1 and F2** into that same card as test-strength additions; both are two-line assertions
   and both defend claims the code comments already make.
3. Leave **F3 and F5** as recorded findings; neither justifies its own card.
4. The lead should note that this branch is unpushed and **no CI verdict exists** for it. The
   darwin/arm64 evidence above is an early signal, not the integration verdict; re-measure in the
   merge tree per §4.1.

---

*Auditor: sync-auditor (independent). Verdict authority: this agent. Evidence: this file.
Tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t551` @ `a4ff25f74`, 2026-09-08.*
