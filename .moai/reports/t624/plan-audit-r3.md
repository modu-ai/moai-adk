# SPEC Review Report: SPEC-SYNC-GATE-FAILSTATE-001

Iteration: 3/3 (lead cap 3; harness Tier M ceiling is 2 per `.moai/config/sections/harness.yaml` L77, so this round runs past the harness ceiling on the lead's explicit authorization)
Verdict: FAIL
Overall Score: 0.86 (Tier M PASS threshold 0.80). Unchanged from round 2 (0.86). No score regression, so no STOP signal.
Escalation: this is the final allowed iteration. Per the Retry Loop Contract, the orchestrator escalates to the user (PASS-with-debt / scope-reduction / explicit override).
Auditor: plan-auditor (independent, adversarial). Reasoning context ignored per M1 Context Isolation. Lead-approved design points (N-1 split, N-6 no-60 s-row, N-4 narrowing, D1-D6, B1-B7, torn-write rows, mtime Gap) were used only as the list of what must be encoded; their existence was not re-scored.
Card: t624 · Branch `WT-sync-gate-failstate` · Tree measured: HEAD `92b9a8680` (re-read immediately before writing this file)
File name: `plan-audit-r3.md`, as the lead instructed. The convention family is `plan-audit-iter<N>.md`; the lead's explicit name wins.

**Why FAIL despite a score above threshold.** Every must-pass criterion passes, N-2 through N-8 are resolved, and no F-finding regressed. One blocking finding remains (D1 / NEW-1). It is the residual of N-1. The new N2a and N2b RED reasons state exact stub-count deltas ("grown by **1**", "grows by **9**"), but the SPEC defines the stub count as an *invocation* count (acceptance.md L146). The unchanged hook invokes the `go` stub **twice** per Go check run: `go vet` at L225 and `go build` at L226, both unconditional. M1 will therefore observe 2 and 18. DoD L570-L572 requires "the observed RED reasons match the named reasons above", so the DoD fails on arrival. The fix is one line of wording; no design is affected. Under M6 a blocking finding is fixed before the verdict is revisited.

## Baseline attribution

| Check | Command | Observed |
|---|---|---|
| tree | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` |
| branch / head | `git branch --show-current`; `git rev-parse --short HEAD` | `WT-sync-gate-failstate`; `92b9a8680` (read at start and again before writing) |
| revision range | `git diff --stat 94d185343 92b9a8680 -- .moai/specs/SPEC-SYNC-GATE-FAILSTATE-001/`; `git log --oneline 94d185343..92b9a8680` | acceptance.md 159, plan.md 30, spec.md 25 (145+/69-); commits `92b9a8680` (revision), `a130f891a` (r2 verdict) |
| target drift since plan pin | `git diff --stat fa96fe644 HEAD -- <both hooks> <both docs> internal/hook internal/template/hook_official_compliance_test.go internal/template/templates/.claude/settings.json.tmpl` | empty, exit 0. Hook readings below hold for `fa96fe644`. |
| lint binary | `command -v moai`; `moai version` | `/Users/goos/go/bin/moai`; `v3.2.0-rc.5  list-974-g84fa4ece4  built 2026-09-09T02:48:44Z` |
| lint binary provenance | `git merge-base --is-ancestor 84fa4ece4 HEAD` | exit 0. The installed build is a strict ancestor of HEAD, **not built from this tree**. Lint rules added after `84fa4ece4` did not run. |
| lint | `moai spec lint SPEC-SYNC-GATE-FAILSTATE-001` | `0 error(s), 0 warning(s)`; one INFO `OwnershipTransitionUnmeasured` (commit `b569c5446` has no `Authored-By-Agent` trailer); exit 0 |
| spec_audit (MCP, `project_root` = worktree) | `mcp__moai__spec_audit filter_spec=SPEC-SYNC-GATE-FAILSTATE-001` | era V3R6, `modern_era_clean: 1`, one INFO `EraAutoDetected` (`H-5 (modern phase or created date)`) |
| REQ numbering | `/usr/bin/grep -noE '^- \*\*REQ-[0-9]+' spec.md` | REQ-001..REQ-015 at L124, 131, 144, 149, 153, 164, 174, 179, 187, 195, 219, 225, 232, 238, 246 |
| mutant numbering | `/usr/bin/grep -noE '^\| M[0-9]+ ' acceptance.md` | M1..M27 at L534..L560, contiguous, no gap or duplicate |
| ledger ids | `/usr/bin/grep -noE '^\| L-[0-9]+' acceptance.md` | L-01..L-15, L-17, L-18, L-19, L-21 in the table; L-16 and L-20 are bullet baselines (L90, L98). No duplicate. |
| D7 | `/usr/bin/grep -rhoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' <spec dir> \| sort \| uniq -c` | `6 SPEC-SYNC-GATE-FAILSTATE-001` only |
| D8 | `/usr/bin/grep -rc syscall <spec dir>` | 0 in all four files |
| MP-7 | `/usr/bin/grep -rn 'NEEDS CLARIFICATION' plan.md research.md` | no match in plan.md; `research.md: No such file or directory` (Tier M) |
| L-18 regex vs template hook | `/usr/bin/grep -cE '(^\|[;&\|(]\|\$\(\|(then\|do)[[:space:]])[[:space:]]*jq([[:space:]]\|$)' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | `0`, exit 1 (matches ledger L-18) |
| L-18 control (re-measured) | same regex on a 5-line file: `x=$(echo "$in" \| jq -r .a)`, `# this hook never calls jq for parsing`, `jq . file`, `if true; then jq -r .x f; fi`, `for f in a; do jq . "$f"; done` | `4`, exit 0 (matches acceptance L93-L97) |
| L-18 false-positive probe | same regex on `# todo jq cleanup` | `1`, exit 0 (see D3) |
| L-21 selector, raw file bytes | `go test ./internal/template/ -list 'TestTemplateNoInternalContentLeak\|TestLeakClassNoDateShaInDefaultTier\|TestC7PackageRestriction'` (literal `\|`, as the cell's raw bytes carry) | only `ok  github.com/modu-ai/moai-adk/internal/template 0.421s`. **Zero tests listed.** |
| L-21 selector, plain pipe (control) | same with plain `\|` → `|` | `TestTemplateNoInternalContentLeak`, `TestLeakClassNoDateShaInDefaultTier`, `TestC7PackageRestriction`, then `ok … 0.267s`. All three names exist (`internal/template/internal_content_leak_test.go` L1535, L1761, L1849). |
| hook reading (stub count) | Read `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` L176-L186, L204-L214, L222-L227 | L179-L186: bare-SHA sentinel written before checks. L225 `run_step go c1 go vet ./...` and L226 `run_step go c2 go build ./...` run unconditionally; `run_step` (L204-L214) invokes the tool whenever `command -v` finds it. **Two stub invocations per Go check run.** |
| counting-stub pattern | Read `internal/hook/stopchain_trim_test.go` L37-L49, L209-L226 | `writeCountingStub` appends one line per invocation (L41); `readCounter` counts non-empty lines (L218-L225). The plan says to reuse this pattern (plan L145-L146). |
| hook-execution probe (N2a/N2b/U4) | scratch fixture: `git init` + commit in the session scratchpad succeeded (HEAD `5dcef3d`); running the unchanged hook against it | **Refused by the worktree-session guard** ("names git in a form too complex to verify"). Not bypassed, and no script-file workaround. D1 rests on the hook and stub-pattern readings above, not on execution. |
| audit_model | not re-measured this round (Grep tool unavailable in this session) | Round 2 observed no `audit_model` key. No cross-model backend was called this round (see Gaps). |

## Must-Pass Results

- [PASS] MP-1 REQ number consistency. spec.md L124-L246 carries REQ-001..REQ-015 in order, with no gap or duplicate. The §D matrix (acceptance L29-L43) carries AC-001..AC-015. Tier M budget: 15/16 REQ, 15/16 AC (acceptance L45).
- [PASS] MP-2 GEARS. Judged on the **REQ layer in spec.md only**; the Given-When-Then ACs are verification layer and were not graded here. The changed REQs still match a pattern: REQ-005 L153 "**When** stdin carries … the gate shall …" (When); REQ-008 L179 "**While** … **when** the gate is invoked, it shall not …" (While+When); REQ-009 L187 "**When** … the gate shall run the checks" (When). The unchanged REQs keep their round-2 classification: Ubiquitous 001, 010-015; When 003, 004, 006, 007; Ubiquitous/When 002.
- [PASS] MP-3 frontmatter. spec.md L2-L13 carries all 12 canonical fields with the correct names and types (`version: "0.1.0"` quoted, `status: draft`, ISO `created`/`updated`, `priority: P1`, `phase: "v3.2.0 target"`, `lifecycle: spec-anchored`, `tags` as a comma string), plus optional `tier: M` (L14). Unchanged by this revision; lint reports 0 errors (older build, see Gaps).
- [PASS] MP-4 language neutrality. REQ-012 L225-L230 keeps detection and checks unchanged for all 16 languages. AC-007 R13 (acceptance L351) is a non-Go branch. `govulncheck` appears only in Out of Scope (spec L276).
- [PASS] MP-5 D7. Only the SPEC's own ID is referenced, so there is no BLOCKING finding.
- [PASS] MP-6 D8. `syscall` count is 0 (auto-pass).
- [PASS] MP-7 clarification gate. No `[NEEDS CLARIFICATION` in plan.md; research.md is absent (Tier M).

## Disposition of round-2 findings (N-1..N-8, O2)

Each was decided from the current files; the SPEC's own disposition notes were not accepted as evidence.

| Id | Class (r2) | Disposition | Evidence |
|---|---|---|---|
| N-1 | blocking | **PARTIAL** → residual is **D1 / NEW-1** (blocking) | Split encoded: N2a at acceptance L489-L490, N2b at L491-L493. The swept set becomes 27 (L506), and M5 now names both halves (L538). Checked against hook L179-L186: N2a's decision count **1**, 8 empty stdouts, 8 of 9 lacking `"systemMessage"`, and final record = bare SHA (L517-L522) are all **true**. N2b's decision count **9** and "each block JSON carries `"systemMessage"`" (L523-L526) are **true**; the blocking printf at hook L383 carries `systemMessage`. **False in both halves:** the stub-count deltas "grown by **1**" (L521) and "grows by **9**" (L525). Each check run invokes the stub twice (hook L225-L226), so the counts are 2 and 18 under the SPEC's own invocation-count definition (L146). The round-2 report's D1 text carried the same miscount ("one stub invocation"); the revision inherited it. |
| N-2 | blocking | **RESOLVED** | The harness deletes the marker before each run (acceptance L302-L303; plan L153-L154). It records "marker observed for this run" or "marker not observed" (L304-L306; plan L158-L162). New non-fatal assertion 2, "run 2 observed its own marker" (L313), and the named RED reason now includes it (L318-L320). Green path re-derived: run 2 is the stale re-run, invokes `go vet` → marker → assertion 2 passes. |
| N-3 | blocking | **RESOLVED** | U4 read step: `chmod 0600` before any read of the record (acceptance L209-L212), with the reason stated. Against hook L179/L185: `cat` fails silently, `echo > file` writes through the 0200 owner-write bit, and after chmod the test reads the bare SHA — the REQ-001 reason at L219-L221. A rename-based or redirect-based M2 write both read cleanly after chmod, so the round-2 green-path fragility is gone. Read from the hook; execution refused (see Gaps). |
| N-4 | blocking | **RESOLVED** | REQ-005 narrowed: "in object-key position … none, one or more spaces, or tabs"; "not required to tell a top-level key from a nested one" (spec L156-L161). Also: F4 tab row (acceptance L179, one literal U+0009 byte); M25 (L558); nested-key Gap (L120-L127); plan §G aligned (L215-L221); DoD disclaims it (L574-L576). No residual claim of top-level matching in spec, plan, or acceptance. |
| N-5 | optional | **RESOLVED** (weakness in D3) | L-18 regex now includes `(then\|do)[[:space:]]` (L71). The control was re-measured this round: `4` on the five-line probe, `0`/exit 1 on the hook. The backtick limit is stated (L97). Residual: the keyword group has no left word boundary (D3). |
| N-6 | optional | **RESOLVED** (stale cross-refs in D4) | REQ-008 "age is at most the stale window … exactly 60 s is fresh" (spec L180, L184-L185). REQ-009 "strictly greater" (L187-L188). b61 row (acceptance L284) classed regression-guard; true today, since `<HEAD> running` ≠ bare SHA. M27 "compare against 65" (L560): age 61 ≤ 65 → notice → b61 red. Exact-60 s Gap (L129-L137) and boundary paragraph (L292-L296) agree with REQ-008/009 and with each other. The DoD counts the Gap (L574-L576). The matrix lists b61 (L34). |
| N-7 | optional | **RESOLVED** | acceptance L227-L230 cite D2 as the ground for the notice outcome and state that today's hook would run and block. |
| N-8 | optional | **RESOLVED** | N1: "the stub writes its marker and then exits 1 immediately, with no sleep", with the marker deleted before each call (acceptance L486-L488). |
| O2 | optional (carried) | **PARTIAL** → residual is **D2 / NEW-2** | L-21 was added as an M1-pending ledger row (acceptance L73) and is listed among the M1 ledger gaps (L102-L106; plan L175-L176). As written, its `-run` selector uses `\|` in the raw file bytes, and that form **lists zero tests** (measured above). The L75 note that `\|` renders as `\|` covers only "the L-18 and L-19 cells". No contingency is stated if M1 finds the SHA class does not cover the two paths. |

## F1-F11 regression spot-check (round 1)

No regression. Each anchor was re-read in the current files:

| Id | Anchor (current lines) | Status |
|---|---|---|
| F1 | AC-006c harness + defect-signal RED reason, acceptance L298-L327; plan L152-L165 | holds (strengthened by N-2) |
| F2 | AC-004 release-blocking a/b/c, matrix L32; RED reason L189-L192 | holds |
| F3 | content anchor: spec L232-L236; acceptance L408-L422 | holds |
| F4 | A7/A8/A9 at L376-L378; M15-M17 at L548-L550 | holds |
| F5 | stdin forms F1-F3 at L176-L178 (+F4 L179); R10a/R10b L347-L348; M14 L547 | holds |
| F6 | payload removed before checks, spec L138-L140; S1 L461; M18 L551 | holds |
| F7 | REQ-007 spec L174-L177; U1-U5 L203-L207; M19 L552 | holds |
| F8 | REQ-010 history invariant spec L195-L217; R13-R17 L351-L355; scope note L359-L362 | holds |
| F9 | retry marker "not an outcome token", plan L80-L82 | holds |
| F10 | M1 covers AC-002/AC-003, plan L166-L167 | holds |
| F11 | a50/b70 rows L283/L285; M20/M21 L553-L554 | holds (b61/M27 added) |

## Lead-decision encoding check

| Decision | Encoded at | Faithful? |
|---|---|---|
| N-1 split (N2a mtime-only / N2b full rewrite) | acceptance L489-L493, L515-L528; M26 L559 | Split faithful. Of the RED reasons, the decision / stdout / record clauses are true and the stub-count magnitudes are false (D1). |
| N-6 no exactly-60 s row, 61 s row, M27, `>` vs `>=` Gap | spec L184-L185, L187-L188; acceptance L284, L292-L296, L560, L129-L137 | Yes, internally consistent |
| N-4 REQ-005 no top-level claim; nested key an explicit Gap not claimed by DoD | spec L156-L161; acceptance L120-L127, L574-L576 | Yes |
| D1-D6 | spec §2 L99-L120 | Yes, unchanged |
| B1-B7 | plan §B L29-L66; spec REQ-002 L131-L142, REQ-006 L164-L172, REQ-007 L174-L177, REQ-009 L187-L193 | Yes, unchanged |
| Torn-write rows | acceptance L224-L273 (TA1-TA5, TB1-TB4, M23/M24) | Yes; N-7 grounding added |
| mtime-unreadable Gap | acceptance L108-L118; DoD L574-L576; plan L245 | Yes |

## Verification-completeness classification (§2 / §2.1)

| AC | Class | §2.1 status at plan phase |
|---|---|---|
| AC-001, AC-004, AC-014, AC-015 | release-blocking (behavioral) | four elements deferred to M1 per B4 (acceptance L12-L17); RED reasons named. AC-015 N2a/N2b named reasons are defective (D1). |
| AC-005 | release-blocking rows L1, TA1, TA4, TB1 and U-row record-format sub-assertion; regression-guard rows otherwise | deferred to M1 per B4 |
| AC-006 | release-blocking a0, a50, c (AC-006c); regression-guard b61, b70, b120 | deferred to M1 per B4 |
| AC-008 | release-blocking A4, A9; regression-guard A1-A3, A5-A8 | deferred to M1 per B4 |
| AC-011 | release-blocking | ledger L-02..L-05, L-08..L-12 carry command / stdout / exit / pin; two local-copy rows are M1 ledger gaps (L435, L437) |
| AC-012 | release-blocking (a); regression-guard (b) | L-06/L-07 carry four elements; (b) rests on L-20 bullet baseline |
| AC-002, AC-003, AC-007, AC-009, AC-010, AC-013 | regression-guard | baselines named; observed as PASS at M1 |

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity | REQ-005 (spec L153-L162) and REQ-008/009 (L179-L193) are now unambiguous. What remains: REQ-002 still says "set the retry marker" and then qualifies it "present if and only if … otherwise it shall be removed" (L138-L140), which a reasonable engineer resolves consistently; plan §C L106 "age within window" versus REQ-008 "at most". |
| Completeness | 1.0 | all present | HISTORY L19, background §1 L46, decisions §2 L99, REQs §3 L122, constraints §4 L252, four `### Out of Scope —` H3s with bullets L274-L301; frontmatter complete. (HISTORY is not updated for round 2; optional, D5.) |
| Testability | 0.75 | one criterion needs correction | Every row is binary with named RED reasons. One release-blocking criterion (AC-015 N2a/N2b, L521/L525) names numeric RED signals that M1 will not observe (D1). |
| Traceability | 1.0 | complete | Every REQ has an AC (matrix L29-L43); every AC names existing REQs. M1-M27 all name existing rows (M25 → F4 L179, M26 → N2a L489, M27 → b61 L284). Referenced ids TB3→U5, TB4→R9/D1, N1→AC-006c assertions 1-2 all exist. |

Harmonic mean: 4 / (1/0.75 + 1 + 1/0.75 + 1) = 4 / 4.667 = 0.857, rounded to 0.86.

## Defects Found (structured defect-list)

D1. NEW-1 (residual of N-1) — acceptance.md:L521, L525 — The N2a RED reason says "a stub count grown by **1**" and the N2b RED reason says "the stub count grows by **9**". The SPEC defines the stub count as the stub `go` **invocation** count (L146), and plan L145-L146 says to reuse the counting-stub pattern, which appends one line per invocation (`internal/hook/stopchain_trim_test.go` L41) and counts lines (L209-L226). The unchanged hook invokes the stub twice per Go check run, `go vet` (L225) and `go build` (L226), both unconditional through `run_step` (L204-L214). M1 will observe 2 for N2a and 18 for N2b, so DoD L570-L572 ("the observed RED reasons match the named reasons") fails on a correct test. Read from the hook and the stub pattern; the execution probe was refused by the worktree-session guard. — Severity: minor — Class: blocking — Required fix: state the deltas in check runs with their invocation count ("grown by one check run — 2 stub invocations, `vet` and `build`" and "9 check runs — 18 invocations"), or drop the magnitudes and keep "grows" / "grows on every call". Leave the decision-count, stdout, and record clauses as they are; they are correct.

D2. NEW-2 (residual of O2) — acceptance.md:L73, L75 — L-21's selector `'TestTemplateNoInternalContentLeak\|TestLeakClassNoDateShaInDefaultTier\|TestC7PackageRestriction'`, run from the raw file bytes (the verbatim rule of verification-completeness §2.1), lists **zero tests**; measured, only `ok … 0.421s`. The plain-pipe control lists all three. The rendering note at L75 covers only L-18 and L-19. Its sweep would be empty and still read as a completed measurement (§1.1). No contingency is stated if M1 finds the SHA class does not cover the two template paths. — Severity: minor — Class: optional (L-21 is a pending measurement, not a criterion; the DoD does not depend on it) — Required fix: extend the L75 note to L-21, or give the L-21 command with plain `|` in a fenced block. Add one sentence naming what M1 does if coverage is absent, for example: add a 7-40-hex check scoped to the two template files, or record REQ-012's SHA clause as an explicit Gap.

D3. NEW-3 (weakness in N-5 fix) — acceptance.md:L71, L398-L400 — AC-009 asserts "A comment naming the tool does not count". The keyword group `(then|do)[[:space:]]` has no left word boundary, so `# todo jq cleanup` matches (measured `1`). The failure is safe (a false red, never a false green), but the stated property is broader than the regex. — Severity: minor — Class: optional — Required fix: anchor the keywords (`(^|[^A-Za-z0-9_])(then|do)[[:space:]]`), or qualify the sentence ("a comment that names the tool without a preceding `then`/`do` token does not count").

D4. NEW-4 — acceptance.md:L3-L6, L479-L480; plan.md:L106, L244 — Cross-references went stale after this revision. The document header still says ledger rows "L-15 to L-19 carry their own pin, `ad0ad6f9b`", but L-18 is now pinned at `a130f891a` and L-21 at "M1 commit". The AC-014 scope note names only "rows a50 and b70". Plan §H's risk row names only "50 s / 70 s boundary rows". Plan §C L106 still says "age within window" where REQ-008 now says "at most". None changes a verdict: criterion-level pins win per §2.1, and b61 adds coverage rather than removing it. — Severity: minor — Class: optional — Required fix: update the header pin sentence, add b61 to both cross-references, and align L106 with "age ≤ window".

D5. NEW-5 — spec.md:L19-L44; plan.md:L14 — HISTORY has no entry for the round-2 revision (N2a/N2b split, F4 tab row, b61/M25-M27, REQ-005 narrowing, REQ-008/009 boundary, two new §D.0 Gaps). Plan §A still points only at the round-1 verdict, saying "every finding is dispositioned in this revision". — Severity: minor — Class: optional — Required fix: add a dated HISTORY bullet for the round-2 revision, and point plan §A at `plan-audit-r2.md` / `plan-audit-r3.md`.

## Regression Check (Iteration 3)

- Round 1 (F1-F11): all still RESOLVED; no regression (table above). O1, O3-O7 were RESOLVED in round 2 and were not re-opened by this revision.
- Round 2 (N-1..N-8, O2): N-2, N-3, N-4, N-5, N-6, N-7, N-8 RESOLVED. N-1 PARTIAL (D1). O2 PARTIAL (D2).
- Stagnation: no defect appears unchanged in all three iterations. There is a **recurring family**: named RED reasons in the interrupted-run and notice criteria drifting from what the current hook does. F1 (r1) was the harness wait, N-1 (r2) the N2 decision count, NEW-1 (r3) the N2a/N2b stub magnitudes. Each round fixed the named instance; the family persisted because each named reason was reasoned about rather than executed, and scratch execution is refused in this worktree.
- Score trend: 0.60 → 0.86 → 0.86 (no regression).

## Recommendation

Final iteration: FAIL with one blocking finding, which is wording-level. No design-level blocker remains.

For manager-spec, if the lead chooses to revise:
1. D1: restate the AC-015 N2a (L521) and N2b (L525) stub-count deltas in check runs with invocation counts, or drop the magnitudes. One-line edits; no row, mutant, or DoD count changes.
2. Optional, same pass: D2 (L-21 selector note plus contingency), D3 (keyword boundary), D4 (stale cross-refs), D5 (HISTORY / plan §A).

Escalation options for the orchestrator (Retry Loop Contract, after iteration 3):
- **PASS-with-debt** is defensible here. D1 is a named-reason mismatch that M1's own RED capture would surface immediately: plan L168-L174 requires comparing each observed failure with its named reason. The debt would be recorded as "AC-015 N2a/N2b stub-count magnitudes to be corrected against the M1 observation before the RED cells are adopted". The risk is a mis-recorded RED cell if the comparison is skipped.
- **Scope reduction**, if the lead prefers to split the card. Suggested boundary:
  - **SPEC A — gate state machine (behavioral).** REQ-001..REQ-012, AC-001..AC-009, AC-013..AC-015, plus REQ-014's hook-comment clause, because it edits the same two hook files that REQ-012 rewrites and REQ-013 keeps byte-identical. This is where every RED-reason defect of all three rounds sits.
    - Sub-split if A is still too large: **A1** = fail/pass record, re-delivery, `stop_hook_active`, mode resolution (REQ-001..007, 010-012; AC-001..005, 007, 008, 009, 013). **A2** = `running` / stale window / retry bound / notices (REQ-008, REQ-009; AC-006, AC-014, AC-015). A2 depends on A1's closed-token record and retry-marker file, and the torn-write rows TA1/TB1 cross the seam, so A2 must land after A1.
  - **SPEC B — documentation wording only.** REQ-013 (document-copy anchor and parity), REQ-014's document clause (doc lines 116, 140-146), and REQ-015 (SX-R05); AC-010's document checks, AC-011's document rows, AC-012. No behavioral RED cells (all ledger-backed) — a Tier S candidate that can land independently of A.
- **Explicit override** to iterate a fourth time. Not recommended; D1 does not need another audit round to be verified.

## Gaps

- **Hook execution refused.** The scratch fixture was created (`git init` + commit succeeded, HEAD `5dcef3d`). Running the unchanged hook against it was refused by the worktree-session guard. D1, N-3, and the N2a/N2b clauses rest on reading hook L176-L186, L204-L214, L222-L227 and the stub pattern at `stopchain_trim_test.go` L37-L49/L209-L226, not on observed execution. The guard was not bypassed.
- **Plan-only evidence.** AC-006c/N1 green paths and the marker lifecycle (N-2) are inferred from the fixture text; no M2 hook exists yet.
- **No Go tests run** beyond two `go test -list` selector listings for L-21.
- **Lint provenance.** Lint and `spec_audit` ran on build `84fa4ece4`, a strict ancestor of HEAD, not a build of this tree.
- **No cross-model audit.** `audit_model` was not re-measured this round (round 2 observed no key), and no codex/GLM backend audit was run.
- **Windows** behavior (AC-006c and U4 skip there) was not assessed.
- **Exit codes on piped commands.** The `rawbytes_exit` / `plainpipe_exit` values recorded above are the pipeline status (`tail`), not `go test`'s own exit code; the listing contents are the evidence.

## Residual risk

- **Heuristic detector.** The jq-free key detector is still heuristic. Escape sequences such as `\\"` immediately before a real key were reasoned about, not probed. A nested `true` key defers re-delivery by one turn (accepted Gap).
- **Boundary nondeterminism.** M27 (window 65) is caught by b61 only when fewer than about 4 s elapse between the mtime write and the hook's read. A slow test run could let that mutant survive once.
- **mtime-unreadable fallback.** Unverified by design (acceptance L108-L118). On such a platform a stale record could keep emitting the fresh notice.
- **Shared state.** Concurrent sessions sharing `.moai/state/` stay out of scope (spec L299-L300).

## Operational Notes (unverified)

- `inferred`: D1 can be settled in one measurement at M1. Run the N2a and N2b loops against the unchanged hook with the counting stub, and read the counter line total before writing the RED cells. Rule applied: hook L225-L226 invoke `go` twice per check run; `readCounter` counts lines.
- `measured`: the L-21 selector copied from raw bytes sweeps zero tests. Command and output are in the Baseline attribution table.
- `assumption`: the recurring RED-reason family would stop recurring if the M1 test-only commit's RED capture were executed before the named reasons are frozen — that is, if RED reasons were written from observation at M1 rather than from reading at plan phase. Measure it by comparing the M1 `progress.md §E.2` observations with each named reason.
