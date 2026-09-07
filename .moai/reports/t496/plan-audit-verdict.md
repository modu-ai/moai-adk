# Plan-Audit Verdict — SPEC-CODEX-EVENT-COVERAGE-001 (card t496, iteration 1)

Auditor: plan-auditor (independent, adversarial stance). Audited artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` at `.moai/specs/SPEC-CODEX-EVENT-COVERAGE-001/`, committed at `b980c3920`.

## Verdict

**FAIL** — Overall Score **0.91** (Tier M threshold 0.80; all 7 must-pass criteria PASS, but 5 blocking-class defects remain — see D1–D5. All five are small, single-spot doc edits; iteration-2 re-audit should be scoped to this defect delta per the Retry Loop Contract. The t401 precedent (FAIL at 0.95 with outstanding blocking defects) applies: score above threshold does not offset blocking defects.)

## 1. Claim

The SPEC's plan artifacts are formally compliant (lint clean, GEARS-clean, traceability complete, must-pass 7/7) and every in-tree factual claim manager-spec made re-measured TRUE in this run — but five blocking defects exist: an AC that under-verifies its requirement while demanding removal of a phrase that stays true (AC-CEV-006), a plan-internal contradiction between M1's file list and AC-CEV-005 (audit.md requires a test file plan.md declares unchanged), a mislabeled evidence premise on the SPEC's central 12-event enumeration (claimed "re-measured in this tree" when it rests on the t494 survey, which does not exist in this tree), an epistemic gap that admits vacuous not-fired verdicts for PreCompact/PostCompact, and a DoD immutability check that cannot detect committed changes.

## 2. Evidence (commands run + verbatim outputs, this run)

| # | Verification | Command | Observed output |
|---|---|---|---|
| E1 | Anchor | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t496`; HEAD `b980c3920` |
| E2 | EventInterrupt 0-hit (spec §C M1) | `grep -rn 'EventInterrupt' internal/` | no output, `rc=1` (0 hits) — claim TRUE |
| E3 | Control (spec §C M1) | `grep -rn 'EventStop' internal/hook/types.go` | `internal/hook/types.go:34: EventStop EventType = "Stop"` — control live |
| E4 | Table census (spec §C M2) | direct read `internal/codexadapter/events.go:51-64` | 11 rows: `true},` ×6 (L52-57), `false},` ×5 (L59-63) — claim TRUE |
| E5 | No interrupt subcommand (spec §C M3) | `grep -n 'interrupt\|Interrupt' internal/cli/hook.go` | no output, `rc=1` — claim TRUE |
| E6 | Install-surface skip (spec §C M4) | read `internal/codexwiring/hooks.go:97-100` | L97 `for _, row := range codexadapter.EventTable {`, L98 `if !row.Adapted {`, L99 `continue`, L100 `}` — claim TRUE |
| E7 | codex-cli version (spec §C M5) | `codex --version` | `codex-cli 0.153.4` — claim TRUE |
| E8 | `hook.EventType` string-based | `grep -n 'type EventType' internal/hook/types.go` | `18:type EventType string` — REQ-CEV-002/D2 premise TRUE (no internal/hook change needed) |
| E9 | `IsInterrupt` disambiguation (plan §B) | `grep -n 'IsInterrupt' internal/hook/types.go` | `249: IsInterrupt bool \`json:"is_interrupt,omitempty"\`` — payload field, not a constant; plan's note accurate |
| E10 | Resolve message false-for-empty-arg (REQ-CEV-003 premise) | read `events.go:80-92` | L86-87 `fmt.Errorf("%w: %q (dispatcher arg %q exists; ...)", ErrUnadapted, codexEvent, row.DispatcherArg)` — with `DispatcherArg=""` renders `dispatcher arg "" exists`, which is false — premise TRUE |
| E11 | `TestDispatcherArgsExist` breaks on empty-arg row (plan §B/D1) | read `dispatcher_registration_test.go:23-46` | loop L40-45 asserts every row's `DispatcherArg` ∈ parsed registered set; `registered[""]` = false → new `{CodexEventInterrupt, "", false}` row fails the test — claim TRUE |
| E12 | `TestEventTableRowCount` pins 11 (plan §B) | read `events_test.go:12-19` | `const wantRows = 11` (L15; plan cites :17 — the use-site, off by 2) — claim TRUE |
| E13 | `TestAdaptedRowCount` unaffected | read `events_test.go:64-77` | pins adapted=6; M1 keeps 6 adapted → correctly omitted from the update list |
| E14 | `ValidateConfig` does not consult EventTable (D3) | read `internal/codexadapter/config.go:53-82` | validates against `acceptedTopLevel`/`acceptedEntryLevel` key sets only; function spans exactly :53-82 — claim TRUE |
| E15 | Runtime Resolve consumers (D3) | read `internal/cli/hook_harness_codex.go:54,58` | `codexadapter.Resolve(...)` both sites wrap refusal via `%w`; no branching on class → new unadapted row refused identically — claim TRUE |
| E16 | Consumer sweep (plan §E) | `grep -rn 'codexadapter\.' internal/ cmd/ pkg/ \| grep -v 'internal/codexadapter/' \| grep -v '_test.go'` | exactly 6 files (codex_readiness.go:116, hook.go:290-comment, doctor_codex.go:132, hook_harness_codex.go:44/54/58/81/90, codexwiring/hooks.go:43/97, wire.go:86) — **plan claims "7파일"** |
| E17 | Test consumers (plan §E) | `grep -rln 'codexadapter' internal/ --include='*_test.go' \| grep -v codexadapter` | 4 files (hook_harness_codex_test.go, codex_readiness_test.go, hooks_test.go, wire_test.go) — **plan claims "5파일"** |
| E18 | Helper callers (plan §D1) | `grep -rn 'IsUnadapted\|IsUnknownEvent' internal/ --include='*.go' \| grep -v codexadapter` | 1 hit: `hook_harness_codex_test.go:313` (a test) — **zero production callers**; plan says "harness의 리퓨절 로깅뿐" |
| E19 | AC-CEV-001 recipe works | `awk '/^var EventTable/,/^}/'` bounds = L51-64; row literals end `true},`/`false},` | recipe is well-formed against actual source formatting; expected post-M1 values (6/6, sum 12) correct |
| E20 | Dispatch trap ('Adapted:' phantom grep) | `grep -rn "Adapted:" .moai/specs/SPEC-CODEX-EVENT-COVERAGE-001/` | no output, `rc=1` — SPEC does not rely on the phantom recipe. TRAP AVOIDED |
| E21 | t494 survey existence | `ls .moai/reports/t494/` (this tree) and primary checkout | `No such file or directory` in BOTH; exists only in sibling worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t494/.moai/reports/t494/codex-doc-survey.md` |
| E22 | Survey content matches SPEC §A | read survey (t494 worktree, read-only) `:76` | "The documentation lists these hook events: `SessionStart`, `SessionEnd`, `SubagentStart`, `SubagentStop`, `PreToolUse`, `PermissionRequest`, `PostToolUse`, `PreCompact`, `PostCompact`, `UserPromptSubmit`, `Stop`, and `Interrupt`." — 12 events, exact match with spec.md:33 |
| E23 | t83 precedent exists | `ls .moai/reports/t83/` | `precondition-measurement.md`, `precondition-measurement-round3.md` present — events.go:11-12 citation resolvable |
| E24 | SPEC lint | `go run ./cmd/moai spec lint .moai/specs/SPEC-CODEX-EVENT-COVERAGE-001/spec.md` | `✓ No findings — all SPEC documents are valid` |
| E25 | AC-CEV-005 test home | `grep -n 'func Test' internal/codexwiring/hooks_test.go` | `TestRenderHooks_*` family lives there (10 tests) — the new "Interrupt absent from rendered bytes" test's only natural home |
| E26 | D7 statuses | `grep '^status:'` both referenced SPECs | `SPEC-CODEX-HOOK-ADAPTER-001: completed`, `SPEC-CODEX-WIRING-001: completed` |
| E27 | D8 syscall scan | `grep -rn 'syscall' .moai/specs/SPEC-CODEX-EVENT-COVERAGE-001/` | no output, `rc=1` |
| E28 | MP-7 markers | `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-EVENT-COVERAGE-001/` | 1 hit: `progress.md:9` — the negation assert ("마커 없음"), not an unresolved marker; plan.md 0 hits; research.md absent (Tier M) |

Comment-fix coverage sweep (dimension 2, last bullet) — full inventory of `events.go` comments that become false/incomplete after M1: (a) L40-44 "All eleven Codex events have a counterpart…" AND its continuation L41-44 "Excluding an event from adaptation is … never an absence of a counterpart — an earlier draft of the SPEC asserted the absence and was wrong"; (b) L46-50 census "Six rows are adapted … Four are held back … and SubagentStop …" (omits Interrupt; note "Six rows are adapted" itself remains TRUE after M1 — 6 adapted unchanged); (c) L11-12 measurement basis 0.147.0 (deferred to M2 re-stamp per plan §D4, with the M1 intermediate state handled in plan F-M1 step 3). plan §D4's ranges (40-44, 46-50, 11-12) COVER all three blocks at plan level — the defect is that AC-CEV-006 mechanically checks only one substring (see D1).

## 3. Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t496` (worktree t496, branch WT-… on top of develop)
- HEAD: `b980c3920` (`docs(spec): add SPEC-CODEX-EVENT-COVERAGE-001 plan artifacts (card t496)`) — the SPEC commit itself; artifacts audited as committed
- All E1-E28 commands executed in this run, 2026-09-07, against this tree. No figure carried from any prior report. "Would fail as claimed" judgments for E11/E12 are code-read derivations (static), not executed RED runs — the SPEC's own constraint (plan §C: "캠페인은 런 페이즈에서 실행하며, plan 페이즈에서 이를 검증 실행하지 않는다") and leaf-worker read-only scope preclude executing the mutation.

## 4. Must-Pass Results

- **[PASS] MP-1 REQ numbering**: REQ-CEV-001…012 sequential, no gaps, no duplicates, consistent naming (spec.md:50-64, 59-64).
- **[PASS] MP-2 GEARS compliance** — judged against the REQUIREMENT layer (REQ-CEV-* in spec.md §B); the Given-When-Then entries in acceptance.md are the verification layer and graded under Group 4 per M3 § Scope, not here. All 12 REQs match GEARS patterns: Ubiquitous (001, 002, 006, 007, 011, 012), Event-driven (003, 004, 010), Unwanted/shall-not (005, 008, 009). No informal-language REQ.
- **[PASS] MP-3 YAML frontmatter**: all 12 canonical fields present with correct types (spec.md:1-16); `tier: M` optional field present; no rejected snake_case aliases; E24 lint clean.
- **[N/A] MP-4 language neutrality**: single-language SPEC (Go, `module: internal/codexadapter`) — auto-pass.
- **[PASS] MP-5 D7 cross-SPEC**: both referenced SPECs exist, `status: completed` (E26) — no retired/superseded/archived, no reconciliation required, no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform**: no `syscall` in SPEC body (E27) — auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate**: plan.md 0 markers; research.md absent (Tier M → N/A for that file); the single progress.md:9 hit is the negation assert (E28).

## 5. Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.95 | 1.0 (minor nits) | REQ layer single-interpretation throughout; measured evidence inline per REQ. Nits: AC-CEV-001 Given mislabels the tree (acceptance.md:16 "Given 본 트리" vs post-M1 expectation — RED-now today, GREEN after M1; sibling ACs 002/004 say "M1 구현 후"); spec.md:51 contains Chinese hanzi "构造" inside Korean prose |
| Completeness | 0.90 | 1.0 (deductions) | All sections present (HISTORY/A-B-C-D-E/F; Out of Scope = 3 H3 topics + specific bullets, spec.md:96-108); frontmatter complete; REQ→AC coverage exhaustive (all 12 REQs covered, all 11 ACs trace — verified pairwise). Deductions: D2 (file-list contradiction), D5 (DoD blind spot) |
| Testability | 0.80 | 0.75-1.0 boundary | M1 ACs are command+strict-criterion (AC-CEV-001..006 machine-verifiable; E19); campaign ACs are record-content checks, binary-readable. Deductions: D1 (AC-CEV-006 under/mis-verification), D4 (2 of 6 campaign events admit vacuous not-fired verdicts), AC-CEV-011's isolation proof has no named command |
| Traceability | 1.00 | 1.0 | Every REQ-CEV-001..012 has ≥1 AC; every AC maps to existing REQs; spec.md §D trace table matches acceptance.md "(maps …)" headers exactly; D3 consumer table verified at every cited line (E14-E16) |

**Overall: 0.91** (mean of four dimensions). Tier M PASS threshold 0.80 — score passes; verdict is driven by blocking defects, not score.

## 6. Defects Found

**D1 — AC-CEV-006 under-verifies REQ-CEV-006 and demands removal of a phrase that stays true** — `acceptance.md:21` — The only executable check is `grep -c 'All eleven' → 0`. Not mechanically covered: the genuinely stale sentence at `events.go:41-44` ("Excluding an event from adaptation is … never an absence of a counterpart — an earlier draft … was wrong" — after M1 this is false in the deepest way: Interrupt is excluded precisely for absence); and the census obligation for `events.go:46-50` (must account for Interrupt). Worse, the AC demands '"Six rows are adapted" 고정 수치' not remain — but 6 adapted rows are unchanged by M1, so that phrase can legitimately remain; demanding its removal is a mis-specified requirement (charitable reading "must be re-authored" still leaves zero executable check for it). An implementer passing AC-CEV-006 as written can ship a comment whose absence-sentence is false. — Severity: **major** — Class: **blocking** — Fix: rewrite AC-CEV-006 to: `grep -c 'All eleven' events.go → 0` AND `grep -c 'never an absence of' events.go → 0` AND the doc comment block contains "Interrupt" with "no MoAI dispatcher counterpart"; delete the "Six rows are adapted" removal demand (replace with "the adapted/held-back census enumerates all 12 rows including Interrupt's reason").

**D2 — plan M1 file list contradicts AC-CEV-005** — `plan.md:85` vs `acceptance.md:20` — M1 declares "파일: internal/codexadapter/{events.go, events_test.go, dispatcher_registration_test.go}. 그 외 무변경." but AC-CEV-005 requires a NEW RenderHooks test ("신설 1개: Interrupt 행이 설치 바이트에 나타나지 않음"), whose only natural home is `internal/codexwiring/hooks_test.go` (E25: the TestRenderHooks_* family). An implementer following the file list skips AC-CEV-005; DoD #1 (all AC-CEV-001~006 GREEN) then cannot close. (Related: acceptance.md D.1 says the hand-wired-Interrupt edge case is "harness 거부 경로 테스트로 커버" — the existing `TestHarnessCodexUnadaptedSubcommandRejected` covers the class via PreCompact only; adding an Interrupt case would touch `hook_harness_codex_test.go`, also outside the declared list.) — Severity: **major** — Class: **blocking** — Fix: add `internal/codexwiring/hooks_test.go` to plan M1's file list and scope "그 외 무변경" to production (non-test) files.

**D3 — The official 12-event enumeration is asserted as re-measured when it rests on an artifact absent from this tree** — `spec.md:21` and `spec.md:33` — The header cites the t494 survey as "조사 입력(인용 아님 — 본 트리에서 재측정함)" — actively denying citation status and claiming in-tree re-measurement. But the §C M1-M5 table re-measures only in-tree facts (E2-E7); none of them can establish "official docs enumerate exactly 12, missing exactly Interrupt" — that axis rests entirely on the t494 survey, which does not exist in this tree or the primary checkout (E21); it exists only in the sibling t494 worktree. The survey content does match spec.md §A exactly (E22), so the claim is well-sourced — but the sourcing is mislabeled, which is precisely the evidence-discipline condition the dispatch flags. — Severity: **major** — Class: **blocking** — Fix: amend spec.md:21 to scope the re-measurement claim to the §C M1-M5 in-tree items, and mark the 12-event enumeration explicitly as t494 §3-sourced (external docs survey, sibling worktree, not re-verifiable in-tree); optionally add a §C row: "공식 12종 열거 — 근거: t494 §3 (외부 문서 주장, 본 트리 재측정 불가)".

**D4 — PreCompact/PostCompact trigger method admits a vacuous not-fired verdict** — `plan.md:96` — The named trigger ("컴팩션이 발화하도록 충분히 긴 컨텍스트로 codex exec 실행") has unproven feasibility in a non-interactive flow, and REQ-CEV-009/010's not-fired disposition ("command + observed output") is satisfiable by a run in which the compaction precondition was never reached — conflating "trigger attempted, event didn't fire" with "trigger condition never achieved". That is the absent-execution-read-as-suppressed-failure shape; it affects 2 of the 6 measurement targets, and the dispatch explicitly asks for honest feasibility statement here. — Severity: **major** — Class: **blocking** — Fix: add to plan M2 (and optionally a REQ-CEV-007 sub-clause): each not-fired verdict must carry independent evidence the trigger precondition was reached (e.g. codex's own compaction notice in its `--json` event stream or session log), or record "trigger-not-achieved" as a disposition distinct from "does not fire"; state plainly that Pre/PostCompact trigger feasibility in `codex exec` is unproven and will be established (or the event reclassified) at run time.

**D5 — DoD #4's internal/hook immutability check cannot detect committed changes** — `acceptance.md:51` — `git diff --stat HEAD -- internal/hook/` compares working tree vs HEAD only: a committed `internal/hook/` edit on the branch passes vacuously. REQ-CEV-005 forbids modification of any file under `internal/hook/` — broader than AC-CEV-002's constant-level guard. — Severity: **major** — Class: **blocking** — Fix: diff against the branch base: `git diff --stat $(git merge-base HEAD origin/develop)..HEAD -- internal/hook/` (or the develop base the card branched from), empty output = pass.

**D6 — plan §E consumer counts off by one** — `plan.md:72` — claims "비테스트 소비자는 위 7파일뿐(테스트 소비자 별도 5파일)"; measured 6 non-test files (E16 — the enumerated list itself is complete and exactly matches the sweep) and 4 test files (E17). Also `plan.md:22` cites `events_test.go:17` for `wantRows = 11`; the const is at :15 (:17 is the use-site). — Severity: minor — Class: optional — Fix: correct counts to 6/4 and the line cite to :15.

**D7 — D1's helper-caller characterization imprecise** — `plan.md:40` — "IsUnadapted/IsUnknownEvent 호출부는 harness의 리퓨절 로깅뿐": measured zero production callers of either helper (E18) — the harness refusal path wraps via `%w` without branching; the only external caller is a test. The no-third-sentinel decision's substance HOLDS (more strongly). — Severity: minor — Class: optional — Fix: reword to "no production caller branches on the refusal class; only a test asserts the marker".

**D8 — AC-CEV-001 Given mislabels the tree** — `acceptance.md:16` — "Given 본 트리의 …" while Then expects post-M1 values (11 today → RED now, GREEN after M1 — the implicit RED-now cell is fine, the label is not). — Severity: minor — Class: optional — Fix: Given → "M1 구현 후 트리".

**D9 — Chinese hanzi typo inside Korean prose of a normative REQ** — `spec.md:51` — "직접构造 가능하다" — Severity: minor — Class: optional — Fix: "직접 생성할 수 있다" (or English "construct directly").

**D10 — progress.md §E.1 artifact count** — `progress.md:7` — "spec.md / plan.md / acceptance.md / progress.md = 4 artifacts": Tier M artifact set is 3 (spec/plan/acceptance); progress.md is the tracking file. — Severity: minor — Class: optional — Fix: "Tier M (3 plan artifacts + progress.md)".

No defects found on: GEARS structure, REQ numbering, frontmatter schema, traceability, Out of Scope convention, milestone separability logic itself (M1's no-behavior-change property is real — verified at E6/E14/E15: install surface skips `!Adapted`, ValidateConfig never sees EventTable, runtime refusal class unchanged for existing consumers; the only behavior delta is `Resolve("Interrupt")` moving from ErrUnknownEvent to ErrUnadapted, which no production caller branches on and no existing test asserts otherwise — E18), the 'Adapted:' phantom-grep trap (E20 — avoided), M3 no-op-on-zero condition (AC-CEV-020(a) lets M1 close without M3).

## 7. Recommendation (numbered, actionable — route fixes from §6)

1. Fix D1 (acceptance.md:21) — replace AC-CEV-006 with the three-command form; drop the "Six rows are adapted" removal demand.
2. Fix D2 (plan.md:85) — add `internal/codexwiring/hooks_test.go` to the M1 file list; scope "그 외 무변경" to production files.
3. Fix D3 (spec.md:21, :33) — mark the 12-event enumeration as t494 §3-sourced; scope the "재측정함" claim to §C M1-M5.
4. Fix D4 (plan.md §F-M2; optionally REQ-CEV-007) — precondition-confirmation requirement or distinct "trigger-not-achieved" disposition; honest feasibility statement for Pre/PostCompact.
5. Fix D5 (acceptance.md:51) — DoD #4 diffs against the merge-base, not HEAD.
6. Optional batch (D6-D10): count corrections (6/4, :15), helper-caller rewording, AC-CEV-001 Given, hanzi typo, §E.1 artifact count.
7. Re-audit scope: defect delta D1-D5 (+ regression check over D6-D10 if touched). No from-scratch re-audit needed.

## 8. Gaps (what was NOT observed)

- The codex official documentation itself was not consulted (no web access used) — the 12-event enumeration is verified only against the t494 survey's quoted list (E22), not against the primary source.
- The M2 campaign was not executed (run-phase concern per the SPEC's own constraint); all campaign-plausibility judgments are design-level.
- The "tests would fail as claimed" assertions (E11/E12) are static code-read derivations; no mutation was compiled or run (leaf-worker read-only scope).
- `golangci-lint` / `go vet` / package test runs were not executed — plan-phase scope; SPEC D.2 assigns them to run-phase.
- Sibling worktree t494 was read at exactly one file (`codex-doc-survey.md` :76 region, read-only) — the rest of that tree was not audited.

## 9. Residual-risk

- The 12-event enumeration could diverge from the current official docs (the survey is same-day but secondhand); if codex added/renamed an event after the survey, M1's "exactly 12" premise shifts — M2's execution-based campaign is the designed backstop, and AC-CEV-010's evidence discipline would surface it.
- M2 outcomes may reclassify events (Adapted flips via M3); the census comments and `TestAdaptedRowCount` (pinned 6) will need a second update pass if any adapt-now lands — the SPEC handles this via M3 but does not name `TestAdaptedRowCount` among M3's touch points.
- The D1-D5 fixes are doc edits; if the lane implements M1 from the current contradictory file list (D2) before the fix lands, AC-CEV-005 will be silently skipped and DoD #1 will stall — the sequencing risk is real until D2 is applied.

## 10. Rubric compliance notes

- M3 anchoring applied: REQ layer graded for GEARS (MP-2), AC layer graded under Group 4 (Given-When-Then is the correct verification-layer format; no GEARS test applied to ACs).
- M6 applied: D6-D10 are optional-class and do NOT contribute to the FAIL; the FAIL rests solely on D1-D5 (correctness / internal-consistency / stated-criterion class).
- M1 Context Isolation: no author reasoning was supplied in the dispatch; audit performed against the four artifact files plus repo sources read for verification.
