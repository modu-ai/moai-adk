# SPEC Review Report: SPEC-SYNC-GATE-FAILSTATE-001

Iteration: 2 (lead cap 3; harness Tier M ceiling 2)
Verdict: FAIL
Overall Score: 0.86 (Tier M PASS threshold 0.80). No score regression (round 1: 0.60), so no STOP signal.
Auditor: plan-auditor (independent, adversarial). Reasoning context ignored per M1 Context Isolation. The lead decisions were used only as a checklist for what must be encoded.
Card: t624 · Branch `WT-sync-gate-failstate` · Tree measured: HEAD `94d185343` (re-read right before writing this file)
File name: `plan-audit-r2.md`, as the lead instructed. The convention's naming family is `plan-audit-iter<N>.md`; the lead's explicit name wins.

Why FAIL despite a score above threshold: all 11 round-1 blocking findings are resolved, and every must-pass criterion passes. The revision still carries four new **blocking** findings (all minor, all text-level fixes). Three contradict the SPEC's own Definition of Done or the green path of a release-blocking criterion, and one leaves two stated REQ-005 properties unpinned. Under M6, blocking findings are fixed before the verdict is revisited. A delta re-audit of N-1 to N-4 is enough to decide round 3.

## Baseline attribution

| Check | Command | Observed |
|---|---|---|
| tree | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` |
| branch / head | `git branch --show-current`; `git rev-parse --short HEAD` | `WT-sync-gate-failstate`; `94d185343` |
| revision range | `git diff --stat 6eda95337 94d185343` | spec.md, plan.md, acceptance.md, plan-audit.md, 3 lock-record files |
| target drift since plan pin | `git diff --stat fa96fe644 HEAD -- <both hooks> <both docs> internal/hook internal/template/hook_official_compliance_test.go internal/template/templates/.claude/settings.json.tmpl` | empty, exit 0. Round-1 readings of these files still hold. |
| hook copy parity | `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | empty, exit 0 |
| doc anchor | `/usr/bin/grep -n '^Purpose: Ensure code has appropriate @MX annotations for AI agent context' <both doc copies>` | both copies: `160:Purpose: … Supports all 16 MoAI-ADK languages.` |
| REQ numbering | `grep -noE '^- \*\*REQ-[0-9]+' spec.md` | REQ-001..REQ-015 at L124, 131, 144, 149, 153, 162, 172, 177, 184, 192, 216, 222, 229, 235, 243 |
| D7 | `grep -rnoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' <spec dir>` | only the SPEC's own ID |
| D8 | `grep -rc syscall <spec dir>` | 0 in all four files |
| MP-7 | `grep -rn 'NEEDS CLARIFICATION' <spec dir>` | no output, exit 1 |
| lint binary | `command -v moai`; `moai version` | `/Users/goos/go/bin/moai`; `v3.2.0-rc.5 list-974-g84fa4ece4 built 2026-09-09T02:48:44Z` |
| lint binary provenance | `git merge-base --is-ancestor 84fa4ece4 HEAD` | exit 0. The installed build is an ancestor of HEAD, **not built from this tree**. |
| lint | `moai spec lint SPEC-SYNC-GATE-FAILSTATE-001` | `0 error(s), 0 warning(s)`; one INFO `OwnershipTransitionUnmeasured` (commit `b569c5446` has no `Authored-By-Agent` trailer); exit 0 |
| spec_audit (MCP, `project_root` = worktree) | `mcp__moai__spec_audit filter_spec=SPEC-SYNC-GATE-FAILSTATE-001` | era V3R6, one INFO `EraAutoDetected` (H-5) |
| audit_model | `grep -rn audit_model .moai/config/sections/` | no match. Claude-only audit; no cross-model backend call. |
| existing gate tests | `grep -nE 'RemoveAll\|…' internal/hook/stopchain_ac004_006b_test.go internal/hook/stopchain_trim_test.go` | each gate invocation uses a fresh `t.TempDir()` fixture (ac004_006b L120) or deletes `.moai/state` first (ac004_006b L129, trim L192). M2 re-delivery does not break them. |
| L-18 jq regex probe | `/usr/bin/grep -cE '<L-18 regex>' <probe file>`; probe lines: `$(… \| jq …)`, a comment naming jq, `jq . file`, `if true; then jq -r .x f; fi`, `for f in a; do jq . "$f"; done` | `2` (lines 1 and 3). The `then jq` and `do jq` invocations are **not** matched. |
| round-1 AC-015 text | `git show 6eda95337:.moai/specs/SPEC-SYNC-GATE-FAILSTATE-001/acceptance.md` (scratch copy), lines 278-310 | round-1 N2 RED reason was the generic "N2 fails both the decision-count and stub-count assertions". The "decision count is 9" wording was added by this revision. |
| scratch fixture probes (N2, U4) | a scratch git repo running the unchanged hook | **refused by the worktree-session guard** ("names git in a form too complex to verify"). N-1 and N-3 therefore rest on reading the hook, not on execution (see Gaps). |

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: spec.md L124-L243 carries REQ-001..REQ-015 in order, with no gaps or duplicates. The acceptance.md §D matrix (L29-L43) carries AC-001..AC-015. Tier M budget: 15/16 REQ, 15/16 AC (acceptance L45).
- [PASS] MP-2 GEARS (judged on the REQ layer in spec.md only). Ubiquitous: REQ-001 L124, REQ-010 L192, REQ-011 L216, REQ-012 L222, REQ-013 L229, REQ-014 L235, REQ-015 L243. When: REQ-003 L144, REQ-004 L149, REQ-005 L153, REQ-006 L162, REQ-007 L172, REQ-009 L184. While+When: REQ-008 L177. Compound Ubiquitous+When: REQ-002 L131. Round-1 O4 is fixed: REQ-010 and REQ-011 now read "The gate shall …". The Given-When-Then ACs sit in the verification layer and were not graded here.
- [PASS] MP-3 frontmatter: spec.md L2-L13 carries all 12 canonical fields with the right names and types (`version: "0.1.0"` quoted, `status: draft`, `created`/`updated` ISO dates, `priority: P1`, `phase: "v3.2.0 target"`, `lifecycle: spec-anchored`, `tags` as a comma string), plus optional `tier: M` (L14). Lint reports 0 errors (older build; see Gaps).
- [PASS] MP-4 language neutrality: REQ-012 L223-L226 keeps early exits, detection, and check commands unchanged for all 16 languages. AC-007 R13 (acceptance L311) adds a non-Go language branch. `govulncheck` appears only in Out of Scope (spec L273).
- [PASS] MP-5 D7: no other SPEC is referenced, so there is no BLOCKING finding.
- [PASS] MP-6 D8: `syscall` count is 0 (auto-pass).
- [PASS] MP-7 clarification gate: no `[NEEDS CLARIFICATION` marker (grep exit 1). research.md does not exist (Tier M).

## Round-1 regression check (F1-F11, O1-O7)

Each finding was re-decided from the current files. The SPEC's own disposition notes were not accepted as evidence.

| Id | Disposition | Evidence |
|---|---|---|
| F1 | RESOLVED | acceptance L266-L285 and L445-L465. The harness now triggers on a marker the stub writes, or on the hook exiting, whichever comes first (plan L152-L157). Against the hook: L183-L186 write the bare SHA **before** L225 invokes the stub, so run 1 still reaches the marker and is killed. Runs 2 and 3 (and N1's 9 runs) take the L179-L182 short-circuit, exit 0 without invoking the stub, and the harness proceeds on exit. The named reason, "exhausted-retry notice absent / stdout empty", is a defect signal, not a wait or timeout. Residual fixture defects: N-1 and N-2. |
| F2 | RESOLVED | acceptance L32 (AC-004 release-blocking for a, b, c), L162-L169 (c shares the b RED reason) |
| F3 | RESOLVED | spec L229-L233 content anchor; acceptance L371-L382 (sed-to-anchor comparison plus unchanged-after-anchor against `fa96fe644`). The anchor measured once per copy at L160. |
| F4 | RESOLVED | acceptance L336-L338 (A7 `BLOCKING=0`, A8 automatic vet-only, A9 automatic vet+build); mutants M15-M17 at L489-L491 |
| F5 | RESOLVED (residual in N-4) | acceptance L152-L156 (single-space, compact, multi-space forms); R10a/R10b at L307-L308; M14 at L488. REQ-005 now also claims tabs and top-level scope that no row pins (N-4). |
| F6 | RESOLVED | spec L138-L140 (payload removed before any check); acceptance L421 (S1 three-call sequence); M18 at L492 |
| F7 | RESOLVED (residual in N-3) | spec L172-L175; acceptance L179-L184 (L1, U1-U5); M19 at L493 |
| F8 | RESOLVED | spec L192-L214 (history-scoped invariant plus equivalence argument); acceptance L311-L315 (R13 Python, R14 `chore: sync`, R15 automatic both-fail, R16 empty opt-out, R17 legacy start), L319-L322 scope note. Re-derived independently: in any gate-written history the new record names the same last gated HEAD, so every blocking invocation of the old hook still reaches the checks. |
| F9 | RESOLVED | plan L80-L82 (retry marker is a separate file, "not an outcome token") |
| F10 | RESOLVED | plan L158-L159 (M1 covers AC-002 and AC-003) |
| F11 | RESOLVED | acceptance L253-L254 (a50, b70), L261-L262; M20 and M21 at L494-L495 |
| O1 | RESOLVED (weakness in N-5) | acceptance L71, L92-L94, L358-L360 (invocation-pattern regex that skips comments). The probe shows it misses `then jq` / `do jq`. |
| O2 | PARTIAL | acceptance L72 and L361-L366 add card-id/date checks on both template files. Commit SHAs are delegated to `TestTemplateNoInternalContentLeak`, whose coverage of these paths is an admitted unknown (L102-L103, L365-L366). |
| O3 | RESOLVED | spec L235-L241 (the manifest-step comment in both hooks is in scope; no `manifest audit` phrase allowed) |
| O4 | RESOLVED | spec L192, L216 |
| O5 | RESOLVED | acceptance L305 (R8 built behaviourally; the new run is proven by a stub-count increase) |
| O6 | RESOLVED | spec L264-L267 (ordering demoted to a process note); REQ-013 L229-L233 has no ordering clause |
| O7 | RESOLVED | spec L141-L142 (payload before `fail`), L174 (`fail` without payload re-gates); plan L58-L66 (B7) |

## Lead-decision encoding check (new since round 1)

| Decision | Encoded at | Faithful? |
|---|---|---|
| B7 — every failed run stores its payload (block or advisory) before the `fail` record; `fail` with no payload re-gates; `fail` with an advisory payload stays silent | spec L133-L142 (REQ-002), L166-L170 (REQ-006), L172-L175 (REQ-007); plan L58-L66, §C table L100-L104 | Yes. Cross-checked against older rows: A1-A3, A6 (advisory payload, silent), A4/A9 (block payload re-emitted), S1 (acceptance L421: an advisory run replaces the stale block payload, then call 3 is silent), U5 (no payload, re-gate). No contradiction found. |
| Torn writes (a) payload present with record not `fail`, and (b) neither written: never a silent pass; rows plus one mutant per direction | acceptance L196-L243; TA1-TA5, TB1-TB4; M23 (a) and M24 (b) at L497-L498 | Yes, with a wording substitution: acceptance L198-L200 reads "re-gate or non-blocking notice" where the lead said "re-gate or today's behaviour". The notice (TA1, TB1) follows from D2's fresh-running rule, and a stale record re-gates (TA2, TB2), so neither state is ever silent. Recorded for lead awareness as N-7. The RED reasons for TA1, TA4, and TB1 match the hook (L179-L186). |
| mtime-unreadable fallback as an explicit Gap | acceptance L105-L115, at the head of §D.0 before any criterion; DoD L512-L513 ("remains an explicit unverified gap; closing the card does not claim it"); plan L235 | Yes. A reader of acceptance.md meets it before the first AC, and the DoD disclaims it. |
| F1 RED reasons are defect signals | see F1 row above | Yes for AC-006c and N1. N2's named reason is wrong (N-1). |

Other constraints: two byte-identical hook copies (REQ-013 L229, AC-010 L370, measured cmp exit 0); content-anchored doc equivalence (F3); all new state under `.moai/state/` (REQ-011 L219-L220, §4 L251-L253); `hookSpecificOutput` first-match rule (§4 L254-L257, plan L118-L120); tier-driven mode (REQ-006, AC-008); jq-free (REQ-005 L156, REQ-012 L223); template neutrality (REQ-012 L227, AC-009 L361-L366); no `make build` (§4 L264-L266, plan L51-L55, DoD L517). All present.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity | REQ-005 L157-L159 claims tab whitespace and top-level scope, but plan L207-L211 describes a key-position match that would also count a nested key (N-4). REQ-002 L138-L140 says "set the retry marker", then qualifies it as "present if and only if …". REQ-008/009 leave the exact-60 s boundary undefined (N-6). |
| Completeness | 1.0 | all present | HISTORY L19, background §1 L46, decisions §2 L99, REQs §3 L122, constraints §4 L249, four `### Out of Scope —` H3s with bullets L271-L298; frontmatter complete |
| Testability | 0.75 | one criterion needs correction | Mostly binary with named RED reasons. Three defects: AC-015 N2 named RED reason misstates what M1 will observe (N-1); AC-006c/N1 marker lifecycle can false-red the green path (N-2); U4 sub-assertion RED reason (N-3) |
| Traceability | 1.0 | complete | Every REQ has an AC (matrix L29-L43); every AC names existing REQs; every §D.16 mutant M1-M24 names existing rows; referenced row ids (TB3→U5, TB4→R9/D1, AC-015→AC-006c, AC-014→a50/b70) all exist |

Harmonic mean 4 / (1/0.75 + 1 + 1/0.75 + 1) = 0.857, rounded to 0.86.

## Defects Found (structured defect-list)

D1. N-1 — acceptance.md:L447-L448, L466-L468 — AC-015 N2 refreshes only the record's **mtime** before each call, yet its named RED reason says "the `decision` count is 9 and the stub count grows". On `fa96fe644`, call 1 finds content differing from the bare SHA, runs the checks, blocks, and **rewrites the record to the bare SHA** (hook L183-L186). Calls 2-9 then take the L179-L182 short-circuit and emit nothing. M1 will observe a decision count of 1, one stub invocation, and 8 empty stdouts, so DoD L509-L510 ("observed RED reasons match the named reasons") is guaranteed to fail. The revision introduced the specific count; round 1 wording was generic. Read from the hook, not executed (see Gaps). — Severity: minor — Class: blocking — Required fix: either restate N2's RED reason to what the hook does ("call 1 blocks and the stub runs; calls 2-9 are empty; the fresh-running notice is absent on all 9"), or change the fixture to rewrite the full `<HEAD> running` content, not just the mtime, before every call. With the rewrite, the existing "9 / grows" reason becomes true.

D2. N-2 — acceptance.md:L266-L270; plan.md:L150-L155 — The interrupted-run harness kills on "the stub's marker file appears", but nothing says the marker is removed, or given a per-run path, between runs. Run 1's marker survives into run 2. On the green path, run 2 is the stale re-run, and the harness can kill it as soon as it starts, before the hook writes the retry marker and the `running` record. Run 3 then finds a stale record with no retry marker, runs the checks, and assertion 2 (stub count 0) goes red for a fixture reason. AC-015 N1 reuses the same harness (L445-L446). There is also no assertion that run 2 reached its own marker. Inferred from the fixture text. — Severity: minor — Class: blocking — Required fix: require the marker to be deleted (or given a per-run path) before each interrupted run, and add a setup-sanity assertion that run 2 observed its own marker, mirroring assertion 1.

D3. N-3 — acceptance.md:L183, L191-L193 — U4 prepares `<HEAD> fail` with mode 0200. The named reason for the record-format sub-assertion is "today rewrites a bare SHA". On `fa96fe644` the hook's `cat` fails silently (L179), and `echo "$HEAD_SHA" > file` (L185) writes through the owner-write bit while **keeping mode 0200**. The test's own read of the record therefore fails with permission denied instead of reading a bare SHA: a wrong-reason RED against DoD L509-L510. The green path is also fragile: REQ-001 does not require a rename-based record write (only plan §C L84-L85 does), so a redirect-based implementation leaves the file unreadable and fails the row on a correct fix. Read from the hook, not executed. — Severity: minor — Class: blocking — Required fix: have the U4 test restore read permission (for example `chmod 0600`) before reading the record, or name the permission-denied observation as U4's RED reason and add "the record is readable after the run" as an explicit green-path expectation.

D4. N-4 — spec.md:L156-L159 (REQ-005) vs acceptance.md:L150-L169, L488 — REQ-005 states two detector properties that no row pins. (i) It recognizes the field "whatever whitespace (none, one or more spaces, tabs)", but forms F1-F3 contain only spaces, so a detector that accepts spaces but not tabs passes every row. (ii) The field must be **top-level**, but no row feeds a nested `{"x":{"stop_hook_active":true}}`, and the plan's suggested match (plan L207-L211, a key-position test) does not enforce top-level, which contradicts the REQ text. Both are properties the SPEC explicitly claims. — Severity: minor — Class: blocking — Required fix: add an F4 tab form (`{"stop_hook_active":\ttrue}`) and a nested-key row, with mutants "spaces-only whitespace" and "accept nested key", **or** narrow REQ-005 to the properties actually tested (drop "tabs"; replace "top-level" with "in object-key position, not inside a string value") and align plan §G.

D5. N-5 — acceptance.md:L71, L92-L94, L358-L360 — The L-18 regex is presented as matching real invocations, but the probe matched 2 of 4 invocation lines. `if …; then jq …; fi` and `for …; do jq …; done` are not matched, so a `then jq` regression passes AC-009's jq check. This is a regression guard, and the hook has 0 jq tokens today (round-1 measurement at the same content, drift since `fa96fe644` empty). — Severity: minor — Class: optional — Required fix: add shell keywords before `jq` (`(^|[;&|(\`]|\$\(|\b(then|do|else|exec|xargs)[[:space:]])`), or keep the regex and extend the L-18 control with the two missed forms, stating the limit.

D6. N-6 — spec.md:L178-L182 (REQ-008), L184-L185 (REQ-009) — "within the stale window" and "older than the stale window" leave an age of exactly 60 s undefined. a50 and b70 cannot tell `<` from `<=`. — Severity: minor — Class: optional — Required fix: state which side 60 s falls on (for example "age < 60 s is fresh").

D7. N-7 — acceptance.md:L198-L200 — The torn-write rows say "re-gate or non-blocking notice", where the lead's decision text said "re-gate or today's behaviour". The substitution follows from D2 and is never silent, but it is a wording change to a lead decision. — Severity: minor — Class: optional — Required fix: add one sentence citing D2 as the reason a fresh torn state reads as a notice, so the lead can confirm the reading.

D8. N-8 — acceptance.md:L445-L451 — For AC-015 N1's 9 runs, the stub's behaviour (still sleeping, or exiting 1 immediately) is unspecified. Under mutant M6 or M11 a sleeping stub would stall each run until its own timeout rather than fail fast. — Severity: minor — Class: optional — Required fix: state that the stub exits 1 immediately for the 9 runs, as in AC-006c run 3 (L272).

D9. O2 (carried, PARTIAL) — acceptance.md:L102-L103, L365-L366 — REQ-012's "no commit SHA in template content" clause is still verified only by delegation to a guard whose coverage of these two paths is unmeasured. — Severity: minor — Class: optional — Required fix: in M1, measure that coverage (or add a 7-40-hex check scoped to the two template files) and record it as a ledger row.

## Regression Check (Iteration 2)

Defects from iteration 1: F1-F11 are all RESOLVED; O1, O3-O7 are RESOLVED; O2 is PARTIAL (optional). Table and evidence above. No defect is unresolved and no progress stalled. Score moved 0.60 → 0.86.

## Recommendation (for manager-spec)

1. Correct AC-015 N2's named RED reason, or its fixture, so the reason matches what `fa96fe644` does (D1).
2. Specify the marker-file lifecycle for the interrupted-run harness and add a run-2 marker sanity assertion (D2).
3. Fix U4's record read (restore permission before reading) or its named reason (D3).
4. Pin or narrow REQ-005's tab and top-level claims, and align plan §G (D4).
5. Optional: D5-D9.

All fixes change row text and fit inside the existing 15 criteria; the Tier M budget is untouched. A round-3 re-audit can be scoped to D1-D4 plus a spot check that no new row id or mutant number broke the M1-M24 sequence.

## Gaps

- The scratch-repo probes for N2 and U4 were refused by the worktree-session guard. D1 and D3 rest on reading hook L176-L186 at `fa96fe644`-identical content (drift empty), not on observed execution. The guard was not bypassed (no script-file workaround).
- D2 is inferred from the fixture text; the green-path race was not executed, because no M2 hook exists yet.
- No Go tests were run.
- The lint and `spec_audit` results come from build `84fa4ece4`, an ancestor of HEAD, not a build of this tree. Lint rules that landed after that commit did not run.
- The runtime Stop-hook stdin spacing and nesting were not observed. D4 rests on the stated REQ text and writable mutants, not on a measured payload.
- No cross-model backend audit was run (no `audit_model` key).
- Windows behaviour (AC-006c, U4 skipped there) was not assessed.

## Residual risk

- The jq-free key detector remains heuristic. Escape sequences such as `\\"` immediately before a real key were reasoned about, not probed.
- The mtime-unreadable fallback is unverified by design (acceptance L105-L115). On such a platform a stale record could keep emitting the fresh notice instead of re-gating.
- Concurrent sessions sharing `.moai/state/` stay out of scope (spec L296-L297). Interleaved records are possible without a lock.

## Operational Notes (unverified)

- `inferred` — D1 can be settled in one measurement at M1: run the N2 loop against the unchanged hook and read the per-call decision count before writing the RED cell. Rule applied: hook L179-L186 rewrite-before-checks.
- `assumption` — The L-18 regex gap (D5) likely also covers backtick command substitution. Measure it by adding a `` `jq .` `` line to the L-18 control.
