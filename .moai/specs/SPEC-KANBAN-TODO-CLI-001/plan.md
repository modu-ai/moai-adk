---
id: SPEC-KANBAN-TODO-CLI-001
title: "Plan — moai todo CLI subcommand with lock-guarded backlog store"
version: "0.1.1"
created: 2026-08-14
updated: 2026-08-14
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, cli, backlog, concurrency, plan"
tier: M
---

# plan.md — SPEC-KANBAN-TODO-CLI-001

## §A Context

Kanban card t12. The `/moai todo` skill currently instructs the model to Read→Write `.moai/state/kanban/backlog.json` directly; with five concurrent sessions on one checkout that window is minutes wide and resolves to last-writer-wins (2026-08-14 incident: 3 cards lost, t4/t5/t6 ID collision — research.md §2). This SPEC replaces the model-driven cycle with a real `moai todo` Go CLI backed by a lock-guarded store.

Evidence base: `research.md` (all claims command-attributed; nothing re-derived here). Affected surfaces:

- New: `internal/kanban/backlog_store.go` (or similarly named), `internal/cli/todo.go` (+ `_test.go` files).
- Modified: `internal/cli/root.go` (register `newTodoCmd()`), `.claude/skills/moai/workflows/todo.md` (shrink), `internal/template/templates/.claude/skills/moai/workflows/todo.md` (same delta), regenerated `internal/template/catalog.yaml` via `make build`.
- Preserved: `internal/kanban/board*.go` (board store, `requireLeadRole`, WIP surfaces — untouched per spec.md §D), `internal/atomicfile` (consumed as-is), `internal/session` (consumed as pattern reference only).

Tier M (spec frontmatter `tier: M`). REQ count is 16 — at the Tier M ceiling. The SPEC was authored at 19 REQs by the first delegation; plan-audit iteration 1 (D2) routed the over-budget set to consolidation, resolved at v0.1.1 by merging three pairs (005+006, 013+014, 017+018 — grouping, no scope loss; spec.md §F History carries the old→new map). AC count is 15 (`AC-TODO-001`, `AC-TODO-003` … `AC-TODO-016`; `AC-TODO-002` was never defined — no filler inserted, count labels corrected), within the Tier M ceiling via REQs that share one verification command.

## §B Known Issues

- **B1 (cross-platform build tags)**: the lock substrate already carries the Unix/Windows split (`board_lock_unix.go` / Windows atomic-create counterpart). The backlog store inherits it by reuse — NO new platform files are authored here. The Windows claim discipline is AC-TODO-016: `GOOS=windows go build` alone never covers `_test.go`; `GOOS=windows go vet` is the gate.
- **B2 (working-tree hygiene)**: five sessions share this checkout (kanban mode). Run phase MUST NOT touch `.moai/state/kanban/backlog.json` (production data) except through read-only observation; tests use `t.TempDir()`.
- **B3 (subagent boundary)**: no AskUserQuestion in `internal/cli` — enforced by the existing static-guard test pattern (`TestNew_NoAskUserQuestion`, research.md §9); a todo-specific guard test is AC-TODO-014.
- **B4 (template mirror hazard)**: the twins are currently byte-identical (research.md §8), so there is no intentionally-neutralized variant to restore — but the rule stands: apply the SAME delta to both, never `cp` one over the other, and verify with `MOAI_TEMPLATE_LEAK_STRICT=1 go test` (neutrality: no SPEC IDs, no internal dates in the shrunk skill).
- **B5 (catalog regeneration)**: after editing the template twin, `make build` regenerates `internal/template/catalog.yaml` — that regeneration must be committed with the template change (prior mirror-commit CI failure precedent).

## §C Pre-flight (run phase, before M1)

```bash
git branch --show-current && git rev-parse HEAD     # Late-Branch: branch/worktree setup is the run session's act (§F Phase 13 note)
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5       # baseline for NEW-vs-pre-existing classification
grep -rn 'Backlog' internal/kanban/*.go | head      # expected: only ColumnBacklog (name space still free)
moai todo 2>&1 | head -3                            # expected: Unknown command "todo" (the RED baseline)
```

## §D Constraints & Design Decisions

Each decision is stated with its justification. Decisions (a)–(h) correspond to the dispatch contract; no open clarification markers remain — spec.md already binds each of them to a REQ, so re-opening any here would contradict the authored requirements.

### D-a Lock substrate — reuse `internal/kanban`, path-parameterized

**Decision**: reuse the in-package `internal/kanban` lock substrate — `acquireBoardLockImpl(lockPath)` is path-parameterized (flock Unix / atomic-create Windows, `BoardLockOwner{PID, CreatedAt}` recorded in the artifact) — pointed at a sibling artifact `.moai/state/kanban/backlog.lock`, acquired through a bounded-retry wrapper mirroring `acquireBoardLockSerialized` (25 ms × 40 ≈ 1 s). `internal/session` `Registry.withLock` is a **pattern reference only** (lock → read → mutate-callback → atomic write → release; reads lock-free): importing `internal/session` from a kanban-scoped store would be a reverse dependency, and its in-process path-keyed mutex layer is registry-specific (research.md §5).

**Why this seam and not another**: the substrate is in-package, already cross-platform, already carries owner identity for diagnostics, and SPEC-KANBAN-BOARD-001 has production-tested it. `internal/lockfile` is in-process-only and explicitly excluded by the substrate's own header. A new lock package would duplicate a reviewed concurrency primitive. Binds REQ-TODO-006/007.

### D-b No lead-role guard — deliberate

**Decision**: backlog writes do NOT apply `requireLeadRole`. The board guard exists because the board has exactly one writer (the lead); the backlog is the operator's queue and any session may append, pick, or complete (spec.md REQ-TODO-011, research.md §4). Stated explicitly so the plan-auditor reads the asymmetry as designed, not as an omission.

### D-c ID issuance inside the lock, persisted high-water mark

**Decision**: ids are issued inside the locked mutation from an additive top-level field `last_seq` (integer). When `last_seq` is absent (a version-1 file predating the field), the store derives it from the maximum present item id on first load and persists it on first write (REQ-TODO-009). Schema handling: **additive within version 1** — no version bump, no new per-item fields (spec.md §E keeps this an explicit out-of-scope).

**Why persisted, not derived**: `done` removes rows, so max-present-id derivation alone would reuse ids of removed items — exactly the t4/t5/t6 collision the incident recorded. The high-water mark must survive removals, therefore it must live in the file.

### D-d Verb semantics, no-prompt constraint

**Decision**: four verbs, none prompts (REQ-TODO-014; `internal/cli/CLAUDE.md` forbids AskUserQuestion in CLI; static guard test pattern `TestNew_NoAskUserQuestion` → AC-TODO-014).

| Verb | Lock | Behavior |
|---|---|---|
| `add "<text>"` | write | Append under lock; print issued id + queue position (REQ-TODO-002). |
| `list` | none | Render queue lock-free; `--json` emits structured records (REQ-TODO-003). |
| `done <n>` | write | Remove addressed row under lock; bare `<n>` normalized to id `t<n>`; explicit id is the preferred form because positions shift under concurrent adds (REQ-TODO-004). |
| `next` (bare) | none | Read-only oldest-first candidate listing for the lead's own AskUserQuestion channel — the pick stays the operator's act ([HARD] in the skill, "The lead never picks for the operator", research.md §10) (REQ-TODO-005). |
| `next <n> [--spec <SPEC-ID>]` | write | Mark addressed item `picked` + attach `spec_id` as ONE locked write (REQ-TODO-005). |

**Why `next <n>` is a locked write and not scope creep**: without it, the picked/spec_id transition remains a model-driven read-modify-write — the residual race this card exists to kill. It is bound by the second When-branch of REQ-TODO-005 in the authored spec; keeping `next` read-only would strand that branch without an owner. Not an open clarification item: the requirement is already dispatched scope.

### D-e File contract preserved verbatim

Missing file = empty queue (`list`/`next` report empty, `add` creates a version-1 file, never an error). Malformed file = error reported, file left byte-identical, never silently reset — the operator's queued intent is the one thing that cannot be regenerated (REQ-TODO-012, contract lines carried verbatim from the current skill, research.md §2).

### D-f Windows posture

The substrate's platform split is inherited by reuse (D-a). Stale-lock posture: bounded retry (~1 s) then a clear error naming the lock artifact path. Automatic stale-lock detection/clearing is OUT of scope at this tier — documented limitation, spec.md §E (on Unix the kernel releases flock on holder death, so the stale artifact case is a Windows-substrate concern only).

### D-g Atomic write

Persistence goes through `internal/atomicfile` (`Replace`; same-directory temp + rename, `ReadFile` absorbs the Windows delete-pending window on the read side — research.md §7). Same shape as `writeBoardAtomic`: `MarshalIndent` + trailing `\n`, temp in the target directory.

### D-h Skill shrink + template mirror

Local `.claude/skills/moai/workflows/todo.md` shrinks to "run the command": keep the queue philosophy, the boundaries, the kanban-dispatch cross-reference, and the [HARD] operator-selection rule; remove the direct Read→Write instructions and the temp-file/rename guidance (now internal to the store). The SAME delta — not a verbatim `cp` — is applied to the template twin `internal/template/templates/.claude/skills/moai/workflows/todo.md` (twins currently byte-identical, research.md §8; the no-verbatim-cp rule guards against overwriting intentionally-neutralized variants if one diverges before run phase). Template neutrality: no SPEC IDs, no internal dates in the shrunk skill body. Run phase MUST `make build` (re-embed + commit regenerated `catalog.yaml`).

### D-i Hard constraints (from spec.md §D)

- No new module dependencies: `internal/kanban` + `internal/atomicfile` only.
- Board surfaces (`board.go`, `board_store.go`, `board_recover.go`) untouched.
- Any new env-var names → constants in `internal/config/envkeys.go` (none anticipated).

## §E Self-Verification (plan phase)

- Frontmatter: all 12 canonical fields present, canonical names (no snake_case aliases), `phase: "v3.1.0 target"` (release target, not a lifecycle token).
- SPEC ID pre-write self-check executed as Bash: `PASS` (research.md §11).
- Requirements: 16, GEARS notation, mirroring spec.md's existing structure (Ubiquitous / When / Where).
- Out of Scope: spec.md §E carries four `### Out of Scope —` H3 sub-headings with bullets (satisfies `OutOfScopeRule`).
- AC coverage: 15 ACs cover all 16 REQs (acceptance.md §A coverage matrix; merged REQs map to two ACs each, grouping noted per AC).

## §F Milestones (ordered by decision-reversibility — data-model decisions first)

Data-model and lock-contract decisions are the ones a reviewer is most likely to overturn; they come first so review concentrates where change is likeliest. Mechanical steps (mirror, docs, verification sweep) come last.

- **M1 — Backlog store + lock + ID issuance** (D-a, D-b, D-c, D-e, D-g): new store file in `internal/kanban`; locked mutation path (serialize → load → mutate → atomic write → release-with-error-join); lock-free load; `last_seq` high-water mark with derive-on-absent; missing/malformed contracts. TDD: RED tests first (concurrent add, ID uniqueness, malformed untouched). Highest change-likelihood: schema field name (`last_seq`) and the derive-on-absent rule.
- **M2 — CLI verbs** (D-d, D-i): `newTodoCmd()` in `internal/cli/todo.go`, registered in `root.go`; `add` / `list --json` / `done` / `next` (bare + `<n> [--spec]`); exit codes 0/1/2; no-prompt static guard test.
- **M3 — Skill shrink + template mirror** (D-h): same delta to both twins; `make build`; commit regenerated `catalog.yaml`; `MOAI_TEMPLATE_LEAK_STRICT=1 go test` neutrality gate.
- **M4 — Cross-platform + full-suite verification**: `GOOS=windows go build ./...` AND `GOOS=windows go vet ./...` (exit codes cited); full `go test ./...` (full-suite rule: affected-package-only self-report is insufficient); lint NEW-vs-baseline classification.

**Phase 13 note (branch/worktree setup)**: deferred to the run session per Late-Branch policy — this shared checkout carries five concurrent kanban sessions and plan phase performs NO git state changes. The run session creates the feature branch / PR per the repo-local all-tier PR policy (`.claude/rules/moai/workflow/repo-local-pr-policy.md`; `main` is protected, Route A disabled). Implementation Kickoff Approval still gates run-phase entry.

## §G Anti-Patterns

- **AP-PL-001 — REQ count over Tier M ceiling (RESOLVED)**: the first delegation authored 19 REQs vs the 16 ceiling. Plan-audit iteration 1 (D2) flagged it as blocking; resolved at v0.1.1 via the **consolidation route** (the auditor's option b, chosen because no user-override channel exists for this session and the merged pairs are true groupings — both behaviors preserved verbatim in each merged REQ): 005+006 → 005 (two When-branches), 013+014 → 012 (two When-branches), 017+018 → 015 (skill + template mirror in one REQ). No scope dropped; every AC stays mapped. Old→new numbering map in spec.md §F History.
- **AP-PL-002 — Applying `requireLeadRole` "for consistency"**: the board guard welded onto the backlog would break every non-lead append — the primary use case (D-b).
- **AP-PL-003 — Deriving ids from max-present-id at issue time**: regresses to id reuse after `done` (the incident's exact failure shape). The persisted high-water mark is the fix, not an optimization.
- **AP-PL-004 — Silent malformed-file recovery**: any "repair on load" path violates REQ-TODO-014 and destroys the one artifact that cannot be regenerated.
- **AP-PL-005 — Verbatim `cp` of the shrunk skill to the template twin**: apply the same delta; `cp` overwrites intentionally-neutralized variants when they exist (B4).
- **AP-PL-006 — Windows claim on `go build` alone**: never compile-verifies `_test.go`; `GOOS=windows go vet` is the claim gate (CLAUDE.local.md §6).

## §H Cross-References

- `spec.md` §C/§D/§E — requirements, constraints, exclusions this plan implements.
- `research.md` — verified evidence for every decision above (§2 incident, §4 substrate, §5 pattern, §7 atomicfile, §8 twins, §9 CLI conventions, §10 dispatch-rule verb fit).
- SPEC-KANBAN-BOARD-001 — substrate provenance; its plan/acceptance conventions mirrored here.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board — the operator-act doctrine the verbs serve.
- `.moai/docs/git-local-workflow-doctrine.md` + `repo-local-pr-policy.md` — all-tier PR route (Route A disabled in this repo).
