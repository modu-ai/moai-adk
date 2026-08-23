# SPEC Review Report: SPEC-WEB-CONSOLE-015

> **This file contains BOTH audit iterations.** Everything above `# Iteration 2` is the iteration-1
> record — verdict FAIL 0.75, MP-7 FAIL with six clarification markers at plan.md:170-194 (line
> numbers as of `f11dd5cb3`, where plan.md was 211 lines).
> **The current verdict is the `# Iteration 2` section below** (an H1, not `##`): PASS 0.94, MP-7
> PASS, zero markers on `cf131b20a`. Reading the iteration-1 must-pass block as the current verdict
> has already misled two readers — check which section a finding belongs to before citing it.

Iteration: 2/3 (Tier L ceiling: 3) — **current verdict below; iteration 1 record preserved**
Auditor: plan-auditor (independent)
Tree audited: `.claude/worktrees/t207` (branch `WT-web-live-todo`) — iter1 @ `f11dd5cb3`, iter2 @ `cf131b20a`
Verdict (iteration 2): **PASS**
Overall Score (iteration 2): **0.94** (Tier L PASS threshold: 0.85; iteration 1: FAIL 0.75)

> Iteration 2 verdict, scores, and evidence: see `## Iteration 2` at the end of this file.
> Implementation Kickoff Approval remains mandatory and score-independent — this PASS never auto-bypasses the plan→run human gate.

Reasoning context ignored per M1 Context Isolation. Every code claim below was verified against the t207 worktree with absolute paths (an earlier relative-path batch had drifted to the primary checkout and was discarded and re-run).

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — 25 REQ definitions (`REQ-WC15-001…051`), zero duplicate definition ids, consistent 3-digit padding, monotone within topic blocks. Numbering is deliberately sparse (x00/x10/x20/x30/x40/x50 per axis), which is the established SPEC-WEB-CONSOLE family convention — verified against sibling SPECs: SPEC-WEB-CONSOLE-014 uses `REQ-WC14-001,002,003,010,011,012,020,021,030,031…` and SPEC-WEB-CONSOLE-013 the same shape. Ruled convention, not gap.
- [PASS] MP-2 EARS/GEARS compliance (requirement layer only) — all 25 REQs match a GEARS pattern: ubiquitous (010, 020, 021, 030, 031, 040, 042, 044, 050), event-driven When (011, 012, 023, 033, 034, 046, 051), state-driven While (045), capability-gate Where (024), unwanted shall-not (001, 002, 022, 032, compound 023/043). REQ-025 is ubiquitous-form with an embedded verification recipe (see D8/D11). ACs are correctly Given-When-Then in the verification layer and were not penalized here.
- [PASS] MP-3 YAML frontmatter — all 12 canonical fields present with correct types (spec.md:1-17): `id`, quoted `title`, quoted semver `version: "0.1.0"`, `status: draft` (enum), ISO `created/updated: 2026-08-24`, `author`, `priority: P2`, `phase: "v3.2.0 target"` (release target, not a prohibited stage name), `module: internal/web`, `lifecycle: spec-anchored`, comma-separated `tags`. No snake_case aliases. Optional `era: V3R6`, `tier: L`, `related_specs` well-formed.
- [N/A] MP-4 language neutrality — SPEC targets the moai-adk-go Go codebase itself (internal/...); single-language scoped, auto-pass.
- [PASS] MP-5 D7 cross-SPEC reconciliation — verification verb executed; all four referenced SPECs exist in the worktree's `.moai/specs/` with non-terminal status: SPEC-KANBAN-TODO-CLI-001 `in-progress`, SPEC-WEB-CONSOLE-REDESIGN-001 `completed`, SPEC-HANDOFF-THRESHOLD-001 `completed`, SPEC-FACTORY-WORKER-FANOUT-001 `implemented`. No retired/superseded/archived reference; no BLOCKING finding.
- [PASS] MP-6 D8 cross-platform discipline — `grep -c syscall` over spec.md/plan.md/acceptance.md/research.md = 0 0 0 0; auto-pass.
- [FAIL] MP-7 clarification gate — `grep -rn '\[NEEDS CLARIFICATION'` finds **6 markers in plan.md §G** (plan.md:170, 173, 179, 184, 188, 194). Score-independent must-pass failure; the orchestrator MUST resolve each via AskUserQuestion before Implementation Kickoff Approval. (The SPEC polices this itself — acceptance.md §G DoD:147-148 — but audit-time state is what binds.)

## Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | minor ambiguity, few items | REQ-021/AC-021 "exactly one exported reader" unpinned; REQ-025 inline verification recipe; AC↔open-decision ambiguity (D3/D4) |
| Completeness | 0.75 | one non-critical artifact missing | design.md absent from Tier L 5-artifact set (plan.md §H:196-202); §C.3 enumeration gaps (D6) |
| Testability | 0.75 | few ACs need minor interpretation | AC-022 false baseline (D2); AC-010b byte-identity input unconstrained; AC-021 half-mechanical; AC set otherwise exceptionally strong (stated baselines, negative assertions, the §A.3 race has a direct regression AC) |
| Traceability | 0.75 | one REQ uncovered | REQ-WC15-025 (spec.md:160-167) maps to no AC; all 25 ACs map to existing REQs (AC-010b→010+051, AC-040→040+041); no orphan ACs |

Aggregate 0.75 < 0.85 Tier L threshold → FAIL on score as well as MP-7.

## Cross-backend convergence (audit_multi, project_root=t207)

`codex` (advisory) reviewed the primary checkout's HEAD commit `a1b1ca696` — an uncommitted-changes code review, not the SPEC documents; **off-target, not counted as evidence**. `glm` (advisory) returned findings about `migrate.py` / `processor.py` — files that do not exist in this SPEC or repo; **off-target hallucination, not counted**. Convergence overall verdict FAIL rests on the claude anchor; both advisories are fail-open no-ops for this audit. Residual risk: no genuine second opinion on the SPEC text was obtained.

## Defects Found

### MUST-FIX (blocking — fix before verdict revisit)

**D1. [critical] plan.md:170-194 — six unresolved `[NEEDS CLARIFICATION]` markers (G-1..G-6)** — MP-7 must-pass failure. Class: blocking. Fix: resolve all six via the orchestrator's AskUserQuestion round(s) (≤4 questions per call → two rounds for six decisions), record the answers, delete the markers. Blocking subset and recommendations in the G-table below.

**D2. [critical] spec.md:160-167 / acceptance.md:48-52 — REQ-WC15-025 (internal/cli/tokens.go duplicate-reader migration) has NO acceptance criterion, and AC-WC15-022's stated baseline is factually false** — Verified in t207: `internal/cli/tokens.go:86` declares `RawPct float64 \`json:"raw_pct"\``, so the **pre-change tree returns 1 hit outside internal/statusline, not zero** as AC-022's baseline sentence claims. Consequences: (a) the one requirement the SPEC itself flags as the silent-break hazard ("no compile error and no runtime error — the snapshot simply stops appearing", verified: `readTokensContextSnapshot` at tokens.go:393-397 returns nil on any read error) is the one requirement with no observing AC; (b) AC-022's post-zero assertion does accidentally observe the removal, but its false baseline mislabels it a preservation test; (c) the positive half — "`moai tokens` still emits a context snapshot for a session that has one" — is tested nowhere. Class: blocking. Fix: add **AC-WC15-025** with both halves (pre-change: `grep -rn '"raw_pct"' internal/cli` returns exactly the tokens.go declaration; post-change: zero, and a `moai tokens` run against a state dir holding a per-session snapshot still embeds the context block), and correct AC-022's baseline sentence to state the true pre-change count (1).

**D3. [major] acceptance.md:40-42, 58-61 vs plan.md:175-179 — AC-WC15-020 and AC-WC15-024 hard-code the hard-cut branch of open decision G-3** — AC-020 asserts `P/.moai/state/context-usage.json` does not exist after a render; AC-024 asserts neither doctrine file contains the bare old path. Under G-3's dual-write branch both ACs FAIL by construction. An open decision cannot coexist with normative ACs that presuppose one answer. Class: blocking. Fix: resolve G-3 (recommendation: hard cut — see G-table) or make both ACs conditional on the resolved branch.

**D4. [major] acceptance.md:65-67 vs plan.md:186-188 — AC-WC15-030 hard-codes the show-dropped-cards branch of open decision G-5** — The fixture includes one `dropped` item and asserts "the fragment contains each item's id and text, its state" — i.e. all three states render. If G-5 resolves to queued-only, AC-030 fails. Class: blocking. Fix: resolve G-5 and align AC-030's fixture/assertion with the answer.

**D5. [major] spec.md:182-185 (REQ-WC15-031) vs spec.md:140-142 (REQ-WC15-002) / 186-187 (REQ-WC15-032) — the "same resolution" the console must reuse embeds a WRITE on the fallback branch** — Verified in t207: `resolveTodoQueueRoot` (todo.go:66-72) → on git-unresolvable contexts → `fallbackTodoQueueRoot` (todo.go:89-102) → `adoptLocalTodoQueue` (todo.go:115-139), which performs `os.MkdirAll` (:124), `os.Rename` (:128), or `os.WriteFile` (:139) — migrating the backlog queue file. A console launched in a non-git directory that calls the relocated shared resolution therefore mutates the backlog, violating REQ-002 ("no write, mutation, or state transition against … the backlog queue") and failing AC-032 (backlog byte-identical, lock untouched). The normal (git-resolvable) path is pure; the conflict is reachable exactly on the fail-open branch the console cannot exclude. Class: blocking. Fix: M4's relocation must **split resolution from adoption** — export a read-only resolver (no adoption side effect) for `internal/web`, keep adoption on the `moai todo` path — and say so in REQ-031/plan §D M4.

**D6. [major] spec.md:241-252 (§C.3) / 171-176 (REQ-WC15-024) — context-usage blast-radius enumeration is incomplete** — Verified in t207: the literal path `state/context-usage.json` is carried by exactly four doctrine files — `context-window-management.md` + **`context-window-management-detail.md`** (local) and **both template mirrors**. §C.3 lists only the main rule and its mirror; the `-detail.md` pair (which carries the snapshot-field list and validity-guard read procedure the main rule points at) is unlisted and outside REQ-024's scope, so it would survive M3 stale. Also unlisted: `internal/cli/tokens_test.go` (tests the duplicate reader being migrated) and `internal/spec/drift_cache.go:24` (comment naming the file — NOTE-level). And §C.4's "Template-First applies to exactly one artifact" (spec.md:254-261) is wrong if M3 edits `.moai/README.md`, whose template mirror also carries the path. Class: blocking (the SPEC's own anti-silent-break discipline — enumerate every consumer — was applied incompletely). Fix: add the `-detail.md` pair + tokens_test.go to §C.3 and REQ-024's scope; correct §C.4's count.

**D7. [major] plan.md:196-202 (§H) — design.md absent from the Tier L 5-artifact set** — Tier L contract = spec/plan/acceptance/design/research; `ls` confirms 5 files present, design.md not among them. The stated rationale ("its content … is carried by spec.md §A/§C and this plan's §D/§G, and a separate design document would restate them") is **acceptable for the schema/join topology content** — I verified §A/§C's code claims and they are accurate — but **not acceptable for the two cross-cutting open decisions (G-1 card-id producer, G-3 compat window)**, which have no design home; the evidence they have no home is precisely the AC self-contradictions D3/D4. Class: blocking before run entry. Fix: write a thin design.md scoped to (i) the resolved G-1/G-3/G-6 decisions and their consequences, (ii) the M3 migration sequence including the tokens.go migration and the resolution/adoption split — or, if the operator prefers no extra artifact, capture the resolved decisions inline in plan.md §G and state the deviation from the Tier L artifact set explicitly in §H for the re-audit to judge.

### SHOULD-FIX (optional — surfaced, orchestrator's discretion)

**D8. acceptance.md:27-29 (AC-WC15-010b)** — "byte-identical" re-marshal needs the input constrained to records produced by the pre-change writer (struct-order Marshal); a hand-edited or differently-indented fixture fails a correct implementation spuriously.

**D9. acceptance.md:44-46 (AC-WC15-021)** — the "exactly one exported reader" half has no mechanical pin (the `grep -c ≥ 1` half only proves use, not exclusivity). Suggest: zero hits for a second exported `Read.*ContextUsage`-shaped identifier in internal/statusline, or an exports test.

**D10. acceptance.md:132-134 (§F)** — the session-id path-separator/`..` rejection for the per-session path is a filesystem write boundary with externally-shaped input; promote from edge-case prose to an AC.

**D11. spec.md:160-167 (REQ-WC15-025)** — the inline "Verification: …" recipe is AC material living in the requirement layer; move it into acceptance.md when adding AC-WC15-025 (D2).

**D12. plan.md:100-111 (M4)** — M4 should state the resolution/adoption split explicitly (extends D5 into the plan's milestone contract, so manager-develop cannot implement the literal move without the read-only seam).

### NOTE

**D13.** spec.md §B.1 lists REQ-WC15-025 out of numeric order (between 022 and 023). Cosmetic.

**D14.** `internal/spec/drift_cache.go:24` comment cites `context-usage.json`; goes stale post-split — fold into M3's comment sweep.

**D15.** Prior-doc citations verified accurate: `.moai/reports/webredesign/moai-web-menu-spec.md` §6.1 does recommend separating producer work ("웹 SPEC과 분리해서 진행하기를 권합니다"), §2.2 is the read-only rule, §4.6 the observed race. plan.md §C's characterization is honest.

**D16.** REQ count 25 / AC count 25 — both exactly at the Tier L ceilings (not over). No violation; no headroom for additions beyond AC-WC15-025 (an AC addition is fine; do not add REQs).

## G-1..G-6 Recommendation Table

| # | Decision | Blocking before run entry? | Recommendation + reasoning |
|---|----------|---------------------------|----------------------------|
| G-1 | Card-id producer (env var / lead-written / worktree-derived) | **YES** — plan itself: "M2 cannot close until that decision lands"; AC-042 depends on it | **(c) worktree-derived** (`git rev-parse --show-toplevel` basename). The dispatch format fixes `wt: .claude/worktrees/<card-id>` and kanban-dispatch.md mandates "the worktree directory keeps the card id" — the data already exists at the only place every card-carrying lane stands. Lanes outside a card worktree (already a dispatch violation) yield the honest "not recorded" marker REQ-012-style honesty already covers. **Reject (a)** launch-time env: factory routes a card whole to a free lane *after* launch, so process env cannot carry a per-card value without relaunching lanes per card — new producer surface AND a new way to be absent. **Reject (b)** lead-written: record goes silently stale when the lead forgets. If wanted later, (a) can be layered as an override with no schema change. |
| G-2 | Promote M3 to a sibling SPEC? | **YES** (decide now — "cheap before run, expensive after") | **Keep in-SPEC iff G-3 resolves to hard cut.** The plan's §C reason 1 is strengthened by D2: even inside this SPEC the migration's positive half lacks an AC — a standalone path-move SPEC would have nothing but that weak shape. Every external consumer is enumerated (after D6's additions) and testable in-tree. **Promote to sibling iff G-3 resolves to dual-write** — a release-spanning compat window is genuinely separate lifecycle work with its own close, and §C's own trigger ("re-deriving the Detection Heuristics read procedure") is already half-met by REQ-024. |
| G-3 | Dual-write window vs hard cut for context-usage path | **YES** — AC-020/AC-024 are incoherent until resolved (D3) | **Hard cut.** (i) Every reader is in-tree and enumerated after D6 — no plugin or external API reads this path; (ii) the file is render-ephemeral session telemetry regenerated on every statusline render — there is no durable data to migrate, a cut loses at most one render's snapshot; (iii) dual-write keeps the last-writer-wins slot alive as exactly the trap §A.3 documents (the observed `368a2bd9…`/`e463a3c9…` race); (iv) dual-write contradicts REQ-024's own instruction to drop the single-slot validation steps (`isFreshForSession`, `writer_pid` discrimination) and would keep that code alive. |
| G-4 | /todo route vs panel on /kanban | No — deferrable to M6 boundary (M1-M5 unaffected) | **Panel on /kanban.** AC-034 keys the section on the existing `data-live="kanban"` refresh area; a separate route either reuses that key awkwardly or requires an event-vocabulary change that §D Out of Scope explicitly forbids. Nav already carries five entries; the queue belongs beside the board it feeds. Batch the question into the same clarification round anyway — it costs nothing next to the blocking four. |
| G-5 | Show dropped cards? | No — deferrable to M6, **but AC-030 must be aligned with the answer** (D4) | **Show all three states with a state badge** (audit view) — matches AC-030 as written, and the kanban lead's card cross-check consumes dropped/picked history as an audit trail. Filtering to `queued` is equally legitimate as a product choice; the defect is only the current AC↔decision mismatch. Whichever answer lands, fix AC-030's fixture or assertion to match. |
| G-6 | `Lane int` vs `*int` | **YES** for M1 specifically (the field type is M1's deliverable) — trivial to decide | **`int` with 0 = not-a-lane.** Factory lanes are numbered from 1 (`lane-1..N`), so 0 is unreachable by legitimate data; the `VerifyRung` pointer precedent does not transfer — there, a *recorded-empty* rung (`*Rung`→`""`) is a reachable state (record.go:85-96), while lane-0 is not. A pointer adds a nil-check at every reader to defend a state that cannot occur, and REQ-051 already renders absent keys as "not recorded" (absent → 0 → honest marker). |

## Audit-Focus Answers (caller's seven questions)

1. **AC quality** — the set is well above the repo median (baselines stated, negative assertions, direct regression test for the observed race). The field-exists-but-nothing-observes failure shape materializes exactly once: REQ-025 (D2) — and AC-022's false baseline actively masks it. Secondary weak spots: D8/D9/D10.
2. **G-1..G-6** — table above. Blocking: G-1, G-2, G-3, G-6. Deferrable: G-4, G-5 (with AC-030 alignment).
3. **M3 / G-3 scoping** — the blast-radius story is the right shape but incomplete (D6: `-detail.md` pair, tokens_test.go). The tokens.go second reader IS covered in the milestone (M3 + §C.3 + REQ-025 "in the same milestone that moves the path") and NOT covered in the ACs (D2 — no AC-025; AC-022's accidental coverage rests on a false baseline and misses the positive half).
4. **M4 queue-root** — the single-queue invariant is genuinely testable: AC-031a (worktree cwd → primary's file, N not 0) is precisely the t106-class regression test, AC-031b keeps `todo.go` structurally delegating. The hazard is the other invariant: the resolver's fallback branch writes (D5).
5. **M6 before M5** — no hidden dependency breaks the ordering; the graph is honest. The one unstated constraint: AC-032 (M6, read-only console) transitively depends on M4 performing the resolution/adoption split (D5) — if M4 does the literal move, M6's AC-032 fails on the fallback branch. M5 and M6 touch different view sections (RoleVM/factory vs kanban panel) — merge adjacency only.
6. **design.md absence** — rationale acceptable for the verified schema/join content, unacceptable for the open cross-cutting decisions (D7). Thin scoped design.md (or inline resolved decisions with an explicit Tier L deviation note) MUST precede run entry.
7. **Scope discipline** — exactly one REQ without an AC (025); zero ACs without a REQ; counts at the Tier L ceilings (25/25); §F edge cases carry no ACs (one promoted, D10).

## Recommendation

1. Orchestrator: run the AskUserQuestion rounds to resolve G-1..G-6 (two rounds of ≤4; carry the G-table recommendations), record answers, remove all six markers (D1).
2. manager-spec: add AC-WC15-025 + fix AC-022's baseline (D2); align AC-020/024 with the resolved G-3 and AC-030 with the resolved G-5 (D3/D4); write the resolution/adoption split into REQ-031 and M4 (D5/D12); extend §C.3/REQ-024/§C.4 (D6); close the design.md gap per D7; optionally D8-D11.
3. Re-audit scope (iteration 2): the enumerated defect delta D1-D7 plus a regression check — not a from-scratch re-audit.

---

# Iteration 2

Tree: t207 @ `cf131b20a` (iter2 SPEC revision `449736deb` + G-4 ruling doc `cf131b20a`). Scope per the retry-loop contract: the D1-D7 defect delta (plus D8-D14 resolution status), a regression check over iter1's passing must-pass criteria, and verification of the two lead-approved deviations. All load-bearing claims re-verified against the t207 tree with absolute-path commands.

## Verdict: PASS — Score 0.94 (threshold 0.85)

## Must-Pass Results (iteration 2)

- [PASS] MP-1 — 25 unique REQ ids, zero duplicate definitions (grep `^- **REQ-WC15-` | sort | uniq -d = 0). REQ-025 renumbered into numeric order (D13 resolved).
- [PASS] MP-2 — all 25 REQs (including the six rewritten ones: 024, 025, 030, 031, 034, 040, 042) still match a GEARS pattern; requirement layer only.
- [PASS] MP-3 — frontmatter unchanged, all 12 canonical fields valid. (See N2: version/HISTORY hygiene note.)
- [N/A] MP-4 — single-language Go codebase SPEC.
- [PASS] MP-5 — 4 related SPECs re-checked, statuses unchanged (in-progress / completed / completed / implemented), none retired/superseded/archived; no BLOCKING.
- [PASS] MP-6 — `grep -c syscall` = 0 across all five artifacts.
- [PASS] MP-7 — `grep -rc 'NEEDS CLARIFICATION'` across the SPEC directory: zero hits. All six markers deleted, not annotated (D1 resolved).

## Category Scores (iteration 2)

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | 1.0 | All six decisions resolved with consequences and rejected alternatives recorded (plan §G, design §1); AC↔open-decision contradictions eliminated; every previously ambiguous AC now mechanically pinned |
| Completeness | 1.0 | Tier L 5-artifact set complete (design.md written, scoped exactly to the iter1 gap); §C.3 blast radius verified complete; §C.6 enumerates the route-decision surface |
| Testability | 0.75 | Every baseline re-measured and TRUE (AC-022 pinned grep returns exactly the 4 listed hits; AC-021 exported-reader baseline = 0 verified). Sole deduction: AC-043's second half ("no new file created under .moai/state/") is measurable with minor interpretation — a dir-listing diff, not a grep as phrased |
| Traceability | 1.0 | All 25 REQs covered (incl. REQ-025 → AC-025); all 29 ACs map to existing REQs; no orphans either direction |

Aggregate 0.94 ≥ 0.85 → PASS. Iter2 score (0.94) > iter1 (0.75) — no STOP signal, no score regression.

## Defect Delta Resolution (iter1 → iter2)

| Iter1 defect | Status | Evidence |
|---|---|---|
| D1 six clarification markers | RESOLVED | plan.md §G rewritten as "Resolved decisions" (G-1..G-6 with reasoning + rejected alternatives); grep = 0 markers |
| D2 REQ-025 uncovered + false AC-022 baseline | RESOLVED | AC-WC15-025 added with both halves (removal grep w/ stated baseline tokens.go:30/:79; positive half asserting the context block survives in `moai tokens` output); AC-022 baseline corrected by pinning the exact command `grep -rn '"raw_pct"' internal/` — **verified verbatim: exactly 4 hits in the 4 listed files** (context_usage.go:63, context_usage_test.go:150, tokens.go:86, tokens_test.go:283). The author's "2 non-test / 4 with tests" = files outside vs inside statusline; my iter1 "1" = non-test hits outside statusline; the original "0" was false. All three counts were scope artifacts — pinning the command in the AC sentence is the correct resolution |
| D3 AC-020/024 vs G-3 | RESOLVED | G-3 resolved: hard cut, four reasons (plan §G, design §1.2); both ACs now annotate themselves "correct by decision, not presupposition" |
| D4 AC-030 vs G-5 | RESOLVED | G-5 resolved: audit view, all three states + badge; badge now asserted in AC-030 |
| D5 adoption write in shared resolver | RESOLVED | REQ-031 requires the split (console entry point mutates nothing on ANY branch; adoption reachable only from `moai todo`); plan M4 mandates two exported entry points and warns a literal move fails AC-031c/032; AC-031c constructs the exact fallback preconditions (no git metadata + local queue present + no fallback queue) in both directions — console leaves disk untouched, `moai todo` still adopts |
| D6 §C.3 blast-radius misses | RESOLVED | Both `-detail.md` files + `tokens_test.go:283` + `drift_cache.go:24` added; `.moai/README.md` resolved as no-change (verified: line 31 carries only the generic `state/` category row, no filename); §C.4 corrected to "two mirror pairs, four files"; AC-024 now four-file scope |
| D7 design.md absent | RESOLVED | Written, deliberately thin, scoped to exactly the named gap: resolved G-1/G-3/G-4/G-6 with consequences and rejected alternatives, M3 migration sequence (steps 1-3 must land together — the silent-break ordering constraint), M4 split design. Does not restate §A/§C (correctly) |
| D8 AC-010b byte-identity | RESOLVED | Input constrained to writer-produced records with rationale |
| D9 AC-021 "exactly one" unpinned | RESOLVED | Mechanical pin: `^func Read.*ContextUsage` count in internal/statusline exactly 1; baseline 0 verified (grep returned nothing on the pre-change tree) |
| D10 path-traversal edge case | RESOLVED | Promoted to AC-052 (separator / `..` / absolute / empty; refused not redirected; render still completes) with the write-boundary rationale |
| D11 REQ-025 inline verification | RESOLVED | Moved to AC-WC15-025; REQ cross-references it |
| D12 M4 split absent from plan | RESOLVED | M4 rewritten ("not a literal move") with the two-entry-point contract |
| D13 REQ-025 out of order | RESOLVED | §B.1 now numeric 020→025 |
| D14 drift_cache comment | RESOLVED | §C.3 row + M3 step 7 sweep |

## Lead-Approved Deviation 1 — G-4 route with vocabulary exclusion held

**Measurement verified TRUE in the tree.** (i) `watchMap["kanban"] = {".moai/state/kanban"}` watches the whole directory (events.go:30; the ruling doc's `:29` is off by one — cosmetic), and the backlog file `.moai/state/kanban/backlog.json` is inside it, so backlog writes already fire the existing event. (ii) `refresh(area)` gates on `hasArea` = `querySelector('[data-live="' + area + '"]')` and re-fetches `window.location.href` (app.js:643-653) — the marker binds a DOM region, not a route, so a `/todo` page carrying `data-live="kanban"` is covered with zero producer change. (iii) `eventFor` (events.go:171-183) uses a strict `if len(p) > bestLen` — verified verbatim — so two equal-length watches tie and the winner falls to Go's randomized map iteration order; the file's own header (events.go:6-9) records exactly that nondeterminism as a previously fixed bug. (iv) `EVENTS` and `watchMap` both carry six entries (app.js:637; events.go:27-32).

**Does AC-034 mechanically pin the no-vocabulary-change?** Yes — the unchanged-diff assertion on `watchMap` and `EVENTS` (six entries, same six names) is a mechanical pin, and AC-035 pins the route surface that DID enter scope (verified baselines: rail() has exactly 5 navRows at shell.templ:130-134, no `todo` case in iconAt, screens.go:23 `Area: area`).

**Auditor note.** The tie-hazard argument strictly applies to same-length watches; a longer file-level watch of `backlog.json` would win longest-match deterministically — but file watches are fragile under atomic-rename writes, so "a new event would have to watch the same directory" is fair engineering judgment, and the no-change solution is strictly simpler. The exclusion stands. The provenance trail (g4-scope-ruling.md + design §1.3, recording that the dispatch inference was withdrawn by measurement and lead-approved, not taken unilaterally) is exactly how a deviation from an audit recommendation should be documented.

## Lead-Approved Deviation 2 — worktree live-event gap recorded, watch-widening excluded

**Auditor opinion: ACCEPT.** (i) The limitation is a staleness window, not incorrect data — correct on load (pinned by AC-031a's worktree resolution test) and on any other `kanban` event; no AC falsely claims live liveness for the worktree case. (ii) The fix is a genuine producer change with its own design questions (which root(s) to watch when served≠primary, double events when equal, fsnotify watch count). (iii) The REQ ceiling is already at 25; the ruling explicitly declines to break it and commits to a separate card if the operator wants the fix. Recorded in three places (spec §C.6, design §1.3, ruling §Residual). Handling is sound.

## New Findings (iteration 2)

**N1. [SHOULD-FIX — optional] acceptance.md — AC count exceeds the Tier L ceiling.** 29 AC definitions (26 numbered ids) vs the Tier L AC ceiling of 25 (spec-workflow.md REQ/AC budget table). The four over-ceiling criteria (AC-025, AC-031c, AC-035, AC-052) were each added under plan-audit direction (D2, D5, G-4 scope) or by promoting an audit-named edge case (D10) — removing any would reopen a defect or leave the route scope unobserved. The ceiling rule exists to stop author-driven over-formalization, not to force deletion of auditor-mandated observations. Required fix: record the deviation explicitly (one line in plan.md §B — "AC count 29 > ceiling 25; four criteria added under plan-audit direction; recorded as a budget deviation"), so the overage is governed rather than silently absorbed.

**N2. [SHOULD-FIX — optional] spec.md frontmatter/HISTORY — revision not versioned.** `version: "0.1.0"` and the HISTORY table are unchanged despite a substantive revision of all four authored artifacts. Bump to `0.2.0` + one HISTORY row.

**N3. [NOTE] g4-scope-ruling.md — `events.go:29` should read `:30`** for the `kanban` watchMap entry (spec.md §C.6 has it right). Cosmetic.

**N4. [NOTE] REQ-WC15-042 — the environment-variable override name is unspecified.** Run-phase detail; repo hardcoding rules will require the constant in `internal/config/envkeys.go`. No plan-phase action needed — flagged so the implementer does not inline a string.

## Recommendation

PASS. Proceed to Implementation Kickoff Approval (mandatory, score-independent). N1/N2 are optional polish that can ride the run-phase kickoff delegation or be fixed in one line each; neither blocks. Per the skip-eligibility contract, this PASS + 0.94 ≥ 0.85 leaves the artifact-hash as the only future gate input — any further plan-artifact edit re-executes the audit.
