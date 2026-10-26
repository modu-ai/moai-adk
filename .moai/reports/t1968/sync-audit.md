auditor-model: glm-5.3

# sync-audit — t1968 (SPEC-FETCH-WEB-001), iteration 1

verdict: FAIL
audited_sha: 537cfac1489a18953f4b03c0161b4968726288f6

Tree: lane worktree `.moai/worktrees/t1968` (branch `WT-jevfetch-a2sys-ai`). Merge base `2a1898ea22`. Date 2026-10-27. Audit plan (`moai verify audit-plan --project-root <toplevel>`): `config_status: ok`, `cross_model_active: true`, gates claude=off / **codex=required** / glm=off (source `config.gates`, `plan_source: config`).

## Verdict

**FAIL — aggregate 0.79** (harmonic mean of 72 / 68 / 90 / 92).

| Dimension | Weight | Score | Verdict | Evidence (verbatim, this run, this tree @537cfac148) |
|---|---|---|---|---|
| Functionality | 40% | 72 | **FAIL** | 14/14 AC tests re-observed green (`ok github.com/modu-ai/moai-adk/internal/fetchweb 0.164s` / `internal/jevmeasure 0.400s` / `internal/jev 1.704s`; cli graded trio + MCP parity + bench legs all `--- PASS`), all 15 repair seams present in code — BUT F2/F3/F4/F5 are requirements-conformance defects the AC suite does not cover (REQ-FW-006/007/001). |
| Security | 25% | 68 | **FAIL** | Off-gate zero-network re-observed (`--- PASS: TestFetchWeb_OffGateZeroNetwork`), no new deps (`git diff 2a1898ea22...HEAD -- go.mod go.sum` empty), `grep -rn tiktoken` 0 hits, no hardcoded credentials — BUT F1 (P1, SSRF/exfil on the activated agentic surface) is a High finding. |
| Craft | 20% | 90 | PASS | `coverage: 95.2%` (fetchweb) / `93.2%` (jevmeasure) / `90.2%` (jev); `golangci-lint run` → `0 issues.`; `gofmt -l` → empty (exit 0); RED-first evidence per milestone (§E.8); gaps recorded honestly. |
| Consistency | 15% | 92 | PASS | `moai spec lint SPEC-FETCH-WEB-001` → `✓ No findings — all SPEC documents are valid` (exit 0); naming quartet verified in shipped templates; template-neutrality guards `--- PASS`; sync commit `be1e8b7e2c` carries full SPEC-ID subject + `3-phase close` + `Authored-By-Agent: manager-docs`. |

Must-pass firewall: Functionality (72) and Security (68) both below pass threshold → FAIL regardless of the other dimensions. The required codex backend also returned **fail** (convergence `overall_verdict: fail`), which independently forces FAIL per the cross-model contract.

## Cross-model convergence (audit_multi, target=baseBranch, project_root=this worktree)

Digest (for the orchestrator's check — write to `<toplevel>/.moai/state/audit-plan-result.json` and run `moai verify audit-plan --project-root <toplevel> --result-file <path>`):

- `overall_verdict`: **fail**
- `gate_unmet`: (empty — no unmet gate; codex answered)
- `plan_source`: `config`
- per_backend_verdicts: `backend=codex, gate=required, verdict=fail` (sole participant; claude/glm gates off per plan)
- `audit_receipt`: `rcpt-8d9725aaf38781dabcaffde4`
- `participant_count`: 1, `disagreement_flag`: null
- `residual_risk_note` (verbatim): `required-backend FAIL: codex`

**This verdict is not final until the orchestrator's convergence check passes.** I independently verified every codex finding against the source before folding it (traces below); none is noise.

## Findings (structured defect-list)

- **F1** [High] [blocking] `internal/fetchweb/fetch.go:37` — No IP/redirect policy: the injected default client follows up to 10 redirects with no validation, and `fetch_web` is an agentic MCP tool (URL sourced from model output, permission rule `Bash(moai fetch web:*)` pre-authorized). A chosen or redirected URL can reach loopback/private/link-local addresses (incl. cloud metadata `169.254.169.254`), and the fetched verbatim content is then forwarded to the external Jev scoring endpoint. Verified: no `CheckRedirect`, no dial-address check anywhere in the package. Required fix: validate the connected IP on every hop (dial `Control` hook + `CheckRedirect` re-validation), refuse loopback/RFC1918/link-local; alternatively an explicit leader/operator scope decision recording this as an accepted, documented boundary plus a follow-up card — the shipped-inert gate (default off) limits but does not remove the activated-surface risk.
- **F2** [Medium] [blocking] `internal/fetchweb/scoring.go:233` and `:272` — An OK scoring response with missing/empty answers maps unanswered question IDs to probability 0 (`probs[missing]` → zero value), so a judgment failure is silently reported as `no_match` exit 1. This violates the code's own stated contract (`scoring.go:50-52`: fetch_error "only for an unavailable scoring call — 'could not judge' is an error, not a no_match") and REQ-FW-007's status mapping. Verified: `answersByID` + direct map lookup, no completeness check in either stage. Required fix: after each `Ask`, assert every requested question ID is present in `res.Answers`; on any missing ID return a stage error → `fetch_error` exit 2.
- **F3** [Medium] [blocking] `internal/fetchweb/blocks.go:292-294` — The `listItemSkip` branch returns without a separator (unlike the chrome branch at :289 which writes `' `'), so text runs adjacent to a skipped subtree inside a list item merge across the hole: `<li>First<ul><li>Second</li></ul>Third</li>` → block "FirstThird" followed by "Second" — word boundary lost AND document order corrupted. REQ-FW-006 verbatim/order guarantee violated. This is the residual the C2 repairer recorded ("`a<pre/>b` → `ab`", card-review.md post-ceiling note); codex independently re-found it. Required fix: write a boundary separator in the skip branch, mirroring the chrome branch.
- **F4** [Medium] [blocking] `internal/fetchweb/blocks.go:343` (`preText`) — Chrome exclusion was added to `collapseText` (N3) but not to `preText`: `<script>`/`<style>` subtrees inside `<pre>` leak their raw text into emitted code blocks (`safe<script>hidden</script>end` → "safehiddenend"). REQ-FW-001 chrome removal violated on the code path; a seam the N3 repair missed. Required fix: skip `chromeTags` subtrees in `preText`'s visitor while preserving interior whitespace/line structure.
- **F5** [Medium] [blocking] `internal/fetchweb/fetch.go:54` — No charset handling: the raw body is treated as UTF-8. A `charset=iso-8859-1` (or EUC-KR — common for Korean legacy pages) page returns mojibake/replacement chars as "verbatim" text with `passthrough` exit 0. REQ-FW-006 verbatim guarantee violated for a real class of pages. Required fix: decode per `Content-Type` charset (falling back to the HTML meta declaration) to UTF-8 before splitting; where decoding fails, `fetch_error` rather than silently corrupted output.
- **F6** [Low] [optional] `internal/cli/mcp_fetch_web.go:28` — `registerFetchWebTool` shows 0.0% coverage under the fetch-web test filter (registration wrapper exercised only by server-construction tests). Informational; no action required for landing.

All five codex findings (P1→F1, P2→F2/F3/F4/F5) were reproduced by source trace in this audit; the defect list is confirmed, not merely relayed.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim 1 — graded trio (AC-FW-007/008/004) green.** Evidence: `go test -run 'TestFetchWeb_OffGateZeroNetwork|TestFetchWeb_E2EFakeServers|TestFetchWeb_StatusExitContract|TestFetchWeb_StaleCacheWarningSurfaced|TestFetchWeb_NoLiveEndpointInTests|TestFetchWebTool|TestFetchWebCommand_Surface' -v -count=1 ./internal/cli/` → `--- PASS: TestFetchWeb_OffGateZeroNetwork`, `--- PASS: TestFetchWeb_E2EFakeServers`, `--- PASS: TestFetchWeb_StatusExitContract` (7 sub-tests incl. both `no_match` branches), `--- PASS: TestFetchWeb_StaleCacheWarningSurfaced`, 4× `TestFetchWebTool_*` PASS, `ok github.com/modu-ai/moai-adk/internal/cli 0.971s`. Baseline: this run, this tree @537cfac148.

**Claim 2 — all 15 post-close repairs (R1-R8, N1-N4, C1-C3) present in code and test-green.** Evidence: source trace — R1 `batchQuestions` (scoring.go:285), R2 `textBoundaryTags` incl. `br` (:321), R3 `itemDepth` guard (blocks.go:161-181), R4 `trimBlankLines` (:362), R5 `[]rune` context (scoring.go:424), R6 slot held through `decode(resp)` (jev.go:733-751), R7 `warningSink` → `Result.Warnings` → stderr (scoring.go:230/269, fetch_web.go:147), R8 `anyFail` → exit 1 (jevmeasure_cmd.go:157-166), N1 `stage1Windows` request accounting (scoring.go:465-492), N2 response-model keying (jev.go:637-659), N3 chrome skip in `collapseText` (:285-291), N4+C1 true-copy aside save (:158), C2 `listItemSkip` incl. pre/table (:335), C3 `hasRun` existence check (bench.go:193-208). Tests: 13 targeted repair tests `--- PASS` across fetchweb/jevmeasure/jev + `TestJevmeasureBench_*` 7 PASS in cli. Baseline: this run, this tree @537cfac148. **Caveat: two repairs left sibling seams open — N3 missed `preText` (F4), C2's skip path lacks the separator (F3).**

**Claim 3 — boundary / requirements spot-checks.** Evidence: `git diff --name-only cb30e05b3e...HEAD -- internal/jevledger` → empty, positive control `-- internal/jev` → 3 files (cache.go, cache_test.go, jev.go) — REQ-FW-016 intact, control non-empty. REQ-FW-002: package note + per-threshold "approximate" comments (blocks.go:5-10,54-59; scoring.go:14-37). Naming: `settings.json.tmpl:569 "Bash(moai fetch web:*)"`; `glm-web-tooling.md:40 "jev active → fetch_web first"`; skill dir present (4 content hits); no jev `enabled: true` in this SPEC's shipped artifacts. CHANGELOG cites 14 live ACs (AC-FW-001..014), matching acceptance.md §F.1. Baseline: this run, this tree @537cfac148.

**Claim 4 — craft/consistency gates.** Evidence: coverage / lint / gofmt / spec-lint outputs quoted in the dimension table; `go build ./...` exit 0 and `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Baseline: this run, this tree @537cfac148.

**Gaps (explicitly NOT observed):**
- The full `internal/cli` suite was not run in this audit (repo §8 discipline: heavy local runs discouraged; lane record shows 10m-timeout under foreign load with 0 test FAILs). The repository-wide test verdict is owned by integration-branch CI and was **not observed** here. Judged acceptable for lane-local discipline — not a blocker — but it remains an unobserved surface.
- Convergence covered codex only (claude/glm gates off per the plan); no second independent model verdict exists beyond codex.
- The orchestrator's `moai verify audit-plan --result-file` check has not yet run; this verdict is not final until it passes.

**Residual-risk:**
- `required-backend FAIL: codex` (convergence residual_risk_note, verbatim).
- Even after F1-F5 are repaired, the verbatim pipeline's fidelity guarantees hold only for UTF-8, chrome-per-`collapseText` documents; exotic encodings/markup (e.g. CDATA-era HTML) remain untested territory.
- CI on the integration branch may surface failures invisible to this lane-local audit (full cli suite never observed green in one local run).

## Iteration history

- iter1 (this audit): FAIL 0.79 — F1-F5 blocking (F6 optional). 14/14 AC re-observed green; 15/15 repairs present; conformance defects F2-F5 sit outside the AC suite's coverage, F1 outside the SPEC's stated requirements but High on the activated agentic surface.

## Recommendations

1. Repair F2 and F4 first (smallest diffs, pure contract/seam gaps: completeness check; chrome skip in `preText`), with RED-first tests per lane TDD discipline.
2. F3: separator in the skip branch + golden test for `<li>text<ul>…</ul>text</li>` order.
3. F5: charset decode or explicit refusal; include a latin-1 golden case.
4. F1: IP-validation on dial + redirect (leader may alternatively take an explicit scope decision with a follow-up card — that disposition belongs to the leader/operator, not the lane).
5. Re-run this sync audit as a delta scoped to F1-F5 after repair; the confirming re-audit is scoped to the enumerated defect delta.
