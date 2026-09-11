# SPEC Review Report: SPEC-SYNC-GATE-FAILSTATE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: FAIL
Overall Score: 0.60 (Tier M PASS threshold 0.80)
Auditor: plan-auditor (independent, adversarial). Reasoning context ignored per M1 Context Isolation — lead decisions were used only as the checklist to verify the encoding.
Card: t624 · Branch `WT-sync-gate-failstate` · Tree measured: HEAD `6eda95337`

## Baseline attribution

| Check | Command | Observed |
|---|---|---|
| tree | `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t624` |
| branch | `git branch --show-current` | `WT-sync-gate-failstate` |
| head | `git log --oneline -6` | `6eda95337 docs(t624): apply lead decisions B1-B6 …` (then `b569c5446`, `fa96fe644`, `d5dc42959`) |
| target drift since plan pin | `git diff --stat fa96fe644 HEAD -- <both hooks> <both docs> internal/hook internal/template/hook_official_compliance_test.go internal/template/templates/.claude/settings.json.tmpl` | empty — plan-phase pin `fa96fe644` still describes these files |
| hook copy parity | `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | empty, exit 0 |
| doc parity 1-161 | `diff <(sed -n 1,161p tmpl) <(sed -n 1,161p local)` | `IDENTICAL_1_161`; first difference at line 162 |
| doc line counts | `wc -l` both docs | template 322, local 326 |
| repro bytes | `wc -c .moai/reports/t624/h01-develop-*.out` | call1 237, call2 0, ctrlA 237, ctrlB1 0, ctrlB2 0 |
| settings timeout | `grep -n sync-phase-quality-gate -A4 settings.json.tmpl` | line 156 `"timeout": 60` |
| no `.sh.tmpl` | `ls templates/.claude/hooks/moai/ \| grep sync-phase` | only `sync-phase-quality-gate.sh` |
| jq in hook | `grep -c jq <template hook>` | `0` |
| auditor rubric | `grep -n "Critical/High" .claude/agents/moai/sync-auditor.md` | L38 `\| Security \| 25% \| OWASP Top 10 compliance \| Any Critical/High finding \|` |
| lint binary | `command -v moai`; `moai version` | `/Users/goos/go/bin/moai`; `v3.2.0-rc.5 list-974-g84fa4ece4 built 2026-09-09T02:48:44Z` |
| lint binary provenance | `git merge-base --is-ancestor 84fa4ece4 HEAD` | exit 0 → the installed build is an older ancestor of HEAD (not built from this tree) |
| lint | `moai spec lint SPEC-SYNC-GATE-FAILSTATE-001` | `0 error(s), 0 warning(s)`, one INFO `OwnershipTransitionUnmeasured` (commit `b569c5446` lacks `Authored-By-Agent` trailer); exit 0 |
| spec_audit (MCP, project_root = worktree) | `mcp__moai__spec_audit filter_spec=SPEC-SYNC-GATE-FAILSTATE-001` | era V3R6, 1 INFO `EraAutoDetected` (H-5) |
| D7 refs | `grep -rnoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' <spec dir>` | only its own ID |
| D8 | `grep -rc syscall <spec dir>` | 0 in all four files |
| MP-7 | `grep -rn "NEEDS CLARIFICATION" <spec dir>` | no match, exit 1 |
| line-anchor probe | `tail -n +161 local.md > a`; `tail -n +162 local.md > b`; `cmp a b` | `differ: char 1, line 1`, exit 1 |

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: spec.md L111-L196 carries REQ-001..REQ-015 in order with no gaps or duplicates; acceptance.md §D matrix L24-L38 carries AC-001..AC-015. Tier M budget 15/16 each.
- [PASS] MP-2 GEARS format (judged on the REQ layer in spec.md only): REQ-001/012/013/014/015 Ubiquitous; REQ-002/003/005/006/007/009 When; REQ-004 two When clauses; REQ-008 While+When. REQ-010 (quantifier prefix) and REQ-011 (gerund subject "Removing …") are loose but still carry a subject + "shall" (optional finding O4). ACs are Given-When-Then (verification layer, not graded here).
- [PASS] MP-3 frontmatter: spec.md L2-L13 carry all 12 canonical fields, correct names and types (`version: "0.1.0"` quoted, `priority: P1`, `lifecycle: spec-anchored`, `phase: "v3.2.0 target"` is not a lifecycle-stage name, `tags` comma string); `tier: M` optional. Lint reports 0 errors (older lint build — see Gaps).
- [PASS] MP-4 language neutrality: the normative change is state logic after language detection; REQ-012 L174-L176 keeps all language branches unchanged. Go stubs appear only in `internal/hook` test fixtures, not in template content. `govulncheck` appears only in Out of Scope (spec.md L222) as an example, not a default.
- [PASS] MP-5 D7: no other SPEC referenced — no BLOCKING finding.
- [PASS] MP-6 D8: `syscall` count 0 — auto-pass.
- [PASS] MP-7 clarification gate: no `[NEEDS CLARIFICATION` markers in plan.md; research.md does not exist (Tier M).

## Decision-encoding check (lead decisions D1-D6, B1-B6)

| Decision | Encoded at | Faithful? |
|---|---|---|
| D1 outcome record + verbatim re-emit; legacy SHA re-gates once | spec L88-L91, REQ-001/002/003/007 | Yes; but plan.md L65-L67 suggests a `running-retry` token that REQ-001 forbids (F9) |
| D2 stale running re-gates once; fresh → notice; never looser | spec L92-L96, REQ-008/009/010 | Encoded, but REQ-010's wording is self-contradictory and the matrix is not exhaustive (F8); window value not behaviourally pinned (F11) |
| B1 one retry, then non-blocking notice, never decision, never re-run | REQ-009, AC-006c, AC-015 | Encoded; RED cells fail for the wrong reason (F1) |
| B2 mode re-resolved at re-delivery; advisory never retroactively blocked | REQ-006, AC-008 | Encoded; mutant gaps (F4) and a reachable retroactive-block sequence (F6) |
| B3 advisory warning once | REQ-006 last sentence, AC-008 A1-A3 | Yes |
| D3 stop_hook_active suppresses stored re-delivery only, jq-free | REQ-005, AC-004, R10 | Encoded; compact-form mutant gap (F5); AC-004c misclassified (F2) |
| D4 retry = delete state file | REQ-011, AC-013 | Encoded; auxiliary state not invalidated (F6) |
| D5 H03 wording only | REQ-014, AC-011 | Yes (doc L116, L140-L146, hook header L3 verified); hook L287 comment left open (O3) |
| D6 sync-auditor rubric canonical, wording only | REQ-015, AC-012 | Yes (doc L71, L136-L137, L156; auditor L38 verified) |
| B4 RED cells adopted at M1 test-only commit | acceptance L10-L15, plan L43-L46, L124-L129 | Yes |
| B5 no `make build` | plan L47-L51, spec L214-L216 | Yes |
| B6 doc L43 out of scope, one line | spec L234 | Yes (L43 claim verified present) |

Current-behavior claims checked against the hook: sentinel written before checks (hook L183-L186) and whole-content equality short-circuit (L179-L182) — confirmed; manifest step L287-L314, `deps_modified` informational only — confirmed; compliance test takes the first line containing `printf`+`decision`+`block` (hook_official_compliance_test.go L96-L101) — confirmed; stopchain tests reset with `os.RemoveAll(.moai/state)` (ac004_006b L129, trim L192) — confirmed.

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | several requirements need interpretation | REQ-010 L164-L166 contradicts REQ-004/006/008 literally (F8); malformed same-HEAD record undefined (F7); plan §C token vs REQ-001 (F9); persisted-block lifetime undefined (F6) |
| Completeness | 0.75 | sections present, state machine gaps | All sections + Out of Scope H3s present (spec L218-L247); frontmatter complete |
| Testability | 0.50 | several criteria unsound | wrong-reason RED AC-006c/AC-015 N1 (F1); AC-004c misclassified (F2); AC-010 line anchors fail under any net line change (F3); writable mutants for REQ-005/006/008 (F4, F5, F11) |
| Traceability | 0.75 | one indirect gap | every REQ has an AC; AC-002/AC-003 assigned to no milestone (F10); REQ-013 ordering clause unverifiable (O6) |

Harmonic mean ≈ 0.60.

## Defects Found

D1. F1 — acceptance.md:L157-L167 (AC-006c) and L283-L305 (AC-015 N1) — Both fixtures kill the hook "once the record reads `running`", but on `fa96fe644` the hook writes a bare SHA before the checks (hook L183-L186), so the record never reads `running`; the M1 RED cell will be a setup wait/timeout, not the stated notice-assertion failure (wrong-reason red, verification-completeness §2). — Severity: major — Class: blocking — Required fix: trigger the kill on a marker the stub `go` writes before sleeping (or on its counter file), and restate the RED reason to match what M1 will actually observe.

D2. F2 — acceptance.md:L131-L134 (AC-004c), plan.md:L127-L128 — AC-004c is classed regression-guard, yet on `fa96fe644` nothing is re-delivered, so the `stop_hook_active:false` call writes 0 bytes and fails the "byte-identical to call 1" assertion; plan M1 says regression-guard rows must PASS on that commit. — Severity: minor — Class: blocking — Required fix: reclassify AC-004c as release-blocking, adopted together with (b), with the same RED reason as AC-004b.

D3. F3 — spec.md:L180-L183 (REQ-013), acceptance.md:L224-L230 (AC-010), plan.md:L89-L91 — Doc parity is anchored to line numbers (1-161 identical; 162 onward unchanged vs `fa96fe644`). The local copy diverges exactly at L162, and `cmp` of `tail -n +161` against `tail -n +162` differs at line 1, so any net line-count change M3 makes — including the D6 reconciliation sentence plan M3 L146-L147 requires — fails AC-010, and a net deletion breaks "identical over 1-161". — Severity: major — Class: blocking — Required fix: anchor on content, not line numbers. Example: both copies identical from the start through the line `Purpose: Ensure code has appropriate @MX annotations …` (the first Phase 9 line, doc L160), and everything after that line unchanged vs `fa96fe644`. Keeps the lead's intent (no edit past the boundary) without forcing a net-zero line count.

D4. F4 — spec.md:L136-L144 (REQ-006), acceptance.md:L196-L211 (AC-008), L315 (M4) — REQ-006 resolves mode at re-delivery from three inputs, but only the tier (fully-autonomous, A5) varies between a blocking call 1 and call 2. Two mutants pass every row: one that ignores `MOAI_SYNC_GATE_BLOCKING=0` at re-delivery, one that treats `automatic` as blocking regardless of composition. — Severity: major — Class: blocking — Required fix: add AC-008 rows — (A7) call 1 unset + vet fail → call 2 `MOAI_SYNC_GATE_BLOCKING=0` → empty; (A8) call 1 unset + vet-only fail → call 2 tier `automatic` → empty; (A9) tier `automatic` with vet and build both failing → re-emit. Pair each with a mutant in §D.16.

D5. F5 — spec.md:L128-L134 (REQ-005), acceptance.md:L124-L134, L187, L287-L290 — Every `stop_hook_active` fixture uses the spaced form `"stop_hook_active": true`. A detector matching only that literal passes AC-004a/b/c, R10 and AC-015, yet misses a compact `"stop_hook_active":true` payload and re-emits the stored block while the flag is set — exactly the block-cap hazard D3 exists to prevent. (The runtime's payload spacing was not verified in this audit.) — Severity: major — Class: blocking — Required fix: add compact-form and extra-whitespace variants to AC-004a/b and a mutant "match spaced literal only" to §D.16.

D6. F6 — spec.md:L117-L144 (REQ-002/003/006), L168-L171 (REQ-011), plan.md:L60-L67 — No requirement invalidates the persisted block (or a separate retry marker) when a later check-running invocation on the same HEAD emits no block. Reachable through the documented retry: call 1 blocking fail (block persisted) → delete `.last` → call 2 tier fully-autonomous (advisory, record `fail`, stale block file remains) → call 3 unset → REQ-003's condition "fail + persisted block exists" holds → old block re-emitted. That retroactively blocks an advisory run, contradicting B2 and REQ-006's second sentence. — Severity: major — Class: blocking — Required fix: require every run that executes the checks to atomically replace or remove all per-HEAD auxiliary state (persisted block, retry marker); add the three-call sequence as an AC-008 or AC-013 row.

D7. F7 — spec.md:L111-L162 (REQ-001..009), plan.md:L71-L82 (§C table) — A same-HEAD record whose content is neither a bare SHA nor one of the three tokens (unknown or extra token, empty file, unreadable file) is undefined. Today's hook (L179) runs the checks and blocks on any content other than the bare SHA; an implementation that parses `<sha> <anything>` as non-fail can go silent. — Severity: major — Class: blocking — Required fix: add to REQ-007: "an unrecognized or unreadable record for the current HEAD is treated like the legacy format: run the checks and rewrite the record", plus one AC-005 variant row.

D8. F8 — spec.md:L94-L96 (D2), L164-L166 (REQ-010), acceptance.md:L170-L194 (AC-007) — REQ-010 claims to cover "every input combination under which the hook on `fa96fe644` emits a block". Read literally it contradicts REQ-004/006/008: today a `<HEAD> pass`, `<HEAD> fail` (no block) or fresh `<HEAD> running` record blocks, because the content differs from the bare SHA, but those REQs require silence or a notice. Read as the §D.7 list, the list is not exhaustive: it omits 3 of the 4 sync-subject patterns (hook L131), every non-Go language branch (C1-only, L228-L284), `automatic` with vet and build both failing (L356), and `MOAI_SYNC_GATE_BLOCKING` set to empty or to a non-opt-out value (L335). — Severity: major — Class: blocking — Required fix: restate the invariant by history — "for any invocation sequence starting from no state or a legacy record, every invocation the `fa96fe644` hook blocks is blocked by the new hook." State the equivalence-class argument (state logic runs only after the unchanged early exits, plan §C L84-L85), and add representative rows: one non-Go language with a failing C1 stub, one `chore: sync` subject, `automatic` with both checks failing.

D9. F9 — plan.md:L65-L67 vs spec.md:L111-L113 (REQ-001) — The plan's example retry-marker token `running-retry` violates REQ-001's closed outcome set {running, pass, fail}. — Severity: minor — Class: blocking — Required fix: drop the example and say the marker lives in a separate file under `.moai/state/` (kept consistent with D6), or widen REQ-001 explicitly.

D10. F10 — plan.md:L122-L123 (M1 coverage list) vs acceptance.md:L16-L18, L103-L121 — AC-002 and AC-003 belong to no milestone's test list, yet the legend requires every regression-guard baseline to be observed as a PASS on the M1 commit; as planned, no test for them is ever written. — Severity: minor — Class: blocking — Required fix: add AC-002 and AC-003 to the M1 test list.

D11. F11 — spec.md:L150-L154 (REQ-008), acceptance.md:L150-L156 (AC-006a/b), L271-L279 (AC-014) — The 60 s window is not pinned by behaviour: AC-006a (age 0) and AC-006b (age 120 s) pass for any window strictly between 0 and 120 s, and AC-014 reads a declared variable, not its use. Mutant: declare the variable as 60 but compare against 100 — every criterion stays green. — Severity: minor — Class: blocking — Required fix: add boundary rows (record age 50 s → notice, no checks run; age 70 s → re-run) and a matching mutant in §D.16.

D12. O1 — acceptance.md:L221-L222 (AC-009) — `grep -c "jq"` printing 0 fails a comment that documents the jq-free detection, which REQ-012 does not forbid (it bans calls). — Severity: minor — Class: optional — Required fix: match an invocation pattern (e.g. `\bjq[[:space:]]`), or state that comments must not name the tool.

D13. O2 — acceptance.md:L221-L222 vs spec.md:L177-L178, plan.md:L97 — Template-neutrality checks cover only the `SPEC-` token; card ids, dates, and SHAs in the hook, plus all neutrality checks on the template doc, go unchecked. — Severity: minor — Class: optional — Required fix: extend AC-009 to the doc copy and to `t[0-9]+` / date / 7-40-hex patterns, or cite the CI neutrality guard as the verifier.

D14. O3 — spec.md:L185-L190 (REQ-014), acceptance.md L61 (L-12 = 2) — The hook's second `manifest audit` comment (hook L287) is neither required to change nor explicitly allowed to stay. — Severity: minor — Class: optional — Required fix: state which.

D15. O4 — spec.md:L164, L168 — REQ-010's quantifier prefix and REQ-011's gerund subject are not canonical GEARS shapes. — Severity: minor — Class: optional — Required fix: rephrase as "The gate shall …".

D16. O5 — acceptance.md:L185, plan.md:L60-L62 — The R8 fixture needs a persisted block, whose shape run-phase does not choose until M2; on the M1 commit R8 degenerates into R7. — Severity: minor — Class: optional — Required fix: build R8 behaviourally (a failing run on the previous commit, then a new commit).

D17. O6 — spec.md:L182-L183 (REQ-013) — "Each edit shall be made in the template copy first" cannot be witnessed from a single commit (verification-claim-integrity §2.3). — Severity: minor — Class: optional — Required fix: drop the ordering clause or require it to be visible in the commit graph.

D18. O7 — spec.md:L141-L144 (REQ-006) — "No persisted block" is equated with "first run was advisory"; a failed block-file write under blocking mode would read as advisory and go silent. — Severity: minor — Class: optional — Required fix: persist the resolved mode in the record, or require the block write to precede the `fail` record write.

## Recommendation (for manager-spec)

1. Fix the AC-006c and AC-015 N1 fixtures and RED reasons (D1), and reclassify AC-004c (D2).
2. Replace the line-number doc anchors in REQ-013 and AC-010 with content anchors (D3).
3. Close the state-machine holes: invalidate auxiliary state on every check-running invocation (D6); define the unrecognized-record path (D7); align the plan §C marker with REQ-001 (D9).
4. Restate REQ-010 as a history-based invariant and add the missing representative AC-007 rows (D8).
5. Add the mutant-closing rows: AC-008 A7-A9 (D4), compact `stop_hook_active` (D5), stale-window boundary (D11); pair each with a §D.16 mutant.
6. Put AC-002 and AC-003 in the M1 test list (D10).

All additions fit inside the existing 15 criteria as extra rows, so the Tier M budget is untouched.

## Gaps

- No Go tests were executed; RED/GREEN claims were judged by reading hook and test code at HEAD `6eda95337`.
- The lint and `spec_audit` results come from build `84fa4ece4`, an ancestor of HEAD, not from a build of this tree; lint rules that landed after that commit did not run.
- The runtime Stop-hook stdin spacing (compact vs spaced JSON) was not observed; D5 rests on the writable mutant, not on a measured payload.
- No cross-model backend audit was run: `grep -rn audit_model .moai/config/sections/` returned nothing, and the backends review code diffs rather than SPEC documents.
- Windows behaviour (AC-006c is Unix-only) was not assessed.

## Residual risk

- Even with the rows added, the new jq-free key detection is only as good as its probes; nested JSON or unusual escapes beyond those probed may still misparse.
- The stale-window logic depends on mtime portability (plan §G); no criterion exercises the "mtime unreadable → treat as stale" fallback.
