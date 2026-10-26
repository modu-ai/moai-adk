auditor-model: glm-5.3

# sync-audit — t1968 (SPEC-FETCH-WEB-001), iteration 1

verdict: FAIL
audited_sha: 80a54e3ef88d6b3ade91893f8c413cfc0644600b

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

---

# sync-audit — t1968 (SPEC-FETCH-WEB-001), iteration 2 (delta: F1-F5 @ 53ffddd499)

Delta audit of the F1-F5 repairs. Audited head `80a54e3ef8` (docs-only on top of `53ffddd499`; `git diff --name-only 53ffddd499..80a54e3ef8` → `.moai/reports/t1968/progress.md` only — confirmed). Audit plan re-read: `config_status: ok`, codex **required** (claude/glm off), `plan_source: config`.

## iter2 Verdict

**FAIL — aggregate 0.86** (harmonic mean of 85 / 76 / 91 / 92). Must-pass firewall: Security (76) below threshold → FAIL. Required codex backend independently returned **fail** (convergence `overall_verdict: fail`), which forces FAIL per the cross-model contract.

| Dimension | Weight | Score | Verdict | Evidence (verbatim, this run, this tree @80a54e3ef8) |
|---|---|---|---|---|
| Functionality | 40% | 85 | PASS (marginal) | `ok github.com/modu-ai/moai-adk/internal/fetchweb 0.577s`; graded trio re-observed (`--- PASS: TestFetchWeb_OffGateZeroNetwork`, `--- PASS: TestFetchWeb_E2EFakeServers`, `--- PASS: TestFetchWeb_StatusExitContract` 7/7 sub-tests, 4× `TestFetchWebTool_*`); `ok internal/jevmeasure 0.553s` / `ok internal/jev 1.714s` — but F5 retains a confirmed silent-substitution slice (F5b) and two pre-existing conformance corners surfaced (F7/F8). |
| Security | 25% | 76 | **FAIL** | F1 substantively repaired (dial-level Control on every resolved IP, 5-hop cap, per-hop scheme refusal, proxy disabled, initial-URL scheme check, harness seam loopback-only; all probe/ssrf-test verified) — BUT `refuseNonPublicIP` allows `100.64.0.0/10` (CGNAT/shared — Tailscale mesh range) and `255.255.255.255`: probe output `PROBE: 100.64.0.1 ALLOWED by refuseNonPublicIP` / `PROBE: 100.100.100.200 ALLOWED` / `PROBE: 255.255.255.255 ALLOWED` (F1b). No new deps (`git diff 2a1898ea22...HEAD -- go.mod go.sum` empty). |
| Craft | 20% | 91 | PASS | `coverage: 94.5% of statements` (fetchweb); `golangci-lint run` → `0 issues.`; `gofmt -l` → clean exit 0; `build-darwin-ok` + `build-windows-ok`. |
| Consistency | 15% | 92 | PASS | `moai spec lint SPEC-FETCH-WEB-001` → `✓ No findings` (exit 0); t1762 boundary intact (`git diff --name-only cb30e05b3e...HEAD -- internal/jevledger` empty, control `internal/jev` → 3 files); templates untouched since iter1 (`git diff 537cfac148..HEAD -- internal/template` empty); MX:ANCHOR added on `NewPageClient`. |

## iter2 F1-F5 dispositions (the enumerated delta)

- **F1 — repaired, one confirmed gap (F1b).** `ssrf.go` implements the canonical pattern: `dialControl` validates the post-DNS resolved IP at connect time on every hop (closes classic check-then-dial rebinding — validation IS the dial-time address; each redirect hop re-dials → re-validates); unparseable address and nil IP fail closed; `checkRedirect` caps at `MaxRedirectHops=5` and refuses non-HTTP(S) per hop; `t.Proxy = nil` (a proxy hop would void dial validation); `DialTLSContext = nil` rides DialContext so HTTPS dials are guarded; `FetchPage` refuses non-HTTP(S) initial URLs; `harnessIPPolicy` admits loopback ONLY (private/link-local/multicast stay refused — proven by the 192.168/169.254 redirect tests); off-gate semantics intact (gate read before `fetchWebPageClientFn` is consulted; REQ-FW-010 injection seam untouched); both production surfaces route through the single construction point (`fetch_web.go:42`, `mcp_fetch_web.go:74` → `RunFetchPipeline(..., fetchWebPageClientFn(), ...)`). 6/6 ssrf tests PASS incl. production-client loopback refusal. **Residual TOCTOU: none across the hook boundary** — Control fires per connection attempt on the exact address the kernel will connect to; the bypass that survives is a RANGE gap (F1b), not a timing gap.
- **F2 — repaired, verified.** Both stages: `p, ok := probs[id]; if !ok { return …("could not judge, not a no-match") }` (scoring.go:234-241, :282-288) → `scoreStream` error → `Result{Status: StatusFetchError}` (scoring.go:110) → exit 2. `TestProcess_Stage1AnswerlessIsError` asserts `StatusFetchError` + the distinguishing condition; stage-2 twin present. PASS observed.
- **F3 — repaired, verified.** Skip branch now writes `' '` (blocks.go:296-301); golden test `TestSplitBlocks_ListItemSkipKeepsWordBoundary` asserts exactly iter1's defect case → `First Third` + `Second` in order. PASS observed.
- **F4 — repaired, verified.** `preText` drops `chromeTags` subtrees with a boundary separator (blocks.go:352-355); `TestSplitBlocks_ScriptInsidePreIsChrome` PASS.
- **F5 — repaired for its core scenario, one slice survives (F5b).** BOM-strip+validate, Content-Type charset honored (`TestFetchPage_HonorsContentTypeCharset`), unknown label refused, strict UTF-8 fallback errors on invalid bytes — all PASS. BUT a declared non-UTF-8 charset whose bytes are corrupted substitutes U+FFFD silently instead of erroring (F5b below): the decoder's error path never fires for such bodies, so the "fetch_error rather than silently corrupted output" clause of the iter1 prescription is only partially delivered.

## iter2 Findings (all four independently reproduced by my own probe in this run, not relayed from codex)

- **F1b** [High] [blocking] `internal/fetchweb/ssrf.go:47` — `refuseNonPublicIP` allows the shared address space `100.64.0.0/10` (CGNAT; the Tailscale mesh range — reachable internal nodes on a common developer-machine configuration) and IPv4 broadcast `255.255.255.255`. Probe: `PROBE: 100.64.0.1 ALLOWED by refuseNonPublicIP` / `PROBE: 100.100.100.200 ALLOWED` / `PROBE: 255.255.255.255 ALLOWED`. Same threat class as F1 (fetched content forwarded verbatim to the external scoring endpoint), environment-gated reachability. Required fix: extend the refusal set with an explicit `100.64.0.0/10` CIDR contains-check (and broadcast/other reserved ranges at discretion), plus regression tests on initial and redirect hops.
- **F5b** [Medium] [blocking] `internal/fetchweb/fetch.go:102` — a declared non-UTF-8 charset (e.g. `charset=euc-kr`) whose bytes are truncated/corrupted decodes with silent U+FFFD substitution and returns `passthrough` exit 0. Probe: `PROBE euc-kr truncated: err=<nil> page="<html><body><p>Price <U+FFFD>"` + `PROBE VERDICT: silent U+FFFD substitution, no fetch_error`. The x/text decoder substitutes instead of erroring, so the declared-charset leg never reaches its error path — a surviving slice of F5's own prescription. Required fix: scan the decoded output for the U+FFFD substitution sequence and return `fetch_error` when it was produced by substitution (single-byte charsets like latin-1 cannot false-positive).
- **F7** [Medium] [blocking] `internal/fetchweb/blocks.go:174` — `<li>Main<aside><p>Note</p></aside></li>` yields one block `kind=list-item text="Main Note" aside=false` (probe-observed): the parent absorbs the aside's text (verbatim contamination) and the aside never forms its own block, so it is absent from the aside scoring stream. Pre-existing seam (present at iter1, not introduced by the repair), but requirement-level (aside stream + verbatim). Required fix: exclude `aside` subtrees from parent list-item extraction and keep the list-paragraph suppression state scoped per aside entry.
- **F8** [Medium] [blocking] `internal/fetchweb/blocks.go:395` — `<td><pre>  first\n    second</pre></td>` yields `kind=table-row text="first second"` (probe-observed): interior indentation and newlines collapse and no code block is emitted — the verbatim guarantee broken for code in table cells, the same class F4 fixed for `pre` under `body`. Pre-existing seam. Required fix: whitespace-preserving extraction for code blocks inside table cells, deduplicated against the row text.

## iter2 Cross-model convergence (audit_multi, target=baseBranch, project_root=this worktree)

Digest for the orchestrator's check (write to `<toplevel>/.moai/state/audit-plan-result.json`, then `moai verify audit-plan --project-root <toplevel> --result-file <path>`):

- `overall_verdict`: **fail**
- `gate_unmet`: (empty — codex answered)
- `plan_source`: `config`
- per_backend_verdicts: `backend=codex, gate=required, verdict=fail` (sole participant; claude/glm gates off per plan)
- `audit_receipt`: `rcpt-d145a5b0a39ff328c52489bb`
- `participant_count`: 1, `disagreement_flag`: null
- `residual_risk_note` (verbatim): `required-backend FAIL: codex`

**This verdict is not final until the orchestrator's convergence check passes.** Every codex finding was reproduced by my own probe before folding; the F1b/F5b/F7/F8 list is confirmed, not relayed.

## iter2 Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim — repairs landed and prior green set holds.** Evidence: `go test -count=1 ./internal/fetchweb/` → `ok github.com/modu-ai/moai-adk/internal/fetchweb 0.577s`; graded trio + MCP parity → `--- PASS` (12 named tests, `ok internal/cli 0.990s`); `ok internal/jevmeasure 0.553s`, `ok internal/jev 1.714s`; jevledger boundary empty with non-empty control; `go.mod`/`go.sum` diff empty; spec-lint `✓ No findings` exit 0; coverage `94.5%`; lint `0 issues.`; gofmt clean; darwin+windows builds ok; templates untouched since iter1. Baseline: this run, this tree @80a54e3ef8.

**Claim — the four new findings are real.** Evidence: transient in-package probe (`zz_audit_probe_test.go`, created, run, deleted; `git status --short` empty after) — outputs quoted inline in F1b/F5b/F7/F8. Baseline: this run, this tree @80a54e3ef8.

**Gaps (explicitly NOT observed):**
- Full `internal/cli` suite still not run here (unchanged from iter1; owned by integration-branch CI).
- Convergence covered codex only (claude/glm off per plan).
- The orchestrator's `--result-file` convergence check has not run; verdict not final until it passes.

**Residual-risk:**
- `required-backend FAIL: codex` (convergence residual_risk_note, verbatim).
- Proxy-required egress environments lose fetch capability entirely (`t.Proxy = nil` — deliberate, documented; correctness over reachability).
- Exotic UTF-8 alias labels outside `isUTF8Family` (e.g. `unicode-1-1-utf-8`) route through `charset.Lookup`'s replacement decoder — adjacent to F5b, fixed by the same substitution scan.
- DNS rebinding across the Control/connect boundary: no gap found — validation is on the connected address itself; the OS does not re-resolve between hook and syscall.

## Iteration history (cont.)

- iter2 (this audit): FAIL 0.86 — F1-F5 all substantively repaired and verified (F2/F3/F4 clean; F1 with range gap F1b; F5 with slice F5b); two pre-existing conformance seams surfaced (F7/F8). Delta for iter3: F1b, F5b, F7, F8 — all localized to ssrf.go / fetch.go / blocks.go, same pattern family as this round's repairs.
