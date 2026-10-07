---
id: SPEC-HOOK-ZONE-BACKSLASH-001
title: "Protected-zone target resolution preserves POSIX component identity — zoneSlash must not invent separators"
version: "0.1.0"
status: draft
created: 2026-10-07
updated: 2026-10-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/hook
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-SELF-IMPROVE-PROTECTED-ZONE-001, SPEC-HOOK-BACKSLASH-SYMLINK-001, SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001]
tags: "hook, security, protected-zone, backslash, symlink, cwe-61, posix, t1556, t1561, t1566"
---

# SPEC-HOOK-ZONE-BACKSLASH-001 — Protected-zone target resolution preserves POSIX component identity

## A. History

| Date | Version | Change |
|------|---------|--------|
| 2026-10-07 | 0.1.0 | Initial draft. Authored by manager-spec for card t1566, from a measured P1 bypass demonstrated by the t1556 codex review gate round 8. Third surface of the POSIX-backslash class (t1556, t1561, this card). |

## B. Problem Statement

The protected-zone guard (SPEC-SELF-IMPROVE-PROTECTED-ZONE-001) decides Write/Edit and
Bash calls by the self-improvement identity set against a manifest + compiled-baseline
zone. Every decision path funnels through `resolveZoneTarget`
(`internal/hook/protected_zone_path.go:182`), which derives the relative forms the zone
is matched against — and `resolveZoneTarget` normalizes its input through `zoneSlash`
(`internal/hook/protected_zone_path.go:35-37`):

```go
func zoneSlash(p string) string {
	return strings.ReplaceAll(filepath.ToSlash(p), "\\", "/")
}
```

The rewrite is unconditional on both platform families. On POSIX, `\` is NOT a path
separator — it is an ordinary, legal filename character (a separator only on Windows).
`zoneSlash` is applied to the raw target before any filesystem step
(`protected_zone_path.go:183`, and to the cwd-joined form at `:190`; `zoneLexicalRel`
applies it again at `:58`-`:59`, and the root resolution of the symlink arm at
`:208`/`:210`).

**Mechanism hypothesis** (stated as a hypothesis per the card mandate; the M1 RED test
pins the truth — the narrative below is derived from the code and matches the measured
observation): a project entry literally named `lnk\dir`, symlinked to the protected
directory `zone_dir`, is destroyed by the rewrite. The validated spelling
`<project>/lnk/dir/secret.md` names a DIFFERENT object — a component `lnk` that does
not exist. The symlink arm's walk (`zoneResolve`, `:103`-`:177`) fails EvalSymlinks on
the whole path, probes `<project>/lnk`, observes it genuinely missing via Lstat
(`:147`-`:150`), and rejoins the tail onto the resolved prefix — minting the fictional
in-project path `<project>/dir/secret.md`. Both the lexical arm (`:203`-`:205`) and the
symlink arm then classify the target OUTSIDE `zone_dir/`, `resolveZoneTarget` returns
zero forms, and `checkProtectedZone` (`protected_zone_guard.go:152`-`:220`) matches no
baseline entry and no manifest entry over an empty form list: decision allow. The actual
Write, however, follows the path the OS walks — the literal component `lnk\dir` IS a
symlink, and the file lands INSIDE the protected zone. The validated path and the
acted-on path diverge — the same divergence class t1530 (lexical `..`) and t1510
(`zoneResolve`) closed.

**Measured evidence (independent gate observation, not a hypothesis)**: round 8 of the
card-t1556 codex review gate (2026-10-07, delta commit `9d78421a4`) — an actual
harness-learner Write call through a backslash-named link received
`decision="allow"` and the protected file recorded `"changed"`. Primary copy:
`.moai/reports/t1556/codex-review-gate-1.md` § 라운드 8 in the t1556 card worktree
(a disposable tree; the durable landing is the t1556 PR). Verbatim finding quote:

> **[P1, 긴급 — SPEC이 t1454 원장으로 이연한 zoneSlash 자매 우회가 실증됨]**
> internal/hook/protected_zone_path.go:36 — POSIX에서 백슬래시 포함 링크로 보호
> 디렉터리를 가리키면 zoneSlash가 다른 경로를 검사해 보호 우회. **실제
> harness-learner Write 호출에서 decision="allow" + 보호 파일에 "changed" 기록
> 관측(게이트 실측)** — 후보가 아니라 실증된 P1.

Root cause class: separator conversion must be platform-appropriate. This is the THIRD
surface of the same POSIX-backslash family:

| Surface | Site | Owner |
|---------|------|-------|
| 1 | `internal/hook/pre_tool.go:1397` (project-boundary walk) | card t1556 — repair in flight, PR open, touches ONLY pre_tool.go (verified disjoint from this card's surface) |
| 2 | `internal/cli/worktree/landing_predicate.go:161` | card t1561 — separate subsystem, separate card |
| 3 | `internal/hook/protected_zone_path.go:36` (`zoneSlash`, this SPEC) | card t1566 — THIS card |

Both decision consumers inherit the defect through the shared resolver:
`checkProtectedZone` (Write/Edit branch, `protected_zone_guard.go:157`) and
`checkProtectedZoneShell` (Bash branch, `protected_zone_shell.go:1014`) — the round-8
measurement exercised the Write/Edit branch; the Bash branch is exposed identically.

## C. Requirements (GEARS)

### REQ-HZB-001 — Zone target resolution preserves POSIX component identity (Event-driven)

**When** a Write/Edit or Bash path on a POSIX host names a component whose literal
spelling contains `\` and that component is a symlink reaching INTO the protected zone,
the zone target resolution (`resolveZoneTarget`) shall resolve and zone-check the
LITERAL path — following the symlink that component actually names — so the resolved
target lands inside the zone and the guard denies the call. It shall not validate any
transformed spelling that the OS would not walk.

### REQ-HZB-002 — Platform-appropriate separator semantics (Ubiquitous)

The slash conversion used by the zone target resolution shall treat the backslash as a
path separator ONLY on Windows; on every POSIX platform it shall treat every `\` in a
path component as an ordinary filename character. The Windows-platform rewriting
behavior is preserved (existing windows-semantic zone tests stay green).

### REQ-HZB-003 — No behavior change for backslash-free paths (Unwanted)

The repair shall not change the resolution or decision behavior of any path that
contains no backslash character: every existing zone test (lexical normalization,
file tools, shell mutation, manifest states, non-regression, deny reason, audit,
baseline, liveness — the `TestProtectedZone` subtest group) stays green.

### REQ-HZB-004 — Fail-closed semantics preserved (Unwanted)

The repair shall not open any fail-open path that did not exist before: the walk's
unresolved-tail rejoin for genuinely missing components, the unreadable-entry
fail-closed return, the symlink depth bound, and the `..` pop over resolved prefixes
keep their current semantics (`protected_zone_path.go:110`-`:177`). A repaired
resolution that cannot vouch for a path keeps failing closed.

### REQ-HZB-005 — Both decision consumers inherit the repair (Event-driven)

**Where** the repair lands in the shared resolver `resolveZoneTarget`, both consumers
— `checkProtectedZone` (Write/Edit branch) and `checkProtectedZoneShell` (Bash
branch) — shall exhibit the REQ-HZB-001 behavior without any per-consumer change. The
plan-phase RED observation is on the Write/Edit branch (the measured surface); the
Bash branch inherits through the shared resolver.

## D. Constraints

- The fix lives in the path-interpretation layer of
  `internal/hook/protected_zone_path.go` (the `zoneSlash` semantics and its call
  sites within `zoneLexicalRel` / `resolveZoneTarget`). It MUST NOT weaken any other
  guard layer: the identity gate, the compiled-baseline floor and its legacy
  sentinels, the manifest fail-closed state machine, and the audit rows are out of
  scope and unchanged.
- **Scope: the named instance only.** Do NOT repair `internal/hook/pre_tool.go`
  (card t1556 owns it; its fix is in flight in a disjoint file) or
  `internal/cli/worktree/landing_predicate.go` (card t1561 owns it). The M3 family
  sweep reads their state as evidence only.
- Zone matching stays slash-normalized: manifest entries and the folded comparison
  (`config.FoldZoneText`, `entry.Match`) operate on `/`-separated relative forms. The
  repair preserves POSIX component identity through RESOLUTION while keeping the
  matching forms slash-separated — the concrete mechanism (runtime GOOS check,
  build-tagged helper, or another shape) is the repair agent's decision, subject to
  REQ-HZB-001/002/003/004.
- Windows root handling caution: `zoneSlash` is also applied to the project root
  (`:59`, `:208`, `:210`); a naive removal of the rewrite changes Windows root
  normalization. REQ-HZB-002 binds this.
- Tests: `t.TempDir()` for every temporary directory; symlink fixtures are
  POSIX-gated and skip gracefully (`t.Skip`) where the platform cannot create
  directory symlinks unprivileged; table-driven where the surface is pure. Never
  `t.Setenv("OTEL_*", ...)` in parallel tests.
- Per repo discipline (AGENTS.local.md §4): run the affected package only —
  `go test -timeout 30m -count=1 ./internal/hook/` — not the full suite locally.
- Go code, comments, godoc: English. Error wrapping `fmt.Errorf("...: %w", err)`.

## E. Success Criteria

1. The reproduction test of AC-HZB-001/002/003 is RED on the pre-repair tree (pinned
   `f97edcc55`, or a tree code-identical to it carrying only SPEC docs) and GREEN
   after the M2 repair — both observed, per the two-cell adoption rule
   (`.claude/rules/moai/development/verification-completeness.md` §2).
2. REQ-HZB-002/003 regression guards pass in the affected package
   (`go test -timeout 30m -count=1 ./internal/hook/` with only the baseline failures
   recorded at M1, if any).
3. The M3 three-surface family sweep evidence is recorded: this card's surface
   repaired (GREEN flip); the two sibling surfaces' state in this tree recorded as
   read-only evidence with their owning cards named.
4. `GOOS=windows go build ./...` exit 0 (the Windows half of the platform split
   compiles).

## F. Scope Boundary — the three-surface family

Repairs of sibling surfaces belong to their own cards. This SPEC adds no repair to
`pre_tool.go` or `landing_predicate.go`; it records their state as M3 evidence. The
t1556 fix touching only `pre_tool.go` is disjoint from this card's surface by file;
no merge-order dependency is assumed, and the M3 sweep re-reads both sibling sites in
THIS tree to state their current class-instance state as evidence.
