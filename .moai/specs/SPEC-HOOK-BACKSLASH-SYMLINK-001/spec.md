---
id: SPEC-HOOK-BACKSLASH-SYMLINK-001
title: "POSIX path-boundary check preserves literal backslashes — the physical walk must not invent separators"
version: "0.1.0"
status: completed
created: 2026-10-07
updated: 2026-10-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-INTERNAL-SECURITY-001, SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001]
tags: "hook, security, path-boundary, backslash, symlink, cwe-61, posix, t1533, t1556"
---

# SPEC-HOOK-BACKSLASH-SYMLINK-001 — POSIX path-boundary check preserves literal backslashes

## A. History

| Date | Version | Change |
|------|---------|--------|
| 2026-10-07 | 0.1.0 | Initial draft. Authored by manager-spec for card t1556, from a verified P1 defect reproduced by the codex review gate round 1 of card t1533. |

## B. Problem Statement

The pre-tool-use file-access guard (`internal/hook/pre_tool.go`, `checkFileAccess` →
`resolveThroughExistingParent` → `resolvePhysicalWalk`) validates that a Write/Edit path
resolves inside the project boundary before the tool runs. On macOS/Linux the walk's
segment split converts every `\` into `/` **unconditionally**
(`internal/hook/pre_tool.go:1397`):

```go
parts := strings.Split(strings.ReplaceAll(filepath.ToSlash(p), "\\", "/"), "/")
```

On POSIX, `\` is an ordinary, legal filename character — it is a path separator ONLY on
Windows. An attacker creates a project directory whose literal name contains a backslash
(e.g. `innocent\dir`) and points it at a location outside the project. The walk then
validates a FICTIONAL path: `innocent\dir/file.txt` is split into the non-existent
`innocent` component, the probe fails as "missing", and the unresolved tail rejoins onto
an in-project prefix — `<project>/innocent/dir/file.txt` — which passes the boundary
check. The actual Write, however, follows the REAL path the OS walks, where the
`innocent\dir` symlink resolves outside the project, and the write lands outside.

Reproduction evidence (verified, twice): card t1533 codex review gate round 1 —
primary copy `.moai/reports/t1556/codex-review-gate-1.md` (in this card worktree;
provenance: `.moai/worktrees/t1533/.moai/reports/t1533/codex-review-gate-1.md`, a
disposable tree), finding
"[P1] POSIX 경로의 실제 백슬래시를 보존 — internal/hook/pre_tool.go:1397". With the
base commit's interpretation, blocking works; the backslash→slash conversion is what
defeats it. The reproduced test wrote `"escaped"` to an external file.

Root cause class: separator conversion must be platform-appropriate — the same defect
family as card t1530 (physical `..` walk) and card t1510 (`zoneResolve`), where the
checker reasoned over a path that is not the path the OS acts on.

## C. Requirements (GEARS)

### REQ-HBS-001 — Platform-appropriate separator semantics (Ubiquitous)

The path-resolution walk used by the pre-tool-use file-access guard shall treat the
backslash as a path separator ONLY on Windows; on every POSIX platform it shall treat
every `\` in a path component as an ordinary filename character.

### REQ-HBS-002 — Validate the path the OS will act on (Event-driven)

**When** a Write/Edit path names a component whose literal spelling contains `\` on a
POSIX host, the file-access guard shall resolve and boundary-check the LITERAL path —
following any symlink that component actually names — and shall not validate any
transformed spelling that the OS would not walk.

### REQ-HBS-003 — Outside-destination deny preserved (Event-detected)

**When** the literal resolution of a Write/Edit path on a POSIX host lands outside the
project boundary (including through a backslash-named directory symlink), the guard
shall deny the tool call before the Write/Edit executes, and no byte shall reach the
external destination.

### REQ-HBS-004 — Windows behavior preserved (Where gate)

**Where** the host platform is Windows, the guard shall keep treating `\` as a path
separator, and every behavior the t1530 physical-walk repair established (new-file tail
rejoin, physical `..` pop over resolved prefixes, symlink depth bound) shall remain
unchanged on both platform families.

### REQ-HBS-005 — No false-positive regression on legitimate POSIX names (Unwanted)

The guard shall not deny, on a POSIX host, a Write/Edit whose literal path is inside the
project merely because a component name contains a `\` character (e.g. a legitimate
new file `weird\name.txt` under the project root), and shall not regress the
in-project new-file resolution guarantee of the t1530 repair.

### REQ-HBS-006 — Fail-closed semantics preserved (Unwanted)

The repair shall not open any fail-open path that did not exist before: when the walk
cannot vouch for a path (`..` above a resolved prefix, unreadable entry, symlink chain
past the depth bound), the guard shall keep the existing unresolved-fallback /
fail-closed behavior.

## D. Constraints

- The fix lives in the path-interpretation layer of `internal/hook` (`resolvePhysicalWalk`
  and its input handling). It MUST NOT weaken any other check in `checkFileAccess`
  (deny patterns, NFC normalization, allowed-external list, project-root EvalSymlinks
  symmetry).
- Platform selection may use a runtime GOOS check or build-tagged helpers — the concrete
  mechanism is the repair agent's decision, subject to REQ-HBS-001/004.
- Windows directory-symlink creation in tests is privileged: tests exercising the
  symlink escape MUST skip gracefully where the platform cannot create the fixture
  (`t.Skip`), while the separator-semantics logic itself MUST be covered by
  platform-independent unit tests on every CI runner.
- Per repo discipline (AGENTS.local.md §4): run the affected package tests only —
  `go test -timeout 30m -count=1 ./internal/hook/` — not the full suite locally.

## E. Success Criteria

1. The reproduction test of AC-HBS-001 is RED on the pre-fix tree (pinned
   `cad44a751`) and GREEN after the fix — both observed, per the two-cell adoption rule
   (`.claude/rules/moai/development/verification-completeness.md` §2).
2. After the fix, no external file is written during the reproduction scenario.
3. `go test -count=1 ./internal/hook/` passes with the fix in place (no regression).
4. `GOOS=windows go build ./...` succeeds (cross-platform compile surface intact).

## F. Out of Scope

### Out of Scope — other findings from the t1533 review round 1

- The 13 other findings (3 P1 + 10 P2; 14 total minus the backslash P1 under repair)
  reported by `.moai/reports/t1556/codex-review-gate-1.md` round 1 — they belong to
  the t1454 residual ledger and follow-up cards. This SPEC repairs ONLY the
  backslash/symlink boundary bypass at `internal/hook/pre_tool.go:1397`.

### Out of Scope — the same conversion pattern in the zone deny path

- `zoneSlash` (`internal/hook/protected_zone_path.go:36`) carries the identical
  unconditional `\`→`/` conversion, and its output feeds `resolveZoneTarget`
  (`protected_zone_path.go:183-218`) → `checkProtectedZone`
  (`protected_zone_guard.go:152-178`) — a DENY-DECISION path, not display matching.
  It is therefore a LIVE SIBLING SURFACE of the same defect family: a backslash-named
  component in a zone-relative path is split into a fictional path before the zone's
  physical walk, with both false-allow and false-deny directions possible. This SPEC
  deliberately defers it to a named follow-up card (recorded in the t1454 residual
  ledger) so one card closes one defect; it is NOT dismissed as benign, and plan.md §G
  forbids fixing it in passing.

### Out of Scope — adjacent resolution surfaces

- `file_changed.go` resolution-recheck path, `agentmemory.go` / `subagent_write_guard.go`
  / `reflective_write.go` `ToSlash` uses, and any hardening of `checkFileAccess` beyond
  the separator semantics of the walk.
