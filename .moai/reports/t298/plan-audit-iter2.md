# SPEC Review Report: SPEC-INTEGRATION-LOCK-LIVENESS-001 (card t298)
Iteration: 2/2 (Tier M ceiling)
Verdict: PASS-WITH-DEBT
Overall Score: 0.895 (Tier M PASS threshold 0.80) — delta vs iter1 0.795 = **+0.100, monotonic**

Reasoning context ignored per M1 Context Isolation. Iteration-2 scope: the delta against the
iter-1 audited state, plus what the delta could have broken. Every claim below was re-measured
in this run, in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t298`.

## §0 Tree-state correction — the dispatch's baseline is stale (report, not a defect of the SPEC)

The dispatch named `d29b8942e` as HEAD. Measured: `git rev-parse --short HEAD` → **`c67a6ea64`**.
Two commits landed on this worktree after the iter-1 audit:

- `afde6ebb3` — merge of card t310 (git doctrine alignment)
- `c67a6ea64` — `docs(...): rescue untracked plan-phase artifacts (t298)`

Three consequences, all verified:

1. **The iter-1 baseline is still valid.** `git merge-base --is-ancestor d29b8942e HEAD` → exit 0,
   and the premise-file drift check
   `git diff --stat d29b8942e HEAD -- internal/kanban/integration_lock.go internal/cli/integration.go internal/hook/integration_lock_guard.go internal/session/session_pid.go internal/session/registry.go CLAUDE.local.md .claude/rules/local/gitflow-lane-protocol.md`
   → **empty output**. Every §B premise file and both doc targets are byte-unchanged. No
   implementation file was touched (`internal/kanban/integration_lock.go` absent from the
   `d29b..HEAD` diffstat).
2. **The dispatch's "delta = `git diff`" framing was incomplete.** The SPEC artifacts were
   untracked at iter-1 and entered history only at `c67a6ea64`, so part of the repair sits
   *inside* that commit rather than in the working-tree diff. Concretely: **AC-INL-012 is new
   since iter-1** — `git show d29b8942e:.../acceptance.md` → the path does not exist there, and
   iter-1's own text audits "AC-INL-001 through AC-INL-011" / "all 11 rows" against a 12-row
   matrix. AC-INL-012 was therefore audited here as part of the delta.
3. **Process finding (not a SPEC defect).** Foreign commits landed on a worktree inside its audit
   window, which `agent-common-protocol.md` § Background Agent Execution names a process defect to
   report rather than absorb quietly. Recorded here; no verdict impact, because the drift check
   above shows nothing the audit depends on moved.

## §1 Must-Pass Results (re-verified on the delta)

- [PASS] **MP-1** REQ number consistency — `grep -o 'REQ-INL-[0-9]*' spec.md | sort -u` →
  REQ-INL-001..010, sequential, no gaps, no duplicates, consistent zero-padding. The delta added
  no REQ (AC-INL-012 explicitly traces to existing REQ-INL-004 + REQ-INL-010).
- [PASS] **MP-2** GEARS compliance (requirement layer only) — unchanged set of 10 `REQ-INL-*`
  entries; the delta touched three of them. REQ-INL-003's label corrected to `(Event-driven)`
  (D7); REQ-INL-006 reworded from a branch to a preserved invariant, still a valid
  Where/capability-gate sentence; REQ-INL-009 reworded, still Event-driven ("**When** the fix
  lands, … shall state …"). Judged against `spec.md`'s requirement layer; the Given-When-Then
  entries in `acceptance.md` are the verification layer and are graded under Group 4.
- [PASS] **MP-3** YAML frontmatter validity — the only frontmatter change in the delta is
  `version: "0.1.0"` → `"0.1.1"`, still a quoted semver string. All 12 canonical fields present
  with canonical names (`created`/`updated`/`tags`; no snake_case alias) plus optional `tier: M`,
  re-read at `spec.md:1-16`. Fresh `moai spec lint` over this worktree (run to completion in the
  background after an initial 120s foreground timeout): **`0 error(s), 64 warning(s)`**, and
  `grep -i 'INTEGRATION-LOCK-LIVENESS'` over that output → **rc=1, zero findings naming this
  SPEC**.
- [N/A] **MP-4** language neutrality — single-language (Go) scope; no template-bound content.
- [PASS] **MP-5** D7 cross-SPEC reconciliation — `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' *.md | sort -u`
  → 1 entry, the SPEC's own ID. The delta introduced no external SPEC reference.
- [PASS] **MP-6** D8 cross-platform discipline — `grep -n 'syscall' *.md` → rc=1, no match.
- [PASS] **MP-7** clarification gate — `grep -rn 'NEEDS CLARIFICATION' .` → rc=1, no match.

## §2 Category Scores (0.0-1.0, rubric-anchored)

| Dimension | iter1 | iter2 | Rubric Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.85 | **0.88** | 0.75-1.0 | D7 (non-canonical label) and D9 (REQ-INL-006 branch ambiguity) both closed — REQ-INL-006 now reads "shall continue to read it exactly as it reads it today … no marker-conditional branch is added", which agrees with plan §B consequence 4. Deduction: N1 (REQ-INL-009 still mandates a claim plan.md M5 forbids). |
| Completeness | 0.80 | **0.88** | 0.75-1.0 | D4's hazard is now named in spec.md §G with its fail-direction, its measured basis, and an explicit out-of-scope disposition; plan M5 gained a [HARD] wording bound and §G gained a matching anti-pattern line. AC-INL-012 additionally closes an acquire-refusal coverage gap iter-1 did not find. Deduction: N1 leaves the requirement layer un-swept. |
| Testability | 0.70 | **0.90** | 0.75-1.0 | D1, D2, D5 all closed. AC-INL-001 gained a mechanically-readable marker assertion; AC-INL-011 replaced an unverifiable Then with two runnable greps carrying expected counts and a measured 0-baseline; AC-INL-012 carries a measured RED cell with a stated measured-by-test rationale. Deduction: N2 (a false internal cross-reference inside AC-INL-012's RED cell). |
| Traceability | 0.85 | **0.92** | 0.75-1.0 | D6 closed — every `REQ-INL-001..010` now returns ≥1 from `grep -c` on acceptance.md (measured per-REQ: 1,3,1,2,1,2,1,1,2,2). REQ-INL-002's marker clause now has an AC (AC-INL-001 clause b). AC-INL-012 traces to existing REQs without inventing one. Deduction: N1 is also a REQ↔plan traceability break. |

Aggregate = harmonic mean(0.88, 0.88, 0.90, 0.92) = **0.895** ≥ 0.80. Monotonic vs 0.795.

## §3 Per-defect closure status

| # | Status | Re-verification performed in this run |
|---|---|---|
| **D1** | **CLOSED** | AC-INL-001's **Then** now requires BOTH (a) `stale == false` AND (b) the record written by the real acquire carries `pid_source == "session-owner"`, read back from the on-disk `.moai/state/integration-lock.json` after the child exits. RED cell (b) cites `grep -n 'PIDSource\|pid_source' internal/kanban/integration_lock.go` — **re-run: rc=1, no match** — and the struct's field list, **re-read at integration_lock.go:71-79: exactly `SessionID, SessionName, PID, Branch, Worktree, AcquiredAt, Card`**, as claimed. **Mutant probe re-run:** an implementation that resolves the owner pid, writes it to `PID`, and never adds the `PIDSource` field now FAILS AC-INL-001 clause (b). The clause's `on-disk … or … status --json` disjunction is *not* the vacuous kind — both disjuncts require the field to exist and carry the value; the disjunction only leaves the read-back surface to M2. It additionally fails AC-INL-012. Closed on both counts. |
| **D2** | **CLOSED** | Every cited measurement reproduced exactly: `grep -n '직렬화' CLAUDE.local.md` → **rc=1, no match**; `grep -n 'integration acquire\|integration release' CLAUDE.local.md` → **exactly 2 hits, lines 312 and 317**; `awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md \| grep -c '세션 프로세스'` → **0**; same pipeline with `'재획득'` → **0**; awk range spans 66 lines (non-empty, so the counts are real zeroes, not an empty-range artifact). `CLAUDE.local.md` is tracked (`git ls-files --error-unmatch` → the path) and unchanged since `d29b8942e`. AC-INL-011's Then is now two greps with expected `≥1`; spec.md §B item 13 replaced the false parenthetical with a measured enshrines-vs-silent distinction, and I confirmed the enshrining literal it now quotes: `grep -n '직렬화는 리드 공지로 한다' .claude/rules/local/gitflow-lane-protocol.md` → **line 43**, inside the 42-43 blockquote. Branch-presence claims also hold: `git cat-file -e origin/main:<path>` → **rc=128** (absent), `origin/develop:<path>` → **rc=0** (present). |
| **D3** | **CLOSED** | `grep -rn '"github.com/modu-ai/moai-adk/internal/kanban"' internal/session/` → **rc=1, 0 hits**; unscoped `grep -rn 'internal/kanban' internal/session/` → **exactly 1 hit, `internal/session/session_pid.go:49`, a comment**. plan.md §C now prescribes the import-scoped form and documents the 1-hit comment so the implementer does not read it as a cycle. §C also correctly dropped the fixed-HEAD assertion in favour of a premise-file drift check — which is the right fix, and which I independently ran (empty, §0 above). |
| **D4** | **PARTIAL** | The hazard itself is properly landed and its evidence reproduces exactly: `grep -n 'Flock\|flock\|LockFile' internal/kanban/integration_lock.go internal/cli/integration.go` → **exactly 5 lines** — prose at 15 and 19, `IntegrationLockFileName` at 38/39/106 — **no call site, and `internal/cli/integration.go` contributes no match**, as stated. `AcquireIntegrationLock` re-read at 146-181: `ReadIntegrationLock` at **155**, `writeIntegrationLock` at **177**, nothing between; `tmp := path + ".tmp"` at **229**. spec.md §G names the hazard, its fail-direction (false-"live-for-me", the one hazard that does NOT resolve toward TREAT-AS-LIVE), and its out-of-scope disposition; plan M5 carries a [HARD] wording bound with an explicit forbidden list, and §G gained a matching anti-pattern. **What is not closed: see N1** — REQ-INL-009 still mandates the exact claim M5 forbids. |
| **D5** | **CLOSED** | progress.md §E.1 now names the command and the pattern. Both re-run: the multi-segment form `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → **PASS**; the single-segment schema form `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` → **NOMATCH**. The note states this honestly and correctly identifies the defect as the unattributed claim rather than the ID. |
| **D6** | **APPLIED** | Full IDs in the §D matrix rows for AC-INL-001 and AC-INL-008. Per-REQ `grep -c` on acceptance.md: 001→1, 002→3, 003→1, 004→2, 005→1, 006→2, 007→1, 008→**1** (was 0), 009→2, 010→2. Matrix is now machine-readable for every REQ. |
| **D7** | **APPLIED** | REQ-INL-003 relabeled `(Event-driven)` at spec.md:128. |
| **D8** | **APPLIED** | Both hazards added to §G with fail-directions, and both evidence claims verified: `IntegrationLock` (integration_lock.go:71-79) carries **no host field**; `session.Entry` (registry.go:86-95) **does** carry `Host string \`json:"host"\``. |
| **D9** | **APPLIED** | REQ-INL-006 reworded to a preserved invariant with an explicit "no marker-conditional branch is added" clause, removing the dead-code invitation. |

## §4 Newly-introduced defects

**N1. REQ-OVERCLAIM-SURVIVES-D4-REPAIR** — spec.md REQ-INL-009 (lines 190-194) vs plan.md §F M5
+ spec.md §G — **the D4 repair bounded the plan and the risk section but left the requirement
itself mandating the forbidden claim.** REQ-INL-009 reads: "the two local documents in scope
**shall state the restored serialization guarantee**". plan.md M5's [HARD] wording bound forbids
precisely this: "Forbidden in the shipped prose: `직렬화를 보장한다` and any equivalent unqualified
guarantee claim", because — per spec.md §G, which the same repair added — serialization is *not*
restored: "the acquire path is an unserialized read-modify-write, and this SPEC does not change
that". Before the repair both surfaces carried the same overclaim (consistent, both wrong); the
repair corrected one and left the other, which converts a shared error into a live contradiction
between the requirement layer and the plan. No AC catches it: AC-INL-010 and AC-INL-011 are both
correctly bounded to "session-anchored guarantee", so an implementer who follows REQ-INL-009's
wording into the shipped prose passes every criterion while landing a false claim in a [HARD]
doctrine document.
— Severity: **major** — Class: **blocking** — Required fix (one phrase): reword REQ-INL-009 to
the bounded form the rest of the artifact set already uses — "shall state the restored
**session-anchored liveness** guarantee and the legacy re-acquire consequence" — so the
requirement, the plan's forbidden list, and §G agree.

**N2. FALSE-INTERNAL-CROSS-REFERENCE** — acceptance.md AC-INL-012 RED-now cell — the cell cites
"`runIntegration` (integration_lock_cli_test.go:26-43, **quoted verbatim in the §E.1 repair
note**)". Measured: `grep -n 'runIntegration\|CLAUDE_PROJECT_DIR' progress.md` → **rc=1, no
match**; progress.md is 110 lines and contains no such quote. The *substantive* claim is true and
I verified it independently — `runIntegration` at integration_lock_cli_test.go:26 runs the cobra
command in-process, and `grep -rln "exec.Command"` over both cited test files returns **rc=1, no
output**, so neither existing refusal test (`TestIntegrationAcquire_RefusesASecondLane` at :106,
`TestAcquireIntegrationLock_RefusesASecondLiveSession` at internal/kanban/integration_lock_test.go:60)
can see the cross-process shape. Only the pointer is false.
— Severity: **minor** — Class: **blocking** — Required fix: drop the "quoted verbatim in the §E.1
repair note" clause, or add the quote to progress.md §E.1.

**N3. STALE-ABSOLUTE-IN-HEADER** — acceptance.md line 3 — "Baseline tree for every RED-now cell:
`d29b8942e`" is now false as an absolute: AC-INL-011's and AC-INL-012's RED cells state they were
measured on `c67a6ea64`. The body reconciles this explicitly (both cells cite the empty drift
diff), so nothing is misleading in practice.
— Severity: **minor** — Class: **optional** — Required fix: soften to "Baseline tree … `d29b8942e`;
cells re-measured after the develop merge name `c67a6ea64` and cite the drift check."

## §5 Delta-breakage check (what the repair could have broken)

- **Traceability matrix after the AC change** — INTACT and improved. All 10 REQs map to ≥1 AC
  (per-REQ counts in D6 above); all 12 ACs reference REQs that exist; AC-INL-012 was added without
  inventing a REQ, and its traceability paragraph names which leg of REQ-INL-004 and REQ-INL-010
  it observes and why no other criterion covers it. No off-by-one, no orphan.
- **Restated premises** — every one re-measured in this run; all reproduce exactly (D1-D5 rows
  above). No restatement introduced a new unverified claim.
- **spec.md ↔ plan.md consistency** — one break found (N1). REQ-INL-006 vs plan §B consequence 4
  now agree (D9's fix); §B item 13 vs AC-INL-010/AC-INL-011 agree; plan §C vs the actual grep
  behaviour agrees.
- **AC RED cells now false?** — none. AC-INL-011's new RED cell is an absence-baseline measured at
  0/0 and re-measured at 0/0 here; AC-INL-001(b)'s is rc=1 and re-measured rc=1; AC-INL-012's
  code-path claims re-verified line by line (`case current.Stale():` at 163, `want.PID = os.Getpid()`
  at 175, the `displaced:` emission at internal/cli/integration.go:196-199).
- **Scope creep into implementation** — NONE. `git status --short` shows exactly the four SPEC
  artifacts modified plus the untracked report directory; the `d29b..HEAD` diffstat carries no
  `internal/kanban/integration_lock.go` entry.

## §6 Recommendation

**PASS-WITH-DEBT at 0.895** (Tier M threshold 0.80; +0.100 monotonic; iteration ceiling reached).
All seven must-pass criteria hold, four of five blocking defects are fully closed, and the fifth
(D4) is closed everywhere except the requirement sentence. The repair also closed a coverage gap
iter-1 missed (AC-INL-012, the acquire-refusal path).

The declared debt, to close before M5 doc work lands — both are one-line edits and neither needs
another audit round:

1. **N1 (major, blocking)** — reword REQ-INL-009 to "restored **session-anchored liveness**
   guarantee". Until this lands, plan.md M5's [HARD] wording bound is the governing instruction
   for the doc prose, and it is correct as written.
2. **N2 (minor, blocking)** — drop or satisfy AC-INL-012's "quoted verbatim in the §E.1 repair
   note" pointer.
3. **N3 (minor, optional)** — orchestrator's discretion.

Run-phase may proceed. The Implementation Kickoff Approval human gate is unaffected by this
verdict and remains required.
