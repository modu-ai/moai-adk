# SPEC Review Report: SPEC-INTEGRATION-LOCK-LIVENESS-001 (card t298)
Iteration: 1/2 (Tier M ceiling)
Verdict: FAIL
Overall Score: 0.795 (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. Every premise below was re-measured on
tree `d29b8942e` @ `WT-integration-lock` in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t298`.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — REQ-INL-001..010, sequential, no gaps, no duplicates,
  consistent zero-padding (`grep -o 'REQ-INL-[0-9]*' spec.md | sort -u` → 001..010).
- [PASS] MP-2 GEARS format compliance (requirement layer only) — all 10 `REQ-INL-*` entries in
  spec.md carry a GEARS pattern: Ubiquitous (001), Event-driven (002, 005, 009), State-driven
  (004), Where/capability-gate (006, 007, 008), Unwanted "shall not" (010). REQ-INL-003's
  sentence is Event-driven ("**When** no owner pid can be resolved … shall record pid 0"); only
  its parenthetical label is non-canonical (D7). Judged against the requirement layer; the
  Given-When-Then entries in acceptance.md are the verification layer and were graded under
  Group 4, not here.
- [PASS] MP-3 YAML frontmatter validity — 12 canonical fields present with canonical names
  (`created`/`updated`/`tags`, no snake_case aliases) plus optional `tier: M`. `moai spec lint`
  over the worktree: `0 error(s), 64 warning(s)`, and **zero** findings naming this SPEC
  (`grep -i 'INTEGRATION-LOCK-LIVENESS' <lint output>` → no finding rows). `phase: "v3.1.4 target"`
  is a release label, not a prohibited lifecycle stage.
- [N/A] MP-4 language neutrality — single-language (Go) project scope; no template-bound or
  multi-language tooling content. Auto-passes.
- [PASS] MP-5 D7 cross-SPEC reconciliation — the only `SPEC-…` token in the artifact set is the
  SPEC's own ID (`grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' *.md | sort -u` → 1 entry). No
  external reference, so nothing to reconcile. No BLOCKING finding.
- [PASS] MP-6 D8 cross-platform discipline — `grep -n 'syscall' *.md` → no match in any of the
  four artifacts. D8 auto-PASS.
- [PASS] MP-7 clarification gate — `grep -rn 'NEEDS CLARIFICATION' <spec dir>/` → rc=1, no match.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | Requirements are implementation-neutral and single-interpretation; plan.md §B names the exact field (`PIDSource`), the exact deleted lines (integration_lock.go:174-176), and the exact seam signature. Deductions: D9 (REQ-INL-006 reads as if `pid_source` gates a distinct evaluation branch, which plan §B.4 contradicts), D7 (label). |
| Completeness | 0.80 | 0.75 | All sections present (HISTORY §H, WHY §A, WHAT §C, HOW plan §B/§F, REQUIREMENTS §C, ACCEPTANCE acceptance.md, four `### Out of Scope — <topic>` H3 sub-headings at spec.md:215/219/224/228 each with `-` bullets). Design-space evaluation with a rejected Option A and a stated verdict. Deductions: D4 (the acquire read-modify-write concurrency hazard is unaddressed anywhere in §D/§G). |
| Testability | 0.70 | 0.50-0.75 | Most ACs name a mechanical shape (child-process exec + `integration status --json` field, `grep`, `GOOS=windows go vet`). Deductions: D2 (AC-INL-011's RED premise is false as measured and its Then carries no command), D1 (no AC asserts the schema marker is written — mutant-satisfiable), D5 (unattributed regex claim in progress.md §E.1). |
| Traceability | 0.85 | 0.75-1.0 | Every REQ-INL-001..010 maps to ≥1 AC; I verified each matrix row semantically against the REQ text — no off-by-one, no mis-mapping. Deductions: D6 (`REQ-INL-004, 008` abbreviated form makes `grep REQ-INL-008 acceptance.md` return 0), D1 (REQ-INL-002's marker clause has no AC). |

Aggregate = harmonic mean(0.85, 0.80, 0.70, 0.85) = 0.795 < 0.80.

## Verified Premises (spec.md §B re-measured, all cited line numbers)

| Premise | Measured result |
|---|---|
| integration_lock.go:174-176 `if want.PID == 0 { want.PID = os.Getpid() }` | CONFIRMED verbatim at those lines |
| integration_lock.go:91-99 `Stale()` (Held gate → PID<=0 false → !FactoryProcessAlive) | CONFIRMED |
| integration_lock.go:84-90 asymmetry doctrine comment | CONFIRMED verbatim |
| integration_lock.go:71-79 struct fields | CONFIRMED |
| internal/cli/integration.go:~148-152 "held by a session that is gone (reclaimable)" | CONFIRMED |
| internal/cli/integration.go:~80-92 `integrationSessionID` chain (flag → env → side-channel) | CONFIRMED |
| internal/hook/integration_lock_guard.go:~95 stale branch allows + reclaim advisory | CONFIRMED |
| session_pid.go resolution order (env-live → nearest non-wrapper ancestor → `os.Getpid()`) | CONFIRMED; wrapper set includes `moai`; `ancestorSessionPID` returns 0 when unresolvable, so a fallback-free seam is genuinely constructible |
| proc_info_other.go (Windows) returns ok=false | CONFIRMED — ancestry unavailable ⇒ pid 0 conservative is a real degradation path, not an assumption |
| internal/cli/integration_lock_cli_test.go `runIntegration` at line 26 | CONFIRMED (spec cites 26-43) |
| internal/cli/launch_session_pid.go exists | CONFIRMED |
| gitflow-lane-protocol.md §3 caveat blockquote at lines 42-43 | CONFIRMED verbatim |
| registry.go `DefaultStaleMinutes = 30`, `PID: resolveSessionPID()` | CONFIRMED |

## Defects Found (structured defect-list)

D1. AC-GAP-PIDSOURCE — acceptance.md §D.1 (all rows) / spec.md REQ-INL-002 — **No acceptance
criterion asserts that the owner-anchor marker is actually written.** REQ-INL-002 requires the
acquire verb to record "an additive owner-anchor marker in the lock record", and plan.md §B
names `pid_source: "session-owner"` as the first (hardest-to-reverse) design consequence. But
every AC's **Then** clause asserts only `stale`/held state and the recorded pid: AC-INL-001 and
AC-INL-002 assert `stale == false` + pid equality; AC-INL-004 and AC-INL-007 mention
`pid_source` only in their **Given** (test-constructed fixtures), never as an outcome.
Mutant probe (verification-completeness.md §2): an implementation that resolves the owner pid,
writes it to `PID`, and never adds the `PIDSource` field at all satisfies AC-INL-001 through
AC-INL-011 while violating REQ-INL-002 and REQ-INL-006's precondition. The schema decision the
plan ranks as least reversible is the one decision no criterion measures.
— Severity: major — Class: blocking — Required fix: add a Then-clause assertion to AC-INL-001
(or a new AC) that the on-disk record written by the real acquire carries
`pid_source == "session-owner"`, verified by reading the JSON record (or `status --json`) after
the child exits. State its RED-now cell (the field does not exist on `d29b8942e`) and its
flipping milestone (M2).

D2. FALSE-RED-PREMISE — acceptance.md AC-INL-011 (lines 138-144) + spec.md §B item 13
(line 99-100) — **AC-INL-011's RED-now cell states a premise that is false on the baseline
tree.** The cell asserts "§4.1.2's operating procedure on `d29b8942e` still instructs the
lead-notice-only workaround". Measured: `CLAUDE.local.md` §4.1.2 (extracted with
`awk '/^#### §4.1.2/,/^#### §4.1.3/'`) is a pure bash procedure block naming
`moai integration acquire --name <lane>` / `EnterWorktree` / `git merge --no-ff` / `git push` /
`moai integration release` — it contains no lead-notice prose at all. Confirming grep:
`grep -n '리드 공지\|리드가 창\|integration acquire' CLAUDE.local.md` returns exactly one line
(312, the `acquire` command inside the block); the lead-notice-only workaround text exists ONLY
in `gitflow-lane-protocol.md` §3, which AC-INL-010 already owns. spec.md §B item 13 repeats the
same false parenthetical ("lead-notice-only serialization prose") for CLAUDE.local.md.
Compounding: AC-INL-011's Then ("the prose describes the restored two-layer serialization")
carries no command and no observable output, so it is unverifiable as written — the one AC in
the set with neither a true RED reason nor a mechanical check.
— Severity: major — Class: blocking — Required fix: re-measure what CLAUDE.local.md §4.1/§4.1.2
actually says, restate the RED cell against that measured text (or drop the AC and fold the
CLAUDE.local.md edit into AC-INL-010's scope), and give the Then a grep with an expected hit
count. Correct spec.md §B item 13's parenthetical in the same pass (cross-layer revision sweep,
verification-completeness.md §3).

D3. PREFLIGHT-EXPECTS-WRONG-OUTPUT — plan.md §C line 97-98 — the pre-flight instructs
`grep -rn 'internal/kanban' internal/session/` → "0 hits expected". Measured: **1 hit** —
`internal/session/session_pid.go:49`, a comment referencing `internal/kanban/factory_slots.go`.
The conclusion the check exists to establish (no import cycle) still holds — it is a comment,
not an import — but the check as written fails at arrival, so the implementer either stops on a
false blocker or learns to read a failing pre-flight as noise.
— Severity: minor — Class: blocking — Required fix: scope the check to imports, e.g.
`grep -rn '"github.com/modu-ai/moai-adk/internal/kanban"' internal/session/` → 0 hits.

D4. UNADDRESSED-CONCURRENCY — spec.md §D / §G (absent) and plan.md M5 — **the acquire path's
read-modify-write is unserialized, and the scheduled doc rewrite is positioned to overclaim.**
Measured: `grep -n 'Flock\|flock\|LockFile' internal/kanban/integration_lock.go
internal/cli/integration.go` returns only three prose occurrences in the package header comment
and no call site. `AcquireIntegrationLock` (integration_lock.go:146-181) does
`ReadIntegrationLock` → decide → `writeIntegrationLock` with nothing between them, so two lanes
acquiring concurrently can both read "free" (or both read the same stale record), both write,
and both believe they hold the window — last write wins silently. The package header's own claim
that "the flock discipline is borrowed only to serialize mutations of that record" is not backed
by the code. Today this is masked (every holder reads reclaimable, so serialization is absent
anyway); after this fix serialization becomes real everywhere EXCEPT this window, and plan.md M5
schedules doc prose asserting "the restored serialization guarantee" and "the record is the
mechanical layer under it". The SPEC names no concurrency hazard anywhere.
— Severity: major — Class: blocking (for the claim; the flock itself is optional/out-of-scope)
— Required fix: name the hazard in spec.md §G Risks with its measured basis, and bound M5's
doc wording so it claims only what lands (liveness anchored to the session process), not
mutual exclusion of concurrent acquires. If serializing the record mutation is wanted, it is a
separate card — do not widen this one.

D5. UNATTRIBUTED-VERIFICATION-CLAIM — progress.md §E.1 line 9 — "SPEC ID pre-write regex check
executed as Bash, verbatim output: `PASS`" names neither the command nor the regex. Measured
against the canonical schema regex (`spec-frontmatter-schema.md`: `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`),
this ID FAILS: `echo 'SPEC-INTEGRATION-LOCK-LIVENESS-001' | grep -Eq '^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$'`
→ non-zero (the multi-hyphen middle segment is not `[A-Z0-9]+`). The claimed PASS was therefore
produced by some other, looser pattern that is not stated (VCI §2 baseline attribution).
— Severity: minor — Class: blocking — Required fix: state the exact command and regex used, or
drop the line. Note: multi-segment IDs are an established repo-wide pattern and `moai spec lint`
emits no ID finding for this SPEC, so the ID itself is not the defect — the unattributed claim is.

D6. ABBREVIATED-TRACE-TOKEN — acceptance.md §D matrix line 21 — the AC-INL-008 row writes
`REQ-INL-004, 008`, so a mechanical `grep -c 'REQ-INL-008' acceptance.md` returns **0** while
every other REQ returns ≥1. The mapping itself is correct (I verified all 11 rows semantically
against the REQ text — no off-by-one), but the traceability table is not machine-readable.
— Severity: minor — Class: optional — Required fix: write full IDs (`REQ-INL-004, REQ-INL-008`).

D7. NON-CANONICAL-PATTERN-LABEL — spec.md REQ-INL-003 (line 117) — labeled "(Event-detected)",
which is not one of the five GEARS pattern names. The sentence itself is a valid Event-driven
requirement, so MP-2 is unaffected.
— Severity: minor — Class: optional — Required fix: relabel "(Event-driven)".

D8. RISK-ENUMERATION-GAP — spec.md §G — two residual hazards of the chosen anchor are unlisted:
(a) **pid reuse** — a recorded owner pid whose session died and whose pid the OS later reassigns
reads live, wedging the window until `--force` (the conservative direction, consistent with the
asymmetry, but unstated); (b) the record carries no host field (`IntegrationLock` struct,
integration_lock.go:71-79) while `session.Entry` does, so the pid probe is implicitly
single-host.
— Severity: minor — Class: optional — Required fix: add both to §G with their fail-direction.

D9. REQ-VS-PLAN-BRANCH-AMBIGUITY — spec.md REQ-INL-006 vs plan.md §B consequence 4 —
REQ-INL-006 reads as though the absent marker selects a distinct evaluation path ("shall
evaluate it with the pre-existing probe semantics"), while plan §B.4 states `Stale()` semantics
are unchanged for both shapes (the existing `PID <= 0` / probe logic already yields the right
answer for each). Outcomes agree; the phrasing invites an implementer to add a `pid_source`
branch that would be dead code.
— Severity: minor — Class: optional — Required fix: reword REQ-INL-006 as an invariant
("shall continue to read as it does today") rather than as a branch.

## Answers to the dispatch's six judgment questions

1. **Failing reproduction first, post-fix assertion is its exact inversion** — YES. plan.md M1
   is the first milestone and its exit criterion is the observed FAILURE
   (`go test ./internal/cli/ -run 'TestIntegrationOwnerLiveness' -v` → both fail, verbatim output
   recorded in progress.md §E.2); AC-INL-001's RED cell (`stale: true` / reclaimable) and green
   cell (`stale: false`, recorded pid == parent's) are exact inversions. The cross-process shape
   is necessary and correctly justified: the existing in-process `runIntegration` helper
   (integration_lock_cli_test.go:26) cannot express "the acquire CLI process exits", and the
   `go build -o` / not-`go run` constraint (plan §D) is correct — an intermediate `go` process is
   not in the wrapper set. The ancestry path is achievable as specified: the child `moai`'s
   nearest non-wrapper ancestor is the test binary, since `moai` IS in `wrapperProcessNames`.
2. **Liveness identity precisely specified** — YES, to the field name, the value, the deleted
   lines, and the seam signature. The conservative fallback is fully specified (pid 0 + marker;
   `Stale()` already returns false for `PID <= 0` at integration_lock.go:95-97; force-release
   only). One residual: D1 means the marker half of that specification has no acceptance
   coverage, and D9 leaves its evaluation role ambiguous.
3. **Mechanically verifiable ACs** — MOSTLY. AC-INL-001/002/003/004/007 name a runnable shape
   with an inspectable outcome; AC-INL-009 names `GOOS=windows GOARCH=amd64 go vet ./internal/...`
   plus a `git diff --stat` check; AC-INL-010 names a grep with an expected hit count. AC-INL-011
   is the exception — no command, no expected output, and a false RED premise (D2).
4. **Scope minimal** — YES. One field added, one three-line fallback deleted, one seam exported,
   tests, two local-only docs. Four `Out of Scope` sub-headings explicitly exclude deny-layer
   enablement, registry-authoritative staleness, writability enforcement, and launcher/hook
   changes. §G anti-patterns forbid the obvious over-reaches (no `--pid` override, no schema
   rename, no guard-pattern widening). No redesign of the lock beyond the liveness identity.
5. **Both doctrine documents covered** — PARTIALLY. gitflow-lane-protocol.md §3 is covered
   correctly and its RED premise is verified true (caveat blockquote present at lines 42-43).
   CLAUDE.local.md §4.1/§4.1.2 is nominally covered by AC-INL-011, but on a premise that does not
   hold (D2), so that half of REQ-INL-009 is not actually gated by a valid criterion.
6. **Concurrency hazards** — one substantive miss: the unserialized acquire read-modify-write
   (D4), which the SPEC never names and whose absence the scheduled doc prose would paper over.
   Secondary misses: pid reuse and the absent host field (D8). The status-then-acquire TOCTOU the
   lane performs is already governed by existing doctrine (the probe is the last check, never the
   first) and is not a new defect of this SPEC.

## Recommendation

FAIL at 0.795 against the Tier M threshold of 0.80. The defect delta is small and surgical —
this is a well-constructed SPEC whose premises are, with the exceptions named, verifiable and
verified. Fix in this order, then re-audit scoped to the delta:

1. **D1** — add the `pid_source` presence assertion to a Then clause (closes the mutant on the
   least-reversible decision).
2. **D2** — re-measure CLAUDE.local.md §4.1/§4.1.2, restate AC-INL-011's RED cell against the
   measured text, give it a grep, and correct spec.md §B item 13's parenthetical.
3. **D4** — add the acquire-race hazard to spec.md §G and bound M5's doc wording to what lands.
4. **D3** — scope the pre-flight grep to imports.
5. **D5** — attribute or drop the progress.md §E.1 regex claim.
6. Optional (orchestrator's discretion): D6, D7, D8, D9.
