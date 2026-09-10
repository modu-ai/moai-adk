# SPEC-RED-NOW-THRESHOLD-001 — Progress

Card: t343 · Branch: `WT-red-now-threshold` · Tier: M
Plan-phase measurement tree: `a6bbbf82b` (every ledger entry E-01..E-17) · Current tree: `15453140a` (origin/develop absorbed, fast-forward, zero conflicts)

## §E.1 Plan-phase Audit-Ready Signal

- `plan_complete_at`: 2026-08-29
- `plan_status`: audit-ready (iteration cap reached; a third audit is the operator's call)
- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set)
- Requirements: **15** (REQ-RNT-001..015) · Acceptance criteria: **16** (13 release-blocking,
  3 regression-guard). Tier M ceiling 16/16 — the AC axis sits **exactly at** the ceiling; a
  further criterion requires tiering up or splitting.
- Measured: `grep -c "^\*\*REQ-RNT-" spec.md` → 15; ledger E-12 → 16; E-13 → 13; E-14 → 3.
- Open decisions: none.
- Known residuals, carried openly: mutant M-3 (survives inside a span), mutant M-6 (a command
  whose premise is false — not caught by L2), and the Go test's own non-execution (tool boundary).

## Audit history

| Iteration | Score | Threshold | Verdict |
|---|---|---|---|
| 1 | 0.75 | 0.80 | FAIL — 8 blocking defects (D1..D8) |
| 2 | **0.800** | 0.80 | FAIL — 1 critical blocking defect (N1); score axis cleared |

**0.75 → 0.800 per-dimension delta** (iteration 2 vs 1):

| Dimension | Iter 1 | Iter 2 | Driver |
|---|---|---|---|
| Clarity | 0.75 | 0.80 | "the command" defined; line numbers removed from REQ-RNT-011 |
| Completeness | 0.75 | 0.80 | continued-firing split into two axes; REQ-RNT-013/014 added |
| Testability | 0.50 | 0.60 | five green paths span-scoped; commands moved to the fenced ledger |
| Traceability | 1.00 | 1.00 | unchanged |

Iteration 2 recorded **zero false completion claims** (every asserted closure reproduced under
re-execution), **zero monotonicity violations**, and MP-8 self-applied reproducing 12/12.
Closure: 7 closed (D1, D2, D3, D6, D8, D9, D10), 3 partially closed (D4, D5, D7).
Reports: `.moai/reports/t343/plan-audit-iter1.md`, `.moai/reports/t343/plan-audit-iter2.md`.

## Iteration-2 defects (N1..N7) — what each closed

| Defect | Sev | Closed by |
|---|---|---|
| **N1** two correctives cancelled; scope predicate reached zero commands | critical | REQ-RNT-001 admits the **ledger as a named carrier** in the requirement layer; REQ-RNT-008's predicate made **carrier-independent** (cell / ledger entry / any fenced block); AC-RNT-008 gains `TestCommandScopeIsCarrierIndependent` + fixture `testdata/red_now/ledger/` + relocation mutant; **M-5** recorded in §D.2; `acceptance.md` §D.5 restated — the previous "regression-guard cells are inside the gate" claim was false and is replaced by a claim that holds |
| **N2** three/four element mismatch | major | REQ-RNT-002 and `plan.md` M1 both now say four (command, stdout, exit code, tree SHA) |
| **N3** AC-RNT-001 overclaim | major | Weakened to what holds: span-scoping defeats a token outside the span, **not** inside it (~41 lines); residual recorded as M-3 |
| **N5** expensive-command hole | major | REQ-RNT-013 gains a timeout disposition reusing the auditor's **existing** Bash bound — refusal + demotion, no new regime |
| **N6** broken cross-reference | minor | Boundary stated in §D.3 where it holds, with the reason no `spec.md` exclusion carries it |
| **N7** self-matching count command | major | Counts moved to ledger E-12/E-13/E-14, anchored `^\| \*\*AC-RNT-`; measured 16 / 13 / 3 |

## Correction round (post-iteration-2, coordinator-relayed)

| Item | Closed by |
|---|---|
| **C1** §A.2's refuting command named an existing test and proved nothing (survived both audits) | Evidence replaced with `-run TestMigrationParityDoesNotExistXYZ` → `ok … [no tests to run]`, exit 0. **Conclusion unchanged, evidence replaced**, recorded in new §A.2.1 rather than swapped silently |
| **C2** framed as one cell | Reframed to the measured census: **nine** criteria on one false premise; AC-TOSQ-011 excluded and asserted neither defective nor sound |
| **C3** L2 verdict could be read off an `ok` token | **REQ-RNT-015** + **AC-RNT-015**: verdict keys on executed-test count. `internal/hook/evidence_writer.go` `deriveFromOutputText` cited as the reason (`hasPrecisePass` returns before inspecting any count); the file is **not touched** — t341's surface |
| **C4** cross-card coordination | `plan.md` §H extended: nine-cell family, both lane-5 findings credited, reciprocal-citation note |
| **C5** failure chain unrecorded | **§A.2.2**: four actors, rule known and actively applied, defect survived every human-judgment layer → the corrective must be mechanical. Bounded honestly — **MP-8 would not have caught this incident either** (mutant **M-6**) |

## Monotonicity

No criterion was weakened, removed, or reclassified in either round. The same three
regression-guards persist (AC-RNT-002, -011, -012). N3 and C1 weaken *claims*, not criteria: N3
drops an immunity assertion while keeping every predicate; C1 replaces evidence while keeping the
conclusion. Scope grew monotonically: 12 → 14 → **15 REQ**, 13 → 15 → **16 AC**.

## Audit debt — `0.800` does not describe the current text

[HARD] **The `0.800` in the table above is the plan-auditor's score for the iteration-2 text, not a
measurement of the artifacts as they now stand.** SSOT for this record: `spec.md` §E.

- **Audits run since the last revision: none.** Iteration 2 returned **FAIL** at 0.800; the SPEC
  was then revised to close one **critical** defect (N1) and four **major** ones (N2, N3, N5, N7),
  plus a minor (N6), followed by a correction round that replaced the §A.2 evidence, reframed the
  census from one cell to nine, and added REQ-RNT-015 / AC-RNT-015. None of that has been audited.
- **Cap reached, no third iteration.** `.moai/config/sections/harness.yaml:77` sets
  `plan_audit_tier_ceilings: M: 2`. The operator ruled to accept within the cap; the coordinator's
  own recommendation was a third audit, and the ruling went the other way. The cost is carried
  explicitly rather than absorbed.
- **N1's closure was confirmed by the author and the coordinator — not by an independent audit.**
  That distinction is the point of the debt. §A.2.1 is the standing counter-example: a claim that
  survived two independent re-executing audits and was still wrong.
- **Reading `0.800` as this SPEC's current quality is the error the SPEC prohibits** — a value
  detached from what it measured. Attaching it to its subject is the same discipline REQ-RNT-001
  imposes on every RED cell.

**Not debt — design facts.** Mutant **M-6** (MP-8 re-executes a command but does not verify the
command measures its stated premise) is a deliberate, recorded boundary of the three-layer
mechanism, not an unpaid item; it is not scheduled for closure, and REQ-RNT-015 narrows only the
`ok … [no tests to run]` case by design. Mutant **M-3** and the Go test's own non-execution are
carried residuals, named where they hold. Closing any of them would add criteria against a 16/16 AC
ceiling — a tier decision belonging to the lead.

## Cross-SPEC state

`SPEC-TODO-LANDING-STATE-001` (card t331), the source of the release-blocking / regression-guard
precedent, has **landed** — `status: completed`, resolvable at
`.moai/specs/SPEC-TODO-LANDING-STATE-001/` on `15453140a`. Its §C sentence was re-read from the
landed copy and is byte-identical to the pre-landing quote (`acceptance.md:93-95`). It is now
citable by path; it remains a precedent, not a dependency.

## §E.2 Run-phase Evidence

Run-phase tree: `15453140a` → `f60403c07` (M1 `3dc185382`, M2 `51d2935d3`,
M3 `5f0b9d7c1`, M4 `f60403c07`). Every figure below was produced by running the
named command in this worktree during this run; the plan-phase ledger was
measured on `a6bbbf82b` and was **re-measured here before any edit**
(`.moai/reports/t343/run-red-baseline.md`) — all eight cited REDs reproduced.

### Milestones

| M | Change | Command | Observed | Exit |
|---|--------|---------|----------|------|
| M1 | `verification-completeness.md` §2.1 — four elements, two carriers, structural-not-lexical, undecidable disposition | `grep -c "RED-now cell content" .claude/rules/moai/development/verification-completeness.md` | `1` (was `0`) | 0 |
| M2 | `plan-auditor.md` MP-8 + Group 4 AC-6 + report row | `grep -c "MP-8" .claude/agents/moai/plan-auditor.md` | `6` (was `0`) | 0 |
| M3 | `internal/spec/red_now_cell_test.go` + 3 fixtures | `go test ./internal/spec/ -count=1` | `ok github.com/modu-ai/moai-adk/internal/spec 36.588s` | 0 |
| M4 | mirrors + `make agents-emit` + `make build` | `make build` | exit 0 (`.moai/reports/t343/m4-make-build.txt`) | 0 |

RED before GREEN, captured verbatim: `.moai/reports/t343/m3-red.txt` (exit 1, 19
`FAIL` lines) is the test suite run **before** M1 and M2 landed;
`.moai/reports/t343/m4-mirror-red.txt` (exit 1) is the three mirror assertions
before M4.

### AC PASS/FAIL matrix

All commands below were run at `f60403c07` unless the row says otherwise.

| AC | Class | Status | Verification command | Actual output |
|----|-------|--------|----------------------|---------------|
| AC-RNT-001 | release-blocking | PASS | `go test ./internal/spec/ -run TestRuleClauseEnumeratesFourElements -count=1` | `ok  github.com/modu-ai/moai-adk/internal/spec` (exit 0) |
| AC-RNT-002 | regression-guard | PASS | `grep -c -e tense -e mood -e counterfactual -e "future.sense" .claude/rules/moai/development/verification-completeness.md .claude/agents/moai/plan-auditor.md` | `…verification-completeness.md:0` / `…plan-auditor.md:0`, exit 1 — unchanged |
| AC-RNT-003 | release-blocking | PASS | `go test ./internal/spec/ -run TestRuleClauseStatesDemotionNotPass -count=1` | `ok` (exit 0) |
| AC-RNT-004 | release-blocking | PASS | `go test ./internal/spec/ -run TestMP8SpanNamesReexecution -count=1` | `ok` (exit 0); the span is contained in the `### M5` section span |
| AC-RNT-005 | release-blocking | PASS | `go test ./internal/spec/ -run TestMP8SpanIsScoreIndependent -count=1` | `ok` (exit 0) |
| AC-RNT-006 | release-blocking | PASS | `go test ./internal/spec/ -run TestMP8SpanCarriesNABranch -count=1` | `ok` (exit 0) |
| AC-RNT-007 | release-blocking | PASS | `go test ./internal/spec/ -run TestGroup4AndReportRowExist -count=1` | `ok` (exit 0); `grep -c "AC-6:" …plan-auditor.md` → `1`, exit 0 (was `0`, exit 1) |
| AC-RNT-008 | release-blocking | PASS | `go test ./internal/spec/ -run 'TestCommandScopeIsCarrierIndependent\|TestMP8SentinelMutantsAreDetected' -count=1 -v` | `M-5 observed: cell-scoped=0 findings, carrier-independent=1 findings [line 17 (ledger-entry) … carries unquoted [\|]]`; zero-pair, two-pair and empty-span mutants each rejected (exit 0) |
| AC-RNT-009a | release-blocking | PASS | `go test ./internal/spec/ -run TestRedNowViolatingFixtureIsReported -count=1 -v` | `violating fixture findings: elements=[AC-VIO-001 (line 29): release-blocking RED-now cell carries no command, stdout and exit code] form=[line 20 (ledger-entry) … carries unquoted [\|]]` (exit 0) |
| AC-RNT-009b | release-blocking | PASS | `go test ./internal/spec/ -run TestRedNowLegitimateFixtureIsClean -count=1` | `ok` (exit 0) — zero findings on a non-empty command set |
| AC-RNT-010 | release-blocking | PASS | `go test ./internal/spec/ -run 'TestMP8MirrorSpanIsByteEqual\|TestRuleMirrorIsByteIdentical' -count=1` + `make build` | both `PASS`; `make build` exit 0 |
| AC-RNT-011 | regression-guard | PASS | `go test ./internal/spec/ -run TestMP8MirrorSpanIsNeutral -count=1` | `PASS` (exit 0); `grep -nE "SPEC-[A-Z]+-[0-9]{3}" internal/template/templates/.claude/agents/moai/plan-auditor.md` → the two pre-existing illustrative placeholders only (lines 474, 492) |
| AC-RNT-012 | regression-guard | PASS | `grep -rl "red_now" internal/template/templates/` | empty stdout, exit 1 — unchanged; `TestRedNowArtifactsDoNotShip` PASS |
| AC-RNT-013 | release-blocking | PASS | `go test ./internal/spec/ -run TestMP8SpanCarriesExecutionDiscipline -count=1` | `ok` (exit 0) |
| AC-RNT-014 | release-blocking | PASS | `go test ./internal/spec/ -run TestMP8LivenessAnchors -count=1` | `PASS` (exit 0) — the MP-8-row deletion mutant was observed failing the liveness predicate |
| AC-RNT-015 | release-blocking | PASS | `go test ./internal/spec/ -run TestMP8SpanKeysOnExecutedCount -count=1` | `ok` (exit 0) |

13 release-blocking PASS, 3 regression-guard PASS, 0 FAIL, 0 PASS-WITH-DEBT.

### Closure gates (`acceptance.md` §D.4)

1. Every release-blocking AC passes with its ledger id, command, output and exit code cited — above.
2. Every regression-guard AC still returns its stated green output — E-02 `0`/`0` exit 1, E-11 empty exit 1, E-10 two placeholders exit 0.
3. Both directions of REQ-RNT-009 observed — AC-RNT-009a and AC-RNT-009b, on separate fixtures.
4. Mutants observed failing, not argued — M-2 (three variants), the MP-8-row deletion, M-4 and M-5.
5. `make build` exit 0 and the mirror span comparison passes.

### Deviations, recorded rather than smoothed over

- **Sentinel spelling.** `plan.md` §F M2 spells the sentinel `# MOAI-REDNOW-BEGIN`.
  It is implemented as `<!-- MOAI-REDNOW-BEGIN -->`. A `# `-prefixed line in
  markdown prose is an H1 heading: it renders visibly and terminates the enclosing
  `### M5` section, which would break AC-RNT-004's containment assertion. The token
  is unchanged, so ledger entries E-07 and E-09 (`grep -c "MOAI-REDNOW-BEGIN"`)
  resolve exactly as written.
- **Two same-SPEC cascades.** The mirror edit invalidated the generated
  `.codex/agents/moai/plan-auditor.toml` (closed by `make agents-emit`) and the
  template catalog hash in `internal/template/catalog.yaml` (closed by `make build`).
  Both are inside the scope envelope; neither is a scope expansion.
- **The plan artifacts were untracked** in this worktree and are committed by M1.

### Boundaries reached during the run, stated where they hold

- **L1 checks the single-invocation form, not read-only-ness.** The mechanical half
  is the extracted metacharacter list; whether a conforming command only reads is
  the auditor's judgment under MP-8's execution-discipline branch. This is stated
  in the MP-8 clause rather than left implied.
- **The quote-aware scan treats a backtick or `$(` inside double quotes as quoted.**
  A real shell expands both there. Narrowing this would need a shell parser; the
  disposition for a refused command is demotion, not an error, so the direction of
  the residual is permissive rather than dangerous.
- **Applied to this SPEC's own `acceptance.md`** (throwaway probe, not committed):
  20 commands collected across two carriers, **zero element findings** — every
  release-blocking row resolves to a ledger entry carrying command, stdout and exit
  code, with the document pin inherited. Three form findings, all on the §D.0.1
  divergence-probe illustration lines, whose `<fixture-A>` / `<fixture-B>`
  placeholders read as redirection to the scanner. That is the form check being
  literally correct on an illustrative line rather than a real citation. It is
  reported, not repaired: repairing it would require either editing `acceptance.md`
  body content (not this agent's artifact) or narrowing the scanner in a way that
  re-opens M-5.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-29
run_commit_sha: "dc817409f"   # a commit cannot cite its own SHA; backfilled in the sync commit
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0   # no file outside the M1-M4 envelope was modified
l44_pre_commit_fetch: "not run — this worktree is a lane tree; the lead owns integration and no push was performed"
l44_post_push_fetch: "not applicable — no push performed (the lead owns integration)"
new_warnings_or_lints_introduced: 0   # golangci-lint run --timeout=5m ./internal/spec/... -> "0 issues.", exit 0
cross_platform_build:
  darwin: "go build ./... -> exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... -> exit 0"
coverage:
  internal/spec: "89.1% of statements (go test -cover ./internal/spec/ -count=1, exit 0)"
total_run_phase_files: 25   # git diff --name-only 15453140a..f60403c07 | wc -l
m1_to_mN_commit_strategy: "one commit per milestone, M1..M4, each naming card t343; no push"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-29
sync_commit_sha: "e0bc3c1f1"   # a commit cannot cite its own SHA; backfilled in the immediately following commit
sync_status: complete
run_commit_sha_backfill: "dc817409f — replaced the prose placeholder that previously occupied the §E.3 run_commit_sha slot (progress.md:197). Verified: `git cat-file -t dc817409f` -> commit; `git merge-base --is-ancestor dc817409f HEAD` -> exit 0; `git branch --contains dc817409f` -> WT-red-now-threshold. The placeholder token is deliberately not quoted anywhere in this directory, so a residue grep returns zero rather than matching this record"
b12_self_test_a: "grep -c 'SPEC-RED-NOW-THRESHOLD-001' CHANGELOG.md -> 0, exit 1 (pre-emission; no duplicate entry from a parallel session)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u -> 17 tokens. 17 is NOT the criterion count and is not what the CHANGELOG states. Two tokens (AC-5, AC-6) are plan-auditor Group 4 checklist identifiers quoted inside AC-RNT-007's cells, not criteria of this SPEC; and AC-RNT-009 is carried as two lettered criteria (AC-RNT-009a / AC-RNT-009b) that the digit-terminated pattern collapses into one. 17 - 2 + 1 = 16, which matches acceptance.md §D.1 (13 release-blocking + 3 regression-guard) and §E.2's 16-row PASS matrix. The CHANGELOG entry states 16"
b12_self_test_c: "ls on every path claimed in the entry -> all present: .claude/rules/moai/development/verification-completeness.md (local + template mirror), .claude/agents/moai/plan-auditor.md (local + template mirror), internal/template/templates/.codex/agents/moai/plan-auditor.toml, internal/spec/red_now_cell_test.go, internal/spec/testdata/red_now/{violating,legitimate,ledger}/acceptance.md, internal/template/catalog.yaml"
changelog_entry_position: "CHANGELOG.md [Unreleased] -> ### Added, first entry; inserted above SPEC-MOVING-REF-GUARD-001"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (merged 3-phase close on this sync commit); updated: 2026-08-29"
  plan_md: "NOT transitioned - carries NO frontmatter block at all (grep -c '^---$' -> 0)"
  acceptance_md: "NOT transitioned - carries NO frontmatter block at all (grep -c '^---$' -> 0)"
  progress_md: "NOT transitioned - carries NO frontmatter block at all (grep -c '^---$' -> 0)"
  sibling_rationale: "spec-frontmatter-schema.md § Artifact Statelessness: plan.md / acceptance.md / design.md / research.md are stateless on the status axis and MUST NOT carry `status:`. All three siblings here carry no frontmatter whatsoever, so there was no field to transition. Adding one to satisfy an 'all four artifacts' reading would have created the exact drift the rule forbids. Recorded rather than silently resolved"
mx_tag_validation:
  scanned: "internal/spec/red_now_cell_test.go — the only Go file this SPEC adds"
  result: "grep -n '@MX:' -> zero matches, and zero is correct here. The file declares no exported identifier (every symbol is a test function or an unexported helper in package spec), so the @MX:ANCHOR / @MX:NOTE triggers for exported surface and high fan_in do not fire. The reused-seam sibling internal/spec/ac_count_clause_test.go (card t338, landed) likewise carries zero. No tag was added, and the absence is a measured disposition rather than an unperformed step"
canary_compliance_check:
  applicable: true
  reason: "this SPEC defines a forward-looking policy — the four-element RED-now cell contract — and the run phase ran L1 against this SPEC's own acceptance.md as a self-application probe"
  result: "20 commands collected across two carriers, ZERO element findings — every release-blocking row resolves to a §D.0 ledger entry carrying command, stdout and exit code, with the document-level pin `a6bbbf82b` inherited. THREE form findings remain OPEN and are carried forward unrepaired: all three sit on the §D.0.1 divergence-probe illustration lines, whose `<fixture-A>` / `<fixture-B>` placeholders the scanner reads as shell redirection. That is the form check being literally correct on an illustrative line rather than on a real citation. Not repaired, deliberately: repairing means either editing acceptance.md body content (outside manager-docs' ownership) or narrowing the scanner in a way that re-opens mutant M-5. Reported as a finding, not smoothed away"
  boundary: "MP-8 re-executes a cited command; it does NOT verify the command measures its stated premise (mutant M-6, spec.md §A.2.2 / §E). MP-8 would NOT have caught the incident recorded in §A.2.1. This is a design fact, not a debt, and it is not scheduled for closure"
carried_residuals_still_open:
  m_3: "a token pasted INSIDE the ~41-line extracted §2 span still satisfies AC-RNT-001's element assertion; span-scoping defeats only a token pasted outside it (acceptance.md §D.2)"
  m_6: "see canary_compliance_check.boundary — design fact, not debt"
  go_test_non_execution: "the repository-local test's own non-execution is a tool boundary the SPEC carries openly; AC-RNT-014's liveness anchors narrow it but do not close it"
  self_application_form_findings: "the three §D.0.1 findings above"
audit_debt: "The plan-auditor score 0.800 (iteration 2, Tier M threshold 0.80, verdict FAIL on critical defect N1) describes the ITERATION-2 TEXT, not the artifacts as they now stand. Audits run since that revision: NONE. The final revision closed one critical (N1) and four major (N2, N3, N5, N7) defects plus a minor (N6), and a later correction round replaced the §A.2 evidence, reframed the census from one cell to nine, and added REQ-RNT-015 / AC-RNT-015 — none of which any audit has seen. The iteration cap is 2 (.moai/config/sections/harness.yaml:77, plan_audit_tier_ceilings: M: 2), it was reached, and the OPERATOR ruled to accept within the cap against the coordinator's own recommendation of a third iteration. N1's closure was confirmed by the author and the coordinator, which is not an independent audit. SSOT: spec.md §E. Reading 0.800 as this SPEC's current quality is the error the SPEC exists to prohibit"
plan_deviation_carried: "plan.md §F M2 specifies the sentinel `# MOAI-REDNOW-BEGIN`; the implementation uses `<!-- MOAI-REDNOW-BEGIN -->`. A `# `-prefixed line is an H1 in markdown and would terminate the enclosing `### M5` section, breaking AC-RNT-004's containment assertion. The TOKEN is unchanged, so ledger entries E-07 and E-09 (`grep -c 'MOAI-REDNOW-BEGIN'`) resolve exactly as written"
tests:
  affected_packages: "go test ./internal/spec/ -count=1 -> ok github.com/modu-ai/moai-adk/internal/spec 34.714s, exit 0 (.moai/reports/t343/sync-test-spec.txt)"
  full_suite: "NOT RUN locally, by instruction and by this repository's [HARD] local-full-suite prohibition. No CI verdict exists — this branch is unpushed"
lint: "golangci-lint run --timeout=5m ./internal/spec/... -> '0 issues.', exit 0 (.moai/reports/t343/sync-lint-spec.txt). The four LSP style hints on red_now_cell_test.go (SplitSeq x3, CutPrefix x1) are hints the configured linter does not raise; they were NOT fixed here — a code edit does not belong in a documentation-transition commit"
cross_platform_build:
  darwin: "go build ./... -> exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... -> exit 0"
mirror_parity: "diff -q on the verification-completeness.md pair -> exit 0 (byte-identical). plan-auditor.md legitimately DIFFERS from its mirror (neutralization), so parity there is asserted span-wise by TestMP8MirrorSpanIsByteEqual, not by diff. grep -c 'MOAI-REDNOW-BEGIN' -> 1 on each of the three carriers (local .md, template .md, codex .toml)"
docs_sync: "no user-facing doc surface enumerates the plan-auditor must-pass firewall. Scanned: README{,.ko,.ja,.zh}.md, docs-site/content/**, .moai/docs/**. README's three plan-auditor hits are agent-catalog rows (a mermaid node, a pipeline node, an evaluator table row) that name the agent, never its MP-N criteria; every docs-site 'Must-Pass' hit is the sync-auditor 4-dimension firewall (Functionality/Security), a different mechanism. grep for 'MP-1'/'MP-7' across docs-site/content -> zero. No README or docs-site edit was made, and none is owed"
spec_audit_pre_sync: "mcp__moai__spec_audit (project_root = this worktree) -> era V3R5, H-3 (§E.2 present, sync_commit_sha missing), severity INFO, zero drift findings. The V3R5 reading is the pre-sync state this section changes; re-classification to V3R6 follows the backfill commit that writes sync_commit_sha"
push_state: "not pushed, not merged, no PR. The lead owns the integration window and gives the order after this report"
```

**What this sync did NOT observe.** The branch is unpushed, so **there is no CI verdict** — the
full-suite judgment belongs to whoever integrates `WT-red-now-threshold`. The full suite was not
run locally, by instruction. No third plan audit was run, so `0.800` remains attached to the
iteration-2 text and nothing has graded the current one. The three self-application form findings
are open. `sync_commit_sha` is written by the immediately following commit.
