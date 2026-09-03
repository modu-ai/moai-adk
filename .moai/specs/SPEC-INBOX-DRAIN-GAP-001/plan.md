---
id: SPEC-INBOX-DRAIN-GAP-001
title: "Implementation plan — distributed lessons-inbox lifecycle"
version: "0.1.0"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
tier: M
---

# SPEC-INBOX-DRAIN-GAP-001 — Implementation Plan

## §A. Context

- Development mode: `tdd` (`.moai/config/sections/quality.yaml`). RED-GREEN-REFACTOR per milestone.
- The collection chain is shipped and must not change its observable append semantics (`internal/hook/failure_observer.go:114-154`): append-only JSONL, 0o600, fail-open, per-append open/close by path. The cap layers onto this path.
- The drain chain is dev-only and untouched: `.claude/settings.local.json` wiring, `session_drain.sh`, `drain.sh` (offset `<state-dir>/drain-offset.json`), `hns-lsel-curator`. NFC-4: on a curator machine, behavior is byte-identical to pre-SPEC.
- This plan is ordered by decision-reversibility: format/data-model decisions first (M1), core behavior second (M2), surface third (M3), hardening last (M4).

## §B. Known Issues

- The inbox has no shipped consumer (verified §B.1 of spec.md) — so rotation deleting archive generations is low-stakes for users, but the archive-rotate (vs pure trim) design still preserves data, honoring the untracked-delete-has-no-safety-net discipline.
- On the maintainer machine the inbox can exceed any sane cap during a drain stall (the t259 3-week stall reached 1.1MB / ~4.2k lines). The stand-down marker is what keeps such a stall from rotating unread stubs out of the curator's offset view — this is why REQ-IBX-002 is state-driven and mandatory, not advisory.
- `moai update` does NOT wipe `.moai/lessons-inbox.jsonl` (it sits at the `.moai` root, outside every `ManagedCleanTargets` entry) — the growth hazard is unbounded across updates, which is exactly what the cap closes.

## §C. Pre-flight (run-phase entry checks)

1. Verb-name collision: confirm no existing `inbox` command in the CLI surface (`grep -rn "inbox" internal/cli/*.go` at run start — plan-phase grep of `internal/cli/` found no `inbox.go`; re-verify).
2. Confirm the cap constants land in `internal/config/defaults.go` as the single source (per CLAUDE.local.md §14 hardcoding rules), named for the inbox domain (e.g. `InboxMaxBytes`, `InboxArchiveGenerations`).
3. Confirm the LSEL stand-down probe target: directory existence of `.moai/state/lsel/` (the state dir `drain.sh` creates). One `os.Stat` per append.
4. Confirm no `.sh.tmpl`/`.sh` pair work is needed at all — this SPEC ships no template files (neutrality doctrine satisfied trivially; no Template-First cycle required).

## §D. Constraints

- No edits under `internal/cli/update/` (REQ-IBX-010; merge-judgment boundary with t239).
- No new template files; no settings.json changes; no SessionStart wiring (REQ-IBX-010).
- Fail-open parity: every new error path logs via `slog.Warn` and swallows, mirroring `appendLessonsInboxStub` (NFC-3).
- Windows: the retention shift must remove-then-rename (NFC-5); verified by CI windows matrix, not locally.
- English code comments (language.yaml `code_comments: en`).

## §E. Self-Verification (planned run-phase evidence)

- E1 AC matrix: `go test ./internal/hook/... ./internal/cli/...` (affected packages only — no full-suite local runs; full suite is CI's verdict) with the AC-IBX-001..010 tests green.
- E2 Cross-platform: windows/macos CI jobs green on the PR head.
- E3 Coverage: `go test -cover ./internal/hook/...` for the new cap/rotation code ≥ package standard.
- E4 Scope guard: `git diff --name-only` vs base shows no `internal/cli/update/` path and no `internal/template/templates/.claude/settings.json.tmpl` change (AC-IBX-009 evidence).
- E5 Race: `go test -race ./internal/hook/ -run Inbox` for the concurrent-append path (AC-IBX-010).
- E6 Behavior parity on curator machines: a test asserting zero rotation when `.moai/state/lsel/` exists (AC-IBX-002 doubles as the NFC-4 proof).
- E7 `golangci-lint run` on touched packages.

## §F. Milestones

### M1 — Schema version + cap constants (data-model decisions first)

- Add `v` integer field to `lessonsInboxStub` (marshal `v:1` on new lines; readers tolerate absence).
- Add cap constants to `internal/config/defaults.go`: max bytes (proposed 1 MiB = 1<<20, just under the measured 1.1MB t259 stall scale) and archive-generation retention (2).
- Deliverable: golden-JSON test (AC-IBX-007) + constants consumed by M2.

### M2 — Write-time cap + rotation + stand-down (core behavior)

- In the append path: stat live size; if ≥ cap and no `.moai/state/lsel/` marker → rotate (`lessons-inbox.jsonl` → `.1`, shift `.1`→`.2`, delete beyond 2 generations), then append to the fresh live file.
- Fail-open: any stat/rotation error → `slog.Warn`, append proceeds best-effort on the existing file (REQ-IBX-009).
- Stand-down: marker present → zero rotation, zero trim (REQ-IBX-002).
- Deliverables: AC-IBX-001, AC-IBX-002, AC-IBX-003, AC-IBX-008, AC-IBX-010 green (`go test -race` included).

### M3 — CLI surface (`moai inbox status` / `moai inbox drain`)

- New `internal/cli/inbox.go` verb file following the flat command-file pattern; shared rotation routine reused (single implementation — the CLI must not fork the collector's rotation logic).
- `status`: size, lines, cap distance, archive generations, ownership regime token (`curator` | `cap-managed`), exit 0.
- `drain`: marker absent → rotate + report, exit 0; marker present → exit non-zero, notice names `session_drain.sh`, zero mutation (REQ-IBX-006/007).
- Deliverables: AC-IBX-004, AC-IBX-005, AC-IBX-006 green.

### M4 — Hardening + scope verification

- Race test on cross-cap concurrent appends (AC-IBX-010 formalized).
- Scope-guard check (E4), lint (E7), coverage (E3).
- No docs-site obligation: `moai inbox` help text is the user documentation surface (keep it self-explanatory).

## §G. Anti-Patterns

- Do NOT put the cap in the hook wrapper shell script — it lives in the binary append path (the wrapper is shipped but the binary is the version-consistent carrier).
- Do NOT add opportunistic drain calls inside `moai update`/`moai doctor` — rejected in spec §E.2; couples modules and moves the merge surface for no benefit.
- Do NOT detect the curator via skill-file presence (`.claude/skills/hns-lsel-curator`) — the ownership marker is the state dir `.moai/state/lsel/` (created by the drain itself, not by skill installation).
- Do NOT fork a second rotation implementation for the CLI — one shared routine.
- Do NOT delete archive generations outside the rotation chain, and do NOT rotate on a curator machine under any cap condition.

## §H. Cross-References

- SPEC-LSEL-DRAIN-STALL-001 (completed) — t259; stand-down lineage; REQ-LDS-009 deferral this SPEC closes; REQ-LDS-010 wrapper hygiene echoed by REQ-IBX-007.
- SPEC-LSEL-LOCAL-EVOLUTION-001 (completed) — LSEL seam provenance; drain writes only under `.moai/state/lsel/` (M1 invariant this SPEC's marker probe relies on).
- SPEC-LLMCFG-PRESERVE-001 (t239) — no code-path collision; recorded as the lane's batch merge-judgment boundary in spec.md §F.
- CLAUDE.local.md §2.3 / §28 — internal dev-only context for the wipe-root analysis and wrapper hygiene (internal citation, not shipped behavior).
