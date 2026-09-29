---
id: SPEC-FACTORY-LANE-JOIN-SOCKET-001
title: "Plan — factory lane join tolerates run-record absence"
version: "0.1.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# plan.md — SPEC-FACTORY-LANE-JOIN-SOCKET-001

## §A Context

- **Card**: t1330 (factory lane join defect repair + `-l/--lead` flag). Join path ONLY; t1332 owns
  auto-dispatch defaulting.
- **Worktree**: `.moai/worktrees/t1330`, branch `WT-join-lead-socket`, base develop `68e37864a`.
- **Tier**: L (rationale: run scope ~20+ files across three packages plus a 16-file docs surface;
  new shared mechanism at the join point; 12 REQ / 16 AC inside Tier L ceilings of 25/25;
  constitutional-adjacent — it layers beside two settled fail-closed SPEC contracts).
- **Cycle**: tdd (RED-GREEN-REFACTOR). RED baseline for the new behavior is in `acceptance.md` §C.
- **Authority chain**: this SPEC layers beside SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-001 and
  SPEC-FACTORY-RUN-RETIRE-001 REQ-014 without amending either (spec.md §C).

### §A.5 PRESERVE list (never modify)

- `internal/factorymsg/store.go` — `ResolveActiveRun`, `ValidateActiveRun`, the refusal strings.
- `internal/homestate/factory_run_retire.go` — reconciliation, retirement bases, `retirable()`.
- `internal/kanban/bootstrap.go` — `NextFactoryLaneNumber` semantics (AC-015 pins them).
- `internal/kanban/factory_slots.go` — atomic claim, bump rule, legacy-row rules.
- `internal/cli/factory.go` parse truth table (`parseLauncherEntry` invariants; the flag parse
  EXTENDS it, never reorders existing checks).
- `internal/hook/factory_messages.go` — the bind chain reads env and records; no behavior change
  is needed there (AC-012 verifies the existing chain binds against the resumed run).
- Existing regression tests' current assertions (extend, never weaken).

## §B Known Issues (relevant subset of the 12 categories)

- **B1 Cross-platform build tags**: candidate-env reading diverges per OS (`ps eww` darwin,
  `/proc/<pid>/environ` linux, no equivalent on windows). Put platform readers behind per-OS
  files with a build-tag-free seam (the RUN-RETIRE REQ-013 pattern); `GOOS=windows go build`
  must stay green.
- **B2 Cross-SPEC policy conflict**: the two fail-closed SPECs are the conflict surface — resolved
  by the parallel-path framing (spec.md §C). If the run phase finds the framing cannot hold, HALT
  with a blocker report naming the clause; do not widen the resolver.
- **B4 Frontmatter schema**: artifacts stateless on the status axis; only spec.md carries
  `status:`.
- **B6 spec-lint headings**: Out of Scope sections use `### Out of Scope — <topic>` H3 (already
  satisfied in spec.md §E; keep the convention in any later amendment).
- **B8 Working-tree hygiene**: no runtime-managed files (`.moai/state/`, `.moai/harness/`) in
  commits; stage by explicit pathspec.
- **B10 Scope discipline**: t1332's auto-dispatch defaulting stays out; numbering stays
  project-scoped.
- **B12 CHANGELOG**: sync-phase (manager-docs) concern; run phase does not touch CHANGELOG.md.

## §C Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD   # expect WT-join-lead-socket / 68e37864a+lane commits
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/cli -run '^(TestFactoryRunSelectionAtomicSlotsAndArgv|TestGLM_FactoryLeadRunIsJoinableByLane)$' -count=1
golangci-lint run --timeout=2m 2>&1 | tail -5              # baseline vs NEW findings
```

Expected at plan close: build green both GOOS; regression pair green (observed: `ok
github.com/modu-ai/moai-adk/internal/cli 3.531s`); lint baseline recorded before M1.

## §D Constraints

- No `AskUserQuestion`; blockers go to the orchestrator as structured reports.
- No `--no-verify`, no force-push, Conventional Commits with `🗿 MoAI` trailer and the card id in
  every commit body.
- Verification is lane-local: targeted packages only (`internal/cli`, `internal/homestate`,
  `internal/kanban`); NEVER `go test ./...` (parallel factory lanes; full suite is CI's).
- REQ-012 isolation in every new test (`t.TempDir()`-based HOME/MOAI_HOME; project dirs under
  TempDir, never inside the repository).
- Discovery is read-then-write: the resume is the ONLY write, and it writes exactly one run row +
  one event, transactionally.

## §E Self-Verification deliverables (run phase)

Per manager-develop template §E with VCI §3 attribution: E1 AC matrix (16 rows), E2 both-GOOS
build, E3 targeted-package coverage, E4 subagent-boundary grep for touched packages, E5 lint
NEW-vs-baseline, E6 branch HEAD + commits (lane does NOT push — lead batch-pushes develop), E7
blockers, E8 RED output for the first GREEN flip (verbatim pre-implementation R-1/R-2/R-3 are
pinned in acceptance.md §C).

## §F Milestones

Ordered by decision-reversibility: data-model and new-type decisions first, mechanical docs last.

- **M1 (Priority High) — homestate resume writer.** New function: restore a run row to `active`
  — insert-when-absent / update-when-retired — stamping a SUPPLIED owner identity (pid +
  process-start), inside one transaction, appending a `run.resumed` event whose payload records
  the verification basis. Does NOT touch `recordFactoryRunStart` (AC-007's mutant separates the
  two writers). Flips AC-006/007/008.
- **M2 (Priority High) — leader discovery primitive with seams.** Candidate enumeration
  (socket-dir pid names + any-status run rows' lead pids) → per-candidate probe (liveness
  fingerprint, leader-label match, canonical-project membership, run-id env read) → verified
  candidate list. Build-tag-free seam taking an injected candidate list + probe functions so the
  classifier matrix (AC-004/005/011) is exercisable on darwin. Platform readers behind per-OS
  files. Flips AC-004/005 (unit level) and defines what M3 wires.
- **M3 (Priority High) — join-point integration + flag + env.** Discovery fallback on the
  `NO_ACTIVE_FACTORY` branch of the shared join gate; `-l/--lead` parse (default
  `kanban.LeaderLabel()`, legacy-value refusal, lane-only, `--factory-run` conflict) in
  `parseLauncherEntry`'s flag surface; resume + re-enter gate; `MOAI_KANBAN_LEAD_NAME` export on
  the discovery path. End-to-end AC-001/002/003/009/010/011/012 flip here.
- **M4 (Priority Medium) — regression + parity matrix.** Update the regression pair for the new
  behavior without weakening current assertions; run the mirror matrix (cc/glm/codex-twin —
  AC-013 including the source-level no-private-copy assert); run the preserved-invariant suites
  (AC-014/015); mutant legs for AC-005/007 recorded verbatim.
- **M5 (Priority Low) — docs and help.** cc/glm help text `-f lane` block (`glm.go:85-98`,
  cc twin); 4-locale `docs-site/content/{en,ja,zh,ko}/advanced/factory-mode.md`,
  `kanban-mode.md`, `cli-reference/launchers.md`; 4 README occurrences. Korean canonical locale
  chain per the oss-docs rules; all locales in one change (parity obligation). Flips AC-016.
  CHANGELOG stays sync-phase.

### Clarifications the operator confirms at the plan gate

1. **[NEEDS CLARIFICATION: --lead default value — card text `lead` vs live canonical `leader`]**
   Recommendation: **default `kanban.LeaderLabel()` = `"leader"`**, diverging from the card
   text. Why: `lead` is the refused legacy spelling in this tree (`role.go:46`,
   `refuseLegacyEntryNames` REQ-RNC-004/-007); no leader can carry that name anymore, so a
   `lead` default would target nothing that can exist; the live leader argv and env both read
   `leader`. The divergence is recorded in spec.md §B REQ-008; one operator confirmation at the
   gate closes it.
2. **[NEEDS CLARIFICATION: amendment-vs-parallel-path framing against
   SPEC-FACTORY-MIXED-HOOK-001 / REQ-014]** Recommendation: **parallel path, no amendment** (the
   framing this SPEC is written in). Why: the refusal contracts stay literally true (the resolver
   is untouched); discovery restores the record and re-enters the same gate; RUN-RETIRE's
   rejected alternative (d) forbids picking among records, not discovering live processes. If the
   operator prefers explicit amendment instead, spec.md §C and REQ-001/004/006 need a revision
   pass before run.
3. **[NEEDS CLARIFICATION: run-id provenance when the record is absent but the lead is alive]**
   Recommendation: **read the run id from the verified leader's own process env**
   (`MOAI_KANBAN_ID`), and decline any candidate whose env cannot be read (fail-closed). Why:
   the run id is the broker address the lead already speaks on — minting or guessing an id joins
   the lane to an empty broker (finding 5 of research.md); env reading is measurable on
   darwin/linux for same-user processes, and windows degradation is an honest refusal. Alternative
   if the operator rejects env reading: discovery via broker peers of any-status runs only
   (weaker — misses the never-recorded case, and stale-prone).

## §G Anti-patterns (refuse during run)

- "Just widen ResolveActiveRun to accept retired rows" — REQ-006/REQ-014; blocker report instead.
- "Reuse recordFactoryRunStart for the resume, it already writes active" — it stamps the caller;
  AC-007's mutant exists precisely for this.
- "Discovery can trust the socket file / the broker peer row" — REQ-002; existence is not
  liveness.
- "Auto-pick the first verified leader when there are two" — REQ-005.
- "Also reset lane numbering per run while we're here" — §A.5 PRESERVE; scope discipline.
- Writing the resume outside a transaction / without the event — REQ-004; audit is not optional.

## §H Cross-references

- spec.md §C (design decision + rejected alternatives), §A.2 (verified mechanism anchors)
- acceptance.md §C (RED ledger, tree pin `68e37864a`), §D (AC set)
- design.md §B (classifier composition), §C (resume writer), §D (seam strategy), §E (platform
  matrix)
- research.md (four-lens fan-out, workflow run wf_7795d774-54e; plan-phase re-verification)
- SPEC-FACTORY-RUN-RETIRE-001 REQ-002/005/006b/014; SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-001/-002
