---
id: SPEC-FACTORY-LANE-JOIN-SOCKET-001
title: "Design — verified leader discovery and run resume"
version: "0.1.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# design.md — SPEC-FACTORY-LANE-JOIN-SOCKET-001

## §A The mechanism in one flow

```
lane join (cc | glm | codex twin)
  └─ enterSelectedFactoryRun(root, explicit, requireActive=true)
       ├─ ResolveActiveRun OK  → gate passes (unchanged path)
       └─ NO_ACTIVE_FACTORY AND explicit == ""
            └─ discovery.DiscoverLeader(root, targetLabel)      [read-only]
                 ├─ 0 verified  → return the original refusal
                 ├─ ≥2 verified → fail closed naming candidates (no mutation)
                 └─ exactly 1   → homestate.ResumeRun(root, runID, leaderPID, leaderStart, basis)
                                  → re-enter enterSelectedFactoryRun (same gate, now passes)
                                  → export MOAI_KANBAN_LEAD_NAME=<leader name> for the child
```

Everything left of `ResumeRun` is read-only probing; the resume is the single write, and it
re-enters the EXISTING gate rather than bypassing it — so every downstream invariant (run id into
`EnvMoaiKanbanID`, lane claim stamping, hook `ValidateActiveRun`) is exercised exactly as an
ordinary join.

## §B Classifier composition (what makes a candidate "verified")

Four facts, all read from the candidate process, each independently decline-able:

1. **Liveness (REQ-002)** — `homestate.ProbeProcessIdentity(pid)` returns live **and** the
   process-start fingerprint is non-empty. Retire-grade: identical predicate family to
   `ClassifyOwnerWith` (`factory_run_retire.go:115-141`). Socket existence, peer rows, and
   registry rows feed only the candidate LIST, never the verdict.
2. **Leader-label match (REQ-008)** — the candidate's argv carries `--name <label>` (or `-n`)
   with label == the targeted leader label (default `kanban.LeaderLabel()`), matched with the
   same label-form tolerance `parseNamedLabel` uses. A `lane-<n>` match is not a leader match.
3. **Project membership (REQ-003)** — the candidate's cwd (same mechanism the t1038 probe used:
   `lsof -d cwd` on darwin, `/proc/<pid>/cwd` on linux) canonicalizes to
   `homestate.CanonicalProjectRoot(root)` — the same key that scopes factory.db and the brokers
   (`paths.go:22-50,151-175`), so worktrees of one repo converge.
4. **Run identity (REQ-003)** — the candidate's env carries a parseable `MOAI_KANBAN_ID`. This
   is the strongest single fact (a moai-launched session carries it; nothing else does) and it
   supplies the resume's run id. Unreadable env → decline.

Sources for the candidate pid list, in order: pid names in the runtime socket directory
(`/tmp/cc-socks/*.sock` — existence = candidacy only), then lead pids from `runs` rows in any
status, then broker `peers` pids of any-status runs. Deduplicated; each candidate bounded by the
probe deadline. The enumeration seam accepts an injected candidate list so tests never depend on
the host's socket population.

## §C Resume writer (the only write)

New `homestate` function (sibling of `RecordRun`, `runtime.go`): given
`(runID, verifiedPID, verifiedStart, basis)` in ONE transaction:

- row absent → INSERT `active` with the supplied owner stamp;
- row `retired` (or otherwise non-active) → UPDATE `status='active'`, overwrite
  `lead_pid`/`lead_process_start` with the verified stamp;
- row already `active` → no-op (discovery would not have fired; defensive branch);
- append `run.resumed` event, payload carrying the basis (probe facts: liveness fingerprint
  present, label match, membership evidence kind) — distinct kind from `run.started`
  (`runtime.go:43` precedent), so an operator audits why the row came back.

Why the owner stamp is the leader's and not the caller's: `recordFactoryRunStart` stamps
`CurrentProcessFingerprint()` of the calling launcher (`factory.go:323`); called from the lane
launcher it would name the lane, and the lane's exit would classify the run dead — retiring a
live lead's run (the exact REQ-005 hazard RUN-RETIRE guards). The verified stamp is what the
probe itself measured, so it is retire-grade correct at write time; if the leader dies later,
`ReconcileActiveRuns` re-retires by the same standard. Self-healing, no new enforcement surface.

## §D Seam strategy (testable without a real leader process)

- Discovery takes injected dependencies: candidate list, liveness probe func, argv/env/cwd
  readers. Tests build candidates from stub processes (the `installFactoryLaunchSeam` pattern)
  or pure fakes; the classifier matrix (AC-004/005/011 edge cases) runs on darwin with zero
  platform dependence.
- Env/argv/cwd readers sit behind per-OS files (`_darwin.go` / `_linux.go` / `_windows.go`)
  behind a build-tag-free interface — the RUN-RETIRE REQ-013 pattern (build-tag-free seam so
  behavior is exercisable from any host).
- E2E legs (AC-001/002/003/012) drive the built launcher binary against a TempDir project with a
  stub child binary, isolated HOME/MOAI_HOME (REQ-012).

## §E Platform matrix

| Capability | darwin | linux | windows |
|---|---|---|---|
| pid liveness + fingerprint | ProbeProcessIdentity (existing) | same | same |
| candidate cwd | `lsof -d cwd` | `/proc/<pid>/cwd` readlink | best-effort → decline |
| candidate env (`MOAI_KANBAN_ID`) | `ps eww` (same-uid) | `/proc/<pid>/environ` (same-uid) | unavailable → decline |
| effect of decline | honest `NO_ACTIVE_FACTORY` (today's behavior) | same | same |

Windows therefore degrades to a refusal, never to a wrong join; if a later SPEC wants windows
discovery, it widens the reader behind the same seam. Pre-merge verification is darwin-only
(lane-local); linux/windows land via the develop-push CI per the established timing table
(RUN-RETIRE §A.2).

## §F Flag surface details

- Parse location: the `-f` flag parse inside `parseLauncherEntry`'s factory surface — new
  `-l/--lead <name>` accepted ONLY when the parsed shape is a lane join; leader entry + `--lead`,
  or `--lead` with `--factory-run`, are parse errors. Legacy value check
  (`kanban.IsLegacyLeaderSpelling`) runs at parse time with the canonical-form error message
  shape `refuseLegacyEntryNames` uses (AC-010).
- Default resolution: empty `--lead` = `kanban.LeaderLabel()`; resolved value exported to the
  child as `MOAI_KANBAN_LEAD_NAME` (existing envkeys constant) on BOTH the ordinary and the
  discovery path (REQ-009; today's ordinary path does not export it for lanes — the flag's
  arrival makes the export unconditional for lane joins, which AC-012 pins).
- `--name` and `--lead` name different sessions (self vs target); no interaction beyond the
  existing `-f lane` + `--name` conflict.

## §G Risks carried into run

- TOCTOU (leader dies between probe and resume) — §C self-healing note; the lane's hook degrades
  to "factory messaging degraded" (existing notice, `factory_messages.go:82-89` shape).
- argv/env format drift — regression pair pins the classifier contract; drift fails AC-001
  loudly.
- Candidate-population latency — per-candidate deadline; measured once in M2 against the live
  ~97-socket population.
- codex twin inheritance — behavior arrives via the shared join point; AC-013's source assert is
  the guard against a launcher-private copy.
