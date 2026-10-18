---
id: SPEC-FACTORY-LEADER-ADOPT-CORE-001
title: "Factory leader start over an orphaned run — promptless refusal with guidance, the deliberate --factory-run adopt pinned to the adoption write, and the guidance text that describes both"
version: "0.3.0"
status: in-progress
created: 2026-10-18
updated: 2026-10-18
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/homestate"
lifecycle: spec-anchored
tier: M
tags: "factory, leader-start, orphan-run, adopt, guidance, card-t1863"
card: t1863
related_specs: [SPEC-FACTORY-LEADER-ADOPT-001, SPEC-FACTORY-RUN-SELECT-CORE-001, SPEC-FACTORY-RUN-RETIRE-001]
---

# SPEC-FACTORY-LEADER-ADOPT-CORE-001 — Factory leader start over an orphaned run (shrunk core)

## A. History

| Date | Version | Change |
|---|---|---|
| 2026-10-18 | 0.3.0 | Plan-audit iteration 2 repair (FAIL 0.94, D7-D8; verdict `.moai/reports/t1863/plan-audit-iter2.md`; D1-D6 verified resolved). D7: AC-FLAC-006 gains the two-orphan rows 13 (`--resume` over A+A2 — the resume adoption admits neither of two candidates, warns and mints, both orphans survive the resume decision) and 14 (`--factory-run` adopting A while A2 survives); the resume-condition and `--factory-run`-condition mutant catches are re-attributed to those rows (on a single orphan the consuming path masks the mutant). D8: AC-FLAC-003's live-owner fixture seed crossed to `claude`, so AC-FLAC-008's legs adopt runs seeded with the OTHER launcher's backend and the backend equality is non-vacuous on both legs; the per-launcher backend-omission mutant named. Counts unchanged (10 REQ, 10 AC). |
| 2026-10-18 | 0.2.0 | Plan-audit iteration 1 repair (FAIL 0.94, D1-D6; verdict `.moai/reports/t1863/plan-audit-iter1.md`). D1: the adoption write is planned to set `runs.lead_backend` in the same transaction (`AdoptRun`'s UPDATE extended) — without it the takeover leaves the prior leader's backend on the row and the AC backend clauses are vacuous; C.3 now names the only two existing `lead_backend` write sites. D2/D3: the finder and guidance fixtures gain a second qualifying orphan and a boot-proof-basis excluded run. D4: the no-stdin guarantee gains a counting-reader instrument. D5/D6: the `-cr` widening row and the `classifyRuns` citation fix. Counts unchanged (10 REQ, 10 AC). |
| 2026-10-18 | 0.1.0 | Initial plan-phase draft (card t1863, Tier M, Class C). The shrunk core of the held predecessor SPEC-FACTORY-LEADER-ADOPT-001 (card t1849, three plan-audit FAILs 0.69/0.78/0.84, held by the leader's shrink ruling, never audited-PASS — nothing from that SPEC is claimed resolved here). Scope per the lane's binding decision D-t1863-1: the kept requirement set is the predecessor's REQ-LDA-001 (finder), 002 (trigger), 004 (reworded: the deliberate adopt rides the existing `--factory-run` path, no prompt choice), 005 (refusal abort), 008 (guidance block, DP-1), 009 (exemptions), 010 (no other writes), 011 (glm parity), 012 (KST rendering) and 014 (docs anchors), renumbered REQ-FLAC-001..010. Dropped: REQ-LDA-003/006/007 (the interactive adopt/new-run/cancel chooser and its choices), REQ-LDA-013 (interactivity detection from the standard streams) and REQ-LDA-015 (the `--spawn` pre-refusal) — all five are deferred to a follow-up card and are named in §F. Built from a fresh read of the code at tree `2178966bb90b28576e8a4c3bce340947b83c7cda`; the predecessor's iter-3 audit findings (D14-D17 defect classes) bind the authoring of this SPEC's ACs and plan. |

## B. Problem Statement

Observed 2026-10-17 (evidence case from the predecessor's card body; the production database was not re-read, see §G): run `tmvgvn` was active with a dead leader and 41 in-progress cards. A new leader started with `moai cc -f` and no `--resume`, and it minted a second run `tn1p4e` over the orphan. A lane join then failed `AMBIGUOUS_FACTORY`, and the operator's way out (retiring `tmvgvn`) stranded the 41 card records.

The defect this SPEC owns is the first link: **a leader start without a resume option never surfaces a dead leader's run, even when that run still carries in-progress cards — it silently mints a second run over it.** Nothing tells an unattended start before the new run id is recorded. The deliberate way out (`--factory-run <run-id>`) exists but takes over through an unconditional restamp that is not atomic against the listed owner and does not republish the leader address (§C.3).

## C. Root Cause (read from code at `2178966bb`)

1. **The leader branch mints unless a resume option is present.** The `cc` leader branch (`internal/cli/cc.go:194-231`) reads the prior run id from the environment (`:199`), enters leader mode (`enterFactoryLeaderMode`, `:201`, which mints the run id into the environment via `leaderRunID`, `internal/cli/factory_launch_helpers.go:81-89`, and stamps the leader address from that id, `internal/cli/factory.go:799`), resolves an explicit run (`enterSelectedFactoryRun`, `:203`, a no-op when no `--factory-run` is given and the caller is not a lane, `factory.go:477-479`), then decides re-adoption (`leaderEntryRunAdoption`, `:214`) and records the start (`recordLeaderRunStart`, `:218`). `leaderEntryRunAdoption` adopts only when `strictResumeCarrierSeen(args)` and the id was freshly minted and no `--factory-run` was given (`internal/cli/factory_run_adopt.go:163-166`); otherwise `recordLeaderRunStart` records the minted id through `recordFactoryRunStart` (`factory_run_adopt.go:173-184`, `factory.go:695-714`), whose `RecordRun` is an unconditional upsert that appends `run.started` (`internal/homestate/runtime.go:28-54`). The `glm` leader branch mirrors this exactly (`internal/cli/glm.go:264-281`, adoption at `:274`, record with `factory.BackendGLM` at `:281`). Before `recordLeaderRunStart` the launcher has changed only environment variables, which the deferred restore functions put back (`cc.go:202-207`) — so a refusal placed before the record writes no run and starts no session.
2. **The adoption primitive is safe and already exists.** `FindResumableRun` (`internal/homestate/factory_run_adopt.go:72-96`) classifies active runs and admits adoption only for a single candidate classified dead on the row's own stamp (`:92`); `AdoptRun` (`:112-145`) restamps the owner with a compare-and-set on the classified stamp (`:128`, refused when no row carries it, `:137-139`) and appends `run.resumed` (`:141`). The launcher half, `adoptResumedLeaderRun` (`internal/cli/factory_run_adopt.go:64-112`), adds the codex-led guard (`:87-93`) and the live-legacy-peer guard (`:100-105`) and republishes both the run id and the leader address after the write (`:109-110`).
3. **The explicit `--factory-run` path does not use that machinery.** With `--factory-run <run-id>`, `enterSelectedFactoryRun` resolves only an active run (`factorymsg.ResolveActiveRun`, `internal/factorymsg/store.go:343-355`), applies the same two guards as errors (`factory.go:489`, `:496-501`), and replaces only `EnvFactoryRunID` (`:502-503`) — the leader address stays the one `enterFactoryLeaderMode` derived from the minted id (`factory.go:799`). The takeover write is then `recordFactoryRunStart`'s unconditional `RecordRun` upsert: it overwrites the owner stamp without checking the previous owner, appends `run.started` rather than `run.resumed`, and clobbers `lane_capacity`. The resume adoption's guarantees (compare-and-set, event kind, address republish) are exactly what the deliberate adopt lacks. `runs.lead_backend` is written today only by `RecordRun` (`internal/homestate/runtime.go:44-46`) and the resume-discovery INSERT (`internal/homestate/factory_run_resume.go:58`); `AdoptRun`'s UPDATE sets only the stamp columns (`internal/homestate/factory_run_adopt.go:128-129`), so this SPEC's adoption write extends it (plan M2) — a takeover that skipped the backend write would leave the prior leader's backend on the row.
4. **A dead-owner run with live records survives lane-join reconciliation.** `ReconcileActiveRuns` retires a dead-owner run only through `retireRunUnlessLive`, which preserves a run holding a non-terminal card or a registered worker (`runHasLiveRecords`, `internal/homestate/factory_run_adopt.go:31-39`). The minted second run therefore sits beside the orphan, and a lane join is refused as ambiguous. Pinned by `TestReconcilePreservesDeadOwnerRunWithLiveRecords`, `TestReconcilePreservesDeadOwnerWithOnlyWorkers`, `TestReconcileRetiresDeadOwnerWithTerminalRecordsOnly` (ledger E4).
5. **The terminal card set is one shared definition**: `{done, failed, abandoned}` (`sqlTerminalStates`, `internal/homestate/factory_run_adopt.go:21`, pinned to `terminalStates` by `TestLiveRecordsTerminalSetMatchesSharedDefinition`).
6. **`moai codex -f` has no leader entry** (refused; the skill's entry table carries the row and `TestFactorySkillAssertionsMatchBehavior` replays it), so the leader start is exactly the `cc` and `glm` branches.
7. **Today's guidance text does not state the refusal.** The guidance anchor this SPEC defines is absent from all ten guidance files (ledger E5), and the four READMEs state that a dead leader's run "is retired automatically when the next lane joins" (`README.md:70` and the three locale twins — the unqualified form, true only for a run without live records per C.4).

## D. Requirements (GEARS)

Terms. **Orphan run**: an `active` run whose owner is classified `dead` on the run row's own owner stamp (the proof basis the adoption's compare-and-set needs, mirroring the single-candidate rule of `FindResumableRun`) and which holds at least one non-terminal card (§C.5). **Resume option**: the explicit vocabulary `strictResumeCarrierSeen` already recognizes (`--resume`, `-r`, the equals form, the attached `-r<value>`; `internal/cli/factory_run_adopt.go:126-155`). **Freshly minted**: the run id the launch would publish was minted by `leaderRunID`'s fallback — a still-set environment marker or a pasted `leader-<run-id>` name is deliberate and not minted (`leaderRunIDWasMinted`, `factory_run_adopt.go:34-42`). **Guidance block**: the one stderr block REQ-FLAC-005 defines. This SPEC adds no interactive surface anywhere: no prompt is rendered and no input is read on any path it adds.

### D.1 Detection

- **REQ-FLAC-001** (Ubiquitous): The factory state layer shall provide one read-only orphan finder that returns every orphan run together with its non-terminal card count, and shall return no run whose status is not `active`, no run whose owner is live or indeterminate, no run whose owner is classified dead only through the lead peer or the boot proof, no run holding only workers, and no run holding only terminal cards.

### D.2 Leader start

- **REQ-FLAC-002** (Event-driven): When a `moai cc -f` or `moai glm -f` launch carries no resume option and no `--factory-run`, would publish a freshly minted run id, and the orphan finder returns at least one orphan run, the launcher shall refuse the launch under REQ-FLAC-005 before any factory run record is written.
- **REQ-FLAC-003** (Event-driven): When a leader start carries `--factory-run <run-id>` naming an active run, the launcher shall take that run over through the same adoption write the resume adoption uses: the takeover shall be atomic against the owner identity the launch itself classified (a compare-and-set on the run row's owner stamp), shall keep the run's id, cards and records (no second run row, no `run.started` event, the run's existing lane capacity unchanged), shall publish the adopted run id and the leader address derived from it exactly as the resume adoption does, and shall refuse a codex-led run and a run holding a live legacy peer.
- **REQ-FLAC-004** (Event-detected): When a deliberate adoption under REQ-FLAC-003 is refused, or the named run's owner changed after the launch classified it so the takeover cannot be made atomically, the launcher shall abort the launch with a non-zero exit and the refusal text, and shall not mint, record or modify any run.
- **REQ-FLAC-005** (Event-driven): When REQ-FLAC-002 applies, the launcher shall write one guidance block to stderr naming each orphan run (run id, owner classification, lead pid, start time rendered under REQ-FLAC-009, non-terminal card count) and the deliberate adopt command of the launcher's own entry (`moai cc -f --factory-run <run-id>` for `moai cc -f`, `moai glm -f --factory-run <run-id>` for `moai glm -f`), and shall then exit non-zero without writing any factory run record or starting a session.
- **REQ-FLAC-006** (Unwanted behavior): When the launch carries a resume option, carries `--factory-run`, publishes a run id that was not freshly minted, or the orphan finder returns no orphan run, the launcher shall neither write the guidance block nor refuse because of an orphan, and the existing behavior of that launch shall hold unchanged.
- **REQ-FLAC-007** (Unwanted behavior): No path this SPEC adds shall retire, restamp or otherwise write to any run other than the single run adopted under REQ-FLAC-003, and no path shall offer to retire a run.
- **REQ-FLAC-008** (Ubiquitous): The `moai glm -f` leader start shall behave identically to `moai cc -f`, the two sharing one gate and one adoption path and differing only in the backend recorded and in the launcher's own command name carried by the guidance block and the refusal text.

### D.3 Rendering

- **REQ-FLAC-009** (Ubiquitous): Every run start time the guidance block renders shall be rendered in KST (UTC+9) as `YYYY-MM-DD HH:MM:SS KST`, and a value that does not parse as an RFC 3339 time shall be rendered unchanged.

### D.4 Guidance text

- **REQ-FLAC-010** (Ubiquitous): The distributed guidance for the factory leader start and lane join — the factory skill and its local twin, the four docs-site factory-mode pages, and the four README dead-leader paragraphs — shall state that a leader start without a resume option over an orphan run is refused with guidance and creates no run; that `--factory-run <run-id>` is the deliberate adopt; and that a lane join does not retire a dead-owner run that still holds in-progress cards or workers; each statement shall be carried as a machine-checkable anchor line the acceptance criteria define.

## E. Decision Points

Full rows with routing labels are in `decision-index.md`.

| DP | Decision | Disposition in this SPEC |
|---|---|---|
| DP-1 | A leader start without a resume option over an orphan run: refuse with guidance, or print guidance and mint. | **Refuse with guidance** (REQ-FLAC-002, REQ-FLAC-005), carrying the predecessor's DP-1 ruling (board `d-20261017T140304Z-62a6`, decided by the leader under delegation) and the lane's shrink instruction. The board record cannot be digest-pinned (`moai decision read` exposes no digest), so the row routes FOUNDER for the operator to confirm (decision-index Q1). |
| DP-2 | Does the unified adoption write on the explicit `--factory-run` path extend to runs whose owner is live or indeterminate, or only to dead-owner orphans. | Authored as **unified** — every `--factory-run` leader start takes over through the adoption write (REQ-FLAC-003), one path, and the address republish fixes C.3 everywhere on the explicit path. FOUNDER (decision-index Q2): the option preserving today's restamp for live-owner takeovers is the ranked Default there. |
| DP-3 | The refusal of REQ-FLAC-005 applies to every fresh-mint leader start over an orphan, terminal or not, until the deferred chooser card lands. | Authored as **terminal-independent** — the shrink defers interactivity detection, so no stream is read and no prompt exists (Terms). FOUNDER (decision-index Q3): the smaller-surface option (refuse only off-terminal) is the ranked Default there. |

## F. Out of Scope

### Out of Scope — The interactive chooser and stream detection (deferred follow-up card)

- The interactive `adopt` / `new run` / `cancel` chooser prompt, the interactivity detection from the standard streams, and the `--spawn` pre-refusal of the predecessor's REQ-LDA-003/006/007/013/015 are deferred to a follow-up card and are NOT specified, reserved, or tested here. Until that card lands, the refusal of REQ-FLAC-005 applies to every fresh-mint leader start over an orphan, interactive terminal or not (DP-3), and no path this SPEC adds reads stdin or renders a prompt.

### Out of Scope — Lane-join chooser and retirement guard

- The interactive lane-join run chooser and the in-progress-card guard on `moai factory runs --retire` belong to the sibling lineage (SPEC-FACTORY-RUN-SELECT-CORE-001; the retire guard has landed as SPEC-FACTORY-RUN-RETIRE-001, `completed` in this tree). This SPEC reads no lane activity and changes no retirement path.

### Out of Scope — Cross-run card-record transfer

- Moving card records, leases, approvals, dispatch rows or broker state between runs belongs to its own card. Nothing here specifies, reserves a hook for, or tests it.

### Out of Scope — New cleanup power over live runs

- Retirement, forced cleanup, takeover or restamp of a run whose leader is live or indeterminate on the paths this SPEC adds is limited to the single deliberate adopt of REQ-FLAC-003; the orphan finder never returns such a run (REQ-FLAC-001), the refusal path writes nothing (REQ-FLAC-005), and nothing offers retirement (REQ-FLAC-007). The owner classifier, the boot proof and `ReconcileActiveRuns` semantics are unchanged.

### Out of Scope — Resume-shaped starts

- A launch carrying a resume option keeps its current behavior, including the warn-and-mint outcomes of `adoptResumedLeaderRun` for several candidates, an indeterminate owner, a codex-led run and a live legacy peer (ledger E3 pins them). A `--continue` or `-c` launch is not part of the resume vocabulary (`resumeFlag` and `resumeFlagShort`, `internal/cli/lane_resume.go:36-37`), so over an orphan it is refused like a bare start.

### Out of Scope — Unattended adoption

- Adopting an orphan run without an operator act. No flag, environment variable or configuration makes a refused start adopt on its own; the explicit `--factory-run <run-id>` command is the operator's deliberate act.

### Out of Scope — Workers-only dead-owner runs

- A dead-owner run that holds registered workers but no non-terminal card is neither refused over nor adopted by the new paths (REQ-FLAC-001). Residual, unchanged from the predecessor: the lane join still preserves such a run (`TestReconcilePreservesDeadOwnerWithOnlyWorkers`), so a duplicate-active-run state can still arise next to it with no guidance. The follow-up chooser card may widen the finder.

### Out of Scope — Always-loaded rule text

- The `factory-dispatch*.md` and `cross-session-messaging*.md` rule files, whose role-core text is injected at session start under a byte budget. Guidance lives in the on-demand skill and the public documents only.

## G. Verification Gaps (not observed in this plan phase)

- The evidence case came from the predecessor's card body; the production factory database was not read.
- No behavioral RED was observed. Every new test is absent today, so each RED cell is the empty sweep (ledger E1, E2); behavioral REDs are recorded in `progress.md` §E.2 at each milestone's RED step.
- The three FOUNDER rulings of §E were decided by the leader under delegation (board records `d-20261017T140304Z-62a6` and the shrink ruling of 2026-10-17); `moai decision read` (installed build `ee981a932`) exposes the records but no digest prefix, so no `board:<id>#<digest>` line can be pinned — the decision-index rows say the same.
- The `--factory-run` takeover mechanics of C.3 were established by reading `cc.go`, `factory.go`, `factory_run_adopt.go` and `runtime.go`, and by the GREEN baselines of ledger E3 (`TestFactoryJoinAndLeadAcceptNonCodexRun` drives today's explicit lead path; `TestFactoryLeadRefusesCodexLedRun` pins its codex refusal); the new adoption-write behavior itself was not run.
- The docs-site Hugo build and the four-locale verify recipe (`hns-oss-docs-verify`) were not run; they are run-phase gates for M4.
- `moai spec lint --strict` was judged by the installed build `ee981a932` (v3.2.0-rc.36-t1822c), a verified ancestor of this tree's HEAD `2178966bb`; a tree-built binary was not produced in the plan phase.
