---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Lanes start a card inside its worktree — mid-session movement prohibition, skill-list duplication measurement, and the card-transition /clear amendment"
version: "0.3.0"
status: draft
created: 2026-09-26
updated: 2026-09-29
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, session-start, context-budget, kanban, clear, skill-list"
tier: M
related_specs: [SPEC-SESSION-WORKTREE-001]
---

# SPEC-SESSION-DOUBLELOAD-001 — Card sessions start inside their worktree (D scope)

## HISTORY

- **2026-09-26** — v0.1.0 plan-phase draft authored for card t1219 (premises from `.moai/reports/t1219/verdict.md`).
- **2026-09-27** — v0.2.0 revision after plan-audit iter-1 (FAIL 0.64). 16 REQ / 16 AC scoped to a measurement + two-tier doctrine change.
- **2026-09-29** — v0.3.0 **D-scope re-authoring** (card t1279). The prior plan round was closed draft+HOLD (lead decision b1, verdict §⑩) after 3 plan-audits + 1 delta audit all FAILed at the Tier cap, because the core premise was refuted by measurement: sessions started inside a nested L1 worktree still load the primary `CLAUDE.local.md` (verdict §① Claim 2), so "launcher start yields a single instruction set" is false and the v0.2.0 two-tier design does not survive. The lead ordered re-authoring **under the same SPEC ID** with a **fresh audit budget** (the prior rounds count as history, not as this SPEC's audit count; source: lead dispatch to factory lane worker-69, 2026-09-29). Scope narrowed to verdict §④ "D 범위" exactly:
  1. mid-session worktree movement prohibition (documents in local AND template copies);
  2. skill-list duplication token measurement (verdict §① Gaps — unmeasured);
  3. the N1 amendment draft (verdict §⑤) — plan phase drafts the text only; the run-phase substitution is gated on re-measuring the t1175 landing at the develop tip.
  - **Reuse-in-place rationale** — the ID is kept rather than issuing a new SPEC ID because (a) the card linkage (t1279 → this SPEC directory) is already recorded in dispatch and queue state, and (b) audit-trail continuity: the three FAIL verdicts and the delta audit under `.moai/reports/t1279/` belong to the ID that produced them; a new ID would orphan that record and present the refuted v0.2.0 design as a separate SPEC's history.
  - **What was kept from v0.2.0 and why** — the evidence-before-doctrine ordering, the probe-cap discipline, the positive-control rule, and template-neutrality all survived the premise refutation unchanged (they are measurement/process requirements, not premise-dependent); they are re-numbered below. **What was dropped** — the two-tier doctrine design (REQ-SDL-008/009/010 of v0.2.0, whose launcher-start premise is refuted by Claim 2), the `AGENTS.md` root-contract surface, and the 7-probe path matrix (A/B/C/D/D′ paths measured the refuted question).

## §A — Problem (measured premises)

All premises below are measured and recorded in `.moai/reports/t1279/verdict.md` (Claude Code 2.1.283, 2026-09-27; every load judgment made from session transcripts, not debug logs). t1243 M1a (`9f32f8077`) independently reproduced Claims 1 and 3 (verdict §⑧).

1. **Instruction-file double load.** Sessions search upward from the start directory and load `CLAUDE.md` / `CLAUDE.local.md` from ancestor directories, without stopping at a git repository boundary, and without content-based deduplication (verdict §① Claim 1, probes 1–2). A session started inside a nested L1 card worktree therefore also loads the primary checkout's `CLAUDE.local.md` (verdict §① Claim 2, probe 3; independently reproduced by t1243 M1a). **Observed, mechanism unexplained:** in that same configuration the primary `CLAUDE.md` is skipped while `CLAUDE.local.md` is not; the exact condition of that skip was not observed and this SPEC does not assert a mechanism.
2. **Cost.** One extra `CLAUDE.local.md` copy costs **+20,855 first-turn input tokens** (input + cache_creation + cache_read; verdict §① Claim 4, probes 3 vs 5).
3. **Skill-list duplication is a mid-session-move problem.** A session started inside the worktree via the launcher carried exactly one project skill source (t1219 Evidence 1 — debug-log based, transcription re-confirmation still owed, verdict §① Gaps). A session started in the primary checkout and moved with `EnterWorktree` gained a second skill list of about 40 worktree-scoped skills while the primary list stayed in place (t1219 Evidence 4). Skill duplication and the `CLAUDE.local.md` double load have different causes (verdict §① (2)).
4. **The skill-list token cost is unmeasured.** The 2,406 B frontmatter-first-line sum (45 skills) underestimates the real list size because skill descriptions are multi-line blocks, and is **ruled unusable as evidence** (verdict §① Gaps).
5. **The D remedy.** Of the compared solutions (verdict §②), D — prohibit mid-session movement; start the lane inside the card worktree via the launcher; when a move is unavoidable, `/clear` after the move — is the one that addresses the skill-list duplication. It does **not** address the `CLAUDE.local.md` double load, which persists at launcher start (premise 1) and is owned by the t1243/t1259 AGENTS.local.md work (out of scope). Whether `/clear` restores the single launcher-start skill set is itself unmeasured (verdict §② D row).
6. **The A alternative is shelved.** Migrating card worktrees to L2 paths (`~/.moai/worktrees/`) is blocked by three verified tool-path blockers (no authorized L2 creation path; WorktreeCreate hook is L1-fixed; EnterWorktree transition restricted, verdict §③). A is out of scope here; its revival conditions are recorded in §E so a future card can pick it up.
7. **The N1 decision.** The card-transition `/clear` clause of the dispatch rule is amended (plan phase drafts; run phase substitutes) to read: card-transition `/clear` exactly once, **after** the move; non-card-transition phase ends unchanged; the obligation is not reduced (verdict §⑤). The t1175 rules-diet card has landed (`7fe658815`) and is absorbed into this branch, so the re-measurement gate is satisfiable at run time; the anchor is located by content because line numbers drift (the audited `:153` target now sits in the `/clear` handoff clause; observed in this tree at the `[HARD]` paragraph "A companion session does not carry one card's context into the next card").

## §B — Goal

A lane session carries the skill set of the card worktree it started in, and the doctrine says so: start a card inside its worktree through the launcher, do not move the session between worktrees mid-session, and when a move is unavoidable, `/clear` once, after the move. The doctrine change is preceded by the measurement that closes the skill-list token gap, and the card-transition `/clear` wording is amended without reducing any obligation.

## §C — Requirements (GEARS)

### C.1 Measurement before doctrine

- **REQ-SDL-001** (Ubiquitous) — The measurement milestone shall complete and commit its evidence before any doctrine file named in §D is edited. *(premise: verdict §④ D 범위 orders the measurement inside this SPEC; §⑤ orders the amendment after its premise is re-measured)*
- **REQ-SDL-002** (Ubiquitous) — The measurement shall record the token cost of the mid-session-move skill-list duplication by a stated method, as the first-turn input delta (input + cache_creation + cache_read) between a session started inside the worktree and a session started in the primary checkout and then moved into it. The 2,406 B frontmatter-first-line byte sum shall not be cited as evidence of the token cost. *(premise: verdict §① Gaps, §① (2))*
- **REQ-SDL-003** (Ubiquitous) — The measurement shall record, by a stated method, whether `/clear` issued after a mid-session move restores the skill set of a launcher-started session in the same tree — by comparing loaded skill sets, never by same-transcript continuity (a real `/clear` opens a new transcript file; observed 369/369, delta audit D2). *(premise: verdict §② D row "미측정")*
- **REQ-SDL-004** (Ubiquitous) — The measurement shall commit its caps before the first probe runs, in a commit that carries no probe output. The caps are: at most 6 probes, at most 2 turns per probe, every probe bounded by an external `timeout`, headless probes declared `--model haiku`, and an aggregate wall-clock cap stated in the caps file. *(premise: verdict header — declared caps, 6 declared / 5 used, haiku, 1 turn)*
- **REQ-SDL-005** (Event-driven) — When a cap is reached before an item of REQ-SDL-002 or REQ-SDL-003 is observed, the measurement shall stop and record that item as a Gap rather than as a result. *(premise: verdict §① Gaps)*
- **REQ-SDL-006** (Unwanted) — The measurement probes shall not: run in the primary checkout or in the card worktree (they run in a disposable scratchpad repository structure outside both); change branch state in the primary checkout; leave a process of their own process group running; run without the `MOAI_KANBAN*` variables scrubbed in the same compound invocation. When a probe records a zero count, the evidence shall carry a positive control that returns a non-zero count on a known two-source input. *(premise: verdict header isolation section; verdict §① Claim 3 positive-control note)*

### C.2 Movement-prohibition doctrine

- **REQ-SDL-007** (Ubiquitous) — The dispatch doctrine shall state, in both the template and local copies of the dispatch rule (template edited first): a lane starts a card session inside the card worktree through the launcher; mid-session movement between worktrees is prohibited; when a mid-session move is unavoidable, `/clear` is issued exactly once, after the move. The added clause shall compose with, not amend, the "A new card starts in a new worktree" clause: that clause's exit-first requirement is satisfied by ending the lane session and launching in the new tree. The local-only lane protocol (`gitflow-lane-protocol.md`) shall carry a one-line pointer to the prohibition — a pointer, not a restatement. *(premise: verdict §④ D 범위, §② D row; t1219 Evidence 1 vs 4)*
- **REQ-SDL-008** (Unwanted) — The doctrine text added by this SPEC shall not claim that launcher-start eliminates the primary `CLAUDE.local.md` load, and shall not claim the prohibition removes the instruction-file double load. *(premise: verdict §① Claim 2 — launcher-start sessions still load it)*
- **REQ-SDL-009** (Where) — Where the Implementation Kickoff decision selects a mechanical movement guard, the SPEC scope shall extend to a guard clause in both copies of the dispatch rule describing the guard's detection rule and its denial message; no Go code is written under this SPEC. Where the kickoff decision does not select it, no guard clause shall be added and the prohibition shall remain documentation-only. The choice is presented in §F and is not pre-decided by this SPEC. *(premise: verdict §④ "가드로 차단할지는 SPEC 에서 선택지로 제시한다")*

### C.3 N1 amendment (draft now, substitute at run)

- **REQ-SDL-010** (Ubiquitous) — This SPEC shall carry the drafted rewording of the card-transition `/clear` clause: on a card transition, `/clear` happens exactly once, after the move; a phase end that is not a card transition clears at the phase boundary unchanged; the obligation is not reduced. The amendment shall be applied to both copies of the dispatch rule, and the two copies shall remain identical where they are identical today. Plan phase drafts the text only; the substitution is run-phase work. *(premise: verdict §⑤ N1)*
- **REQ-SDL-011** (Event-driven) — When the run phase begins the substitution, it shall first re-measure at the develop tip that t1175 has landed on develop and is absorbed into the working branch, and shall re-locate the amendment anchor by content (the `[HARD]` paragraph beginning "A companion session does not carry one card's context into the next card") in both copies; the substitution shall not proceed from a remembered line number. *(premise: verdict §⑤ ordering condition, §⑦ ND2, §⑨ "대상 줄은 착지한 사본에도 그대로"; observed anchor drift in this tree)*
- **REQ-SDL-012** (Unwanted) — The lines this SPEC adds to any template file shall not carry card ids, SPEC ids, dated incident references, or commit hashes. *(premise: template-neutrality doctrine; carried over from v0.2.0 REQ-SDL-013 — process rule, premise-independent)*

## §D — Affected Surfaces

| Surface | Path | Kind |
|---|---|---|
| T | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | template, edited first |
| L | `.claude/rules/moai/workflow/kanban-dispatch.md` | local mirror of T (byte-identical today) |
| G | `.claude/rules/local/gitflow-lane-protocol.md` | local-only, pointer line |
| E | `.moai/reports/t1279/` | measurement evidence |

## §E — Exclusions (What NOT to Build)

### Out of Scope — A: L2 worktree migration (shelved, revival conditions recorded)

- A moves card worktrees outside the primary folder to eliminate the `CLAUDE.local.md` double load at source (verdict §② A row — the only solution that removes Claim 1's cause for instruction files). Shelved per verdict §④ because three tool paths block it (verdict §③).
- **Revival conditions (both must hold, for a future card):**
  1. an authorized L2 creation path exists (`moai worktree new` L2 mode or equivalent), and the WorktreeCreate hook supports L2;
  2. a lane-transition alternative exists — an authorized path for a session already inside a worktree to move to another L2 or integration tree (e.g. a launcher relaunch procedure), or the Claude Code EnterWorktree constraint is lifted.
- Unverified at shelving time (verdict §③ Gaps): `moai integration acquire/release` and the `git -C` guard on L2 paths.

### Out of Scope — B: primary CLAUDE.local.md slimming

- t1243/t1259 (CLAUDE.local.md → AGENTS.local.md migration) address the same double-load cost; this SPEC does not touch the primary file (verdict §④).

### Out of Scope — C: `--setting-sources` adjustment

- Rejected: `user,project` disables `CLAUDE.local.md` loading wholesale, losing the worktree copy too (verdict §② C row, t1219 audit N1).

### Out of Scope — candidate cards #1 and #2 (separate queued candidates)

- Redefinition of the `/clear`-continuity measurement (real `/clear` opens a new transcript, 369/369; delta audit D2) — candidate #1, verdict §⑩.
- SPEC-SESSION-MIDMOVE-001 residual defects D1 / D3 / D4 — candidate #2, verdict §⑩.

### Out of Scope — instruction-file double load at launcher start

- The +20,855-token `CLAUDE.local.md` double load persists on launcher-started nested L1 sessions (verdict §① Claim 2 + Claim 4). This SPEC does not fix it; B's owners do. The reason `CLAUDE.md` alone is skipped in that configuration is unobserved and is not investigated here.

### Out of Scope — Go or runtime changes

- No `.go` edits, no hook code, no launcher changes. The mechanical guard of REQ-SDL-009, if selected, is a documentation clause only; any Go-based enforcement is a separate SPEC.

### Out of Scope — integration-window worktree moves and the detail companion

- The brief entry into the release/integration worktree to merge a card branch keeps its current wording. `kanban-dispatch-detail.md` is unchanged by this SPEC.

## §F — Kickoff Decision Pending

- **Mechanical movement guard (REQ-SDL-009)** — one binary choice surfaced at Implementation Kickoff:
  - **Select:** the dispatch rule gains a guard clause (detection rule + denial message, documentation only). Cost: one more clause in an always-adjacent rule file; the guard has no mechanical teeth under this SPEC.
  - **Do not select:** the prohibition is doctrine-only. Cost: compliance rests on the doctrine and review, not on a gate.
- Per the lane's standing rule, the lane applies Implementation Kickoff autonomously after the plan-auditor PASSes; this SPEC stays `status: draft` until then.
