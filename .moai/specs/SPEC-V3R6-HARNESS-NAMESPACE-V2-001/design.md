---
id: SPEC-V3R6-HARNESS-NAMESPACE-V2-001
title: "Harness namespace doctrine-code drift catch-up — design (atomic transition + substring-conflict + generator boundary)"
version: "0.1.0"
status: draft
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/harness, internal/template, .claude/skills/moai-meta-harness"
lifecycle: spec-anchored
tags: "namespace, harness, doctrine-drift, atomic-transition, generator-emission"
---

# Design — SPEC-V3R6-HARNESS-NAMESPACE-V2-001

## §A. Purpose of This Document

This design doc justifies three load-bearing decisions that the plan-phase must lock before run-phase:

1. **Atomic-transition sequencing** — why all 5 surfaces must land in one PR, and the exact ordering within the PR.
2. **Substring-conflict analysis** — why `harness-*` and `moai-harness-*` do not collide under exact `HasPrefix`, and the forbidden patterns.
3. **Generator emission boundary** — why the emission-prefix edit is a string edit in template Markdown, NOT new generator authoring, and the exact files/lines in scope.

These three questions are interlocked: the atomicity constraint exists BECAUSE the generator emission and the Go enforcement must switch together (otherwise generated `harness-*` skills are unprotected); the substring constraint exists BECAUSE the new `harness-*` prefix shares a substring with the template-managed `moai-harness-*`; the generator boundary exists BECAUSE the emission prefix lives in a surprising place (template Markdown, not a Go constant) and over-reading the edit scope would violate the run-phase constraint.

## §B. Atomic-Transition Sequencing

### §B.1 Why atomic (the partial-migration failure mode)

The 5 surfaces form a producer-enforcer-sentinel triangle:

```
 ┌─────────────────┐        emits         ┌──────────────────┐
 │ moai-meta-harness│ ────────────────────►│  harness-* skill │ (user project)
 │ (generator)     │   prefix = harness-*  │  on disk         │
 └─────────────────┘                       └────────┬─────────┘
                                                    │
                                           protects │
                                                    ▼
 ┌─────────────────┐        scans          ┌──────────────────┐
 │ isUserOwned     │ ◄──────────────────── │  moai update     │
 │  Namespace()    │   prefix = harness-*  │  (enforcer)      │
 └─────────────────┘                       └────────┬─────────┘
                                                    │
                                          detects   │
                                          leak      ▼
 ┌─────────────────┐                     ┌──────────────────┐
 │ CI sentinel     │ ◄───────────────────│  template tree   │
 │ TestNamespace   │    pattern=harness- │  (must NOT leak) │
 └─────────────────┘                     └──────────────────┘
```

**Failure mode A (enforcer switches, generator doesn't)**: `isUserOwnedNamespace` recognizes `harness-*`, but `moai-meta-harness` still emits `my-harness-*`. A user generates a new skill → it lands as `my-harness-*` → the enforcer (now only recognizing `harness-*` + legacy) still protects it via the deprecation window. This is the SAFE direction, but it means the doctrine's `harness-*` target is still not actually produced. The migration is incomplete.

**Failure mode B (generator switches, enforcer doesn't)**: `moai-meta-harness` emits `harness-*`, but `isUserOwnedNamespace` still recognizes only `my-harness-*`. A user generates a new skill → it lands as `harness-foo` → `moai update` runs → `isUserOwnedNamespace(".claude/skills/harness-foo/SKILL.md")` returns `false` → the skill is NOT backed up → a clean-reinstall cycle could lose it. **This is the user-data-loss path.** It is the HARD failure mode the atomicity constraint prevents.

**Failure mode C (sentinel switches, enforcement doesn't)**: The CI sentinel scans for `harness-*` leak, but the enforcer still uses `my-harness-*`. A `harness-*` skill leaked into template would be caught by CI, but a user's `harness-*` skill would not be protected. Inconsistent state.

### §B.2 PR-internal ordering (within the atomic commit group)

The PR is one squash-merge, but the commit group has an internal order to make review tractable:

1. **M1 — Go enforcement** (skills-side predicate migration + backward-compat). Lands the dual-recognition (`harness-*` canonical + `my-harness-*` legacy). After M1, BOTH prefixes are protected, so the enforcer is safe regardless of what the generator emits.
2. **M2 — Generator emission prefix** (template Markdown edit). After M2, the generator emits `harness-*`. Safe because M1 already protects it.
3. **M3 — CI sentinel pattern**. After M3, CI detects `harness-*` leak. Safe because M1+M2 are consistent.
4. **M4 — Test fixtures** (mechanical rename). After M4, all tests reference `harness-*`. The backward-compat fixtures (legacy `my-harness-*` recognition tests) are ADDED in M1, not renamed away.
5. **M5 — Integration verification**. Full `go test ./...` + lint + atomic-PR gate.

**Key invariant**: after M1, the enforcer is dual-recognition. This means M2/M3/M4 can proceed in any order without risking user-data loss, because BOTH `harness-*` and `my-harness-*` are protected. The atomicity constraint (single PR) is belt-and-suspenders: even if the PR is reviewed milestone-by-milestone, no intermediate revision is unsafe.

### §B.3 Why single-PR (not multi-PR staged)

Multi-PR staging (e.g., M1 in PR-1, M2 in PR-2) would mean PR-1 merges and ships while the generator still emits `my-harness-*`. Users on the PR-1 release get a `harness-*`-recognizing enforcer but a `my-harness-*`-emitting generator — an inconsistent shipped state. The doctrine's §24.5 freeze ("새 `harness-*` prefix로 실제 skill generate 금지 — Go code가 protection 안 하므로") exists precisely to prevent this. The atomic single-PR ensures the shipped release has a consistent producer-enforcer pair.

## §C. Substring-Conflict Analysis

### §C.1 The two namespaces sharing a substring

| Namespace | Owner | Example | Doctrine line |
|-----------|-------|---------|---------------|
| `harness-*` | user-owned | `harness-ios-patterns` | doctrine §24.1 line 15 |
| `moai-harness-*` | template-managed (builder) | `moai-meta-harness`, `moai-harness-learner` | doctrine §24.1 line 14 |

Both contain the substring `harness-`. A naive classifier using `strings.Contains(name, "harness-")` would misclassify `moai-harness-learner` as user-owned → it would not be overwritten on `moai update` → stale builder skill persists → drift.

### §C.2 The exact-`HasPrefix` solution

Go's `strings.HasPrefix` is a literal byte-prefix comparison. It does NOT do substring matching. Therefore:

```
HasPrefix("harness-ios-patterns", "harness-")        = true   (user-owned ✓)
HasPrefix("moai-harness-learner", "harness-")        = false  (template-managed ✓)
HasPrefix("harness-ios-patterns", "moai-harness-")   = false  (not template-managed ✓)
HasPrefix("moai-harness-learner", "moai-harness-")   = true   (template-managed ✓)
```

All four directions classify correctly. The `moai-` prefix on the template-managed namespace acts as a natural disambiguator: no string starting with `moai-` can also start with `harness-` (the `m` vs `h` first byte differs).

### §C.3 Forbidden patterns

- **FORBIDDEN**: `strings.Contains(name, "harness-")` — matches both namespaces.
- **FORBIDDEN**: `filepath.Glob(".claude/skills/*harness-*")` — matches both namespaces.
- **FORBIDDEN**: regex `.*harness-.*` — matches both namespaces.
- **REQUIRED**: `strings.HasPrefix(name, "harness-")` — exact byte prefix.
- **REQUIRED**: `strings.HasPrefix(name, "moai-harness-")` — exact byte prefix for the template-managed side (already used by `cleanMoaiManagedPaths` via the `moai*` glob, which is even broader but safe because it only targets `moai-*`).

### §C.4 The `moai*` deletion glob is already safe

`cleanMoaiManagedPaths` (update.go:1706-1708) deletes `.claude/skills/moai*`. This glob matches `moai-harness-*` (correct — template-managed, should be deleted before reinstall) and does NOT match `harness-*` (correct — user-owned, must not be deleted). The deletion path needs NO change. The risk surface is only the backup + overlay-write predicate (`isUserOwnedNamespace`), which M1 migrates.

## §D. Backward-Compat Deprecation Window

### §D.1 Why a window is needed

Existing user projects may have `my-harness-*` skill directories generated before this SPEC ships. If M1 removes `my-harness-*` recognition cold-turkey:

1. A user runs `moai update` on a project with `.claude/skills/my-harness-legacy/`.
2. `isUserOwnedNamespace(".claude/skills/my-harness-legacy/SKILL.md")` returns `false` (no longer recognized).
3. The backup step (`backupUserOwnedNamespace`) does NOT back it up.
4. If a clean-reinstall cycle removes it (unlikely via the `moai*` glob, but possible via other cleanup paths), the user loses the skill.

The deprecation window keeps `my-harness-*` recognized as user-owned, so step 2 returns `true` and step 3 backs it up.

### §D.2 Window definition

During the window, `isUserAreaPath` and `isUserOwnedNamespace` recognize BOTH:

```go
// Canonical (doctrine-declared, this SPEC's target)
if strings.HasPrefix(norm, ".claude/skills/harness-") {
    return true
}
// Legacy (pre-migration, recognized for backward compat — REQ-HNS-005)
if strings.HasPrefix(norm, ".claude/skills/my-harness-") {
    return true
}
```

Both branches are documented with an `@MX:NOTE` citing REQ-HNS-005 and the sunset trigger.

### §D.3 Sunset trigger

The legacy branch is removed in a **follow-up chore SPEC** (out of scope here) after a grace period of **2 release cycles** (proposal — the follow-up SPEC confirms the exact window). The sunset chore:

1. Greps user projects (impossible in-tree — this is a user-communication task, not a code task) — actually, the sunset is purely a code-side removal: drop the `my-harness-*` branch from the predicates.
2. The risk of sunset: a user who never re-ran `moai-meta-harness` to regenerate their skill under `harness-*` loses backward-compat. Mitigation: the 2-cycle grace period + release-notes warning.

This SPEC does NOT close the sunset. The sunset is a forward-looking follow-up (acceptance.md §D.4).

### §D.4 Why not sunset in this SPEC

The sunset requires user communication (release notes warning users to regenerate). Bundling it into the migration SPEC would force a single PR to both add and remove recognition, which is contradictory. The clean separation: this SPEC adds `harness-*` + keeps `my-harness-*` legacy; the follow-up removes `my-harness-*` legacy.

## §E. Generator Emission Boundary

### §E.1 Where the emission prefix actually lives

The `moai-meta-harness` generator is a Claude skill (template Markdown), NOT Go code. At runtime, the Claude generator reads the skill body and emits skill directories per the contract documented in:

- `internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md` § 6.4.1 (emission template + example)
- `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` lines 164-177 (emission contract + drift note)

There is NO Go-side prefix constant. The prefix `my-harness-*` appears as a literal string in the template Markdown, consumed by the Claude generator at runtime.

### §E.2 What M2 edits (in scope)

M2 edits the literal prefix string in:

- `meta-harness.md`: `my-harness-<domain>-patterns` → `harness-<domain>-patterns`, `my-harness-<domain>-best-practices` → `harness-<domain>-best-practices`, and all prose references (~15 occurrences).
- `moai-meta-harness/SKILL.md`: emission contract line 164 (`harness-*` prefix ONLY), drift note line 168 (remove — this SPEC resolves it), companion-skill preload line 174 (`harness-<domain>-*`), smoke-gate reference line 177 (`harness-*`).

### §E.3 What M2 does NOT edit (out of scope — NFR-HNS-003)

- The 7-Phase workflow body (Phases 1-7 of the meta-harness workflow) is UNCHANGED.
- The layer architecture (`internal/harness/layer1.go`-`layer5.go`, `chaining_rules.go`, `types.go`) logic is UNCHANGED — only comment string updates.
- The skill-skeleton templates (the actual SKILL.md body template emitted into `harness-*/SKILL.md`) are UNCHANGED.
- No new generator phases, no new layers, no new skeleton templates.

### §E.4 How the run-phase constraint binds

The handoff constraint "생성 코드 새로 작성 안 함 — namespace rename만" (do not author new generator code — namespace rename only) is operationalized as:

- **Diff gate (AC-HNS-003)**: `git diff --stat <base>..<head> -- internal/harness/` shows NO new files. `git diff <base>..<head> -- internal/harness/` shows NO `+func ` additions.
- **Template diff gate**: `git diff <base>..<head> -- internal/template/templates/.claude/skills/moai-meta-harness/ internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md` shows ONLY prefix-string edits and comment/drift-note updates, NOT new sections or new workflow phases.

This boundary is load-bearing: the rejected predecessor SPEC violated the generator HARD contract by attempting to relocate builder skills to a nonexistent `moai-builder-*` namespace, which would have required new generator logic. This SPEC does not.

## §F. Alternatives Considered

### §F.1 Alternative 1 — Policy reversal to `moai-harness-*` (REJECTED)

The rejected predecessor's choice. Failed because:
- Collides with existing template-managed `moai-meta-harness` + `moai-harness-learner`.
- `moai-meta-harness/SKILL.md:164` HARD-forbids emitting `moai-harness-*`.
- Requires inventing `moai-builder-*` (0 codebase refs) to relocate the colliding builders.
- Contradicts all 3 namespace SSOTs.

### §F.2 Alternative 2 — Cold-turkey migration (no deprecation window) (REJECTED)

Switch `isUserOwnedNamespace` to `harness-*` only, drop `my-harness-*` recognition. Failed because:
- Existing user `my-harness-*` skills lose backup recognition → user-data loss risk.
- No user-communication mechanism to warn before the drop.

### §F.3 Alternative 3 — Multi-PR staged migration (REJECTED)

M1 (enforcement) in PR-1, M2 (generator) in PR-2, etc. Failed because:
- PR-1 ships an inconsistent producer-enforcer pair (enforcer recognizes `harness-*`, generator still emits `my-harness-*`).
- The doctrine's §24.5 freeze exists precisely to prevent this shipped inconsistency.

### §F.4 Chosen approach — Atomic single-PR with dual-recognition window (THIS SPEC)

- Single PR: all 5 surfaces land together. No shipped inconsistency.
- M1 adds `harness-*` AND keeps `my-harness-*` (dual recognition). The window makes the intermediate milestone states safe.
- Sunset is a follow-up chore SPEC with user communication. Not bundled here.

## §G. Open Questions (for plan-auditor / user review)

- **Q1**: Is the 2-release-cycle deprecation window (§D.3) the right length? The user may prefer 1 cycle or 3. This is a policy preference; the SPEC proposes 2 as a default.
- **Q2**: Should the legacy `my-harness-*` recognition in `isUserAreaPath` (the older predicate) be kept in sync with `isUserOwnedNamespace` (the newer superset predicate)? Currently both need the dual-recognition edit. Answer: yes, both are edited in M1 for consistency, even though `isUserOwnedNamespace` is the authoritative one.

No other open questions. The design is deterministic given the doctrine (which is fixed).
